// Package bench implements `odachi bench`: the instrument that measures how a
// person gets stuck.
//
// It runs a corpus of journeys — each one a sequence of invocations of the binary
// under test, in a throwaway directory with a controlled HOME — and classifies
// every non-zero exit by what its message named: a command that resolves the
// block (in_band), no command (out_of_band), a command that does not resolve it
// (dead_end), or a command with a `<placeholder>` (templated). Any dead end fails
// the run.
//
// It is pure orchestration over two ports: Files (the sandbox directories) and
// Exec (the process runner). The rules and the corpus live in internal/bench, so
// the instrument can grow a journey without growing any logic.
//
// Two invariants are worth stating because they are what make the reading
// trustworthy: the bench only ever runs the binary under test or `odachi`
// commands a message named — never an arbitrary command — and it writes only
// inside its own temporary directories, which it removes.
package bench

import (
	"fmt"
	"path/filepath"
	"strings"

	benchcore "github.com/biggs-100/odachi/internal/bench"
)

// Exit codes. A dead end and a bench that could not run share one code because
// both mean "do not trust this tree yet".
const (
	// CodeOK means every block named something that resolves.
	CodeOK = 0
	// CodeFailed means a dead end exists, or the bench could not run.
	CodeFailed = 1
)

// sandboxPattern is the prefix of every temporary directory. It is recognisable
// on purpose: a crashed run leaves one behind and a human has to know what it is.
const sandboxPattern = "odachi-bench-"

// homeDir is the directory inside the sandbox that plays the user's home.
const homeDir = "home"

// Options is one `odachi bench` invocation.
type Options struct {
	// Binary is the binary every journey drives, and the one a named continuation
	// runs. Empty is a block: the bench has no default of its own, because the
	// only honest default is "the binary that is running", and only the CLI knows
	// which one that is.
	Binary string

	// Only filters the corpus by name. Empty means every journey.
	Only []string

	// Corpus is the corpus to run. Nil means benchcore.Journeys(): the shipped one.
	// It is injectable so a test can measure four known blocks instead of the
	// product's own state.
	Corpus []benchcore.Journey

	// Env is the environment the sandbox starts from. The run replaces HOME,
	// USERPROFILE and the git config with paths inside the sandbox.
	Env []string
}

// Files is the sandbox port: create a temporary directory, put files in it, and
// remove it when the journey is over.
type Files interface {
	MkdirTemp(pattern string) (string, error)
	MkdirAll(path string) error
	Write(path string, data []byte) error
	RemoveAll(path string) error
}

// Exec is the process port: run a program in a directory with an explicit
// environment and return what it printed and what it exited. A non-zero exit is a
// result, not an error; err means the process could not be run at all.
type Exec interface {
	Run(dir string, env []string, name string, args ...string) (output string, exitCode int, err error)
}

// Result is what the CLI prints and exits with. The four counts are separate
// because only one of them fails the run, and the dead ends carry the command
// that did not resolve and the message that named it.
type Result struct {
	Code        int
	Message     string
	Lines       []string
	Journeys    []JourneyResult
	Steps       int
	Blocks      int
	InBand      int
	OutOfBand   int
	Templated   int
	Interactive int
	DeadEnd     int
	DeadEnds    []DeadEnd
}

// JourneyResult is what one journey found. Surprises are the steps that exited
// something other than what the journey says: the product moved, and a bench that
// stays quiet about it would be measuring a product that no longer exists.
//
// Failed says the journey could not be measured at all — no sandbox, a setup step
// that did not run, a binary that could not be executed. Nothing was learned from
// it, so it never reads as success.
type JourneyResult struct {
	Name      string
	Steps     int
	Blocks    int
	Surprises []string
	Failed    bool
}

// DeadEnd is one block whose message named a command that did not resolve it.
type DeadEnd struct {
	Journey string
	Command string
	// Exit is what the named command exited.
	Exit int
	// Message is what the named command printed, on one line: the evidence a
	// human reads to judge the classification.
	Message string
}

// Run performs one `odachi bench`.
func Run(opts Options, files Files, exec Exec) Result {
	binary := strings.TrimSpace(opts.Binary)
	if binary == "" {
		return failed("I do not know which binary to run: pass --binary <path> (without it the bench cannot measure anything)")
	}

	corpus := opts.Corpus
	if corpus == nil {
		corpus = benchcore.Journeys()
	}

	selected, unknown := benchcore.Named(corpus, opts.Only)
	if len(unknown) > 0 {
		return failed("unknown " + quoted(unknown) + "; the journeys are: " + names(corpus))
	}
	if len(selected) == 0 {
		return failed("there are no journeys to run; the journeys are: " + names(corpus))
	}

	r := &runner{opts: opts, binary: binary, files: files, exec: exec}
	for _, journey := range selected {
		r.run(journey)
	}
	return r.report()
}

// runner holds one run.
type runner struct {
	opts   Options
	binary string
	files  Files
	exec   Exec

	results []JourneyResult
	blocks  []block
}

