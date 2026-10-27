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
	"github.com/biggs-100/odachi/internal/app/progress"
	reviewapp "github.com/biggs-100/odachi/internal/app/review"
	"github.com/biggs-100/odachi/internal/app/start"
	"github.com/biggs-100/odachi/internal/app/summary"
	"github.com/biggs-100/odachi/internal/app/undo"
	"github.com/biggs-100/odachi/internal/domain/agent"
	"github.com/biggs-100/odachi/internal/domain/persona"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
	bundledskills "github.com/biggs-100/odachi/skills"
)

// version is the release identifier this binary reports. A plain `go build`
// leaves it as "dev"; goreleaser injects it with -ldflags -X main.version=<tag>,
// so the number a human reads always comes from the release, never from a hand
// edit of a source file.
var version = "dev"

// versionLine is the only shape the two surfaces print. It takes the identifier
// as a parameter instead of reading the global, so a test can pin what ldflags
// injects without writing state another parallel test reads.
func versionLine(v string) string { return "odachi " + v + "\n" }

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run runs the CLI in the current directory. It returns the exit code instead
// of calling os.Exit so the behaviour is testable.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// The version ask comes before anything that can fail: a deleted working
	// directory must not stop a human from asking what they installed.
	if isVersionAsk(args) {
		fmt.Fprint(stdout, versionLine(version))
		return start.CodeOK
	}

	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stderr, "odachi: could not read the current directory: "+oneLine(err.Error()))
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
		fmt.Fprint(stdout, versionLine(version))
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
	case "progress":
		return runProgress(dir, args[1:], stdout, stderr)
	case "undo":
		return runUndo(dir, args[1:], stdout, stderr)
	case "check":
		return runCheck(dir, args[1:], stdout, stderr)
	case "init":
		return runInit(dir, args[1:], stdout, stderr)
	case "setup":
		return runSetup(args[1:], stdout, stderr)
	case "skills":
		return runSkills(dir, args[1:], stdout, stderr)
	case "skill":
		return runSkill(dir, args[1:], stdout, stderr)
	case "role":
		return runRole(dir, args[1:], stdout, stderr)
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
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi version takes no arguments\n", args[1])
		return start.CodeBlocked
	default:
		fmt.Fprintf(stderr, "odachi: unknown command: %s; available commands: start (odachi start), ask (odachi ask \"<question>\"), plan (odachi plan \"<text>\" | odachi plan --done T2), level (odachi level), summary (odachi summary), progress (odachi progress), undo (odachi undo), check (odachi check), review (odachi review), doctor (odachi doctor), bench (odachi bench), init (odachi init), skills (odachi skills), skill (odachi skill new <slug>), role (odachi role new <slug>), pack (odachi pack), version (odachi --version)\n", args[0])
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
	// Flag errors are translated into one English line below, so the standard
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

	flags.StringVar(&opts.Outcome, "outcome", "", "answer 1: what has to exist when we finish")
	flags.StringVar(&opts.Audience, "audience", "", "answer 2: phone | cpu | both")
	flags.Var(&done, "done", "answer 3: how you will know it worked (repeatable)")
	flags.Var(&exclude, "exclude", "answer 4: no-goal by catalog key or free text (repeatable)")
	flags.StringVar(&opts.Sensitive, "sensitive", "", "answer 5: none | names | health | money")
	flags.StringVar(&opts.Deadline, "deadline", "", "answer 6: today | week | relaxed")
	flags.Var(&learn, "learn", "answer 7: what you want to learn (repeatable)")
	flags.StringVar(&brief, "brief", "", "read an existing brief instead of asking")
	flags.BoolVar(&force, "force", false, "rewrite an existing brief.md")
	flags.BoolVar(&opts.WaiveReview, "waive-review", false, "record that this feature lands without a review verdict (the proof and the fence still apply)")
	flags.BoolVar(&asJSON, "json", false, "machine output")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi start -h to see the flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi start only takes flags (run odachi start -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}

	opts.Dir = dir
	// The feature starts from the commit that exists now: the task doc records
	// it, and the closing gate asks for the review of that diff. A repository
	// with no commit records nothing, and the gate keeps asking per tree.
	if head, err := (gitRepo{}).Head(dir); err == nil {
		opts.Base = head
	}
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

	agents, agentsCode, agentsMessage := storedAgents(dir, "odachi start")
	if agentsCode != start.CodeOK {
		return report(start.Result{Code: agentsCode, Message: agentsMessage}, asJSON, stdout, stderr)
	}
	opts.Agents = agents

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

	flags.StringVar(&opts.Answer, "answer", "", "the human's answer to an open question")
	flags.StringVar(&opts.For, "for", "", "P<n> of the question being answered (required when two or more are open)")
	flags.IntVar(&opts.Feature, "feature", 0, "the feature number to write to (default: the highest)")
	flags.StringVar(&opts.Craft, "craft", "", "decision the agent took without asking")
	flags.StringVar(&opts.Because, "because", "", "why the agent took that decision (required with --craft)")
	flags.StringVar(&opts.Change, "change", "", "a requirement the human changed, in their words")
	flags.StringVar(&opts.Requirement, "req", "", "R<n> the change names (required with --change)")
	flags.BoolVar(&asJSON, "json", false, "machine output")

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
		fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi ask -h to see the flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi ask takes a single question in quotes\n", flags.Arg(0))
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

	flags.StringVar(&opts.Requirement, "req", "", "R<n> of the requirement the new step is tied to")
	flags.StringVar(&opts.Done, "done", "", "T<n> of the step that is marked done")
	flags.StringVar(&opts.Commit, "commit", "", "commit that closed the step (goes with --done)")
	flags.IntVar(&opts.Feature, "feature", 0, "the feature number to write to (default: the highest)")
	flags.BoolVar(&asJSON, "json", false, "machine output")

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
		fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi plan -h to see the flags")
		return start.CodeBlocked
	}

	// The text may also come after the flags: `plan --req R1 "texto"`.
	rest := flags.Args()
	if opts.Text == "" && len(rest) == 1 {
		opts.Text = rest[0]
		rest = nil
	}
	if len(rest) > 0 {
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi plan takes a single text in quotes\n", rest[0])
		return start.CodeBlocked
	}

	opts.Dir = dir
	opts.Git = gitRepo{}
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

	flags.StringVar(&to, "to", "", "profile to apply: .odachi/profile.toml is backed up and rewritten")
	flags.BoolVar(&asJSON, "json", false, "machine output")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi level -h to see the flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi level only takes flags (run odachi level -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}

	current, code, message := currentProfile(dir, "odachi level", asJSON || !isTerminal(stdin))
	if code != start.CodeOK {
		return reportLevel(level.Result{Code: code, Message: message}, asJSON, stdout, stderr)
	}

	agents, agentsCode, agentsMessage := storedAgents(dir, "odachi level")
	if agentsCode != start.CodeOK {
		return reportLevel(level.Result{Code: agentsCode, Message: agentsMessage}, asJSON, stdout, stderr)
	}

	return reportLevel(level.Run(level.Options{Dir: dir, Current: *current, To: to, Agents: agents}, osFiles{}), asJSON, stdout, stderr)
}

