// Package skilleval runs one skill's exam: the same prompt twice, once with the
// skill available and once without it, and the score the event stream decides.
//
// It is the second instrument after the lens probe, and it is built the same way:
// nothing here calls a model by itself. Without --exec the verb prints the prompt
// and runs nobody; with it, the command the human wrote is what runs, and the
// exam only reads what it wrote.
package skilleval

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/biggs-100/odachi/internal/domain/oneline"
	domain "github.com/biggs-100/odachi/internal/domain/skilleval"
	corpus "github.com/biggs-100/odachi/internal/skilleval"
)

// Codes answer what the verb did: nothing was measured, or it was.
const (
	CodeOK      = 0
	CodeBlocked = 1
)

// The two arms: the one that carries the skill, and the one that does not. They
// are the whole measurement, and their names are what the record says.
const (
	With    = "with"
	Without = "without"
)

// Options is one exam run.
type Options struct {
	// Scratch is a folder this run owns; every arm and draw is built under it.
	Scratch string
	// Skill is the exam and the skill it measures, loaded by the caller.
	Exam  corpus.Exam
	Skill []corpus.File
	// Exec is the command that runs an agent. {prompt} is replaced by the path of
	// the prompt file; a command without it receives the prompt on stdin.
	Exec string
	// Runs is how many draws each arm answers.
	Runs int
	// Timeout is how long one draw may take, in seconds.
	Timeout int
	// MoveHome is the human's permission to take the skill out of the home for
	// the length of the run. Without it, a skill installed there is a refusal.
	MoveHome bool
	// Recover puts back whatever a dead run left out, and measures nothing.
	Recover bool
	// Record is where this run files what it measured, when the caller has a
	// clone to file it in. Empty means the run prints and files nothing.
	Record string
	// Identity names the bytes being measured: the skill's text and the exam.
	Identity string
	// Again is the human's permission to pay for a measurement that is already
	// in the store. Without it, a record for the same bytes is a refusal.
	Again bool
	// Now is the instant the record carries; a test sets it, a run does not.
	Now time.Time
}

// Files is the disk this run writes the projects to, and the clone's records.
type Files interface {
	Write(path string, data []byte) error
	MkdirAll(path string) error
	Exists(path string) (bool, error)
}

// Shell runs the human's command.
type Shell interface {
	RunInput(dir, command, input string, timeoutSeconds int) (stdout, stderr string, exitCode int, timedOut bool, err error)
	// Quote is how a path reaches the command line: the prompt lives somewhere with
	// spaces in it on every OS this ships to.
	Quote(path string) string
}

// Home is the skills folder the client reads, and the journal that remembers what
// left it. The exam asks; the adapter moves.
type Home interface {
	Installed(name string) (path string, ok bool, err error)
	Take(name string) (domain.Mark, error)
	Restore(mark domain.Mark) error
	Recover() ([]string, error)
}

// Result is what the verb prints and what the caller files.
type Result struct {
	Code    int
	Message string
	Lines   []string
	Draws   []domain.Draw
}