// block is one invocation that exited non-zero, before it is classified.
type block struct {
	journey string
	output  string

	command []string
	kind    benchcore.Kind
	// continuation is what the named command printed when it was run, and what it
	// exited: the evidence behind the classification.
	continuation     string
	continuationExit int
}

// run executes one journey and records its result. A journey can fail in ways that
// are not blocks — no temp directory, a setup command that cannot run — and every
// one of them is said instead of being swallowed.
func (r *runner) run(journey benchcore.Journey) {
	measured := r.execute(journey)
	result := JourneyResult{
		Name:      journey.Name,
		Steps:     len(journey.Steps),
		Blocks:    len(measured.blocks),
		Surprises: measured.surprises,
		Failed:    !measured.measured,
	}
	r.results = append(r.results, result)
	r.blocks = append(r.blocks, measured.blocks...)
}

// outcome is what one journey produced: the blocks it found, what surprised it,
// and whether it could be measured at all.
type outcome struct {
	blocks    []block
	surprises []string
	measured  bool
}

// execute runs one journey in its own sandbox and always removes it.
func (r *runner) execute(journey benchcore.Journey) (result outcome) {
	dir, err := r.files.MkdirTemp(sandboxPattern)
	if err != nil {
		return outcome{surprises: []string{"could not create the temporary directory: " + oneLine(err.Error())}}
	}
	defer func() {
		// The bench leaves the machine as it found it. A sandbox that cannot be
		// removed is a surprise, not a silent leak.
		if err := r.files.RemoveAll(dir); err != nil {
			result.surprises = append(result.surprises, "could not delete the temporary directory "+filepath.ToSlash(dir)+": "+oneLine(err.Error()))
		}
	}()

	home := filepath.Join(dir, homeDir)
	if err := r.files.MkdirAll(home); err != nil {
		return outcome{surprises: []string{"could not prepare the sandbox home: " + oneLine(err.Error())}}
	}
	env := benchcore.SandboxEnv(r.opts.Env, home)

	// Classification runs inside the sandbox and before it is removed (defers run
	// last in, first out), so a journey that fails halfway still reports the
	// blocks it found: a partial reading is a reading, and dropping it would hide
	// the dead end that caused the failure.
	defer func() { r.classify(dir, env, result.blocks) }()

	for _, action := range journey.Setup {
		if problem := r.act(dir, env, action); problem != "" {
			return outcome{surprises: []string{"the setup failed: " + problem}}
		}
	}

	for i, step := range journey.Steps {
		output, code, err := r.exec.Run(dir, env, r.binary, step.Args...)
		if err != nil {
			// The binary could not be run at all: nothing was measured.
			return outcome{
				blocks:    result.blocks,
				surprises: append(result.surprises, fmt.Sprintf("could not run step %d: %s", i+1, oneLine(err.Error()))),
			}
		}

		if code != step.Want {
			result.surprises = append(result.surprises, fmt.Sprintf("step %d exited %d and expected %d", i+1, code, step.Want))
		}
		if code != 0 {
			result.blocks = append(result.blocks, block{journey: journey.Name, output: output})
		}

		for _, action := range step.Then {
			if problem := r.act(dir, env, action); problem != "" {
				result.surprises = append(result.surprises, fmt.Sprintf("what follows step %d failed: %s", i+1, problem))
				return result
			}
		}
	}

	result.measured = true
	return result
}

// act performs one setup action inside the sandbox.
func (r *runner) act(dir string, env []string, action benchcore.Action) string {
	switch {
	case len(action.Command) > 0:
		output, code, err := r.exec.Run(dir, env, action.Command[0], action.Command[1:]...)
		if err != nil {
			return "could not run " + strings.Join(action.Command, " ") + ": " + oneLine(err.Error())
		}
		if code != 0 {
			return strings.Join(action.Command, " ") + " exited " + fmt.Sprint(code) + ": " + oneLine(output)
		}
		return ""
	case action.Path != "":
		path := filepath.Join(dir, filepath.FromSlash(action.Path))
		if err := r.files.MkdirAll(filepath.Dir(path)); err != nil {
			return "could not create the directory of " + action.Path + ": " + oneLine(err.Error())
		}
		if err := r.files.Write(path, []byte(action.Content)); err != nil {
			return "could not write " + action.Path + ": " + oneLine(err.Error())
		}
		return ""
	default:
		return ""
	}
}

// classify decides what each block's message names, and runs the named command in
// the same sandbox to find out whether it resolves.
//
// Only commands a message named are ever run, and only when they are `odachi`
// commands: the bench is not a shell. Each distinct command runs once per journey,
// so the same message cannot report two different verdicts.
func (r *runner) classify(dir string, env []string, blocks []block) {
	type outcome struct {
		output string
		exit   int
	}
	ran := map[string]outcome{}

	for i := range blocks {
		command := benchcore.Command(blocks[i].output)
		blocks[i].command = command

		if benchcore.IsTemplated(command) || len(command) == 0 {
			// Nothing to run: either a placeholder the bench cannot fill or no
			// command at all.
			blocks[i].kind = benchcore.KindOf(command, -1)
			continue
		}

		key := strings.Join(command, "\x00")
		result, ok := ran[key]
		if !ok {
			output, code, err := r.exec.Run(dir, env, r.binary, command[1:]...)
			if err != nil {
				output = oneLine(err.Error())
				code = -1
			}
			result = outcome{output: output, exit: code}
			ran[key] = result
		}

		blocks[i].continuation = result.output
		blocks[i].continuationExit = result.exit
		blocks[i].kind = benchcore.KindOf(command, result.exit)
	}
}