// runSummary parses the flags and prints the derived report.
func runSummary(dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi summary", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var asJSON bool
	flags.BoolVar(&asJSON, "json", false, "machine output")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi summary -h to see the flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi summary takes no arguments (run odachi summary -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}

	return reportSummary(summary.Run(summary.Options{Dir: dir}, osFiles{}), asJSON, stdout, stderr)
}

// runProgress prints where the open feature stands. It reads the project's
// review switch the same way check does, so the report never claims a gate the
// project turned off, and it writes nothing.
func runProgress(dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi progress", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var asJSON bool
	flags.BoolVar(&asJSON, "json", false, "machine output: the same report as one object")
	feature := flags.Int("feature", 0, "the feature number to report on (default: the highest)")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi progress -h to see the flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi progress takes no arguments (run odachi progress -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}

	enabled, err := reviewapp.Enabled(dir, osFiles{}, reviewModeCodec{})
	if err != nil {
		fmt.Fprintln(stderr, "odachi: "+oneLine(err.Error()))
		return start.CodeBlocked
	}

	result := progress.Run(progress.Options{
		Dir:        dir,
		ReviewMode: enabled,
		Feature:    *feature,
		// The gate's own rule, asked instead of re-derived: one owner, one answer.
		Review: func() (bool, string, error) {
			return check.ReviewOwed(check.Options{Dir: dir, ReviewMode: enabled, Feature: *feature}, osFiles{}, gitRepo{})
		},
	}, osFiles{}, gitRepo{})

	if asJSON {
		return reportProgressJSON(result, stdout, stderr)
	}
	for _, line := range result.Lines {
		fmt.Fprintln(stdout, line)
	}
	return result.Code
}

