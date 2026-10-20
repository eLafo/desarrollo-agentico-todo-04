#!/usr/bin/env bash
# Gates de la CI (.github/workflows/ci.yaml) en local, con un registro por gate.
#
#   check.sh [--out <directorio en .adw/>] [gate ...]
#
# Gates, en el orden de la CI: gofmt, vet, migrate, test, build, e2e y docker. Sin gates, todos.
# e2e son los tests de navegador (npx playwright test): arrancan la aplicación con GO_ENV=test.
# docker es informativo: en local no cambia el veredicto; en la CI sí bloquea.
#
# Con --out deja en el directorio un registro por gate (gate-<gate>.txt): comando,
# código de salida, duración y salida recortada.
#
# Imprime una línea por gate (gate, estado, código de salida, registro) y el veredicto:
#   GATES: GREEN | GATES: RED (<gates>) | GATES: ERROR (<gates>)
# Estados: pass, fail, skip (no se ejecuta porque falló un gate del que depende) y
# error (falta una herramienta, una dependencia de Node o el navegador, o la base de datos de
# test no es la de este proyecto o no responde: no es culpa del cambio).
#
# Como la CI, migrate empieza con una base de test vacía: la borra, la crea y la migra.
# Sale con 0 en verde, 1 si falla un gate bloqueante, 2 si el uso es incorrecto y 3 si hay error.
set -o pipefail

all=(gofmt vet migrate test build e2e docker)
root=$(git rev-parse --show-toplevel) || exit 2
cd "$root" || exit 2
# En un worktree de bin/tree, su .env da los puertos y el proyecto de Compose. go test y Playwright
# no lo leen solos: se exporta para todo lo que se ejecute desde aquí.
# shellcheck disable=SC1091
[[ -f .env ]] && { set -a; . ./.env; set +a; }

usage() {
  echo "uso: check.sh [--out <directorio en .adw/>] [${all[*]}]" >&2
  exit 2
}

