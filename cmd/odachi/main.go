// Command odachi is the whole product: one static binary, no daemon, no
// network. `odachi` with no arguments is `odachi start`.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/biggs-100/odachi/internal/app/ask"
	"github.com/biggs-100/odachi/internal/app/check"
	initapp "github.com/biggs-100/odachi/internal/app/init"
	"github.com/biggs-100/odachi/internal/app/level"
	"github.com/biggs-100/odachi/internal/app/start"
	"github.com/biggs-100/odachi/internal/app/summary"
)

// version is the release identifier this binary reports. A plain `go build`
// leaves it as "dev"; goreleaser injects it with -ldflags -X main.version=<tag>,
// so the number a human reads always comes from the release, never from a hand
// edit of a source file.
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run runs the CLI in the current directory. It returns the exit code instead
// of calling os.Exit so the behaviour is testable.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// The version ask comes before anything that can fail: a deleted working
	// directory must not stop a human from asking what they installed.
	if isVersionAsk(args) {
		fmt.Fprintln(stdout, "odachi "+version)
		return start.CodeOK
	}

	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stderr, "odachi: no pude leer el directorio actual: "+oneLine(err.Error()))
		return start.CodeBlocked
	}
	return runIn(dir, args, stdin, stdout, stderr)
}

// runIn dispatches one command. `odachi` and `odachi --flag ...` are both
// `odachi start`; an unknown command lists what exists.
func runIn(dir string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// The version ask comes before the dispatch: a leading dash is otherwise a
	// `odachi start` flag, and `--version` is the one a human types first.
	if isVersionAsk(args) {
		fmt.Fprintln(stdout, "odachi "+version)
		return start.CodeOK
	}

	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return runStart(dir, args, stdin, stdout, stderr)
	}

	switch args[0] {
	case "start":
		return runStart(dir, args[1:], stdin, stdout, stderr)
	case "ask":
		return runAsk(dir, args[1:], stdout, stderr)
	case "level":
		return runLevel(dir, args[1:], stdin, stdout, stderr)
	case "summary":
		return runSummary(dir, args[1:], stdout, stderr)
	case "check":
		return runCheck(dir, args[1:], stdout, stderr)
	case "init":
		return runInit(dir, args[1:], stdout, stderr)
	case "version":
		// Reached only with arguments: `odachi version` alone was answered above.
		fmt.Fprintf(stderr, "odachi: sobra un argumento: %s; odachi version no toma argumentos\n", args[1])
		return start.CodeBlocked
	default:
		fmt.Fprintf(stderr, "odachi: comando desconocido: %s; comandos disponibles: start (odachi start), ask (odachi ask \"<pregunta>\"), level (odachi level), summary (odachi summary), check (odachi check), init (odachi init), version (odachi --version)\n", args[0])
		return start.CodeBlocked
	}
}

// isVersionAsk reports whether this is exactly the version request. `--version`
// and `version` are the only two surfaces; anything longer is a mistake the
// normal error path explains.
func isVersionAsk(args []string) bool {
	if len(args) != 1 {
		return false
	}
	return args[0] == "--version" || args[0] == "version"
}

// runStart parses the flags, wires the two adapters and reports the result.
func runStart(dir string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi start", flag.ContinueOnError)
	// Flag errors are translated into one Spanish line below, so the standard
	// library's own usage dump is discarded.
	flags.SetOutput(io.Discard)

	var (
		opts    start.Options
		done    stringList
		exclude stringList
		learn   stringList
		brief   string
		force   bool
		asJSON  bool
	)

	flags.StringVar(&opts.Outcome, "outcome", "", "respuesta 1: qué tiene que existir cuando terminemos")
	flags.StringVar(&opts.Audience, "audience", "", "respuesta 2: phone | cpu | both")
	flags.Var(&done, "done", "respuesta 3: cómo vas a saber que funcionó (repetible)")
	flags.Var(&exclude, "exclude", "respuesta 4: no-goal por clave del catálogo o texto libre (repetible)")
	flags.StringVar(&opts.Sensitive, "sensitive", "", "respuesta 5: none | names | health | money")
	flags.StringVar(&opts.Deadline, "deadline", "", "respuesta 6: today | week | relaxed")
	flags.Var(&learn, "learn", "respuesta 7: qué querés aprender (repetible)")
	flags.StringVar(&brief, "brief", "", "leer un brief existente en vez de preguntar")
	flags.BoolVar(&force, "force", false, "reescribir un brief.md existente")
	flags.BoolVar(&asJSON, "json", false, "salida para máquinas")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: flag inválido: "+oneLine(err.Error())+"; corré odachi start -h para ver los flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: sobra un argumento: %s; odachi start solo toma flags (corré odachi start -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}

	opts.Dir = dir
	opts.Done = done
	opts.Exclude = exclude
	opts.Learn = learn
	opts.BriefPath = brief
	opts.Force = force
	// No terminal (or machine output) means no human to ask: the run uses the
	// flags or exits 2 naming them.
	opts.Machine = asJSON || !isTerminal(stdin)

	loaded, code, message := loadProfile(dir, "odachi start")
	if code != start.CodeOK {
		return report(start.Result{Code: code, Message: message}, asJSON, stdout, stderr)
	}
	opts.Profile = loaded

	var interview start.Interview
	if !opts.Machine {
		interview = &terminalInterview{in: bufio.NewReader(stdin), out: stdout}
	}

	return report(start.Run(opts, interview, osFiles{}), asJSON, stdout, stderr)
}