// reportProgressJSON prints the report as one object: the same facts the lines
// carry, for a renderer that lives in another program. The names are the contract.
func reportProgressJSON(result progress.Result, stdout, stderr io.Writer) int {
	payload := struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		State   *progress.State `json:"state,omitempty"`
	}{Code: result.Code, Message: result.Message, State: result.State}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Fprintln(stderr, "odachi: could not print the report as JSON: "+oneLine(err.Error()))
		return start.CodeBlocked
	}
	fmt.Fprintln(stdout, string(data))
	return result.Code
}

// runCheck parses the flags, wires the three adapters and reports the gate.
func runCheck(dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi check", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var (
		opts     check.Options
		asJSON   bool
		noReview bool
	)

	flags.StringVar(&opts.Run, "run", "", "run the command, capture the output and write proof/<NNNN>-<slug>.txt")
	flags.StringVar(&opts.Proof, "proof", "", "use a proof that is already written (agents and CI)")
	flags.StringVar(&opts.Base, "base", "HEAD", "what to compare the fence against (empty: against nothing, everything counts as new)")
	flags.StringVar(&opts.FenceFrom, "fence-from", "", "read the fence from this git ref instead of the worktree (CI: the base branch)")
	flags.IntVar(&opts.TimeoutSeconds, "timeout", check.DefaultTimeout, "seconds --run may take")
	flags.BoolVar(&opts.Quiet, "quiet", false, "prints nothing when ready (for the hook)")
	flags.BoolVar(&opts.RequireReview, "require-review", false, "requires a review verdict for this tree (default: the project mode)")
	flags.BoolVar(&noReview, "no-review", false, "does not require review even when the project has it on")
	flags.BoolVar(&asJSON, "json", false, "machine output")
	flags.IntVar(&opts.Feature, "feature", 0, "the feature number to judge (default: the highest)")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi check -h to see the flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi check only takes flags (run odachi check -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}
	if opts.TimeoutSeconds <= 0 {
		fmt.Fprintln(stderr, "odachi: --timeout has to be greater than zero seconds; run odachi check -h to see the flags")
		return start.CodeBlocked
	}
	if opts.Run != "" && opts.Proof != "" {
		fmt.Fprintln(stderr, "odachi: --run and --proof do not go together: use one of the two (odachi check --run \"<command>\" or odachi check --proof <file>)")
		return start.CodeBlocked
	}
	if opts.RequireReview && noReview {
		fmt.Fprintln(stderr, "odachi: --require-review and --no-review do not go together: choose one; without flags the project mode decides")
		return start.CodeBlocked
	}

	// The project owns the default; the flags are the exception for one run.
	// A malformed switch blocks instead of quietly reading as off: a human who
	// wrote `mode = "on"` must not get a gate that never opened.
	opts.NoReview = noReview
	if !opts.RequireReview && !noReview {
		enabled, err := reviewapp.Enabled(dir, osFiles{}, reviewModeCodec{})
		if err != nil {
			fmt.Fprintln(stderr, "odachi: "+oneLine(err.Error()))
			return start.CodeBlocked
		}
		opts.ReviewMode = enabled
	}

	opts.Dir = dir
	// Only a run records who ran it. Resolving the model costs a process, so a
	// check that judges an existing proof (the hook, CI) pays nothing.
	if opts.Run != "" {
		opts.Model = hostModel{}.Model(dir)
	}
	return reportCheck(check.Run(opts, osFiles{}, gitRepo{}, osShell{}), opts.Quiet, asJSON, stdout, stderr)
}

