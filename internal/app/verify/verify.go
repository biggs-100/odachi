// Package verify runs the independent verifier of one tree and stores the record
// it answers.
//
// The verb is the producer of the record the gate asks for. It composes the
// question -the requirement in the human's words, what the feature promised, and
// the risk the commits declared- hands it to one command, and seals what came back
// under the tree the tool read. It never decides whether the tree is good: the
// verifier answers, and the tool only refuses an answer nobody signed.
//
// Like `review`, the runner underneath is one command with the prompt on stdin and
// `{prompt}` as a path, so a caller with an agent CLI and a caller with a script
// use the same door. Without a command the verb prints the prompt and runs nobody:
// the question a human pays a model to answer is worth reading first.
package verify

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/biggs-100/odachi/internal/domain/verify"
)

// The codes a caller returns. They match the rest of the verbs.
const (
	CodeOK      = 0
	CodeBlocked = 1
)

// DefaultTimeout is what one verifier may take. It matches `review --exec`: a
// verifier that runs a suite is slow on purpose, and a timeout shorter than the
// work would report a slow machine as a failure of the tree.
const DefaultTimeout = 1800

// Files is the disk this verb needs: the store is written, never listed.
type Files interface {
	Exists(path string) (bool, error)
	Read(path string) ([]byte, error)
	Write(path string, data []byte) error
}

// Git answers the one fact the verb takes from the repository: the tree this
// worktree is. It is the same read `check` and `review` make, so the record names
// the bytes the gate will name.
type Git interface {
	TreeHash(root string) (string, error)
}

// Stop names the clock that ended a run. The verb needs its own word for it:
// this package decides what a human reads and does not import the shell. Silence
// and the total deadline are different facts, and telling them apart is what
// keeps the next step honest.
type Stop int

const (
	// StopNone is a run that ended on its own: its exit code is the result.
	StopNone Stop = iota
	// StopDeadline is the total deadline expiring.
	StopDeadline
	// StopSilence is the silence clock firing before the total deadline.
	StopSilence
)

// Shell runs the verifier: one command, the prompt on stdin, and `{prompt}`
// replaced by a path to the same bytes. It names the clock that stopped the run,
// because the two budgets ask the human for different things.
type Shell interface {
	RunInputStopped(dir, command, input string, timeoutSeconds int, firstOutput, idle time.Duration) (stdout, stderr string, exitCode int, stop Stop, err error)
}

// Options is what the caller knows: who to run, what the feature promised, and the
// risks the commits declared.
type Options struct {
	// Dir is the worktree the verifier reads.
	Dir string
	// Store is the clone's state, where the record lives.
	Store string
	// Exec is the command that answers. Empty prints the prompt and runs nobody.
	Exec string
	// Requirement is R1 in the human's words, quoted and not rewritten.
	Requirement string
	// DoneWhen is what the feature promised, in the order the task document keeps.
	DoneWhen []string
	// Risks are the names the commits declared. Empty means the commits said
	// `none`, and the prompt says so.
	Risks []string
	// TimeoutSeconds bounds the command.
	TimeoutSeconds int
	// FirstOutput is how long the command may stay silent before its first byte.
	FirstOutput time.Duration
	// Idle is how long it may go quiet between bytes.
	Idle time.Duration
	// Now is the clock the record is sealed with.
	Now time.Time
}

// Result is what the human reads, and the code the shell returns.
type Result struct {
	Code    int
	Message string
	Lines   []string
}

// Run composes the question, runs it when there is a command, and stores the
// answer sealed under the tree.
func Run(opts Options, files Files, git Git, shell Shell) Result {
	if strings.TrimSpace(opts.Dir) == "" {
		opts.Dir = "."
	}
	if opts.TimeoutSeconds <= 0 {
		opts.TimeoutSeconds = DefaultTimeout
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}
	// The record lives in the clone's store, and a folder with no repository has
	// nowhere to put it: writing beside the worktree would put evidence inside the
	// tree it certifies.
	if strings.TrimSpace(opts.Store) == "" {
		return blocked("this folder is not a git repository, and a verification lives in the clone's store; run odachi verify inside the clone")
	}

	tree, err := git.TreeHash(opts.Dir)
	if err != nil {
		return blocked(oneLine(err.Error()) + "; check the git tree and run odachi verify again")
	}
	rel, err := verify.Path(tree)
	if err != nil {
		return blocked(oneLine(err.Error()) + "; check the git tree and run odachi verify again")
	}

	prompt := Prompt(opts, tree)
	if strings.TrimSpace(opts.Exec) == "" {
		lines := append([]string{"the prompt the verifier will read:"}, strings.Split(prompt, "\n")...)
		lines = append(lines, `next: odachi verify --exec "<command>"`)
		return Result{Code: CodeOK, Message: "printed the prompt and ran nobody: add --exec to verify", Lines: lines}
	}

	stdout, stderr, exit, stop, err := shell.RunInputStopped(opts.Dir, opts.Exec, prompt, opts.TimeoutSeconds, opts.FirstOutput, opts.Idle)
	if err != nil {
		return blocked("the verifier could not run: " + oneLine(err.Error()))
	}
	if stop == StopSilence {
		// The silence clock fires long before the total deadline, so a run it
		// stopped is not a run that needed more time: sending the human to
		// --timeout would name a flag that was not the cause.
		message := fmt.Sprintf("the verifier went silent and the silence clock stopped it; it was allowed %s, and the %d-second total deadline was not the reason",
			silenceBudget(opts), opts.TimeoutSeconds)
		return stopped(message, "next: run the verifier by hand with the prompt on stdin and keep what it says")
	}
	if stop != StopNone {
		message := fmt.Sprintf("the verifier did not answer in %d seconds, the total deadline; the run was stopped and nothing was verified", opts.TimeoutSeconds)
		return stopped(message, fmt.Sprintf(`next: odachi verify --exec "<command>" --timeout %d`, opts.TimeoutSeconds*2))
	}
	if exit != 0 {
		detail := firstLine(stderr)
		if detail == "" {
			detail = firstLine(stdout)
		}
		message := fmt.Sprintf("the verifier exited %d", exit)
		if detail != "" {
			message += ": " + detail
		}
		return blocked(message + "; a verifier that did not run verified nothing")
	}

	record, err := verify.Parse([]byte(stdout))
	if err != nil {
		return refused("the verifier's answer cannot be read: " + oneLine(err.Error()))
	}
	// The tree is the tool's fact, not the verifier's: an answer that names another
	// tree speaks for bytes nobody read, and one that names none is stamped here.
	if named := strings.TrimSpace(record.Tree); named != "" && !record.Covers(tree) {
		return refused(fmt.Sprintf("the answer names the tree %s, and this worktree is %s: a record for another tree covers nothing", named, tree))
	}
	record.Tree = tree
	record = record.Seal(opts.Now)
	if err := record.Validate(); err != nil {
		return refused("the verifier's answer is not a verification: " + oneLine(err.Error()))
	}

	path := filepath.Join(opts.Store, filepath.FromSlash(rel))
	if err := files.Write(path, verify.Render(record)); err != nil {
		return blocked("the verification could not be stored: " + oneLine(err.Error()))
	}

	message := fmt.Sprintf("the verifier answered %s: %s", record.Verdict, rel)
	lines := []string{message, "by: " + record.By, "at: " + record.At}
	for _, note := range record.Notes {
		lines = append(lines, "note: "+note)
	}
	lines = append(lines, `next: odachi check --run "<command>"`)
	return Result{Code: CodeOK, Message: message, Lines: lines}
}