// runAsk parses the flags, wires the Files adapter and reports the result.
func runAsk(dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi ask", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var (
		opts   ask.Options
		asJSON bool
	)

	flags.StringVar(&opts.Answer, "answer", "", "respuesta del humano a una pregunta abierta")
	flags.StringVar(&opts.For, "for", "", "P<n> de la pregunta que se contesta (obligatorio si hay dos o más abiertas)")
	flags.StringVar(&opts.Craft, "craft", "", "decisión que el agente tomó sin preguntar")
	flags.StringVar(&opts.Because, "because", "", "por qué el agente tomó esa decisión (obligatorio con --craft)")
	flags.BoolVar(&asJSON, "json", false, "salida para máquinas")

	// The documented shape starts with the question, so a leading argument is
	// the question and the flags may come after it.
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		opts.Question = args[0]
		args = args[1:]
	}

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: flag inválido: "+oneLine(err.Error())+"; corré odachi ask -h para ver los flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: sobra un argumento: %s; odachi ask toma una sola pregunta entre comillas\n", flags.Arg(0))
		return start.CodeBlocked
	}

	opts.Dir = dir
	return reportAsk(ask.Run(opts, osFiles{}), asJSON, stdout, stderr)
}

// runLevel parses the flags, reads the profile in force and reports the result.
func runLevel(dir string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi level", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var (
		to     string
		asJSON bool
	)

	flags.StringVar(&to, "to", "", "perfil a aplicar: se respalda .odachi/profile.toml y se reescribe")
	flags.BoolVar(&asJSON, "json", false, "salida para máquinas")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: flag inválido: "+oneLine(err.Error())+"; corré odachi level -h para ver los flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: sobra un argumento: %s; odachi level solo toma flags (corré odachi level -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}

	current, code, message := currentProfile(dir, "odachi level", asJSON || !isTerminal(stdin))
	if code != start.CodeOK {
		return reportLevel(level.Result{Code: code, Message: message}, asJSON, stdout, stderr)
	}

	return reportLevel(level.Run(level.Options{Dir: dir, Current: *current, To: to}, osFiles{}), asJSON, stdout, stderr)
}

// runSummary parses the flags and prints the derived report.
func runSummary(dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi summary", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var asJSON bool
	flags.BoolVar(&asJSON, "json", false, "salida para máquinas")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: flag inválido: "+oneLine(err.Error())+"; corré odachi summary -h para ver los flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: sobra un argumento: %s; odachi summary no toma argumentos (corré odachi summary -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}

	return reportSummary(summary.Run(summary.Options{Dir: dir}, osFiles{}), asJSON, stdout, stderr)
}

// runCheck parses the flags, wires the three adapters and reports the gate.
func runCheck(dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi check", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var (
		opts   check.Options
		asJSON bool
	)

	flags.StringVar(&opts.Run, "run", "", "corre el comando, captura la salida y escribe proof/<NNNN>-<slug>.txt")
	flags.StringVar(&opts.Proof, "proof", "", "usa una prueba ya escrita (agentes y CI)")
	flags.StringVar(&opts.Base, "base", "HEAD", "contra qué comparar el cerco (vacío: contra nada, todo cuenta como nuevo)")
	flags.IntVar(&opts.TimeoutSeconds, "timeout", check.DefaultTimeout, "segundos que puede tardar --run")
	flags.BoolVar(&opts.Quiet, "quiet", false, "no imprime nada si está listo (para el hook)")
	flags.BoolVar(&asJSON, "json", false, "salida para máquinas")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: flag inválido: "+oneLine(err.Error())+"; corré odachi check -h para ver los flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: sobra un argumento: %s; odachi check solo toma flags (corré odachi check -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}
	if opts.TimeoutSeconds <= 0 {
		fmt.Fprintln(stderr, "odachi: --timeout tiene que ser mayor que cero segundos; corré odachi check -h para ver los flags")
		return start.CodeBlocked
	}
	if opts.Run != "" && opts.Proof != "" {
		fmt.Fprintln(stderr, "odachi: --run y --proof no van juntos: usá uno de los dos (odachi check --run \"<comando>\" o odachi check --proof <archivo>)")
		return start.CodeBlocked
	}

	opts.Dir = dir
	return reportCheck(check.Run(opts, osFiles{}, gitRepo{}, osShell{}), opts.Quiet, asJSON, stdout, stderr)
}

