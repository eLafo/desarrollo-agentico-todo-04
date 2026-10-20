package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"go.yaml.in/yaml/v3"
)

// Etiquetas de control de la skill adw: ver .claude/skills/adw/references/adapters/github.md.
const (
	waitingLabel = "adw:waiting-for-human"
	stagePrefix  = "stage:"
)

// defaultMaxRejections es el tope de rechazos seguidos de la verificación si config/adw.yml no lo fija.
const defaultMaxRejections = 3

// defaultMaxParallel son las sesiones a la vez si config/adw.yml no lo fija: de una en una.
const defaultMaxParallel = 1

// config refleja config/adw.yml. MaxRejections no lo usa el trigger, sino la skill, que lo lee
// del mismo fichero; está aquí para que KnownFields(true) lo admita y para validarlo al arrancar.
type config struct {
	Repo            string `yaml:"repo"`
	Label           string `yaml:"label"`
	IntervalSeconds int    `yaml:"interval_seconds"`
	MaxRejections   int    `yaml:"max_rejections"`
	MaxParallel     int    `yaml:"max_parallel"`
	Prompt          string `yaml:"prompt"`
}

type issue struct {
	URL    string `json:"url"`
	Number int    `json:"number"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
	Assignees []struct {
		Login string `json:"login"`
	} `json:"assignees"`
}

func readConfig(path string) (config, error) {
	// El decodificador solo cambia los campos presentes: si se omite, queda el valor por defecto.
	cfg := config{MaxRejections: defaultMaxRejections, MaxParallel: defaultMaxParallel}
	file, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer file.Close()
	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("configuración YAML: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return cfg, fmt.Errorf("la configuración debe contener un solo documento YAML")
	}
	if strings.TrimSpace(cfg.Label) == "" || cfg.IntervalSeconds < 1 {
		return cfg, fmt.Errorf("label no puede estar vacío e interval_seconds debe ser positivo")
	}
	if cfg.MaxRejections < 1 {
		return cfg, fmt.Errorf("max_rejections debe ser mayor o igual que 1")
	}
	if cfg.MaxParallel < 1 {
		return cfg, fmt.Errorf("max_parallel debe ser mayor o igual que 1")
	}
	if !strings.Contains(cfg.Prompt, "{{payload}}") {
		return cfg, fmt.Errorf("prompt debe contener {{payload}}")
	}
	return cfg, nil
}

// agentTurn dice si el turno del caso es del agente: nadie espera al humano,
// no hay un stage abierto y la issue está libre o asignada a esta cuenta.
func agentTurn(it issue, me string) bool {
	for _, label := range it.Labels {
		if label.Name == waitingLabel || strings.HasPrefix(label.Name, stagePrefix) {
			return false
		}
	}
	for _, assignee := range it.Assignees {
		if !strings.EqualFold(assignee.Login, me) {
			return false
		}
	}
	return true
}

var envLine = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)=(.*)$`)

// readEnv lee el .env que escribe bin/tree: una variable por línea, sin comillas.
func readEnv(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var env []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !envLine.MatchString(line) {
			return nil, fmt.Errorf("%s: línea no válida: %q", path, line)
		}
		env = append(env, line)
	}
	return env, nil
}

// prefixWriter antepone una etiqueta a cada línea, para distinguir la salida de cada sesión.
// Las líneas completas se escriben de una vez, con un candado compartido entre sesiones.
type prefixWriter struct {
	lock   *sync.Mutex
	out    io.Writer
	prefix string
	buf    []byte
}

func (w *prefixWriter) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			return len(p), nil
		}
		w.lock.Lock()
		fmt.Fprintf(w.out, "%s%s\n", w.prefix, w.buf[:i])
		w.lock.Unlock()
		w.buf = w.buf[i+1:]
	}
}

func (w *prefixWriter) flush() {
	if len(w.buf) > 0 {
		w.Write([]byte("\n"))
	}
}

// trigger lanza las sesiones, cada una en el worktree de su caso, y recuerda las que están en
// marcha hasta que terminan: entre el lanzamiento y la etiqueta stage:* que pone la skill pasa un
// rato, y sin esto la consulta siguiente lanzaría el caso otra vez.
type trigger struct {
	cfg  config
	tree string // bin/tree del checkout
	top  string // raíz del checkout

	out      sync.Mutex // la salida del trigger y de las sesiones, línea a línea
	mu       sync.Mutex // running y failure
	running  map[int]bool
	failure  error
	sessions sync.WaitGroup
	slots    chan struct{}
	warned   map[int]bool // worktrees que no se pudieron borrar, avisados una vez
}