// runReview parses the flags, wires the four adapters and reports the verdict.
func runReview(dir string, args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "mode" {
		return runReviewMode(dir, args[1:], stdout, stderr)
	}

	flags := flag.NewFlagSet("odachi review", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var (
		opts   reviewapp.Options
		asJSON bool
	)

	flags.StringVar(&opts.Exec, "exec", "", "run this command with the prompt on stdin and save the verdict it answers")
	flags.StringVar(&opts.Model, "model", "", "model that replaces {model} in --exec (default: the session's, or ODACHI_MODEL)")
	flags.StringVar(&opts.Attach, "attach", "", "save a verdict or a refutation already written in this file")
	flags.BoolVar(&opts.Show, "show", false, "which verdict exists for this tree, or that there is none")
	flags.StringVar(&opts.Lens, "lens", "", "explicit lens: risk | readability | reliability | resilience | all (repeatable or comma-separated)")
	flags.StringVar(&opts.Refute, "refute", "", "print the refuter's prompt for this verdict")
	flags.StringVar(&opts.Base, "base", reviewapp.DefaultBase, "what to compare the candidate against (empty: against nothing, everything counts as new)")
	flags.BoolVar(&asJSON, "json", false, "machine output")

	if err := flags.Parse(withoutEmptyLens(args)); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi review -h to see the flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi review only takes flags (run odachi review -h)\n", flags.Arg(0))
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
		fmt.Fprintln(stderr, `odachi: --exec, --attach, --show and --refute do not go together: choose a single form (odachi review | odachi review --exec "<command>" | odachi review --attach <file> | odachi review --show | odachi review --refute <verdict>)`)
		return start.CodeBlocked
	}

	opts.Dir = dir
	// The project's lens table is read here, not inside the verb: a malformed file
	// blocks with the path and the fix, which is the CLI's job to say.
	config, err := reviewapp.Config(dir, osFiles{}, reviewModeCodec{})
	if err != nil {
		fmt.Fprintln(stderr, "odachi: "+oneLine(err.Error()))
		return start.CodeBlocked
	}
	opts.LensModels = config.Lenses

	return reportReview(reviewapp.Run(opts, osFiles{}, gitRepo{}, osShell{}, reviewLenses{}, hostModel{}), asJSON, stdout, stderr)
}

// runReviewMode parses the small mode grammar by hand: the action is a word,
// not a flag value, and `--json` should work before or after it.
func runReviewMode(dir string, args []string, stdout, stderr io.Writer) int {
	const usageLine = "usage: odachi review mode [status|on|off] [--json]"

	action := reviewapp.ModeStatus
	actionSeen := false
	asJSON := false

	for _, arg := range args {
		switch arg {
		case "-h", "--help":
			fmt.Fprintln(stdout, usageLine)
			return start.CodeOK
		case "--json", "-json":
			asJSON = true
		case reviewapp.ModeStatus, reviewapp.ModeOn, reviewapp.ModeOff:
			if actionSeen {
				fmt.Fprintln(stderr, "odachi: the review mode is chosen once; "+usageLine)
				return start.CodeBlocked
			}
			action = arg
			actionSeen = true
		default:
			fmt.Fprintf(stderr, "odachi: I do not understand %q; %s\n", arg, usageLine)
			return start.CodeBlocked
		}
	}

	result := reviewapp.Mode(reviewapp.ModeOptions{Dir: dir, Action: action}, osFiles{}, reviewModeCodec{})
	return reportReview(result, asJSON, stdout, stderr)
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

	flags.BoolVar(&opts.List, "list", false, "what can be undone, with its date")
	flags.BoolVar(&opts.Collect, "collect", false, "move old backups to .odachi/backups")
	flags.StringVar(&opts.File, "file", "", "only that artifact: brief | profile | fence | task")
	flags.BoolVar(&asJSON, "json", false, "machine output")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi undo -h to see the flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi undo only takes flags (run odachi undo -h)\n", flags.Arg(0))
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
			fmt.Fprintln(stderr, "odachi: could not build the JSON output: "+oneLine(err.Error()))
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
			fmt.Fprintln(stderr, "odachi: could not build the JSON output: "+oneLine(err.Error()))
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

	var (
		asJSON      bool
		skillsScope string
		agents      string
		global      bool
	)
	flags.BoolVar(&asJSON, "json", false, "machine output")
	flags.StringVar(&skillsScope, "skills", "project", "where to install the skills: project | home")
	flags.StringVar(&agents, "agents", "", "which clients receive the native agent: opencode | claude | both")
	flags.BoolVar(&global, "global", false, "install the home door, the one that shows up in any folder, and make it OpenCode's default agent")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi init -h to see the flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi init only takes flags (run odachi init -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}
	switch skillsScope {
	case "project", "home":
	default:
		fmt.Fprintf(stderr, "odachi: --skills has to be \"project\" or \"home\", not %q; project leaves the skills in the project, home leaves them in ~/.agents/skills\n", skillsScope)
		return start.CodeBlocked
	}
	if strings.TrimSpace(agents) != "" {
		if _, err := persona.NormalizeSelection(agents); err != nil {
			fmt.Fprintf(stderr, "odachi: %s; use opencode, claude or both\n", oneLine(err.Error()))
			return start.CodeBlocked
		}
	}

	created, err := ensureHistory(dir)
	if err != nil {
		fmt.Fprintln(stderr, "odachi: "+err.Error())
		return start.CodeBlocked
	}

	if global && os.Getenv("ODACHI_NO_SETUP") != "" {
		fmt.Fprintln(stderr, "odachi: ODACHI_NO_SETUP=1 is set: the home is not touched; run init without --global or unset the variable")
		return start.CodeBlocked
	}

	result := initapp.Run(initapp.Options{
		Dir:            dir,
		Exe:            hookExe(),
		PathMissing:    pathMissing(),
		Profiles:       writtenProfile{},
		Agents:         agents,
		AgentSelection: agentSelection{},
		Home:           osHome{},
		Global:         global,
	}, osFiles{}, &initSkills{dir: dir, scope: skillsScope})
	if created {
		line := "project history: I created it in this folder (git init); it is what lets me prove that a proof belongs to this state of the code"
		result.Lines = append([]string{line}, result.Lines...)
	}
	if len(result.Files) > 0 || created {
		result.Files = append([]string{".git/"}, result.Files...)
	}

	return reportInit(result, asJSON, stdout, stderr)
}