// reportCheck prints the gate's verdict. A ready run is silent under --quiet,
// because the pre-commit hook runs it on every commit; a blocked run always
// says what blocks and which command unblocks it.
func reportCheck(result check.Result, quiet, asJSON bool, stdout, stderr io.Writer) int {
	if asJSON {
		files := result.Files
		if files == nil {
			files = []string{}
		}
		payload := struct {
			Code    int      `json:"code"`
			Reason  string   `json:"reason"`
			Message string   `json:"message"`
			Next    string   `json:"next,omitempty"`
			Lines   []string `json:"lines,omitempty"`
			Proof   string   `json:"proof,omitempty"`
			Files   []string `json:"files"`
		}{
			Code:    result.Code,
			Reason:  result.Reason,
			Message: result.Message,
			Next:    result.Next,
			Lines:   result.Lines,
			Proof:   result.Proof,
			Files:   files,
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, "odachi: no pude armar la salida JSON: "+oneLine(err.Error()))
			return start.CodeBlocked
		}
		fmt.Fprintln(stdout, string(data))
		return result.Code
	}

	if result.Code == start.CodeOK && quiet {
		return result.Code
	}

	lines := result.Lines
	if len(lines) == 0 {
		lines = []string{result.Message}
	}

	out := stdout
	if result.Code != start.CodeOK {
		out = stderr
	}
	for i, line := range lines {
		if i == 0 && result.Code != start.CodeOK {
			line = "odachi: " + line
		}
		fmt.Fprintln(out, line)
	}
	return result.Code
}

// runInit parses the flags and wires the Files adapter.
func runInit(dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi init", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var asJSON bool
	flags.BoolVar(&asJSON, "json", false, "salida para máquinas")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: flag inválido: "+oneLine(err.Error())+"; corré odachi init -h para ver los flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: sobra un argumento: %s; odachi init solo toma flags (corré odachi init -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}

	return reportInit(initapp.Run(initapp.Options{Dir: dir, Exe: hookExe()}, osFiles{}), asJSON, stdout, stderr)
}

// reportInit prints what init did. Unlike the other verbs it always prints its
// block on a block, because the block contains the three lines the human has to
// paste into a hook odachi did not write.
func reportInit(result initapp.Result, asJSON bool, stdout, stderr io.Writer) int {
	if asJSON {
		files := result.Files
		if files == nil {
			files = []string{}
		}
		payload := struct {
			Code    int      `json:"code"`
			Message string   `json:"message"`
			Next    string   `json:"next,omitempty"`
			Lines   []string `json:"lines,omitempty"`
			Files   []string `json:"files"`
		}{
			Code:    result.Code,
			Message: result.Message,
			Next:    result.Next,
			Lines:   result.Lines,
			Files:   files,
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, "odachi: no pude armar la salida JSON: "+oneLine(err.Error()))
			return start.CodeBlocked
		}
		fmt.Fprintln(stdout, string(data))
		return result.Code
	}

	lines := result.Lines
	if len(lines) == 0 {
		lines = []string{result.Message}
	}

	out := stdout
	if result.Code != start.CodeOK {
		out = stderr
	}
	for i, line := range lines {
		if i == 0 && result.Code != start.CodeOK {
			line = "odachi: " + line
		}
		fmt.Fprintln(out, line)
	}
	return result.Code
}

// outcome is what every verb hands the printer: one line for a human, or JSON
// for a machine. start, ask, level and summary all return this shape.
type outcome struct {
	Code    int
	Message string
	Lines   []string
	Profile string
	Task    string
	Files   []string
}

// report prints the result of `odachi start`.
func report(result start.Result, asJSON bool, stdout, stderr io.Writer) int {
	return reportOutcome(outcome{
		Code:    result.Code,
		Message: result.Message,
		Profile: result.Profile,
		Task:    result.Task,
		Files:   result.Files,
	}, asJSON, stdout, stderr)
}

// reportAsk prints the result of `odachi ask`.
func reportAsk(result ask.Result, asJSON bool, stdout, stderr io.Writer) int {
	return reportOutcome(outcome{
		Code:    result.Code,
		Message: result.Message,
		Lines:   result.Lines,
		Task:    result.Task,
		Files:   result.Files,
	}, asJSON, stdout, stderr)
}

// reportLevel prints the result of `odachi level`.
func reportLevel(result level.Result, asJSON bool, stdout, stderr io.Writer) int {
	return reportOutcome(outcome{
		Code:    result.Code,
		Message: result.Message,
		Lines:   result.Lines,
		Profile: result.Profile,
		Task:    result.Task,
		Files:   result.Files,
	}, asJSON, stdout, stderr)
}

