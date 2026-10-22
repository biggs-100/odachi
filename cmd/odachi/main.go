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
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/app/ask"
	"github.com/biggs-100/odachi/internal/app/check"
	initapp "github.com/biggs-100/odachi/internal/app/init"
	"github.com/biggs-100/odachi/internal/app/level"
	"github.com/biggs-100/odachi/internal/app/pack"
	"github.com/biggs-100/odachi/internal/app/plan"
	reviewapp "github.com/biggs-100/odachi/internal/app/review"
	"github.com/biggs-100/odachi/internal/app/start"
	"github.com/biggs-100/odachi/internal/app/summary"
	"github.com/biggs-100/odachi/internal/app/undo"
	"github.com/biggs-100/odachi/internal/domain/agent"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
	bundledskills "github.com/biggs-100/odachi/skills"
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
	case "plan":
		return runPlan(dir, args[1:], stdout, stderr)
	case "level":
		return runLevel(dir, args[1:], stdin, stdout, stderr)
	case "summary":
		return runSummary(dir, args[1:], stdout, stderr)
	case "undo":
		return runUndo(dir, args[1:], stdout, stderr)
	case "check":
		return runCheck(dir, args[1:], stdout, stderr)
	case "init":
		return runInit(dir, args[1:], stdout, stderr)
	case "skills":
		return runSkills(dir, args[1:], stdout, stderr)
	case "skill":
		return runSkill(dir, args[1:], stdout, stderr)
	case "pack":
		return runPack(dir, args[1:], stdout, stderr)
	case "review":
		return runReview(dir, args[1:], stdout, stderr)
	case "doctor":
		return runDoctor(dir, args[1:], stdout, stderr)
	case "bench":
		return runBench(dir, args[1:], stdout, stderr)
	case "version":
		// Reached only with arguments: `odachi version` alone was answered above.
		fmt.Fprintf(stderr, "odachi: sobra un argumento: %s; odachi version no toma argumentos\n", args[1])
		return start.CodeBlocked
	default:
		fmt.Fprintf(stderr, "odachi: comando desconocido: %s; comandos disponibles: start (odachi start), ask (odachi ask \"<pregunta>\"), plan (odachi plan \"<texto>\" | odachi plan --done T2), level (odachi level), summary (odachi summary), undo (odachi undo), check (odachi check), review (odachi review), doctor (odachi doctor), bench (odachi bench), init (odachi init), skills (odachi skills), skill (odachi skill new <slug>), pack (odachi pack), version (odachi --version)\n", args[0])
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

	return report(start.Run(opts, interview, osFiles{}, &osLock{}), asJSON, stdout, stderr)
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

// runPlan parses the flags, wires the Files adapter and reports the result.
func runPlan(dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi plan", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var (
		opts   plan.Options
		asJSON bool
	)

	flags.StringVar(&opts.Requirement, "req", "", "R<n> del requisito al que se ata el paso nuevo")
	flags.StringVar(&opts.Done, "done", "", "T<n> del paso que se marca hecho")
	flags.StringVar(&opts.Commit, "commit", "", "commit que cerró el paso (acompaña a --done)")
	flags.BoolVar(&asJSON, "json", false, "salida para máquinas")

	// The documented shape starts with the step text, so a leading argument is
	// the text and the flags may come after it.
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		opts.Text = args[0]
		args = args[1:]
	}

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: flag inválido: "+oneLine(err.Error())+"; corré odachi plan -h para ver los flags")
		return start.CodeBlocked
	}

	// The text may also come after the flags: `plan --req R1 "texto"`.
	rest := flags.Args()
	if opts.Text == "" && len(rest) == 1 {
		opts.Text = rest[0]
		rest = nil
	}
	if len(rest) > 0 {
		fmt.Fprintf(stderr, "odachi: sobra un argumento: %s; odachi plan toma un solo texto entre comillas\n", rest[0])
		return start.CodeBlocked
	}

	opts.Dir = dir
	return reportPlan(plan.Run(opts, osFiles{}), asJSON, stdout, stderr)
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
	flags.BoolVar(&opts.RequireReview, "require-review", false, "exige un veredicto de review para este árbol (apagado por defecto)")
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