out=""
gates=()
while (($#)); do
  case $1 in
    --out)
      [[ $# -ge 2 ]] || usage
      out=${2%/}
      shift 2
      ;;
    -*) usage ;;
    *)
      [[ " ${all[*]} " == *" $1 "* ]] || usage
      gates+=("$1")
      shift
      ;;
  esac
done
((${#gates[@]})) || gates=("${all[@]}")

if [[ -n $out ]]; then
  out=${out#"$root"/}
  [[ $out == .adw/* && $out != *..* ]] || { echo "check.sh: --out debe estar dentro de .adw/" >&2; exit 2; }
  mkdir -p "$out" || exit 2
fi

soda=$(command -v soda || true)
[[ -n $soda ]] || { [[ -x "$(go env GOPATH 2>/dev/null)/bin/soda" ]] && soda="$(go env GOPATH)/bin/soda"; }

red=()
errors=()

# Estado de cada gate (pass, fail, skip o error), sin arrays asociativos: vale para bash 3.2.
set_status() { printf -v "status_$1" '%s' "$2"; }
get_status() { local name="status_$1"; printf '%s' "${!name}"; }

# record <gate> <comando mostrado> <código> <segundos> <salida>
record() {
  local gate=$1 shown=$2 code=$3 seconds=$4 output=$5 file=-
  if [[ -n $out ]]; then
    file="$out/gate-$gate.txt"
    {
      echo "\$ $shown"
      echo "exit: $code"
      echo "duración: ${seconds}s"
      echo "---"
      local lines
      lines=$(printf '%s\n' "$output" | wc -l | tr -d ' ')
      if ((lines > 200)); then
        printf '%s\n' "$output" | head -n 50
        echo "… ($((lines - 200)) líneas omitidas) …"
        printf '%s\n' "$output" | tail -n 150
      else
        printf '%s\n' "$output"
      fi
    } > "$file"
  fi
  printf '%s\t%s\texit=%s\t%s\n' "$gate" "$(get_status "$gate")" "$code" "$file"
}

# run <gate> <bloquea: yes|no> <comando mostrado> <comando…>
run() {
  local gate=$1 blocking=$2 shown=$3 start output code
  shift 3
  start=$SECONDS
  output=$("$@" 2>&1)
  code=$?
  if ((code == 0)); then
    set_status "$gate" pass
  elif environment_error "$gate" "$output"; then
    set_status "$gate" error
    errors+=("$gate")
  else
    set_status "$gate" fail
    [[ $blocking == yes ]] && red+=("$gate")
  fi
  record "$gate" "$shown" "$code" $((SECONDS - start)) "$output"
}

fail_with() { # <gate> <motivo>: error de precondición
  set_status "$1" error
  errors+=("$1")
  record "$1" "-" - 0 "$2"
}

# Los ficheros que se registrarían, sin lo ignorado ni lo borrado: lo que vería la CI.
gofmt_gate() {
  local file files=() unformatted
  while IFS= read -r file; do
    [[ -f $file ]] && files+=("$file")
  done < <(git ls-files -co --exclude-standard -- '*.go')
  ((${#files[@]})) || return 0
  unformatted=$(gofmt -l "${files[@]}" 2>&1) || { echo "$unformatted"; return 1; }
  [[ -z $unformatted ]] && return 0
  echo "Archivos sin formatear:"
  echo "$unformatted"
  return 1
}

# Un fallo de conexión o de credenciales con la base de datos es del entorno, no del cambio.
# En e2e también lo son que falte el navegador o que el puerto de la aplicación esté ocupado.
environment_error() { # <gate> <salida>
  [[ $1 == migrate || $1 == test || $1 == e2e ]] || return 1
  if [[ $1 == e2e ]] && grep -qiE "Executable doesn't exist|is already used|Cannot find module '@playwright/test'" <<< "$2"; then
    return 0
  fi
  grep -qiE 'connection refused|password authentication failed|dial tcp|no such host|could not connect|server closed the connection unexpectedly|the database system is (starting up|shutting down)|is being accessed by other users|database "?todo_test"? does not exist' <<< "$2"
}

# migrate, test y e2e usan la base de datos del entorno de este proyecto: la del servicio db de
# compose.yaml, en DB_PORT (5432 en el checkout principal). Si el puerto es de otro proyecto, no se
# toca. Con DB_HOST, se usa ese servidor.
test_db_problem() {
  if [[ -z ${DB_HOST:-} ]] && ! docker compose ps --status running --services 2> /dev/null | grep -qx db; then
    echo "El servicio db de este proyecto no está en marcha (docker compose ps). Arranca el entorno (env:start)."
  elif ! db_reachable; then
    echo "La base de datos de test no responde en ${DB_HOST:-127.0.0.1}:${DB_PORT:-5432}. Arranca el entorno (env:start)."
  fi
}

# Si el host no contesta, /dev/tcp espera al timeout de TCP: se corta a los 3 segundos.
db_reachable() {
  (exec 3<>"/dev/tcp/${DB_HOST:-127.0.0.1}/${DB_PORT:-5432}") 2>/dev/null &
  local pid=$! tries=0
  while kill -0 "$pid" 2>/dev/null; do
    if ((tries++ >= 15)); then
      kill "$pid" 2>/dev/null
      wait "$pid" 2>/dev/null
      return 1
    fi
    sleep 0.2
  done
  wait "$pid"
}

migrate_gate() {
  local dropped
  dropped=$("$soda" drop -e test 2>&1)
  if [[ $? -ne 0 && $dropped != *"does not exist"* ]]; then
    echo "$dropped"
    return 1
  fi
  echo "$dropped"
  "$soda" create -e test && "$soda" migrate -e test
}

for gate in "${gates[@]}"; do
  case $gate in
    gofmt)
      run gofmt yes "gofmt -l <ficheros .go de git ls-files -co --exclude-standard>" gofmt_gate
      ;;
    vet)
      run vet yes "go vet ./..." go vet ./...
      ;;
    migrate)
      problem=$(test_db_problem)
      if [[ -z $soda ]]; then
        fail_with migrate "soda no está instalado: go install github.com/gobuffalo/pop/v6/soda@v6.4.1"
      elif [[ -n $problem ]]; then
        fail_with migrate "$problem"
      else
        run migrate yes "soda drop -e test; soda create -e test && soda migrate -e test" migrate_gate
      fi
      ;;
    test)
      if [[ $(get_status migrate) == fail || $(get_status migrate) == error ]]; then
        set_status test skip
        record test "-" - 0 "No se ejecuta: falló migrate."
      elif problem=$(test_db_problem) && [[ -n $problem ]]; then
        fail_with test "$problem"
      else
        run test yes "GO_ENV=test go test -p 1 ./..." env GO_ENV=test go test -p 1 ./...
      fi
      ;;
    build)
      run build yes "go build ./..." go build ./...
      ;;
    e2e)
      if [[ $(get_status migrate) == fail || $(get_status migrate) == error ]]; then
        set_status e2e skip
        record e2e "-" - 0 "No se ejecuta: falló migrate."
      elif [[ ! -d node_modules/@playwright/test ]]; then
        fail_with e2e "Faltan las dependencias de Node de los tests de navegador: npm ci"
      elif problem=$(test_db_problem) && [[ -n $problem ]]; then
        fail_with e2e "$problem"
      else
        run e2e yes "npx playwright test" npx playwright test
      fi
      ;;
    docker)
      if ! docker info > /dev/null 2>&1; then
        fail_with docker "Docker no responde."
      else
        run docker no "docker build -t todo ." docker build -t todo .
      fi
      ;;
  esac
done

# docker es informativo: su error no cuenta para el veredicto.
blocking_errors=()
for gate in ${errors[@]+"${errors[@]}"}; do
  [[ $gate == docker ]] || blocking_errors+=("$gate")
done

if ((${#blocking_errors[@]})); then
  echo "GATES: ERROR (${blocking_errors[*]})"
  exit 3
elif ((${#red[@]})); then
  echo "GATES: RED (${red[*]})"
  exit 1
fi
echo "GATES: GREEN"