// reportSummary prints the result of `odachi summary`.
func reportSummary(result summary.Result, asJSON bool, stdout, stderr io.Writer) int {
	return reportOutcome(outcome{
		Code:    result.Code,
		Message: result.Message,
		Lines:   result.Lines,
		Task:    result.Task,
		Files:   result.Files,
	}, asJSON, stdout, stderr)
}

// reportOutcome prints the result and returns the exit code. Machine output is
// JSON; human output is the block the verb built, on stderr when the run is
// blocked.
func reportOutcome(o outcome, asJSON bool, stdout, stderr io.Writer) int {
	if asJSON {
		files := o.Files
		if files == nil {
			files = []string{}
		}
		payload := struct {
			Code    int      `json:"code"`
			Message string   `json:"message"`
			Lines   []string `json:"lines,omitempty"`
			Profile string   `json:"profile,omitempty"`
			Task    string   `json:"task,omitempty"`
			Files   []string `json:"files"`
		}{
			Code:    o.Code,
			Message: o.Message,
			Lines:   o.Lines,
			Profile: o.Profile,
			Task:    o.Task,
			Files:   files,
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, "odachi: no pude armar la salida JSON: "+oneLine(err.Error()))
			return start.CodeBlocked
		}
		fmt.Fprintln(stdout, string(data))
		return o.Code
	}

	if o.Code == start.CodeOK {
		for _, line := range humanLines(o) {
			fmt.Fprintln(stdout, line)
		}
		return o.Code
	}
	fmt.Fprintln(stderr, "odachi: "+o.Message)
	return o.Code
}

// humanLines is the block a human reads: whatever the verb listed, or the one
// line it reported when the verb has no block.
func humanLines(o outcome) []string {
	if len(o.Lines) > 0 {
		return o.Lines
	}
	return []string{o.Message}
}

// usage is the Spanish help a human reads with `odachi start -h`.
func usage(out io.Writer) {
	fmt.Fprint(out, `odachi · dos manos: vos decidís, el agente construye

uso:
  odachi                      lo mismo que odachi start
  odachi start [flags]
  odachi ask "<pregunta>"
  odachi ask --answer "<respuesta>" [--for P1]
  odachi ask --craft "<decisión>" --because "<razón>"
  odachi level [--to <perfil>]
  odachi summary
  odachi check [flags]
  odachi init [--json]
  odachi --version
  odachi version

la versión:
  --version       imprime "odachi <versión>" y sale con 0 (un build sin ldflags dice dev)

flags de check:
  --run s         corre el comando y escribe la prueba de este árbol
  --proof path    usa una prueba ya escrita (agentes y CI)
  --base ref      contra qué comparar el cerco (default HEAD; vacío: contra nada)
  --timeout n     segundos que puede tardar --run (default 120)
  --quiet         no imprime nada si está listo (para el hook)
  --json          salida para máquinas

flags:
  --outcome s     respuesta 1: qué tiene que existir cuando terminemos
  --audience s    respuesta 2: phone | cpu | both
  --done s        respuesta 3 (repetible): cómo vas a saber que funcionó
  --exclude s     respuesta 4 (repetible): no-goal por clave o texto libre
  --sensitive s   respuesta 5: none | names | health | money
  --deadline s    respuesta 6: today | week | relaxed
  --learn s       respuesta 7 (repetible): qué querés aprender
  --brief path    leer un brief existente en vez de preguntar
  --force         reescribir un brief.md existente (se respalda primero)
  --json          salida para máquinas

después de start:
  odachi ask "<pregunta>"                       anotar una pregunta abierta
  odachi ask --answer "<respuesta>" [--for P1]  anotar tu respuesta
  odachi ask --craft "<decisión>" --because "<razón>"  anotar una decisión del agente
  odachi level                                  ver tus hechos; propone senior cuando se gana
  odachi level --to <perfil>                    aplicar un perfil (se respalda el anterior)
  odachi summary                                qué hay hecho en cada feature
  odachi check                                  ¿está listo? brief, cerco, preguntas y prueba
  odachi check --run "<comando>"                corre la prueba y la guarda con su árbol
  odachi init                                   AGENTS.md + hook pre-commit

códigos de salida: 0 artefactos escritos · 1 bloqueado · 2 falta una respuesta humana
`)
}

// isTerminal reports whether the input is a terminal. Anything else (a pipe, a
// file, a test) means no human is there to answer.
func isTerminal(in io.Reader) bool {
	file, ok := in.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// stringList is a repeatable string flag: --done a --done b.
type stringList []string

func (l *stringList) String() string {
	return strings.Join(*l, ", ")
}

func (l *stringList) Set(value string) error {
	*l = append(*l, value)
	return nil
}

// oneLine keeps a CLI message on one line whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