// runSetup parses the flags and installs the global half. It takes no directory:
// the point is that the folder the human is in is not touched. The installer
// runs it after copying the binary, so the first conversation already has an
// agent that knows the verbs.
func runSetup(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi setup", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var (
		agents string
		asJSON bool
	)
	flags.StringVar(&agents, "agents", "", "opencode | claude | both (default: the clients installed in the home)")
	flags.BoolVar(&asJSON, "json", false, "machine output")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi setup -h to see the flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi setup only takes flags (run odachi setup -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}

	if os.Getenv("ODACHI_NO_SETUP") != "" {
		fmt.Fprintln(stderr, "odachi: ODACHI_NO_SETUP=1 is set: this machine's home is not touched; unset it to run setup")
		return start.CodeBlocked
	}

	result := initapp.RunSetup(initapp.Options{
		Profiles: writtenProfile{},
		Agents:   agents,
		Home:     osHome{},
	}, osFiles{}, &initSkills{dir: ".", scope: "home"})
	return reportInit(result, asJSON, stdout, stderr)
}

// initSkills is the join between `odachi init` and `odachi pack`. init decides
// where the skills go — the project's cross-client directory by default, or the
// human's home when `--skills home` asks for one copy per machine — and pack owns
// how a skill is installed: the destination table, the backup before a differing
// file, the identical bytes skipped.
//
// The join lives here, in the CLI, and not in internal/app/init importing
// internal/app/pack: the CLI is the only layer that already wires pack's payload
// and its two adapters, and an app that imports another app would inherit its
// error strings and its destination table for a job that is not its own.
type initSkills struct {
	dir   string
	scope string
}

