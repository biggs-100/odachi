// Package role implements `odachi role`: one source per professional, and the
// command that renders the agent and the skill from it. The parsing and the
// rules live in internal/domain/role; this package only orchestrates and
// phrases what a human reads.
package role

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/layout"
	"github.com/biggs-100/odachi/internal/domain/role"
)

// Exit codes: the same two every verb uses.
const (
	// CodeOK means the command did its job.
	CodeOK = 0
	// CodeBlocked means a violation, an existing slug or an unreadable file.
	CodeBlocked = 1
)

// The layout this command owns, relative to the project root.
const (
	// Dir is where the sources live: one file per role, the only thing a human
	// edits.
	Dir = layout.Roles
	// FileExt is the suffix of a source file.
	FileExt = ".md"
)

// Files is the disk port: enough to write a source and, when rendering, to read
// every source.
type Files interface {
	Exists(path string) (bool, error)
	Read(path string) ([]byte, error)
	Write(path string, data []byte) error
	List(dir string) ([]string, error)
}

// Result is what the CLI prints and exits with.
type Result struct {
	Code    int
	Message string
	Lines   []string
	Files   []string
}

// New writes the skeleton of one role: the eight sections with the question each
// one answers. It validates the slug before it touches anything, refuses an
// existing role, and reads the file back to prove it parses.
func New(dir, slug string, files Files) Result {
	if !role.ValidSlug(slug) {
		return blocked(fmt.Sprintf(
			"the slug %q does not work: it must match %s (lowercase letters, numbers and dashes, starting with a letter); choose one and run odachi role new <slug> again",
			slug, role.NamePattern))
	}

	rel := Dir + "/" + slug + FileExt
	path := filepath.Join(rootOf(dir), Dir, slug+FileExt)

	exists, err := files.Exists(path)
	if err != nil {
		return blocked("could not check " + rel + ": " + oneLine(err.Error()) + "; check the permissions and run odachi role new <slug> again")
	}
	if exists {
		return blocked("already exists " + rel + ": not touching it; choose another slug, or edit the one that is there and run odachi role " + slug)
	}

	text := role.Template(slug)
	if err := files.Write(path, []byte(text)); err != nil {
		return blocked("could not write " + rel + ": " + oneLine(err.Error()) + "; check the permissions and run odachi role new <slug> again")
	}

	// The skeleton is the first thing a human edits, so it has to be valid
	// before they start: a template that fails the check teaches the wrong
	// lesson on the first read.
	written := role.Parse(rel, text)
	if problems := written.Problems(slug); len(problems) > 0 {
		lines := make([]string, 0, len(problems))
		for _, problem := range problems {
			lines = append(lines, problem.File+": "+problem.Reason)
		}
		return Result{Code: CodeBlocked, Message: "the skeleton does not pass its own check", Lines: lines, Files: []string{rel}}
	}

	return Result{
		Code:    CodeOK,
		Message: "wrote " + rel + "; fill the eight sections, then run odachi role " + slug,
		Files:   []string{rel},
	}
}

// rootOf is the project root the caller resolved.
func rootOf(dir string) string {
	if dir == "" {
		return "."
	}
	return dir
}

// blocked is a Result with exit code 1.
func blocked(message string) Result {
	return Result{Code: CodeBlocked, Message: message}
}

// oneLine folds a message onto one line so a Result never prints a stack of
// paths the human cannot read.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
