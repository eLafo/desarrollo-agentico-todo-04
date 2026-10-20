#!/usr/bin/env bash
# Logs de las sesiones de adw, uno por invocación: .adw/logs/<issue>/<iteración>_<sesión>.jsonl
#
#   log.sh hook                    Lo llaman los hooks del proyecto con la entrada del hook en stdin.
#   log.sh case <sesión> <issue>   Asocia la sesión a su issue y mueve allí lo registrado.
#
# Una sesión se registra cuando invoca adw, por el prompt (/adw …) o por la herramienta Skill.
# Hasta que el coordinador la asocia a su issue, escribe en unknown-<n>.

# En el checkout principal, también si la sesión corre en un worktree de bin/tree: así los logs
# sobreviven al borrado del worktree.
project=${CLAUDE_PROJECT_DIR:-$(pwd)}
if common=$(git -C "$project" rev-parse --path-format=absolute --git-common-dir 2> /dev/null); then
  project=$(dirname "$common")
fi
logs="$project/.adw/logs"

register() {
  local session=$1 count
  mkdir -p "$logs/.sessions"
  count=$(( $(cat "$logs/.unknown-count" 2>/dev/null || echo 0) + 1 ))
  echo "$count" > "$logs/.unknown-count"
  mkdir -p "$logs/unknown-$count"
  echo "unknown-$count/1_$session.jsonl" > "$logs/.sessions/$session"
}

hook() {
  local input session
  input=$(cat)
  session=$(sed -n 's/^{"session_id":"\([A-Za-z0-9-]*\)".*/\1/p' <<<"$input")
  [[ -n $session ]] || return 0
  if [[ ! -f $logs/.sessions/$session ]]; then
    if grep -qE '"hook_event_name":"UserPromptSubmit".*"prompt":"/adw( |")' <<<"$input" ||
      { grep -q '"tool_name":"Skill"' <<<"$input" && grep -q '"skill":"adw"' <<<"$input"; }; then
      register "$session"
    else
      return 0
    fi
  fi
  printf '%s\n' "$input" >> "$logs/$(cat "$logs/.sessions/$session")"
}

bind() {
  local session=$1 issue=${2##*/} current target iteration
  issue=${issue#\#}
  if [[ ! $session =~ ^[A-Za-z0-9-]+$ || ! $issue =~ ^[0-9]+$ ]]; then
    echo "uso: log.sh case <sesión> <número de issue>" >&2
    return 2
  fi
  [[ -f $logs/.sessions/$session ]] || register "$session"
  current=$(cat "$logs/.sessions/$session")
  if [[ $current == "$issue"/* ]]; then
    echo "Sesión ya asociada: .adw/logs/$current"
    return 0
  fi
  mkdir -p "$logs/$issue"
  iteration=$(( $(find "$logs/$issue" -name '*.jsonl' | wc -l) + 1 ))
  target="$issue/${iteration}_$session.jsonl"
  if [[ -f $logs/$current ]]; then
    mv "$logs/$current" "$logs/$target"
  fi
  rmdir "$logs/$(dirname "$current")" 2>/dev/null
  echo "$target" > "$logs/.sessions/$session"
  echo "Sesión asociada: .adw/logs/$target"
}

case ${1:-} in
  hook) hook || true ;;
  case) bind "${2:-}" "${3:-}" ;;
  *) echo "uso: log.sh hook | log.sh case <sesión> <issue>" >&2; exit 2 ;;
esac
