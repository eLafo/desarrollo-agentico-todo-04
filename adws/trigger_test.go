package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	newIssue      = "https://github.com/example/todo/issues/12"
	oldIssue      = "https://github.com/example/todo/issues/3"
	defaultConfig = "repo: example/todo\nlabel: adw\ninterval_seconds: 20\nprompt: /adw {{payload}}\n"
)

var queryJSON = []string{"--json", "url,number,title,labels,assignees,updatedAt"}

// snapshot devuelve la foto JSON compacta de una issue, como la pasa el trigger.
func snapshot(url string, labels []string, assignees ...string) string {
	type label struct {
		Name string `json:"name"`
	}
	type user struct {
		Login string `json:"login"`
	}
	number, err := strconv.Atoi(path.Base(url))
	if err != nil {
		panic(err)
	}
	it := struct {
		URL       string  `json:"url"`
		Number    int     `json:"number"`
		Labels    []label `json:"labels"`
		Assignees []user  `json:"assignees"`
	}{URL: url, Number: number, Labels: []label{{Name: "adw"}}, Assignees: []user{}}
	for _, name := range labels {
		it.Labels = append(it.Labels, label{Name: name})
	}
	for _, login := range assignees {
		it.Assignees = append(it.Assignees, user{Login: login})
	}
	data, err := json.Marshal(it)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func list(snapshots ...string) string {
	return "[" + strings.Join(snapshots, ",") + "]\n"
}

func issueURL(number int) string {
	return fmt.Sprintf("https://github.com/example/todo/issues/%d", number)
}

var (
	newSnapshot = snapshot(newIssue, nil)
	oldSnapshot = snapshot(oldIssue, nil)
)

func TestTriggerHelperProcess(t *testing.T) {
	if os.Getenv("TEST_TRIGGER_HELPER") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			if err := run(os.Args[i+1:]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(exitCode(err))
			}
			os.Exit(0)
		}
	}
	os.Exit(90)
}

