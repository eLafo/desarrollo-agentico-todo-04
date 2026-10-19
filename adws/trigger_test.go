package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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
	it := struct {
		URL       string  `json:"url"`
		Labels    []label `json:"labels"`
		Assignees []user  `json:"assignees"`
	}{URL: url, Labels: []label{{Name: "adw"}}, Assignees: []user{}}
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

func runTrigger(t *testing.T, issues, configBody string, env []string, args ...string) (string, string, string, error) {
	t.Helper()
	project, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if configBody == "" {
		configBody = defaultConfig
	}
	write("adw.yml", configBody)
	write("issues", issues)
	write("gh", `#!/usr/bin/env bash
set -euo pipefail
if [[ "$*" == "api user --jq .login" ]]; then
  [[ "${TEST_GH_USER_EXIT:-0}" == 0 ]] || exit "$TEST_GH_USER_EXIT"
  printf 'agente\n'
  exit 0
fi
printf '%s\n' "$@" > "$TEST_ROOT/gh-args"
[[ "${TEST_GH_EXIT:-0}" == 0 ]] || exit "$TEST_GH_EXIT"
cat "$TEST_ROOT/issues"
`)
	write("claude", `#!/usr/bin/env bash
set -euo pipefail
[[ $# == 4 && "$1" == -p && "$3" == --permission-prompts && "$4" == none ]] || exit 90
if IFS= read -r unexpected; then exit 91; fi
[[ "${CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS:-}" == 0 ]] || exit 92
printf 'START:%s\n' "$2" >> "$TEST_ROOT/events"
printf '%s' "$PWD" > "$TEST_ROOT/cwd"
[[ "${TEST_CLAUDE_EXIT:-0}" == 0 ]] || exit "$TEST_CLAUDE_EXIT"
/bin/sleep 0.02
printf 'END:%s\n' "$2" >> "$TEST_ROOT/events"
`)
	write("git", `#!/usr/bin/env bash
set -euo pipefail
[[ "$*" == "status --porcelain" ]] || exit 90
[[ "${TEST_GIT_DIRTY:-0}" == 0 ]] || printf ' M main.go\n'
`)
	write("sleep", `#!/usr/bin/env bash
set -euo pipefail
printf 'SLEEP:%s\n' "$1" >> "$TEST_ROOT/events"
if [[ -f "$TEST_ROOT/slept" ]]; then exit 73; fi
touch "$TEST_ROOT/slept"
`)
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
		"PATH="+root+string(os.PathListSeparator)+os.Getenv("PATH"),
		"TEST_ROOT="+root, "TEST_TRIGGER_HELPER=1", "TEST_GH_EXIT=0", "TEST_CLAUDE_EXIT=0",
	)
	cmd.Env = append(cmd.Env, env...)
	output, runErr := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("trigger exceeded test timeout: %s", output)
	}
	read := func(name string) string {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(root, name))
		if os.IsNotExist(err) {
			return ""
		}
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	if cwd := read("cwd"); cwd != "" && cwd != project {
		t.Fatalf("unexpected working directory: %s", cwd)
	}
	return string(output), read("events"), read("gh-args"), runErr
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
		t.Fatalf("want sequential invocations %q, got %q", want, events)
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
	pass := "START:" + prompt + "\nEND:" + prompt + "\nSLEEP:7\n"
	if events != pass+pass || !strings.Contains(args, "--label\nmi-etiqueta\n") {
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
		snapshot("https://github.com/example/todo/issues/1", []string{"adw:waiting-for-human"}),
		snapshot("https://github.com/example/todo/issues/2", []string{"stage:planning"}),
		snapshot("https://github.com/example/todo/issues/3", nil, "otra-cuenta"),
		snapshot("https://github.com/example/todo/issues/4", nil, "agente", "otra-cuenta"),
		snapshot("https://github.com/example/todo/issues/5", []string{"feature", "plan:ready"}, "Agente"),
		snapshot("https://github.com/example/todo/issues/6", []string{"issue:failed"}),
	)
	output, events, _, err := runTrigger(t, issues, "", nil, "--once")
	mine := snapshot("https://github.com/example/todo/issues/5", []string{"feature", "plan:ready"}, "Agente")
	free := snapshot("https://github.com/example/todo/issues/6", []string{"issue:failed"})
	want := "START:/adw " + mine + "\nEND:/adw " + mine + "\nSTART:/adw " + free + "\nEND:/adw " + free + "\n"
	if err != nil || events != want {
		t.Fatalf("only free or own cases without waiting or stage labels must launch: err=%v, events=%q, output=%s", err, events, output)
	}
}

func TestTriggerStopsOnDirtyCheckout(t *testing.T) {
	output, events, _, err := runTrigger(t, list(newSnapshot), "", []string{"TEST_GIT_DIRTY=1"}, "--once")
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 2 || events != "" || !strings.Contains(output, "checkout") {
		t.Fatalf("a dirty checkout must stop before Claude: err=%v, events=%q, output=%s", err, events, output)
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

func TestTriggerPollsWithoutRememberingIssues(t *testing.T) {
	_, events, _, err := runTrigger(t, list(newSnapshot), "", nil)
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 73 {
		t.Fatalf("unexpected exit: %v", err)
	}
	pass := "START:/adw " + newSnapshot + "\nEND:/adw " + newSnapshot + "\nSLEEP:20\n"
	if events != pass+pass {
		t.Fatalf("this minimal trigger deliberately has no memory: %q", events)
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

func TestTriggerMaxRejections(t *testing.T) {
	cases := []struct {
		name string
		body string
		want int
	}{
		{name: "default", body: defaultConfig, want: 3},
		{name: "loop-disabled", body: defaultConfig + "max_rejections: 1\n", want: 1},
		{name: "configured", body: defaultConfig + "max_rejections: 5\n", want: 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "adw.yml")
			if err := os.WriteFile(path, []byte(tc.body), 0o600); err != nil {
				t.Fatal(err)
			}
			cfg, err := readConfig(path)
			if err != nil {
				t.Fatalf("max_rejections must accept integers from 1 and default to 3: %v", err)
			}
			if cfg.MaxRejections != tc.want {
				t.Fatalf("want max_rejections %d, got %d", tc.want, cfg.MaxRejections)
			}
		})
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
