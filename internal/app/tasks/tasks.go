// Package tasks finds and reads the task documents of a project: one file per
// feature under odd/tasks, named NNNN-<slug>.md.
//
// The verbs `ask`, `level` and `summary` all need the same answer to "which
// feature are we talking about?", so the answer lives here once instead of
// three times. It is pure orchestration over a small disk port: no os, no
// adapters.
package tasks

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// Dir is where task documents live, relative to the project root.
const Dir = "odd/tasks"

// Document is one task document on disk: the file, its number, its slug, the
// bytes as written and the parsed content.
type Document struct {
	Path   string
	Rel    string
	Number int
	Slug   string
	Raw    string
	Doc    taskdoc.TaskDoc
}

// Files is the disk port this package needs: enough to find and read a task
// document. Every Files adapter that can read a project satisfies it.
type Files interface {
	List(dir string) ([]string, error)
	Read(path string) ([]byte, error)
}

// All returns every task document, sorted by number, so callers can report them
// in the order the human opened them.
func All(files Files, dir string) ([]Document, error) {
	names, err := files.List(filepath.Join(dir, Dir))
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", Dir, err)
	}

	docs := make([]Document, 0, len(names))
	for _, name := range names {
		number, slug, ok := parseName(name)
		if !ok {
			continue
		}

		rel := Dir + "/" + name
		path := filepath.Join(dir, Dir, name)
		data, err := files.Read(path)
		if err != nil {
			return nil, fmt.Errorf("could not read %s: %w", rel, err)
		}

		raw := string(data)
		docs = append(docs, Document{
			Path:   path,
			Rel:    rel,
			Number: number,
			Slug:   slug,
			Raw:    raw,
			Doc:    taskdoc.Parse(raw),
		})
	}

	slices.SortFunc(docs, func(a, b Document) int {
		if a.Number != b.Number {
			return a.Number - b.Number
		}
		return strings.Compare(a.Slug, b.Slug)
	})
	return docs, nil
}

// Skipped returns the documents in the task folder that odachi does not read:
// the ones whose name is not NNNN-<slug>.md. A project that keeps another shape
// - ODD names a feature without a number - hears the truth here instead of
// wondering why summary counts one.
func Skipped(files Files, dir string) ([]string, error) {
	names, err := files.List(filepath.Join(dir, Dir))
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", Dir, err)
	}

	var skipped []string
	for _, name := range names {
		if !strings.HasSuffix(strings.ToLower(name), ".md") {
			continue
		}
		if _, _, ok := parseName(name); !ok {
			skipped = append(skipped, name)
		}
	}
	slices.Sort(skipped)
	return skipped, nil
}

// Planned is the documents odachi cannot read that carry the shape of a plan: the
// ones a human meant as a feature and the tool will ignore. A document with none of
// a plan's headers is prose -it was never meant to be read- and naming it is enough.
//
// The distinction matters because only the first is a loss: a misnamed feature is
// invisible to `check`, and a note in the folder of the plan is just in the wrong
// folder, which is the human's to tidy.
func Planned(files Files, dir string) ([]string, error) {
	skipped, err := Skipped(files, dir)
	if err != nil {
		return nil, err
	}

	planned := make([]string, 0, len(skipped))
	for _, name := range skipped {
		data, err := files.Read(filepath.Join(dir, Dir, name))
		if err != nil {
			return nil, fmt.Errorf("could not read %s/%s: %w", Dir, name, err)
		}
		if LooksLikeAPlan(string(data)) {
			planned = append(planned, name)
		}
	}
	return planned, nil
}

// LooksLikeAPlan says this text carries the shape a task document has: a base
// commit, a requirement, a step or a history line. A name odachi cannot read plus
// one of those is a feature that would be silently ignored; a name it cannot read
// and none of them is prose.
func LooksLikeAPlan(text string) bool {
	doc := taskdoc.Parse(text)
	return doc.Base != "" || len(doc.Requirements) > 0 || len(doc.Steps) > 0 || len(doc.History) > 0
}

// SkippedLine is the one sentence that names the documents odachi does not read,
// or "" when the folder holds none: the tool says what it skips instead of
// letting a project wonder why summary counts one.
func SkippedLine(skipped []string) string {
	if len(skipped) == 0 {
		return ""
	}
	return fmt.Sprintf("%d document(s) in %s are not features odachi reads: %s (the name is not NNNN-slug.md)",
		len(skipped), Dir, fewNames(skipped))
}

// fewNames lists the first three names and counts the rest, so a folder with
// forty foreign documents still prints one line.
func fewNames(names []string) string {
	const limit = 3
	if len(names) <= limit {
		return strings.Join(names, ", ")
	}
	return strings.Join(names[:limit], ", ") + fmt.Sprintf(" and %d more", len(names)-limit)
}

// Current returns the highest-numbered task document: the feature the human is
// working on now. A project with no feature yet is an error a human can read,
// and it names the command that creates the first one.
func Current(files Files, dir string) (Document, error) {
	docs, err := All(files, dir)
	if err != nil {
		return Document{}, err
	}
	if len(docs) == 0 {
		return Document{}, fmt.Errorf("there is no task doc in %s yet; run odachi start to create the first one", Dir)
	}
	return docs[len(docs)-1], nil
}

// Feature returns the feature with this number, or the current one when the
// number is zero. A number the project does not have is an error that names the
// features it does: a selector that answers nothing leaves the human guessing
// which numbers exist, and the highest is not an answer to "give me 0001".
func Feature(files Files, dir string, number int) (Document, error) {
	if number == 0 {
		return Current(files, dir)
	}

	docs, err := All(files, dir)
	if err != nil {
		return Document{}, err
	}
	for _, doc := range docs {
		if doc.Number == number {
			return doc, nil
		}
	}
	if len(docs) == 0 {
		return Document{}, fmt.Errorf("there is no task doc in %s yet; run odachi start to create the first one", Dir)
	}

	return Document{}, fmt.Errorf("%s", NotFound(docs, number))
}

// NotFound is the sentence for a number the project does not have: a selector
// that answers nothing leaves the human guessing which numbers exist, and the
// highest is not an answer to "give me 0001".
func NotFound(docs []Document, number int) string {
	names := make([]string, 0, len(docs))
	for _, doc := range docs {
		names = append(names, fmt.Sprintf("%04d-%s", doc.Number, doc.Slug))
	}
	return fmt.Sprintf("there is no feature %04d in %s; the ones odachi reads are %s", number, Dir, fewNames(names))
}

// parseName reads NNNN-<slug>.md. Anything else in the directory — a backup, a
// note, a document the human named by hand — is housekeeping, not a feature.
func parseName(name string) (number int, slug string, ok bool) {
	if !strings.HasSuffix(name, ".md") {
		return 0, "", false
	}
	base := strings.TrimSuffix(name, ".md")

	prefix, rest, found := strings.Cut(base, "-")
	if !found || len(prefix) != 4 {
		return 0, "", false
	}
	number, err := strconv.Atoi(prefix)
	if err != nil || number <= 0 {
		return 0, "", false
	}
	slug = strings.TrimSpace(rest)
	if slug == "" {
		return 0, "", false
	}
	return number, slug, true
}