// Los ejecutables simulados. Escriben lo que les pasa en ficheros de TEST_ROOT:
//   - events: START y END de cada sesión, SLEEP de cada espera y REMOVE de cada borrado;
//   - gh-args: los argumentos de la última consulta de issues;
//   - sessions: el directorio y el puerto de cada sesión;
//   - tree-log: las órdenes de bin/tree.
const (
	fakeGH = `#!/usr/bin/env bash
set -euo pipefail
if [[ "$*" == "api user --jq .login" ]]; then
  [[ "${TEST_GH_USER_EXIT:-0}" == 0 ]] || exit "$TEST_GH_USER_EXIT"
  printf 'agente\n'
  exit 0
fi
if [[ "$1 $2" == "issue view" ]]; then
  printf '%s\n' "$*" >> "$TEST_ROOT/gh-views"
  cat "$TEST_ROOT/state-$3" 2>/dev/null || printf 'OPEN\n'
  exit 0
fi
printf '%s\n' "$@" > "$TEST_ROOT/gh-args"
[[ "${TEST_GH_EXIT:-0}" == 0 ]] || exit "$TEST_GH_EXIT"
calls=$(( $(cat "$TEST_ROOT/list-calls" 2>/dev/null || echo 0) + 1 ))
echo "$calls" > "$TEST_ROOT/list-calls"
if [[ -f "$TEST_ROOT/issues.$calls" ]]; then cat "$TEST_ROOT/issues.$calls"; else cat "$TEST_ROOT/issues"; fi
`
	fakeClaude = `#!/usr/bin/env bash
set -euo pipefail
[[ $# == 4 && "$1" == -p && "$3" == --permission-prompts && "$4" == none ]] || exit 90
if IFS= read -r unexpected; then exit 91; fi
[[ "${CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS:-}" == 0 ]] || exit 92
printf 'START:%s\n' "$2" >> "$TEST_ROOT/events"
printf '%s' "$PWD" > "$TEST_ROOT/cwd"
printf 'CWD:%s APP_PORT:%s\n' "$PWD" "${APP_PORT:-}" >> "$TEST_ROOT/sessions"
echo "hola desde $(basename "$PWD")"
[[ "${TEST_CLAUDE_EXIT:-0}" == 0 ]] || exit "$TEST_CLAUDE_EXIT"
[[ "${TEST_CLAUDE_FAIL_IN:-}" != "$(basename "$PWD")" ]] || exit 9
# Espera a que haya TEST_CLAUDE_PEERS sesiones empezadas: solo pasa si van en paralelo.
for _ in $(seq 1 250); do
  [[ $(grep -c '^START:' "$TEST_ROOT/events") -ge ${TEST_CLAUDE_PEERS:-1} ]] && break
  /bin/sleep 0.02
done
[[ $(grep -c '^START:' "$TEST_ROOT/events") -ge ${TEST_CLAUDE_PEERS:-1} ]] || exit 95
# Con TEST_CLAUDE_HOLD, sigue en marcha hasta que exista el fichero release.
if [[ -n "${TEST_CLAUDE_HOLD:-}" ]]; then
  until [[ -f "$TEST_ROOT/release" ]]; do /bin/sleep 0.02; done
fi
/bin/sleep "${TEST_CLAUDE_SLEEP:-0.02}"
printf 'END:%s\n' "$2" >> "$TEST_ROOT/events"
`
	fakeGit = `#!/usr/bin/env bash
set -euo pipefail
if [[ "$*" == "rev-parse --show-toplevel" ]]; then
  printf '%s\n' "$TEST_ROOT"
  exit 0
fi
[[ $# == 4 && "$1" == -C && "$3 $4" == "status --porcelain" ]] || exit 90
[[ "${TEST_GIT_DIRTY:-}" != "$(basename "$2")" ]] || printf ' M main.go\n'
`
	fakeTree = `#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >> "$TEST_ROOT/tree-log"
dir="$TEST_ROOT/trees/issue-$2"
case $1 in
  create)
    mkdir -p "$dir"
    [[ -f "$dir/.env" ]] || printf '# generado\nCOMPOSE_PROJECT_NAME=todo-issue-%s\nAPP_PORT=%s\n' "$2" $((13000 + $2)) > "$dir/.env"
    printf '%s\n' "$dir"
    ;;
  remove)
    if [[ "${TEST_TREE_REMOVE_FAIL:-}" == "$2" ]]; then
      echo "bin/tree: trees/issue-$2 tiene cambios sin registrar (--force los descarta)" >&2
      exit 1
    fi
    rm -rf "$dir"
    printf 'REMOVE:%s\n' "$2" >> "$TEST_ROOT/events"
    echo "Borrado trees/issue-$2"
    ;;
esac
`
	// Por defecto, la primera espera vuelve y la segunda termina el trigger con 73. Con
	// TEST_SLEEP_RELEASE, la segunda suelta las sesiones retenidas y espera a que acabe una, y es
	// la tercera la que termina.
	fakeSleep = `#!/usr/bin/env bash
set -euo pipefail
printf 'SLEEP:%s\n' "$1" >> "$TEST_ROOT/events"
calls=$(( $(cat "$TEST_ROOT/sleep-calls" 2>/dev/null || echo 0) + 1 ))
echo "$calls" > "$TEST_ROOT/sleep-calls"
last=2
[[ -z "${TEST_SLEEP_RELEASE:-}" ]] || last=3
if [[ -n "${TEST_SLEEP_RELEASE:-}" && $calls == 2 ]]; then
  touch "$TEST_ROOT/release"
  until grep -q '^END:' "$TEST_ROOT/events"; do /bin/sleep 0.02; done
  /bin/sleep 0.3 # a que el trigger recoja la sesión
fi
if (( calls >= last )); then exit 73; fi
`
)

type result struct {
	output, events, args string
	err                  error
}