func (t *trigger) say(number int, format string, args ...any) {
	t.out.Lock()
	defer t.out.Unlock()
	if number > 0 {
		fmt.Printf("[#%d] ", number)
	}
	fmt.Printf(format+"\n", args...)
}

func (t *trigger) failed() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.failure
}

func (t *trigger) isRunning(number int) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.running[number]
}

func (t *trigger) repoArgs() []string {
	if repo := strings.TrimSpace(t.cfg.Repo); repo != "" {
		return []string{"--repo", repo}
	}
	return nil
}

// prune borra los worktrees de las issues cerradas que no tienen una sesión en marcha.
func (t *trigger) prune(open map[int]bool) {
	entries, err := os.ReadDir(filepath.Join(t.top, "trees"))
	if err != nil {
		return
	}
	for _, entry := range entries {
		number, err := strconv.Atoi(strings.TrimPrefix(entry.Name(), "issue-"))
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "issue-") || err != nil || open[number] || t.isRunning(number) {
			continue
		}
		args := append([]string{"issue", "view", strconv.Itoa(number)}, t.repoArgs()...)
		state, err := exec.Command("gh", append(args, "--json", "state", "--jq", ".state")...).Output()
		if err != nil {
			t.say(number, "no se sabe si la issue está cerrada: %v", err)
			continue
		}
		if strings.TrimSpace(string(state)) != "CLOSED" {
			continue
		}
		output, err := exec.Command(t.tree, "remove", strconv.Itoa(number)).CombinedOutput()
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if err != nil {
			if !t.warned[number] {
				t.say(number, "issue cerrada, pero no se borra su worktree: %s", lines[len(lines)-1])
				t.warned[number] = true
			}
			continue
		}
		t.say(number, "issue cerrada: %s", lines[len(lines)-1])
	}
}