// Run works the exam and answers what each arm showed. A run that cannot keep its
// promise leaves nothing half done: the skill it took out goes back before the
// answer, whatever happened.
func Run(opts Options, files Files, shell Shell, home Home) (result Result) {
	// Whatever a dead run left out comes home before anything is measured: a
	// journal with marks is a skill that never came back, and measuring on top of
	// it would refuse the run with a folder that only looks occupied.
	back, err := home.Recover()
	if err != nil {
		return blocked("could not put back what an earlier run left out: " + err.Error())
	}
	if opts.Recover {
		if len(back) == 0 {
			return ok([]string{"nothing was left out: the journal is empty"})
		}
		return ok([]string{"put back " + strings.Join(back, ", ")})
	}

	if strings.TrimSpace(opts.Exec) == "" {
		return ok([]string{
			"the exam for " + opts.Exam.Skill + " asks for this, twice -once with the skill and once without it:",
			"",
			opts.Exam.Task,
			"next: odachi skills eval " + opts.Exam.Skill + " --exec \"<the command that runs your agent>\" --runs 3",
		})
	}

	runs := opts.Runs
	if runs < 1 {
		runs = 1
	}

	// What was already paid for is not paid again: the record is the receipt, and
	// a skill that changed answers a different identity, so only the same bytes
	// over the same exam meet this refusal.
	if opts.Record != "" && !opts.Again {
		paid, err := files.Exists(opts.Record)
		if err != nil {
			return blocked("could not look at the store for a record: " + err.Error())
		}
		if paid {
			return blocked(fmt.Sprintf(
				"%s was measured already: %s; pass --again to pay for a second run, or change the skill and measure what changed",
				opts.Exam.Skill, opts.Identity))
		}
	}

	path, installed, err := home.Installed(opts.Exam.Skill)
	if err != nil {
		return blocked(err.Error())
	}
	var taken domain.Mark
	if installed {
		if !opts.MoveHome {
			return blocked(fmt.Sprintf(
				"%s is installed in the home, and the arm without the skill is that folder absent: pass --move-home to take it out for the run and put it back, or remove it yourself first",
				path))
		}
		taken, err = home.Take(opts.Exam.Skill)
		if err != nil {
			return blocked("could not take the skill out of the home: " + err.Error())
		}
		// The permission covers one run, and the skill coming home is part of the
		// answer: a restore that failed leaves the machine missing a skill, and no
		// green line may say the run was clean.
		defer func() {
			if err := home.Restore(taken); err != nil {
				result = blocked("the exam ran and the skill did not come home: " + err.Error())
			}
		}()
	}

	arms := []string{With, Without}
	if !installed && len(opts.Skill) == 0 {
		return blocked("the skill " + opts.Exam.Skill + " is neither installed in the home nor bundled with this binary: there is nothing to measure")
	}

	var draws []domain.Draw
	lines := []string{}
	for _, arm := range arms {
		for n := 1; n <= runs; n++ {
			draw := drawOnce(opts, files, shell, arm, n)
			draws = append(draws, draw)
			lines = append(lines, describe(draw))
		}
	}

	for _, arm := range arms {
		met, of := 0, 0
		for _, draw := range draws {
			if draw.Arm != arm {
				continue
			}
			met += domain.Met(draw.Behaviours)
			of += len(draw.Behaviours)
		}
		lines = append(lines, fmt.Sprintf("%s: %d of %d behaviours across %d draw(s)", arm, met, of, runs))
	}

	// A run whose draws all failed measured nothing: it files no record, because a
	// refusal that leaves a receipt would keep the next run from trying again.
	failed := 0
	for _, draw := range draws {
		if draw.Problem != "" {
			failed++
		}
	}
	if failed == len(draws) {
		return blocked("no draw answered: " + draws[0].Problem + "; nothing was recorded")
	}

	// The record is the whole point of the run: a score nobody can compare is a
	// sentence, and a skill that gets worse looks exactly like one that never
	// moved.
	if opts.Record != "" {
		record := domain.Record{
			Schema:   domain.RecordSchema,
			Skill:    opts.Exam.Skill,
			Identity: opts.Identity,
			At:       opts.Now.UTC().Format(time.RFC3339),
			Draws:    draws,
		}
		data, err := domain.RenderRecord(record)
		if err != nil {
			return blocked("the draws ran but the record cannot be written: " + err.Error())
		}
		if err := files.Write(opts.Record, data); err != nil {
			return blocked("the draws ran but the record cannot be written: " + err.Error())
		}
		lines = append(lines, "record: "+record.Summarize())
	}

	return Result{Code: CodeOK, Message: fmt.Sprintf("the exam for %s ran %d draw(s) per arm", opts.Exam.Skill, runs), Lines: lines, Draws: draws}
}