// runReview parses the flags, wires the four adapters and reports the verdict.
func runReview(dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi review", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var (
		opts   reviewapp.Options
		asJSON bool
	)

	flags.StringVar(&opts.Exec, "exec", "", "corre este comando con el prompt por stdin y guarda el veredicto que contesta")
	flags.StringVar(&opts.Attach, "attach", "", "guarda un veredicto o una refutación ya escrita en este archivo")
	flags.BoolVar(&opts.Show, "show", false, "qué veredicto hay para este árbol, o que no hay")
	flags.StringVar(&opts.Lens, "lens", "", "lente explícita: risk | readability | reliability | resilience | all (repetible o separada por comas)")
	flags.StringVar(&opts.Refute, "refute", "", "imprime el prompt del refutador para este veredicto")
	flags.StringVar(&opts.Base, "base", reviewapp.DefaultBase, "contra qué comparar el candidato (vacío: contra nada, todo cuenta como nuevo)")
	flags.BoolVar(&asJSON, "json", false, "salida para máquinas")

	if err := flags.Parse(withoutEmptyLens(args)); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: flag inválido: "+oneLine(err.Error())+"; corré odachi review -h para ver los flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: sobra un argumento: %s; odachi review solo toma flags (corré odachi review -h)\\n", flags.Arg(0))
		return start.CodeBlocked
	}

	// The four forms answer different questions, and asking for two at once is a
	// mistake worth naming instead of resolving by precedence.
	forms := 0
	for _, used := range []bool{opts.Exec != "", opts.Attach != "", opts.Show, opts.Refute != ""} {
		if used {
			forms++
		}
	}
	if forms > 1 {
		fmt.Fprintln(stderr, `odachi: --exec, --attach, --show y --refute no van juntos: elegí una sola forma (odachi review | odachi review --exec "<comando>" | odachi review --attach <archivo> | odachi review --show | odachi review --refute <veredicto>)`)
		return start.CodeBlocked
	}

	opts.Dir = dir
	return reportReview(reviewapp.Run(opts, osFiles{}, gitRepo{}, osShell{}, reviewLenses{}), asJSON, stdout, stderr)
}

// withoutEmptyLens is what makes `odachi review --lens --show` work: the flag
// package would read `--show` as the value of `--lens`, and a lens with no value
// means "the lenses the evidence asks for". Only the bare flag is rewritten, so
// `--lens risk` is untouched.
func withoutEmptyLens(args []string) []string {
	normalized := make([]string, 0, len(args))
	for i, arg := range args {
		if arg == "-lens" || arg == "--lens" {
			if i+1 == len(args) || strings.HasPrefix(args[i+1], "-") {
				normalized = append(normalized, arg+"=")
				continue
			}
		}
		normalized = append(normalized, arg)
	}
	return normalized
}

// runUndo parses the flags, wires the Files adapter and reports the result.
func runUndo(dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi undo", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var (
		opts   undo.Options
		asJSON bool
	)

	flags.BoolVar(&opts.List, "list", false, "qué se puede deshacer, con fecha")
	flags.StringVar(&opts.File, "file", "", "solo ese artefacto: brief | profile | fence | task")
	flags.BoolVar(&asJSON, "json", false, "salida para máquinas")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: flag inválido: "+oneLine(err.Error())+"; corré odachi undo -h para ver los flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: sobra un argumento: %s; odachi undo solo toma flags (corré odachi undo -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}

	opts.Dir = dir
	return reportUndo(undo.Run(opts, osFiles{}, &osLock{}), asJSON, stdout, stderr)
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