// Install writes the skills job and turns what pack reports into the lines a
// human reads. Project scope is the default and keeps the old two destinations:
// the curriculum in the project and bootstrap in the home. Home scope puts all
// ten in the home; it never deletes a project copy an earlier run left.
func (s initSkills) Install() initapp.SkillReport {
	if _, ok := bundledskills.Bootstrap(); !ok {
		return initapp.SkillReport{
			Blocked: "this binary does not carry the " + bundledskills.BootstrapName + " skill: it is an incomplete installation; download odachi again",
			Next:    "odachi init",
		}
	}

	if s.scope == "home" {
		all, err := bundledskills.All()
		if err != nil {
			return initapp.SkillReport{Blocked: oneLine(err.Error()) + "; download odachi again", Next: "odachi init"}
		}
		home := pack.Run(pack.Options{Agents: agent.Standard, Scope: string(agent.ScopeHome), Dir: s.dir, Skills: payload(all)}, osFiles{}, osHome{})
		report := s.report([]pack.Result{home})
		if standard, err := agent.Resolve(agent.Standard); err == nil {
			exists, err := (osFiles{}).Exists(filepath.Join(s.dir, filepath.FromSlash(standard.Dest)))
			if err == nil && exists {
				report.Notice = append(report.Notice, standard.Display()+": it already had a copy from an earlier init; I did not touch it: now the skills live in "+standard.In(agent.ScopeHome).Display())
			}
		}
		return report
	}

	curriculum, err := bundledskills.Curriculum()
	if err != nil {
		return initapp.SkillReport{Blocked: oneLine(err.Error()) + "; download odachi again", Next: "odachi init"}
	}
	bootstrap, _ := bundledskills.Bootstrap()

	// The project gets the craft, in the cross-client location every client
	// reads. The home gets the one skill that teaches how to arrive, because
	// before a project exists the home is the only place a skill can live.
	project := pack.Run(pack.Options{Agents: agent.Standard, Dir: s.dir, Skills: payload(curriculum)}, osFiles{}, osHome{})
	home := pack.Run(pack.Options{Agents: agent.Standard, Scope: string(agent.ScopeHome), Dir: s.dir, Skills: payload([]bundledskills.Skill{bootstrap})}, osFiles{}, osHome{})
	return s.report([]pack.Result{project, home})
}

