// Package summary implements `odachi summary`: the derived report of every
// feature — its slug, its first requirement verbatim, the steps, the open
// questions and the tail of its history — followed by one line of totals.
//
// It reports what is on disk and narrates nothing: the teaching about the week
// is the agent's job (T8), not the binary's.
package summary

import (
	"fmt"
	"strings"

	"github.com/biggs-100/odachi/internal/app/tasks"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// Exit codes (Contract IV).
const (
	// CodeOK means the report was printed.
	CodeOK = 0
	// CodeBlocked means the project could not be read.
	CodeBlocked = 1
)

// historyWindow is how many history lines per feature the report shows: enough
// to remember where it stopped, short enough to read in one screen.
const historyWindow = 3

// firstRequirement is the requirement the report quotes: the user's own words.
const firstRequirement = "R1"

// Options is one `odachi summary` invocation.
type Options struct {
	Dir string
}

// Files is the disk port: the report only reads.
type Files interface {
	List(dir string) ([]string, error)
	Read(path string) ([]byte, error)
}

// Result is what the CLI prints and exits with. It has the same shape as
// start.Result; Lines carries the block a human reads and Message the one-line
// headline.
type Result struct {
	Code    int
	Message string
	Lines   []string
	Task    string
	Files   []string
}

// Run performs one `odachi summary`.
func Run(opts Options, files Files) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}

	docs, err := tasks.All(files, opts.Dir)
	if err != nil {
		return Result{Code: CodeBlocked, Message: oneLine(err.Error())}
	}
	if len(docs) == 0 {
		// A project with nothing built yet is a fact, not a failure.
		return Result{Code: CodeOK, Message: "todavía no hay features: corré odachi start para crear la primera"}
	}

	lines := make([]string, 0, len(docs)*4+2)
	var totalDone, totalSteps, totalOpen int

	for i, doc := range docs {
		if i > 0 {
			lines = append(lines, "")
		}

		done, steps := doc.Doc.Progress()
		open := len(doc.Doc.OpenQuestions())
		totalDone += done
		totalSteps += steps
		totalOpen += open

		lines = append(lines, fmt.Sprintf("%04d-%s · pasos: %d de %d · %s",
			doc.Number, doc.Slug, done, steps, plural(open, "pregunta abierta", "preguntas abiertas")))
		lines = append(lines, "  "+requirementLine(doc.Doc))
		lines = append(lines, "  historial:")

		history := tail(doc.Doc.History, historyWindow)
		if len(history) == 0 {
			lines = append(lines, "    (todavía no hay líneas)")
		}
		for _, line := range history {
			lines = append(lines, "    "+line)
		}
	}

	totals := fmt.Sprintf("total: %s · %d pasos hechos de %d · %s",
		plural(len(docs), "feature", "features"), totalDone, totalSteps,
		plural(totalOpen, "pregunta abierta", "preguntas abiertas"))
	lines = append(lines, "", totals)

	return Result{Code: CodeOK, Message: totals, Lines: lines}
}

// requirementLine quotes R1 verbatim: it is the user's own sentence, and the
// report must not paraphrase it.
func requirementLine(doc taskdoc.TaskDoc) string {
	for _, requirement := range doc.Requirements {
		if requirement.ID == firstRequirement {
			return firstRequirement + ` · "` + requirement.Text + `"`
		}
	}
	if len(doc.Requirements) > 0 {
		first := doc.Requirements[0]
		return first.ID + ` · "` + first.Text + `"`
	}
	return firstRequirement + " · (todavía no hay requisito)"
}

// tail returns the last n lines of the history.
func tail(lines []string, n int) []string {
	if len(lines) <= n {
		return lines
	}
	return lines[len(lines)-n:]
}

// plural renders "1 pregunta abierta" and "2 preguntas abiertas".
func plural(n int, singular, many string) string {
	if n == 1 {
		return "1 " + singular
	}
	return fmt.Sprintf("%d %s", n, many)
}

// oneLine keeps every message to a single line, whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