// reportReview prints the review. The prompt is the deliverable of the bare form,
// so it is printed as one block exactly as the reviewer will read it; a failed
// review prints the command's own output, because that is the diagnosis.
func reportReview(result reviewapp.Result, asJSON bool, stdout, stderr io.Writer) int {
	if asJSON {
		files := result.Files
		if files == nil {
			files = []string{}
		}
		payload := struct {
			Code       int                   `json:"code"`
			Message    string                `json:"message"`
			Lines      []string              `json:"lines,omitempty"`
			Prompt     string                `json:"prompt,omitempty"`
			Output     string                `json:"output,omitempty"`
			Verdict    *domainreview.Verdict `json:"verdict,omitempty"`
			Lenses     []string              `json:"lenses,omitempty"`
			Refutation any                   `json:"refutation,omitempty"`
			Files      []string              `json:"files"`
		}{
			Code:       result.Code,
			Message:    result.Message,
			Lines:      result.Lines,
			Prompt:     result.Prompt,
			Output:     result.Output,
			Verdict:    result.Verdict,
			Lenses:     result.Lenses,
			Refutation: result.Refutation,
			Files:      files,
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, "odachi: no pude armar la salida JSON: "+oneLine(err.Error()))
			return start.CodeBlocked
		}
		fmt.Fprintln(stdout, string(data))
		return result.Code
	}

	if result.Prompt != "" {
		fmt.Fprint(stdout, result.Prompt)
		if !strings.HasSuffix(result.Prompt, "\n") {
			fmt.Fprintln(stdout)
		}
		return result.Code
	}

	if result.Code == start.CodeOK {
		for _, line := range result.Lines {
			fmt.Fprintln(stdout, line)
		}
		if len(result.Lines) == 0 {
			fmt.Fprintln(stdout, result.Message)
		}
		return result.Code
	}

	fmt.Fprintln(stderr, "odachi: "+result.Message)
	if len(result.Lines) > 1 {
		for _, line := range result.Lines[1:] {
			fmt.Fprintln(stderr, line)
		}
	}
	if result.Output != "" {
		fmt.Fprint(stderr, result.Output)
		if !strings.HasSuffix(result.Output, "\n") {
			fmt.Fprintln(stderr)
		}
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

	created, err := ensureHistory(dir)
	if err != nil {
		fmt.Fprintln(stderr, "odachi: "+err.Error())
		return start.CodeBlocked
	}

	result := initapp.Run(initapp.Options{Dir: dir, Exe: hookExe(), PathMissing: pathMissing()}, osFiles{}, &initSkills{dir: dir})
	if created {
		line := "historial del proyecto: lo creé en esta carpeta (git init); es lo que me deja probar que una prueba corresponde a este estado del código"
		result.Lines = append([]string{line}, result.Lines...)
	}
	if len(result.Files) > 0 || created {
		result.Files = append([]string{".git/"}, result.Files...)
	}

	return reportInit(result, asJSON, stdout, stderr)
}

// initSkills is the join between `odachi init` and `odachi pack`. init decides
// that two destinations get skills — the project's cross-client directory and the
// human's home, so an agent can learn how to start a project before one exists —
// and pack owns how a skill is installed: the destination table, the backup
// before a differing file, the identical bytes skipped.
//
// The join lives here, in the CLI, and not in internal/app/init importing
// internal/app/pack: the CLI is the only layer that already wires pack's payload
// and its two adapters, and an app that imports another app would inherit its
// error strings and its destination table for a job that is not its own.
type initSkills struct {
	dir string
}