func runTrigger(t *testing.T, issues, configBody string, env []string, args ...string) (string, string, string, error) {
	t.Helper()
	r := runTriggerIn(t, t.TempDir(), issues, configBody, env, args...)
	return r.output, r.events, r.args, r.err
}

func writeFile(t *testing.T, name, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(name)
	if os.IsNotExist(err) {
		return ""
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func runTriggerIn(t *testing.T, root, issues, configBody string, env []string, args ...string) result {
	t.Helper()
	project, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	if configBody == "" {
		configBody = defaultConfig
	}
	writeFile(t, filepath.Join(root, "adw.yml"), configBody)
	writeFile(t, filepath.Join(root, "issues"), issues)
	writeFile(t, filepath.Join(root, "fake", "gh"), fakeGH)
	writeFile(t, filepath.Join(root, "fake", "claude"), fakeClaude)
	writeFile(t, filepath.Join(root, "fake", "git"), fakeGit)
	writeFile(t, filepath.Join(root, "fake", "sleep"), fakeSleep)
	writeFile(t, filepath.Join(root, "bin", "tree"), fakeTree)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	childArgs := []string{"-test.run=^TestTriggerHelperProcess$", "--", "--config", filepath.Join(root, "adw.yml")}
	cmd := exec.CommandContext(ctx, os.Args[0], append(childArgs, args...)...)
	cmd.Dir = project
	for _, value := range env {
		if value == "TEST_USE_WRAPPER=1" {
			wrapperArgs := []string{filepath.Join(project, "adws/trigger-github.sh"), "--config", filepath.Join(root, "adw.yml")}
			cmd = exec.CommandContext(ctx, "bash", append(wrapperArgs, args...)...)
			cmd.Dir = root
		}
	}
	cmd.Env = append(os.Environ(),
		"PATH="+filepath.Join(root, "fake")+string(os.PathListSeparator)+os.Getenv("PATH"),
		"TEST_ROOT="+root, "TEST_TRIGGER_HELPER=1", "TEST_GH_EXIT=0", "TEST_CLAUDE_EXIT=0",
	)
	cmd.Env = append(cmd.Env, env...)
	output, runErr := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("trigger exceeded test timeout: %s", output)
	}
	if cwd := readFile(t, filepath.Join(root, "cwd")); cwd != "" && !strings.HasPrefix(cwd, filepath.Join(root, "trees", "issue-")) {
		t.Fatalf("claude must run in the worktree of its case, not in %s", cwd)
	}
	return result{string(output), readFile(t, filepath.Join(root, "events")), readFile(t, filepath.Join(root, "gh-args")), runErr}
}

// withoutSleeps quita las esperas de los eventos: en modo continuo, las sesiones corren mientras
// el trigger espera, y el orden entre unas y otras no está fijado.
func withoutSleeps(events string) string {
	var kept []string
	for _, line := range strings.Split(strings.TrimSpace(events), "\n") {
		if !strings.HasPrefix(line, "SLEEP:") {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}

// maxOverlap cuenta el máximo de sesiones en marcha a la vez según los eventos.
func maxOverlap(events string) int {
	running, max := 0, 0
	for _, line := range strings.Split(events, "\n") {
		switch {
		case strings.HasPrefix(line, "START:"):
			running++
		case strings.HasPrefix(line, "END:"):
			running--
		}
		if running > max {
			max = running
		}
	}
	return max
}

func TestTriggerOnceRunsIssuesInSequence(t *testing.T) {
	output, events, args, err := runTrigger(t, list(newSnapshot, oldSnapshot), "", nil, "--once")
	if err != nil {
		t.Fatalf("%v: %s", err, output)
	}
	wantArgs := append([]string{"issue", "list", "--repo", "example/todo", "--state", "open", "--label", "adw", "--search", "sort:created-desc", "--limit", "100"}, queryJSON...)
	if got := strings.Split(strings.TrimSpace(args), "\n"); !reflect.DeepEqual(got, wantArgs) {
		t.Fatalf("query must request newest open adw issues: %q", got)
	}
	want := "START:/adw " + newSnapshot + "\nEND:/adw " + newSnapshot + "\nSTART:/adw " + oldSnapshot + "\nEND:/adw " + oldSnapshot + "\n"
	if events != want {
		t.Fatalf("with max_parallel 1 (the default), want sequential invocations %q, got %q", want, events)
	}
}

func TestTriggerRunsEachCaseInItsWorktree(t *testing.T) {
	root := t.TempDir()
	r := runTriggerIn(t, root, list(newSnapshot, oldSnapshot), "", nil, "--once")
	if r.err != nil {
		t.Fatalf("%v: %s", r.err, r.output)
	}
	sessions := readFile(t, filepath.Join(root, "sessions"))
	for _, n := range []int{12, 3} {
		want := fmt.Sprintf("CWD:%s APP_PORT:%d\n", filepath.Join(root, "trees", fmt.Sprintf("issue-%d", n)), 13000+n)
		if !strings.Contains(sessions, want) {
			t.Fatalf("each session must run in its worktree with the variables of its .env: want %q in %q", want, sessions)
		}
		if line := fmt.Sprintf("[#%d] hola desde issue-%d\n", n, n); !strings.Contains(r.output, line) {
			t.Fatalf("each session's output must be prefixed with its issue: want %q in %s", line, r.output)
		}
	}
	if log := readFile(t, filepath.Join(root, "tree-log")); log != "create 12\ncreate 3\n" {
		t.Fatalf("the worktree of each case must be created with bin/tree before its session: %q", log)
	}
}

func TestTriggerRunsCasesInParallel(t *testing.T) {
	body := defaultConfig + "max_parallel: 2\n"
	output, events, _, err := runTrigger(t, list(newSnapshot, oldSnapshot), body, []string{"TEST_CLAUDE_PEERS=2"}, "--once")
	if err != nil {
		t.Fatalf("with max_parallel 2, both sessions must run at the same time: %v: %s", err, output)
	}
	if maxOverlap(events) != 2 {
		t.Fatalf("want two sessions at once: %q", events)
	}
}

func TestTriggerCapsParallelSessions(t *testing.T) {
	body := defaultConfig + "max_parallel: 2\n"
	issues := list(newSnapshot, oldSnapshot, snapshot(issueURL(7), nil))
	output, events, _, err := runTrigger(t, issues, body, []string{"TEST_CLAUDE_SLEEP=0.3"}, "--once")
	if err != nil {
		t.Fatalf("%v: %s", err, output)
	}
	if strings.Count(events, "END:") != 3 || maxOverlap(events) != 2 {
		t.Fatalf("--once must run every case, never more than max_parallel at once: %q", events)
	}
}

func TestTriggerWaitsForBackgroundAgentsWithoutCeiling(t *testing.T) {
	env := []string{"CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS=600000"}
	output, events, _, err := runTrigger(t, list(newSnapshot), "", env, "--once")
	if err != nil {
		t.Fatalf("claude must run with CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS=0, even if the environment sets another value: %v: %s", err, output)
	}
	if want := "START:/adw " + newSnapshot + "\nEND:/adw " + newSnapshot + "\n"; events != want {
		t.Fatalf("want %q, got %q", want, events)
	}
}

func TestTriggerRepoOptionalOmitsRepoFlag(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{name: "omitted", body: strings.Replace(defaultConfig, "repo: example/todo\n", "", 1)},
		{name: "empty", body: strings.Replace(defaultConfig, "example/todo", "\"\"", 1)},
		{name: "whitespace", body: strings.Replace(defaultConfig, "example/todo", "'  '", 1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, events, args, err := runTrigger(t, list(newSnapshot), tc.body, nil, "--once")
			wantArgs := append([]string{"issue", "list", "--state", "open", "--label", "adw", "--search", "sort:created-desc", "--limit", "100"}, queryJSON...)
			wantEvents := "START:/adw " + newSnapshot + "\nEND:/adw " + newSnapshot + "\n"
			if err != nil || events != wantEvents || !reflect.DeepEqual(strings.Split(strings.TrimSpace(args), "\n"), wantArgs) {
				t.Fatalf("empty/omitted repo must delegate resolution to gh: err=%v, args=%q, events=%q, output=%s", err, args, events, output)
			}
		})
	}
}

func TestTriggerRepoOverride(t *testing.T) {
	body := strings.Replace(defaultConfig, "example/todo", "\" custom/todo \"", 1)
	output, _, args, err := runTrigger(t, "", body, []string{"GH_REPO=example/ambient"}, "--once")
	if err != nil || !strings.Contains(args, "--repo\ncustom/todo\n") {
		t.Fatalf("configured repo must be an explicit override: err=%v, args=%q, output=%s", err, args, output)
	}
}

func TestTriggerDefaultRepoErrorDoesNotLaunchClaude(t *testing.T) {
	body := strings.Replace(defaultConfig, "repo: example/todo\n", "", 1)
	output, events, args, err := runTrigger(t, list(newSnapshot), body, []string{"TEST_GH_EXIT=7"}, "--once")
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 7 || events != "" || strings.Contains(args, "--repo\n") || !strings.Contains(output, "consulta de issues") {
		t.Fatalf("gh context/query error must stop before Claude: err=%v, args=%q, events=%q, output=%s", err, args, events, output)
	}
}

func TestTriggerShellWrapper(t *testing.T) {
	output, events, _, err := runTrigger(t, list(newSnapshot), "", []string{"TEST_USE_WRAPPER=1", "GOFLAGS=-mod=readonly"}, "--once")
	want := "START:/adw " + newSnapshot + "\nEND:/adw " + newSnapshot + "\n"
	if err != nil || events != want {
		t.Fatalf("wrapper must load YAML and launch the configured prompt: err=%v, events=%q, output=%s", err, events, output)
	}
}

func TestTriggerConfigurablePromptIsLiteral(t *testing.T) {
	body := `repo: example/todo
label: mi-etiqueta
interval_seconds: 7
prompt: >-
  /mi-flujo {{payload}} --remote origin;
  $(exit 99) "{{payload}}"
`
	_, events, args, err := runTrigger(t, list(newSnapshot), body, []string{"ADW_PROMPT=este valor antiguo no se usa"})
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 73 {
		t.Fatalf("unexpected exit: %v", err)
	}
	template := `/mi-flujo {{payload}} --remote origin; $(exit 99) "{{payload}}"`
	prompt := strings.ReplaceAll(template, "{{payload}}", newSnapshot)
	if !strings.Contains(events, "START:"+prompt+"\n") || !strings.Contains(events, "END:"+prompt+"\n") || strings.Count(events, "SLEEP:7\n") != 2 || !strings.Contains(args, "--label\nmi-etiqueta\n") {
		t.Fatalf("YAML config must control the literal prompt, label and interval: %q, %q", events, args)
	}
}

func TestTriggerPromptCannotOverridePermissionPrompts(t *testing.T) {
	body := strings.Replace(defaultConfig, "prompt: /adw {{payload}}", "prompt: /adw {{payload}} --permission-prompts terminal", 1)
	output, events, _, err := runTrigger(t, list(newSnapshot), body, nil, "--once")
	prompt := "/adw " + newSnapshot + " --permission-prompts terminal"
	want := "START:" + prompt + "\nEND:" + prompt + "\n"
	if err != nil || events != want {
		t.Fatalf("prompt must stay literal while CLI permission prompts remain none: err=%v, events=%q, output=%s", err, events, output)
	}
}

func TestTriggerEmptyListDoesNotInvokeClaude(t *testing.T) {
	for _, issues := range []string{"", "[]\n"} {
		output, events, _, err := runTrigger(t, issues, "", nil, "--once")
		if err != nil || events != "" {
			t.Fatalf("issues=%q: err=%v, events=%q, output=%s", issues, err, events, output)
		}
	}
}

func TestTriggerOnlyLaunchesCasesInAgentTurn(t *testing.T) {
	issues := list(
		snapshot(issueURL(1), []string{"adw:waiting-for-human"}),
		snapshot(issueURL(2), []string{"stage:planning"}),
		snapshot(issueURL(3), nil, "otra-cuenta"),
		snapshot(issueURL(4), nil, "agente", "otra-cuenta"),
		snapshot(issueURL(5), []string{"feature", "plan:ready"}, "Agente"),
		snapshot(issueURL(6), []string{"issue:failed"}),
	)
	output, events, _, err := runTrigger(t, issues, "", nil, "--once")
	mine := snapshot(issueURL(5), []string{"feature", "plan:ready"}, "Agente")
	free := snapshot(issueURL(6), []string{"issue:failed"})
	want := "START:/adw " + mine + "\nEND:/adw " + mine + "\nSTART:/adw " + free + "\nEND:/adw " + free + "\n"
	if err != nil || events != want {
		t.Fatalf("only free or own cases without waiting or stage labels must launch: err=%v, events=%q, output=%s", err, events, output)
	}
}

func TestTriggerSkipsDirtyWorktree(t *testing.T) {
	output, events, _, err := runTrigger(t, list(newSnapshot, oldSnapshot), "", []string{"TEST_GIT_DIRTY=issue-12"}, "--once")
	want := "START:/adw " + oldSnapshot + "\nEND:/adw " + oldSnapshot + "\n"
	if err != nil || events != want || !strings.Contains(output, "[#12] ") || !strings.Contains(output, "cambios sin registrar") {
		t.Fatalf("a dirty worktree must skip its case and let the others run: err=%v, events=%q, output=%s", err, events, output)
	}
}

func TestTriggerStopsWhenGitHubAccountIsUnknown(t *testing.T) {
	output, events, args, err := runTrigger(t, list(newSnapshot), "", []string{"TEST_GH_USER_EXIT=4"}, "--once")
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 4 || events != "" || args != "" || !strings.Contains(output, "cuenta de gh") {
		t.Fatalf("without the gh account the trigger cannot tell its own cases: err=%v, events=%q, args=%q, output=%s", err, events, args, output)
	}
}

func TestTriggerStopsOnGitHubError(t *testing.T) {
	_, events, _, err := runTrigger(t, list(newSnapshot), "", []string{"TEST_GH_EXIT=7"}, "--once")
	if err == nil || events != "" {
		t.Fatalf("err=%v, events=%q", err, events)
	}
}

func TestTriggerStopsOnClaudeError(t *testing.T) {
	_, events, _, err := runTrigger(t, list(newSnapshot, oldSnapshot), "", []string{"TEST_CLAUDE_EXIT=9"}, "--once")
	if err == nil || events != "START:/adw "+newSnapshot+"\n" {
		t.Fatalf("must stop before launching second issue: err=%v, events=%q", err, events)
	}
}

func TestTriggerSessionErrorWaitsForTheOthers(t *testing.T) {
	body := defaultConfig + "max_parallel: 2\n"
	issues := list(newSnapshot, oldSnapshot, snapshot(issueURL(7), nil))
	output, events, _, err := runTrigger(t, issues, body, []string{"TEST_CLAUDE_FAIL_IN=issue-12", "TEST_CLAUDE_SLEEP=0.3"})
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 9 {
		t.Fatalf("a failed session must end the trigger with its exit code: %v: %s", err, output)
	}
	if !strings.Contains(events, "END:/adw "+oldSnapshot+"\n") || strings.Contains(events, "issues/7") {
		t.Fatalf("after a failure, the trigger must wait for the running sessions and launch no more: %q", events)
	}
}

func TestTriggerRemembersOnlyRunningSessions(t *testing.T) {
	_, events, _, err := runTrigger(t, list(newSnapshot), "", []string{"TEST_CLAUDE_HOLD=1", "TEST_SLEEP_RELEASE=1"})
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 73 {
		t.Fatalf("unexpected exit: %v", err)
	}
	session := "START:/adw " + newSnapshot + "\nEND:/adw " + newSnapshot
	if withoutSleeps(events) != session+"\n"+session || strings.Count(events, "SLEEP:") != 3 {
		t.Fatalf("a running case must not be launched again, and a finished one must: %q", events)
	}
}

func TestTriggerRemovesWorktreesOfClosedIssues(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"issue-12", "issue-40", "issue-41", "notas"} {
		if err := os.MkdirAll(filepath.Join(root, "trees", dir), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	writeFile(t, filepath.Join(root, "state-40"), "CLOSED\n")
	writeFile(t, filepath.Join(root, "state-41"), "OPEN\n")
	r := runTriggerIn(t, root, list(newSnapshot), "", nil, "--once")
	if r.err != nil {
		t.Fatalf("%v: %s", r.err, r.output)
	}
	if log := readFile(t, filepath.Join(root, "tree-log")); log != "remove 40\ncreate 12\n" {
		t.Fatalf("only the worktree of the closed issue must be removed, before launching: %q", log)
	}
	if views := readFile(t, filepath.Join(root, "gh-views")); strings.Contains(views, "view 12 ") || !strings.Contains(views, "view 40 --repo example/todo --json state --jq .state\n") {
		t.Fatalf("the state is only asked for worktrees whose issue is not in the open list: %q", views)
	}
	if !strings.Contains(r.output, "[#40] issue cerrada: Borrado trees/issue-40\n") {
		t.Fatalf("the removal must be reported: %s", r.output)
	}
}

func TestTriggerKeepsWorktreeOfRunningCase(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "issues.2"), "[]\n")
	writeFile(t, filepath.Join(root, "issues.3"), "[]\n")
	writeFile(t, filepath.Join(root, "state-12"), "CLOSED\n")
	r := runTriggerIn(t, root, list(newSnapshot), "", []string{"TEST_CLAUDE_HOLD=1", "TEST_SLEEP_RELEASE=1"})
	if exit, ok := r.err.(*exec.ExitError); !ok || exit.ExitCode() != 73 {
		t.Fatalf("unexpected exit: %v: %s", r.err, r.output)
	}
	want := "START:/adw " + newSnapshot + "\nEND:/adw " + newSnapshot + "\nREMOVE:12"
	if withoutSleeps(r.events) != want {
		t.Fatalf("the worktree of a running case must be kept until its session ends: %q", r.events)
	}
}

func TestTriggerWarnsOnceWhenWorktreeCannotBeRemoved(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "trees", "issue-40"), 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "state-40"), "CLOSED\n")
	r := runTriggerIn(t, root, "[]\n", "", []string{"TEST_TREE_REMOVE_FAIL=40"})
	if exit, ok := r.err.(*exec.ExitError); !ok || exit.ExitCode() != 73 {
		t.Fatalf("a worktree that cannot be removed must not stop the trigger: %v: %s", r.err, r.output)
	}
	if log := readFile(t, filepath.Join(root, "tree-log")); log != "remove 40\nremove 40\n" {
		t.Fatalf("the removal must be retried on each poll: %q", log)
	}
	if strings.Count(r.output, "[#40] issue cerrada, pero no se borra su worktree: bin/tree: trees/issue-40 tiene cambios sin registrar") != 1 {
		t.Fatalf("the refusal must be reported once: %s", r.output)
	}
}

