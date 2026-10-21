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

	"github.com/biggz/odachi/internal/domain/taskdoc"
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
		return nil, fmt.Errorf("no pude leer %s: %w", Dir, err)
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
			return nil, fmt.Errorf("no pude leer %s: %w", rel, err)
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

// Current returns the highest-numbered task document: the feature the human is
// working on now. A project with no feature yet is an error a human can read,
// and it names the command that creates the first one.
func Current(files Files, dir string) (Document, error) {
	docs, err := All(files, dir)
	if err != nil {
		return Document{}, err
	}
	if len(docs) == 0 {
		return Document{}, fmt.Errorf("todavía no hay ningún task doc en %s; corré odachi start para crear el primero", Dir)
	}
	return docs[len(docs)-1], nil
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