// Install writes the two halves of the skills job and turns what pack reports
// into the two lines a human reads.
func (s initSkills) Install() initapp.SkillReport {
	curriculum, err := bundledskills.Curriculum()
	if err != nil {
		return initapp.SkillReport{Blocked: oneLine(err.Error()) + "; volvé a bajar odachi", Next: "odachi init"}
	}
	bootstrap, ok := bundledskills.Bootstrap()
	if !ok {
		return initapp.SkillReport{
			Blocked: "este binario no trae la skill " + bundledskills.BootstrapName + ": es una instalación incompleta; volvé a bajar odachi",
			Next:    "odachi init",
		}
	}

	// The project gets the craft, in the cross-client location every client
	// reads. The home gets the one skill that teaches how to arrive, because
	// before a project exists the home is the only place a skill can live.
	project := pack.Run(pack.Options{Agents: agent.Standard, Dir: s.dir, Skills: payload(curriculum)}, osFiles{}, osHome{})
	home := pack.Run(pack.Options{Agents: agent.Standard, Scope: string(agent.ScopeHome), Dir: s.dir, Skills: payload([]bundledskills.Skill{bootstrap})}, osFiles{}, osHome{})

	report := initapp.SkillReport{}
	for _, result := range []pack.Result{project, home} {
		for _, dest := range result.Dests {
			report.Dests = append(report.Dests, initapp.SkillDest{
				Label:     dest.Dest,
				Installed: dest.Installed,
				Unchanged: dest.Unchanged,
				Files:     dest.Files,
			})
		}
		if result.Code == pack.CodeOK || report.Blocked != "" {
			continue
		}
		report.Blocked = blockedSkills(result)
		report.Next = "odachi init"
		// A destination inside the project leaves the project unready; the home
		// is the only half whose failure does not fail the run.
		report.Fatal = destScope(result) == string(agent.ScopeProject)
	}

	report.Notice = s.otherProjectCopy()
	return report
}

// destScope is the scope pack resolved for a run: the project copy says so, the
// home copy says `home`, and a run with no destination falls back to the project.
func destScope(result pack.Result) string {
	for _, dest := range result.Dests {
		return dest.Scope
	}
	return string(agent.ScopeProject)
}

// otherProjectCopy acknowledges a second project location: somebody who ran
// `pack --agents opencode` before init has the curriculum in that client's native
// directory too. init writes one project location and never deletes, so it says
// the other copy is there instead of pretending it is not.
func (s initSkills) otherProjectCopy() []string {
	files := osFiles{}
	standard, err := agent.Resolve(agent.Standard)
	if err != nil {
		return nil
	}
	lines := []string{}
	for _, a := range agent.Supported() {
		if a.Scope != agent.ScopeProject {
			continue
		}
		exists, err := files.Exists(filepath.Join(s.dir, filepath.FromSlash(a.Dest)))
		if err != nil || !exists {
			continue
		}
		lines = append(lines, a.Display()+": ya tiene una copia de las skills (la puso odachi pack --agents "+a.Name+"); no la toco: la del proyecto es "+standard.Display())
	}
	return lines
}

// blockedSkills is the sentence for a destination that could not be written:
// the file, the reason the adapter gave, and the command that retries it.
func blockedSkills(result pack.Result) string {
	for _, dest := range result.Dests {
		if dest.Failure == nil {
			continue
		}
		return "no pude escribir " + dest.Failure.Path + ": " + dest.Failure.Reason + "; revisá los permisos y volvé a correr odachi init"
	}
	return oneLine(result.Message)
}