func TestTriggerInvalidConfigurationDoesNotLaunch(t *testing.T) {
	cases := []struct {
		name string
		body string
		args []string
	}{
		{name: "missing-label", body: strings.Replace(defaultConfig, "label: adw", "label: ''", 1)},
		{name: "zero-interval", body: strings.Replace(defaultConfig, "interval_seconds: 20", "interval_seconds: 0", 1)},
		{name: "negative-interval", body: strings.Replace(defaultConfig, "interval_seconds: 20", "interval_seconds: -1", 1)},
		{name: "zero-max-rejections", body: defaultConfig + "max_rejections: 0\n"},
		{name: "negative-max-rejections", body: defaultConfig + "max_rejections: -2\n"},
		{name: "non-integer-max-rejections", body: defaultConfig + "max_rejections: tres\n"},
		{name: "zero-max-parallel", body: defaultConfig + "max_parallel: 0\n"},
		{name: "negative-max-parallel", body: defaultConfig + "max_parallel: -1\n"},
		{name: "non-integer-max-parallel", body: defaultConfig + "max_parallel: dos\n"},
		{name: "missing-payload-marker", body: strings.Replace(defaultConfig, "{{payload}}", "", 1)},
		{name: "old-issue-marker", body: strings.Replace(defaultConfig, "{{payload}}", "{{issue_url}}", 1)},
		{name: "malformed-yaml", body: "repo: [\n"},
		{name: "unknown-field", body: defaultConfig + "unexpected: true\n"},
		{name: "duplicate-field", body: defaultConfig + "repo: otro/repo\n"},
		{name: "multiple-documents", body: defaultConfig + "---\n" + defaultConfig},
		{name: "unknown-option", args: []string{"--unknown"}},
		{name: "old-repo-argument", args: []string{"example/todo"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, events, ghArgs, err := runTrigger(t, list(newSnapshot), tc.body, nil, tc.args...)
			if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 2 || events != "" || ghArgs != "" {
				t.Fatalf("invalid configuration must stop before querying: err=%v, events=%q, args=%q, output=%s", err, events, ghArgs, output)
			}
		})
	}
}