// drawOnce builds one arm, runs the command in it, and scores the stream.
func drawOnce(opts Options, files Files, shell Shell, arm string, n int) domain.Draw {
	dir := filepath.Join(opts.Scratch, arm, fmt.Sprintf("r%d", n))
	draw := domain.Draw{Arm: arm, Run: n}

	if err := build(opts, files, dir, arm); err != nil {
		draw.Problem = "could not build the arm: " + err.Error()
		return draw
	}

	promptPath := filepath.Join(dir, "prompt.md")
	if err := files.Write(promptPath, []byte(opts.Exam.Task)); err != nil {
		draw.Problem = "could not write the prompt: " + err.Error()
		return draw
	}

	command := strings.ReplaceAll(opts.Exec, "{prompt}", shell.Quote(promptPath))
	stdout, stderr, exitCode, timedOut, err := shell.RunInput(dir, command, opts.Exam.Task, opts.Timeout)
	if err != nil {
		draw.Problem = "the command did not run: " + err.Error()
		return draw
	}
	if timedOut {
		draw.Problem = fmt.Sprintf("the command was still running at %ds; its answer was thrown away", opts.Timeout)
		return draw
	}
	if exitCode != 0 {
		draw.Problem = fmt.Sprintf("the command answered %d: %s", exitCode, oneLine(stderr))
		return draw
	}

	events, err := domain.ParseStream([]byte(stdout))
	if err != nil {
		draw.Problem = "the stream cannot be read: " + err.Error()
		return draw
	}
	draw.Behaviours = domain.Score(events)
	for _, event := range events {
		if event.Tool == "skill" && event.Name == opts.Exam.Skill {
			draw.Loaded = true
		}
	}
	return draw
}

// build writes the exam's project, and -in the arm that has it- the skill, where
// the client discovers it: the project's own .agents/skills.
func build(opts Options, files Files, dir, arm string) error {
	if err := files.MkdirAll(dir); err != nil {
		return err
	}
	names := make([]string, 0, len(opts.Exam.Files))
	for name := range opts.Exam.Files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := files.MkdirAll(filepath.Dir(path)); err != nil {
			return err
		}
		if err := files.Write(path, opts.Exam.Files[name]); err != nil {
			return err
		}
	}
	if arm != With {
		return nil
	}
	for _, file := range opts.Skill {
		path := filepath.Join(dir, ".agents", "skills", opts.Exam.Skill, filepath.FromSlash(file.Path))
		if err := files.MkdirAll(filepath.Dir(path)); err != nil {
			return err
		}
		if err := files.Write(path, file.Content); err != nil {
			return err
		}
	}
	return nil
}

// describe is one line of the table: the arm, the run, the score, and the event
// that decided each behaviour, so a reader can check the answer without trusting
// the number.
func describe(draw domain.Draw) string {
	if draw.Problem != "" {
		return fmt.Sprintf("%-7s r%d  -    %s", draw.Arm, draw.Run, draw.Problem)
	}
	loaded := "the skill did not load"
	if draw.Loaded {
		loaded = "the skill loaded"
	}
	parts := make([]string, 0, len(draw.Behaviours))
	for _, behaviour := range draw.Behaviours {
		answer := "no"
		if behaviour.Met {
			answer = "yes"
		}
		parts = append(parts, behaviour.ID+": "+answer+" ("+behaviour.Why+")")
	}
	return fmt.Sprintf("%-7s r%d  %d/%d  %s · %s",
		draw.Arm, draw.Run, domain.Met(draw.Behaviours), len(draw.Behaviours),
		strings.Join(parts, " · "), loaded)
}

func ok(lines []string) Result {
	return Result{Code: CodeOK, Lines: lines}
}

func blocked(message string) Result {
	return Result{Code: CodeBlocked, Message: message, Lines: []string{message}}
}

func oneLine(s string) string {
	return oneline.Line(s)
}