// payload converts the embedded skills into the shape pack installs.
func payload(skills []bundledskills.Skill) []pack.Skill {
	out := make([]pack.Skill, 0, len(skills))
	for _, s := range skills {
		out = append(out, pack.Skill{Name: s.Name, Content: s.Content})
	}
	return out
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
		skills := make([]initSkillDest, 0, len(result.SkillDests))
		for _, dest := range result.SkillDests {
			destFiles := dest.Files
			if destFiles == nil {
				destFiles = []string{}
			}
			skills = append(skills, initSkillDest{
				Dest:      dest.Label,
				Installed: dest.Installed,
				Unchanged: dest.Unchanged,
				Files:     destFiles,
			})
		}
		personas := make([]initPersonaDest, 0, len(result.PersonaDests))
		for _, dest := range result.PersonaDests {
			destFiles := dest.Files
			if destFiles == nil {
				destFiles = []string{}
			}
			personas = append(personas, initPersonaDest{
				Client:    dest.Client,
				Dir:       dest.Dir,
				Written:   dest.Written,
				Unchanged: dest.Kept,
				Files:     destFiles,
			})
		}
		payload := struct {
			Code     int               `json:"code"`
			Message  string            `json:"message"`
			Next     string            `json:"next,omitempty"`
			Lines    []string          `json:"lines,omitempty"`
			Files    []string          `json:"files"`
			Skills   []initSkillDest   `json:"skills"`
			Personas []initPersonaDest `json:"personas"`
		}{
			Code:     result.Code,
			Message:  result.Message,
			Next:     result.Next,
			Lines:    result.Lines,
			Files:    files,
			Skills:   skills,
			Personas: personas,
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

// initSkillDest is one row of the machine report: the same destination a human
// reads on the same run, with what it got.
type initSkillDest struct {
	Dest      string   `json:"dest"`
	Installed int      `json:"installed"`
	Unchanged int      `json:"unchanged"`
	Files     []string `json:"files"`
}

// initPersonaDest is one client's persona directory in the same report: which
// client, where it reads personas from, what this run did there, and the files it
// owns.
type initPersonaDest struct {
	Client    string   `json:"client"`
	Dir       string   `json:"dir"`
	Written   int      `json:"written"`
	Unchanged int      `json:"unchanged"`
	Files     []string `json:"files"`
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
		Lines:   result.Lines,
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

// reportPlan prints the result of `odachi plan`.
func reportPlan(result plan.Result, asJSON bool, stdout, stderr io.Writer) int {
	return reportOutcome(outcome{
		Code:    result.Code,
		Message: result.Message,
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

// reportUndo prints the result of `odachi undo`.
func reportUndo(result undo.Result, asJSON bool, stdout, stderr io.Writer) int {
	return reportOutcome(outcome{
		Code:    result.Code,
		Message: result.Message,
		Lines:   result.Lines,
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
  odachi plan "<texto>" [--req R1]
  odachi plan --done T2 [--commit abc1234]
  odachi level [--to <perfil>]
  odachi summary
  odachi undo [--list] [--file brief]
  odachi check [flags]
  odachi review [--exec "<comando>" | --attach <archivo> | --show | --refute <veredicto>]
  odachi review [--lens <lente>]
  odachi doctor [--json]
  odachi bench [--journeys <nombres>] [--binary <ruta>] [--json]
  odachi init [--json]
  odachi skills [--check | --write]
  odachi skill new <slug>
  odachi pack [--agents claude,codex|standard] [--scope project|home] [--dest <carpeta>] [--dry-run]
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
  --require-review  exige un veredicto de review para este árbol (apagado por defecto)
  --json          salida para máquinas

flags de review:
  --exec s        corre este comando con el prompt por stdin (y {prompt} como ruta a ese
                  mismo archivo) y guarda el veredicto que contesta
  --attach path   guarda un veredicto o una refutación ya escrita (agente o humano)
  --show          qué veredicto hay para este árbol, o que no hay
  --lens s        lente explícita: risk | readability | reliability | resilience | all
  --refute path   imprime el prompt del refutador para ese veredicto
  --base ref      contra qué comparar el candidato (default HEAD; vacío: contra nada)
  --json          salida para máquinas

flags de plan:
  --req R<n>      requisito al que se ata el paso nuevo
  --done T<n>     paso que se marca hecho
  --commit sha    commit que cerró el paso (acompaña a --done)
  --json          salida para máquinas

flags de skills:
  --check         valida cada skill y SKILLS.md; sale con 1 si algo no cumple
  --write         regenera SKILLS.md desde los frontmatter
  --json          salida para máquinas

flags de pack:
  --agents s      solo estos agentes, separados por coma (default: los detectados);
                  standard es la carpeta que leen todos los clientes (.agents/skills)
  --scope s       project | home: el mismo agente, en la otra raíz (default: el de la tabla)
  --dest path     instala todo en esta carpeta: absoluta, o relativa al proyecto
  --dry-run       qué haría, sin escribir nada
  --json          salida para máquinas

flags de bench:
  --journeys s    solo estos recorridos, separados por coma
  --binary path   el binario que corre cada recorrido (default: el que está corriendo)
  --json          salida para máquinas

flags de init:
  --json          salida para máquinas, con los destinos de las skills y las personas

qué deja escrito init, en un solo comando:
  AGENTS.md              el bloque de reglas que el agente lee
  .gitignore             /proof/ y /.odachi/reviews/ no ensucian el árbol que certifican
  .git/hooks/pre-commit  no se commitea sin odachi check --quiet en verde
  .agents/skills         las 8 skills del proyecto: la carpeta que leen todos los clientes
  ~/.agents/skills       la skill de arranque: cómo empezar un proyecto desde cero
  .opencode/agents       las dos personas, novato y senior: se eligen con Tab
  .claude/output-styles  las mismas dos personas: se eligen con /config

las personas se escriben siempre, para cada cliente que sepa elegirlas: init no
sabe quién sos, y la persona la elegís vos con la tecla de tu cliente. El
ceremonial lo sigue decidiendo .odachi/profile.toml, y el perfil no cambia la
persona: son dos ejes.

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
  odachi plan "<texto>" [--req R1]              agrega el próximo paso al task doc
  odachi plan --done T2 [--commit abc1234]      marca un paso hecho y registra el commit
  odachi level                                  ver tus hechos; propone senior cuando se gana
  odachi level --to <perfil>                    aplicar un perfil (se respalda el anterior)
  odachi summary                                qué hay hecho en cada feature
  odachi undo                                   restaura lo último que escribió odachi (lo actual queda respaldado)
  odachi undo --list                            qué se puede deshacer, con fecha
  odachi undo --file brief                      solo ese: brief | profile | fence | task
  odachi check                                  ¿está listo? brief, cerco, preguntas y prueba
  odachi check --run "<comando>"                corre la prueba y la guarda con su árbol
  odachi review                                 el prompt de review de este candidato
  odachi review --exec "<comando>"              corre el revisor con el prompt por stdin y guarda su veredicto
  odachi review --exec "agente {prompt}"        igual, pero el revisor recibe la ruta del prompt
  odachi review --attach <archivo>              guarda un veredicto o una refutación ya escrita
  odachi review --show                          qué veredicto hay para este árbol
  odachi review --lens risk                     agrega una lente a las que pide la evidencia
  odachi review --lens --show                   qué lentes pide este árbol y cuáles faltan
  odachi review --refute <veredicto>            el prompt del refutador para ese veredicto
  odachi check --require-review                 exige review para este árbol y sus lentes
  odachi doctor                                 diez chequeos de salud, de solo lectura
  odachi bench                                  mide cómo se traba una persona (in-band / dead-end)
  odachi init                                   AGENTS.md + hook + las skills del proyecto en
                                                 .agents/skills (y la de arranque en ~/.agents/skills)
                                                 + las dos personas en cada cliente que las elige
  odachi skills                                 las skills del proyecto y su trigger
  odachi skills --check                         valida (lo corre CI); exit 1 si algo no cumple
  odachi skills --write                         regenera SKILLS.md desde los frontmatter
  odachi skill new <slug>                       crea skills/<slug>/SKILL.md y actualiza el registro
  odachi pack --agents claude,codex             instala las skills incluidas en cada agente
  odachi pack --agents standard                 en ./.agents/skills, la carpeta que leen todos los clientes
  odachi pack --agents standard --scope home    en ~/.agents/skills: una copia para todos los proyectos
  odachi pack --agents opencode --scope home    el agente que es de proyecto, instalado en tu home
  odachi pack --agents mi-cliente --dest ./skills-x  un cliente con su propia ruta
  odachi pack --dry-run                         qué copiaría, sin escribir nada

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
