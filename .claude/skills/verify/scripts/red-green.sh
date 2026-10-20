#!/usr/bin/env bash
# Ejecuta tests nuevos sobre la base y sobre el cambio: la evidencia de que podían fallar.
#
#   red-green.sh --out <directorio en .adw/> --name <NN_afirmación-detalle> --base <ref> \
#                --run <patrón de -run> [--extra <fichero>]... <paquete> <fichero de test> ...
#   red-green.sh --out <directorio en .adw/> --name <NN_afirmación-detalle> --base <ref> \
#                [--run <patrón de --grep>] [--extra <fichero>]... e2e/<test>.spec.ts ...
#
# Crea un worktree desprendido de <ref> en .adw/base/, copia allí los ficheros de test
# indicados y los --extra que necesiten (fixtures, testdata), tal como están en la copia
# de trabajo, y ejecuta en los dos lados:
#   - tests de Go:        GO_ENV=test go test -v -p 1 -count=1 -run <patrón> <paquete>
#   - tests de navegador: npx playwright test [--grep <patrón>] <tests>
#     Cada lado arranca su propia aplicación (playwright.config.ts) y la prueba en el navegador.
#     A la base se le copian también playwright.config.ts y e2e/fixtures.ts de la copia de
#     trabajo, y se le enlaza node_modules: son el arnés de los tests, no el comportamiento.
# Guarda cada ejecución en <out>/<name>-base.txt y <out>/<name>-cambio.txt y borra el worktree.
#
# Imprime los tests seleccionados, el resultado de cada lado y el veredicto:
#   RED-GREEN: DISCRIMINANTE                rojo en la base porque fallan todos los tests
#                                           seleccionados, y verde en el cambio
#   RED-GREEN: DISCRIMINANTE (compilación)  la base no compila con los tests de Go: vale si lo que
#                                           falta es el comportamiento que pide el criterio
#   RED-GREEN: NO DISCRIMINANTE (<motivo>)
# Sale con 0 si es discriminante, 1 si no, 2 si el uso es incorrecto y 3 si no se puede preparar la base.
set -o pipefail

usage() {
  {
    echo "uso: red-green.sh --out <directorio en .adw/> --name <NN_afirmación-detalle> --base <ref> --run <patrón> [--extra <fichero>]... <paquete> <fichero de test> ..."
    echo "     red-green.sh --out <directorio en .adw/> --name <NN_afirmación-detalle> --base <ref> [--run <patrón>] [--extra <fichero>]... e2e/<test>.spec.ts ..."
  } >&2
  exit 2
}

root=$(git rev-parse --show-toplevel) || exit 2
cd "$root" || exit 2
# En un worktree de bin/tree, su .env da los puertos: los dos lados heredan E2E_PORT y DB_PORT.
# shellcheck disable=SC1091
[[ -f .env ]] && { set -a; . ./.env; set +a; }

out="" name="" base="" run=""
extras=()
while (($#)); do
  case $1 in
    --out | --name | --base | --run | --extra)
      [[ $# -ge 2 ]] || usage
      case $1 in
        --out) out=${2%/} ;;
        --name) name=$2 ;;
        --base) base=$2 ;;
        --run) run=$2 ;;
        --extra) extras+=("$2") ;;
      esac
      shift 2
      ;;
    -*) usage ;;
    *) break ;;
  esac