// Prompt is the question the verifier reads: the tree, the requirement in the
// human's words, what the feature promised, the risk that asked for it, and the
// one JSON object it answers with.
func Prompt(opts Options, tree string) string {
	var b strings.Builder
	b.WriteString("You are odachi-verify, the independent verifier of one tree. You did not write it, and you do not fix it: you look for the requirement failing and you say so.\n\n")
	b.WriteString("Tree: " + tree + "\n\n")
	if requirement := strings.TrimSpace(opts.Requirement); requirement != "" {
		b.WriteString("The requirement, in the human's words:\n" + requirement + "\n\n")
	}
	if len(opts.DoneWhen) > 0 {
		b.WriteString("The feature promised:\n")
		for _, item := range opts.DoneWhen {
			b.WriteString("- " + item + "\n")
		}
		b.WriteString("\n")
	}
	risks := "none: the commits declared no risk, and nothing asks for a verification"
	if len(opts.Risks) > 0 {
		risks = strings.Join(opts.Risks, ", ")
	}
	b.WriteString("The commits declared this risk: " + risks + "\n\n")
	b.WriteString("Read the worktree and derive the probes that would catch the requirement unmet: the command that exposes it, the file that contradicts it, the edge the code claims to handle. Run the ones you can, read the rest, and do not edit, commit or delegate.\n\n")
	// The whole suite belongs to the gate, and a verifier that runs it competes
	// with the machine for twelve minutes to answer a question one package would
	// answer (the `tests` skill says the same: focused commands). The first field
	// run spent its whole timeout there and answered nothing.
	b.WriteString("Run focused probes, not the whole suite: the gate runs `go test ./...` and pays for it. One package, one test, one file, one command at a time, and a probe you cannot run is `unverified`, not a guess.\n\n")
	b.WriteString("Answer with one JSON object and nothing else:\n")
	b.WriteString(`{"schema":1,"by":"<the model or the command that verified>","verdict":"verified|refuted","notes":["<what you ran and what it said>"]}` + "\n\n")
	b.WriteString(`"verified" means the requirement holds in these bytes. "refuted" means it does not, and the notes name the probe that showed it. The tool stamps the tree and the hour; do not write them.` + "\n")
	return b.String()
}

// blocked is a run that did not reach a verification, or could not store one.
func blocked(message string) Result {
	return Result{Code: CodeBlocked, Message: message, Lines: []string{message}}
}

// stopped is a run one of the clocks cut, with the line that says what to do
// next: the message names the clock, and the second line names the step.
func stopped(message, next string) Result {
	return Result{Code: CodeBlocked, Message: message, Lines: []string{message, next}}
}

// silenceBudget says how long the run was allowed to stay quiet, in the numbers
// the clock was given: one when both budgets are the same, both when the run was
// given different ones, so the message never names a number the clock did not use.
func silenceBudget(opts Options) string {
	first := int(opts.FirstOutput / time.Second)
	idle := int(opts.Idle / time.Second)
	if first == idle {
		return fmt.Sprintf("%d seconds without output", first)
	}
	return fmt.Sprintf("%d seconds without its first output and %d seconds between outputs", first, idle)
}

// refused is an answer that arrived and is not a verification.
func refused(message string) Result {
	return Result{
		Code:    CodeBlocked,
		Message: message,
		Lines:   []string{message, `answer the four keys: {"schema":1,"by":"<who>","verdict":"verified|refuted","notes":["<what you ran>"]}`},
	}
}

// oneLine keeps a reason inside the sentence that carries it.
func oneLine(text string) string { return strings.Join(strings.Fields(text), " ") }

// firstLine is the part of a stream a one-line refusal can carry.
func firstLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			return oneLine(trimmed)
		}
	}
	return ""
}