// launch prepara el worktree del caso y lanza su sesión sin esperar a que termine. Dice si la ha
// lanzado. Un error al preparar el worktree es fatal; un worktree con cambios sin registrar solo
// salta el caso.
func (t *trigger) launch(it issue, snapshot json.RawMessage) (bool, error) {
	create := exec.Command(t.tree, "create", strconv.Itoa(it.Number))
	create.Stderr = os.Stderr
	output, err := create.Output()
	if err != nil {
		return false, fmt.Errorf("worktree de %s: %w", it.URL, err)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	dir := lines[len(lines)-1]
	status, err := exec.Command("git", "-C", dir, "status", "--porcelain").Output()
	if err != nil {
		return false, fmt.Errorf("estado de %s: %w", dir, err)
	}
	if len(bytes.TrimSpace(status)) > 0 {
		t.say(it.Number, "%s tiene cambios sin registrar: déjalo limpio para que el trigger lance el caso", dir)
		return false, nil
	}
	env, err := readEnv(filepath.Join(dir, ".env"))
	if err != nil {
		return false, err
	}
	var payload bytes.Buffer
	if err := json.Compact(&payload, snapshot); err != nil {
		return false, err
	}
	prompt := strings.ReplaceAll(t.cfg.Prompt, "{{payload}}", payload.String())

	agent := exec.Command("claude", "-p", prompt, "--permission-prompts", "none")
	agent.Dir = dir
	// Las variables del worktree, como si se hubiera cargado su .env. Sin tope de espera a los
	// subagentes en segundo plano: si el coordinador termina su turno mientras uno trabaja,
	// claude -p corta la sesión a los 10 minutos y deja el caso a medias.
	agent.Env = append(append(os.Environ(), env...), "PWD="+dir, "CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS=0")
	prefix := fmt.Sprintf("[#%d] ", it.Number)
	stdout := &prefixWriter{lock: &t.out, out: os.Stdout, prefix: prefix}
	stderr := &prefixWriter{lock: &t.out, out: os.Stderr, prefix: prefix}
	agent.Stdout, agent.Stderr = stdout, stderr
	// Stdin queda cerrado: no hay conversación ni evaluación de shell.

	t.mu.Lock()
	t.running[it.Number] = true
	t.mu.Unlock()
	t.say(it.Number, "Lanzando %s en %s", it.URL, dir)
	t.sessions.Add(1)
	go func() {
		defer t.sessions.Done()
		err := agent.Run()
		stdout.flush()
		stderr.flush()
		t.mu.Lock()
		delete(t.running, it.Number)
		if err != nil && t.failure == nil {
			t.failure = fmt.Errorf("Claude para %s: %w", it.URL, err)
		}
		t.mu.Unlock()
		<-t.slots
		if err != nil {
			t.say(it.Number, "La sesión falla: %v. No se lanzan más casos.", err)
		}
	}()
	return true, nil
}

// poll consulta las issues, borra los worktrees de las cerradas y lanza los casos en turno del
// agente que no tienen una sesión en marcha. Con wait, espera a que haya sitio para cada caso;
// sin él, lanza los que quepan y deja el resto para la consulta siguiente.
func (t *trigger) poll(me string, wait bool) error {
	queryArgs := append([]string{"issue", "list"}, t.repoArgs()...)
	queryArgs = append(queryArgs, "--state", "open", "--label", t.cfg.Label,
		"--search", "sort:created-desc", "--limit", "100", "--json", "url,number,title,labels,assignees,updatedAt")
	query := exec.Command("gh", queryArgs...)
	query.Stderr = os.Stderr
	output, err := query.Output()
	if err != nil {
		return fmt.Errorf("consulta de issues: %w", err)
	}
	var snapshots []json.RawMessage
	if len(bytes.TrimSpace(output)) > 0 {
		if err := json.Unmarshal(output, &snapshots); err != nil {
			return fmt.Errorf("respuesta de la consulta de issues: %w", err)
		}
	}
	issues := make([]issue, len(snapshots))
	open := map[int]bool{}
	for i, snapshot := range snapshots {
		if err := json.Unmarshal(snapshot, &issues[i]); err != nil {
			return fmt.Errorf("issue de la consulta: %w", err)
		}
		if issues[i].Number < 1 {
			return fmt.Errorf("issue sin número en la consulta: %s", issues[i].URL)
		}
		open[issues[i].Number] = true
	}
	t.prune(open)
	for i, it := range issues {
		if !agentTurn(it, me) || t.isRunning(it.Number) {
			continue
		}
		if wait {
			t.slots <- struct{}{}
		} else {
			select {
			case t.slots <- struct{}{}:
			default:
				return nil
			}
		}
		if t.failed() != nil {
			<-t.slots
			return nil
		}
		started, err := t.launch(it, snapshots[i])
		if !started {
			<-t.slots
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func run(args []string) error {
	flags := flag.NewFlagSet("trigger-github", flag.ContinueOnError)
	path := flags.String("config", "config/adw.yml", "fichero YAML de configuración")
	once := flags.Bool("once", false, "procesa una sola consulta, espera a sus sesiones y termina")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("configura el prompt y las opciones en %s", *path)
	}
	cfg, err := readConfig(*path)
	if err != nil {
		return err
	}
	for _, tool := range []string{"gh", "claude", "git"} {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("no se encuentra %s en PATH", tool)
		}
	}
	top, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return fmt.Errorf("raíz del checkout: %w", err)
	}
	t := &trigger{
		cfg:     cfg,
		top:     strings.TrimSpace(string(top)),
		running: map[int]bool{},
		slots:   make(chan struct{}, cfg.MaxParallel),
		warned:  map[int]bool{},
	}
	t.tree = filepath.Join(t.top, "bin", "tree")
	login := exec.Command("gh", "api", "user", "--jq", ".login")
	login.Stderr = os.Stderr
	output, err := login.Output()
	if err != nil {
		return fmt.Errorf("cuenta de gh: %w", err)
	}
	me := strings.TrimSpace(string(output))

	// Al salir, por fin o por error, se espera a las sesiones en marcha: nadie más recogería su
	// resultado.
	finish := func(err error) error {
		t.sessions.Wait()
		if err != nil {
			return err
		}
		return t.failed()
	}
	for {
		if err := t.poll(me, *once); err != nil || *once || t.failed() != nil {
			return finish(err)
		}
		if err := exec.Command("sleep", strconv.Itoa(cfg.IntervalSeconds)).Run(); err != nil {
			return finish(err)
		}
	}
}

func exitCode(err error) int {
	var process *exec.ExitError
	if errors.As(err, &process) && process.ExitCode() > 0 {
		return process.ExitCode()
	}
	return 2
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitCode(err))
	}
}