done
(($#)) || usage

# Tests de navegador si el primer fichero es de e2e/; si no, tests de Go con su paquete.
if [[ $1 == e2e/*.spec.ts ]]; then
  mode=e2e
  pkg=""
else
  mode=go
  pkg=$1
  shift
  [[ -n $run && -n $pkg ]] || usage
  (($#)) || usage
fi
[[ -n $out && -n $name && -n $base ]] || usage

out=${out#"$root"/}
[[ $out == .adw/* && $out != *..* ]] || { echo "red-green.sh: --out debe estar dentro de .adw/" >&2; exit 2; }
[[ $name =~ ^[0-9][0-9]_[a-z0-9-]+$ ]] || { echo "red-green.sh: --name debe tener la forma NN_afirmación-detalle" >&2; exit 2; }
if [[ $mode == go ]]; then
  [[ $pkg =~ ^\./[A-Za-z0-9_./-]*$ && $pkg != *..* && -d $pkg ]] || { echo "red-green.sh: el paquete debe ser un directorio del repositorio, como ./actions" >&2; exit 2; }
fi
commit=$(git rev-parse --verify --quiet "$base^{commit}") || { echo "red-green.sh: $base no es un commit" >&2; exit 2; }
for test in "$@"; do
  if [[ $mode == go ]]; then
    [[ $test == *_test.go && $test != *..* && $test != /* && -f $test ]] || { echo "red-green.sh: $test no es un fichero de test del repositorio" >&2; exit 2; }
  else
    [[ $test == e2e/*.spec.ts && $test != *..* && -f $test ]] || { echo "red-green.sh: $test no es un test de navegador de e2e/" >&2; exit 2; }
  fi
done
for extra in ${extras[@]+"${extras[@]}"}; do
  [[ $extra != *..* && $extra != /* && $extra != .* && -f $extra ]] || { echo "red-green.sh: $extra no es un fichero del repositorio" >&2; exit 2; }
done
if [[ $mode == e2e ]]; then
  [[ -f playwright.config.ts && -d node_modules/@playwright/test ]] || { echo "red-green.sh: faltan playwright.config.ts o las dependencias de Node (npm ci)" >&2; exit 3; }
fi

mkdir -p "$out" .adw/base || exit 3
# Worktrees que dejó una ejecución interrumpida.
git worktree prune
worktree="$root/.adw/base/$(basename "$out")-$$"
tmp=$(mktemp -d) || exit 3
trap 'git worktree remove --force "$worktree" > /dev/null 2>&1; rm -rf "$tmp"' EXIT
git worktree add --detach --quiet "$worktree" "$commit" || { echo "red-green.sh: no se puede crear el worktree de la base" >&2; exit 3; }
harness=()
if [[ $mode == e2e ]]; then
  harness=(playwright.config.ts)
  [[ -f e2e/fixtures.ts ]] && harness+=(e2e/fixtures.ts)
  ln -s "$root/node_modules" "$worktree/node_modules" || exit 3
fi
for file in "$@" ${extras[@]+"${extras[@]}"} ${harness[@]+"${harness[@]}"}; do
  mkdir -p "$worktree/$(dirname "$file")" && cp "$file" "$worktree/$file" || exit 3
done

if [[ $mode == go ]]; then
  shown="GO_ENV=test go test -v -p 1 -count=1 -run '$run' $pkg"
else
  shown="npx playwright test${run:+ --grep '$run'} $*"
fi

# side <etiqueta> <directorio> <fichero> <json>: ejecuta los tests y guarda la salida. Devuelve su código.
side() {
  local label=$1 dir=$2 file=$3 json=$4 start output code
  start=$SECONDS
  if [[ $mode == go ]]; then
    output=$(cd "$dir" && env GO_ENV=test go test -v -p 1 -count=1 -run "$run" "$pkg" 2>&1)
  else
    output=$(cd "$dir" && env PLAYWRIGHT_JSON_OUTPUT_FILE="$json" npx playwright test --reporter=list,json ${run:+--grep "$run"} "${tests[@]}" 2>&1)
  fi
  code=$?
  {
    echo "\$ $shown"
    echo "en: $label"
    echo "exit: $code"
    echo "duración: $((SECONDS - start))s"
    echo "---"
    printf '%s\n' "$output"
  } > "$file"
  return "$code"
}

# results <json>: una línea por test de navegador, «<título completo>\t<estado>», con el estado
# de Playwright (expected, unexpected, flaky o skipped). Los errores que impiden ejecutar tests
# (la aplicación no arranca, un import que falta) salen como «!error\t<mensaje>».
results() {
  [[ -s $1 ]] || return 0
  node -e '
    const report = JSON.parse(require("fs").readFileSync(process.argv[1], "utf8"));
    const walk = (suite, path) => {
      for (const child of suite.suites || []) walk(child, [...path, child.title]);
      for (const spec of suite.specs || [])
        for (const test of spec.tests || [])
          console.log([...path, spec.title].filter(Boolean).join(" › ") + "\t" + test.status);
    };
    for (const suite of report.suites || []) walk(suite, [suite.title]);
    for (const error of report.errors || [])
      console.log("!error\t" + String(error.message || "").split("\n")[0]);
  ' "$1"
}

tests=("$@")
base_file="$out/$name-base.txt"
change_file="$out/$name-cambio.txt"
side "base $(git rev-parse --short "$commit") con los tests del cambio" "$worktree" "$base_file" "$tmp/base.json"
base_code=$?
side "cambio (copia de trabajo)" "$root" "$change_file" "$tmp/cambio.json"
change_code=$?

if [[ $mode == e2e ]]; then
  change_results=$(results "$tmp/cambio.json")
  base_results=$(results "$tmp/base.json")
  selected=$(printf '%s\n' "$change_results" | grep -v '^!error' | cut -f1 | sed '/^$/d' | sort -u)
else
  # Tests seleccionados: los que se ejecutan a la profundidad del patrón (Suite/Test_X son dos niveles).
  depth=$(($(printf '%s' "$run" | tr -cd '/' | wc -c) + 1))
  selected=$(grep -E '^ *=== RUN' "$change_file" | awk '{print $3}' | awk -F/ -v d="$depth" 'NF == d' | sort -u)
fi

echo "tests seleccionados: ${selected:-ninguno}" | tr '\n' ';' | sed 's/;$//; s/;/; /g'
echo
echo "base:   exit=$base_code $base_file"
echo "cambio: exit=$change_code $change_file"

# status <resultados> <test>: el estado de un test de navegador en un lado.
status() { printf '%s\n' "$1" | awk -F'\t' -v t="$2" '$1 == t {print $2; exit}'; }

if [[ $mode == e2e ]]; then
  if printf '%s\n' "$change_results" | grep -q '^!error'; then
    echo "RED-GREEN: NO DISCRIMINANTE (el cambio no ejecuta los tests: revisa $change_file)"
    exit 1
  fi
  if [[ -z $selected ]]; then
    echo "RED-GREEN: NO DISCRIMINANTE (no se selecciona ningún test)"
    exit 1
  fi
  while IFS= read -r test; do
    if [[ $(status "$change_results" "$test") != expected ]]; then
      echo "RED-GREEN: NO DISCRIMINANTE ($test no pasa en el cambio)"
      exit 1
    fi
  done <<< "$selected"
  if ((change_code != 0)); then
    echo "RED-GREEN: NO DISCRIMINANTE (el cambio no está en verde)"
    exit 1
  fi
  if ((base_code == 0)); then
    echo "RED-GREEN: NO DISCRIMINANTE (los tests también pasan en la base)"
    exit 1
  fi
  # Un rojo porque la aplicación de la base no arranca o le falta un fichero no demuestra el comportamiento.
  if printf '%s\n' "$base_results" | grep -q '^!error'; then
    echo "RED-GREEN: NO DISCRIMINANTE (la base no ejecuta los tests: no arranca o le falta un fichero, que se pasa con --extra; revisa $base_file)"
    exit 1
  fi
  while IFS= read -r test; do
    if [[ $(status "$base_results" "$test") != unexpected ]]; then
      echo "RED-GREEN: NO DISCRIMINANTE ($test no falla en la base: revisa $base_file)"
      exit 1
    fi
  done <<< "$selected"
  echo "RED-GREEN: DISCRIMINANTE"
  exit 0
fi

if grep -qE '\[(build|setup) failed\]' "$change_file"; then
  echo "RED-GREEN: NO DISCRIMINANTE (el cambio no compila con los tests)"
  exit 1
fi
if [[ -z $selected ]]; then
  echo "RED-GREEN: NO DISCRIMINANTE (el patrón no selecciona ningún test)"
  exit 1
fi
for test in $selected; do
  if ! grep -qF -- "--- PASS: $test " "$change_file"; then
    echo "RED-GREEN: NO DISCRIMINANTE ($test no pasa en el cambio)"
    exit 1
  fi
done
if ((change_code != 0)); then
  echo "RED-GREEN: NO DISCRIMINANTE (el cambio no está en verde)"
  exit 1
fi
if ((base_code == 0)); then
  echo "RED-GREEN: NO DISCRIMINANTE (los tests también pasan en la base)"
  exit 1
fi
if grep -qE '\[(build|setup) failed\]' "$base_file"; then
  echo "RED-GREEN: DISCRIMINANTE (compilación)"
  exit 0
fi
for test in $selected; do
  if ! grep -qF -- "--- FAIL: $test " "$base_file"; then
    echo "RED-GREEN: NO DISCRIMINANTE ($test no falla en la base: revisa $base_file)"
    exit 1
  fi
done
# Un rojo porque a la base le falta un fixture o un fichero no demuestra el comportamiento.
if grep -qiE 'could not find (a )?scenario|no such file or directory' "$base_file"; then
  echo "RED-GREEN: NO DISCRIMINANTE (la base falla por un fichero que no tiene: pásalo con --extra)"
  exit 1
fi
echo "RED-GREEN: DISCRIMINANTE"
exit 0