func TestTriggerConfigDefaults(t *testing.T) {
	cases := []struct {
		name           string
		body           string
		wantRejections int
		wantParallel   int
	}{
		{name: "default", body: defaultConfig, wantRejections: 3, wantParallel: 1},
		{name: "loop-disabled", body: defaultConfig + "max_rejections: 1\n", wantRejections: 1, wantParallel: 1},
		{name: "configured", body: defaultConfig + "max_rejections: 5\nmax_parallel: 3\n", wantRejections: 5, wantParallel: 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "adw.yml")
			if err := os.WriteFile(path, []byte(tc.body), 0o600); err != nil {
				t.Fatal(err)
			}
			cfg, err := readConfig(path)
			if err != nil {
				t.Fatalf("max_rejections and max_parallel must accept integers from 1: %v", err)
			}
			if cfg.MaxRejections != tc.wantRejections || cfg.MaxParallel != tc.wantParallel {
				t.Fatalf("want max_rejections %d and max_parallel %d, got %d and %d", tc.wantRejections, tc.wantParallel, cfg.MaxRejections, cfg.MaxParallel)
			}
		})
	}
}

func TestTriggerReadsWorktreeEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	writeFile(t, path, "# comentario\n\nCOMPOSE_PROJECT_NAME=todo-issue-22\nAPP_PORT=13022\n")
	env, err := readEnv(path)
	if err != nil || !reflect.DeepEqual(env, []string{"COMPOSE_PROJECT_NAME=todo-issue-22", "APP_PORT=13022"}) {
		t.Fatalf("readEnv must return the variables of the .env: %q, %v", env, err)
	}
	writeFile(t, path, "export APP_PORT=13022\n")
	if _, err := readEnv(path); err == nil {
		t.Fatal("readEnv must reject lines that are not KEY=VALUE")
	}
	if env, err := readEnv(filepath.Join(t.TempDir(), "missing")); err != nil || env != nil {
		t.Fatalf("a missing .env means no variables: %q, %v", env, err)
	}
}

func TestTriggerMissingConfig(t *testing.T) {
	if _, err := readConfig(filepath.Join(t.TempDir(), "missing.yml")); err == nil {
		t.Fatal("missing config must report an error")
	}
}

func TestTriggerProjectConfig(t *testing.T) {
	if _, err := readConfig("../config/adw.yml"); err != nil {
		t.Fatalf("invalid project config: %v", err)
	}
}