// report turns the pack results into one SkillReport. It is shared by the two
// scopes so the failure sentences and the fatal rule cannot drift.
func (s initSkills) report(results []pack.Result) initapp.SkillReport {
	report := initapp.SkillReport{}
	for _, result := range results {
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
	report.Notice = append(report.Notice, s.otherProjectCopy()...)
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
		lines = append(lines, a.Display()+": it already has a copy of the skills (odachi pack --agents "+a.Name+" put it there); I do not touch it: the project one is "+standard.Display())
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
		return "could not write " + dest.Failure.Path + ": " + dest.Failure.Reason + "; check the permissions and run odachi init again"
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
				Scope:     dest.Scope,
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
			fmt.Fprintln(stderr, "odachi: could not build the JSON output: "+oneLine(err.Error()))
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

// initPersonaDest is one client's native agent in the same report: which client,
// which scope, where it reads agents from, what this run did there, and the file
// it owns. Project and home are two rows because they are two files with two
// jobs: the mirror obeys the project, the door works in any folder.
type initPersonaDest struct {
	Client    string   `json:"client"`
	Scope     string   `json:"scope"`
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
			fmt.Fprintln(stderr, "odachi: could not build the JSON output: "+oneLine(err.Error()))
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

// usage is the help a human reads with `odachi start -h`.
func usage(out io.Writer) {
	fmt.Fprint(out, `odachi · two hands: you decide, the agent builds

usage:
  odachi                      the same as odachi start
  odachi start [flags]
  odachi ask "<question>"
  odachi ask --answer "<answer>" [--for P1]
  odachi ask --craft "<decision>" --because "<reason>"
  odachi ask --change "<words>" --req R1
  odachi plan "<text>" [--req R1]
  odachi plan --done T2 [--commit abc1234]
  odachi level [--to <profile>]
  odachi summary
  odachi progress
  odachi undo [--list | --collect] [--file brief]
  odachi check [flags]
  odachi review [--exec "<command>" | --attach <file> | --show | --refute <verdict>]
  odachi review [--lens <lens>]
  odachi review mode [status|on|off]
  odachi doctor [--json]
  odachi bench [--journeys <names>] [--binary <path>] [--json]
  odachi init [--skills project|home] [--agents opencode|claude|both] [--global] [--json]
  odachi setup [--agents opencode|claude|both] [--json]
  odachi skills [--check | --write]
  odachi role new <slug>
  odachi role [<slug>]
  odachi skill new <slug>
  odachi pack [--agents claude,codex|standard] [--scope project|home] [--dest <folder>] [--dry-run]
  odachi --version
  odachi version

the version:
  --version       prints "odachi <version>" and exits 0 (a build without ldflags says dev)

check flags:
  --run s         run the command and write this tree's proof
  --proof path    use a proof that is already written (agents and CI)
  --base ref      what to compare the fence against (default HEAD; empty: against nothing)
  --fence-from ref  read the fence from this git ref instead of the worktree (CI: the base branch)
  --timeout n     seconds --run may take (default 120)
  --quiet         prints nothing when ready (for the hook)
  --require-review  requires a review verdict for this tree (default: the project mode)
  --no-review     does not require review even when the project has it on
  --json          machine output

review flags:
  --exec s        run this command with the prompt on stdin (and {prompt} as the path to that
                  same file) and save the verdict it answers
  --model s       model that replaces {model} in --exec (default: the session's, or
                  ODACHI_MODEL)
  --attach path   save a verdict or a refutation that is already written (agent or human)
  --show          which verdict exists for this tree, or that there is none
  --lens s        explicit lens: risk | readability | reliability | resilience | all
  --refute path   print the refuter's prompt for that verdict
  --base ref      what to compare the candidate against (default HEAD; empty: against nothing)
  --json          machine output

review mode:
  odachi review mode status   whether the project requires review (default: off)
  odachi review mode on       check requires the verdict for this tree
  odachi review mode off      check does not require it, even when the hook runs check
  the mode lives in .odachi/review.toml; --require-review and --no-review override it for one run

plan flags:
  --req R<n>      requirement the new step is tied to
  --done T<n>     step that is marked done
  --commit sha    commit that closed the step (goes with --done; its tree is recorded when git can read it)
  --json          machine output

undo flags:
  --list          what can be undone, with its date
  --collect       move old backups to .odachi/backups; it deletes no version
  --file s        only that artifact: brief | profile | fence | task
  --json          machine output

skills flags:
  --check         validate each skill and SKILLS.md; exits 1 when something does not comply
  --write         regenerate SKILLS.md from the frontmatter
  --json          machine output

pack flags:
  --agents s      only these agents, comma-separated (default: the detected ones);
                  standard is the folder every client reads (.agents/skills)
  --scope s       project | home: the same agent, in the other root (default: the one in the table)
  --dest path     install everything in this folder: absolute, or relative to the project
  --dry-run       what it would do, without writing anything
  --json          machine output

bench flags:
  --journeys s    only these journeys, comma-separated
  --binary path   the binary each journey runs (default: the one that is running)
  --json          machine output

init flags:
  --skills s      project | home (default project): the project carries the craft skills;
                  home leaves all of them in ~/.agents/skills and writes no copy in the project
  --agents s      opencode | claude | both (default both): which clients receive the native
                  agent and the home door; the choice stays in .odachi/agents.toml
                  and does not delete the files another client already has
  --global        install the home door: the agent that shows up in any folder,
                  not only in this project, and make it OpenCode's default agent.
                  It is opt-in: without the flag nothing is written
                  to your home, because the selector of any folder is yours
  --json          machine output, with the destinations of the skills and of the agent

setup flags:
  odachi setup    the global half only, from any folder: the bootstrap skill in
                  ~/.agents/skills, the door for the clients found in the home,
                  and OpenCode's default agent. It writes nothing here. The
                  installer runs it, so the user only installs odachi and talks.
  --agents s      opencode | claude | both (default: the clients installed in the home)
  --json          machine output

what init leaves written, in one command (--skills project and --agents both, the defaults):
  AGENTS.md              the rules block the agent reads, with the voice at the end
  .opencode/agents       odachi.md (chosen with Tab) + odachi-worker.md and odachi-verify.md (subagents)
  .claude/output-styles  odachi.md: the same block, to choose it with /config in this project
  .claude/agents         odachi-worker.md and odachi-verify.md: the Claude Code subagents
  .gitignore             /proof/, /.odachi/reviews/ and *.bak-* do not dirty the tree
  .git/hooks/pre-commit  no commit without odachi check --quiet green
  .agents/skills         the project's skills: the folder every client reads
  ~/.agents/skills       the bootstrap skill: how to start a project from scratch

the home door -the agent that shows up in any folder- is opt-in: an init without
--global does not write it. With --global it goes to ~/.config/opencode/agents and
~/.claude/output-styles, one per chosen client, with the subagents it invokes
next to it, and OpenCode gets default_agent
in ~/.config/opencode/opencode.json (or the .jsonc that already exists) unless a
default you picked is already there.

with --agents opencode nothing is written to .claude, and with --agents claude nothing is
written to .opencode; the files of a client that is already installed are not deleted. The choice
is kept in .odachi/agents.toml and a later init preserves it; without the file, the default
is still both.

with --skills home, all the skills go to ~/.agents/skills and the project gets no copy;
if it already had .agents/skills from an earlier init, it is not deleted: it is reported.

the odachi-worker and odachi-verify subagents are invoked by the agent, they are not chosen with
a key. The parent hands them the task doc, the step, the surfaces, the commands and
the skill paths; they come back with evidence and ≤2k tokens, and they do not commit.

the voice —how the agent talks to you— is the last section of the block, and the binary
writes it: novato when there is no .odachi/profile.toml yet, otherwise the one of the profile
it names. The native agent has two copies per chosen client: the project
mirror, with the same rules and the same voice, and the home door, which shows up
in any folder and reads the project block when it exists; when it does not exist, it guides you to start with
odachi init and odachi start. odachi level --to <profile> changes the profile, that
section and the project mirror, and touches nothing else in the block.

flags:
  --outcome s     answer 1: what has to exist when we finish
  --audience s    answer 2: phone | cpu | both
  --done s        answer 3 (repeatable): how you will know it worked
  --exclude s     answer 4 (repeatable): no-goal by key or free text
  --sensitive s   answer 5: none | names | health | money
  --deadline s    answer 6: today | week | relaxed
  --learn s       answer 7 (repeatable): what you want to learn
  --brief path    read an existing brief instead of asking
  --force         rewrite an existing brief.md (it is backed up first)
  --json          machine output

after start:
  odachi ask "<question>"                       record an open question
  odachi ask --answer "<answer>" [--for P1]     record your answer
  odachi ask --craft "<decision>" --because "<reason>"  record a decision the agent took
  odachi ask --change "<words>" --req R1        record the requirement the human changed; its steps reopen
  odachi plan "<text>" [--req R1]               add the next step to the task doc
  odachi plan --done T2 [--commit abc1234]      mark a step done; the tree of its commit is recorded
  odachi level                                  see your facts; it proposes senior when earned
  odachi level --to <profile>                   apply a profile and its voice (both are backed up)
  odachi summary                                what is done in each feature
  odachi undo                                   restore the last thing odachi wrote (the current one is backed up)
  odachi undo --list                            what can be undone, with its date
  odachi undo --file brief                      only that one: brief | profile | fence | task
  odachi check                                  is it ready? brief, fence, questions and proof
  odachi check --run "<command>"                run the proof and save it with its tree
  odachi review                                 the review prompt for this candidate
  odachi review --exec "<command>"              run the reviewer with the prompt on stdin and save its verdict
  odachi review --exec "agent {prompt}"         the same, but the reviewer receives the prompt path
  odachi review --attach <file>                 save a verdict or a refutation that is already written
  odachi review --show                          which verdict exists for this tree
  odachi review --lens risk                     add a lens to the ones the evidence asks for
  odachi review --lens --show                   which lenses this tree asks for and which are missing
  odachi review --refute <verdict>              the refuter's prompt for that verdict
  odachi check --require-review                 requires review for this tree and its lenses
  odachi doctor                                 twelve health checks, read-only
  odachi bench                                  measures how a person gets stuck (in-band / dead-end)
  odachi init                                   AGENTS.md (with the profile voice) + the project
                                                 agent and the home door (chosen with Tab)
                                                 + hook + the project skills in .agents/skills,
                                                 and the bootstrap one in ~/.agents/skills
  odachi skills                                 the project's skills and their trigger
  odachi skills --check                         validate (CI runs it); exit 1 when something does not comply
  odachi skills --write                         regenerate SKILLS.md from the frontmatter
  odachi skill new <slug>                       create skills/<slug>/SKILL.md and update the registry
  odachi pack --agents claude,codex             install the bundled skills into each agent
  odachi pack --agents standard                 in ./.agents/skills, the folder every client reads
  odachi pack --agents standard --scope home    in ~/.agents/skills: one copy for every project
  odachi pack --agents opencode --scope home    the project agent, installed in your home
  odachi pack --agents my-client --dest ./skills-x  a client with its own path
  odachi pack --dry-run                         what it would copy, without writing anything

exit codes: 0 artifacts written · 1 blocked · 2 a human answer is missing
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