// report builds the block a human reads and the numbers the contract fixes. The
// order is the order of the corpus, so two runs over the same corpus print the
// same report.
func (r *runner) report() Result {
	lines := []string{}

	steps := 0
	counts := map[benchcore.Kind]int{}
	deadEnds := []DeadEnd{}
	surprises := []string{}

	for _, journey := range r.results {
		steps += journey.Steps
		surprises = append(surprises, journey.Surprises...)
	}

	for _, block := range r.blocks {
		counts[block.kind]++
		if block.kind != benchcore.DeadEnd {
			continue
		}
		deadEnds = append(deadEnds, DeadEnd{
			Journey: block.journey,
			Command: strings.Join(block.command, " "),
			Exit:    block.continuationExit,
			Message: firstLine(block.continuation),
		})
	}

	unmeasured := 0
	for _, journey := range r.results {
		if journey.Failed {
			unmeasured++
		}
	}

	result := Result{
		Code:        CodeOK,
		Journeys:    r.results,
		Steps:       steps,
		Blocks:      len(r.blocks),
		InBand:      counts[benchcore.InBand],
		OutOfBand:   counts[benchcore.OutOfBand],
		Templated:   counts[benchcore.Templated],
		Interactive: counts[benchcore.Interactive],
		DeadEnd:     counts[benchcore.DeadEnd],
		DeadEnds:    deadEnds,
	}
	if result.DeadEnd > 0 || unmeasured > 0 {
		result.Code = CodeFailed
	}

	lines = append(lines, fmt.Sprintf("bench: %s, %s, %s",
		plural(len(r.results), "journey", "journeys"), plural(result.Steps, "step", "steps"), plural(result.Blocks, "block", "blocks")))
	lines = append(lines, fmt.Sprintf("in_band %d · out_of_band %d · templated %d · interactive %d · dead_end %d",
		result.InBand, result.OutOfBand, result.Templated, result.Interactive, result.DeadEnd))

	for _, journey := range r.results {
		line := fmt.Sprintf("%s · %s · %s",
			journey.Name, plural(journey.Steps, "step", "steps"), blocksLine(journey.Blocks))
		if len(journey.Surprises) > 0 {
			line += " · " + strings.Join(journey.Surprises, "; ")
		}
		lines = append(lines, line)
	}

	if len(deadEnds) > 0 {
		lines = append(lines, "dead ends:")
		for _, dead := range deadEnds {
			lines = append(lines, fmt.Sprintf("- %s · %s exited %d: %s", dead.Journey, dead.Command, dead.Exit, dead.Message))
		}
	}

	if len(surprises) > 0 {
		lines = append(lines, fmt.Sprintf("surprises: %d steps did not exit the way the journey expected", len(surprises)))
	}

	result.Message = verdict(result.DeadEnd, unmeasured)
	lines = append(lines, result.Message)
	result.Lines = lines
	return result
}

// verdict is the one line that says whether the claim held. A journey that could
// not run is its own verdict: a bench that measured nothing may not report
// success.
func verdict(deadEnds, unmeasured int) string {
	switch {
	case unmeasured > 0:
		return fmt.Sprintf("could not measure %s: check the output of each journey", plural(unmeasured, "journey", "journeys"))
	case deadEnds > 0:
		return plural(deadEnds, "dead end", "dead ends") + ": each one sends you to a place with no way out"
	default:
		return "no dead ends: every block names a command that resolves it"
	}
}

// blocksLine is how a journey's block count reads.
func blocksLine(blocks int) string {
	if blocks == 0 {
		return "no blocks"
	}
	return plural(blocks, "block", "blocks")
}

// failed is a bench that could not run at all.
func failed(message string) Result {
	return Result{Code: CodeFailed, Message: message, Lines: []string{message}}
}

// names lists the corpus by name.
func names(journeys []benchcore.Journey) string {
	out := make([]string, 0, len(journeys))
	for _, journey := range journeys {
		out = append(out, journey.Name)
	}
	return strings.Join(out, ", ")
}

// quoted wraps the unknown names so a typo reads as a typo.
func quoted(items []string) string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, "«"+item+"»")
	}
	return strings.Join(out, ", ")
}

// plural is how an English sentence counts things.
func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return fmt.Sprintf("%d %s", n, many)
}

// firstLine is the part of a message that says what happened. A report line is one
// line: the rest of a multi-line message is still in the tool's own output.
func firstLine(text string) string {
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if line = oneLine(line); line != "" {
			return line
		}
	}
	return "no output"
}

// oneLine keeps every message on one line, whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
