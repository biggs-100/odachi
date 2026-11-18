// Package skilleval carries the exams that ship with odachi: the task an agent is
// given, and the small project it is given it in. The skills are the product, so
// the exam for one of them travels with the binary, the way the lens probe's
// fixture does: a project's copy of a skill is odachi's copy, and a measurement
// that lives in the project measures the project.
package skilleval

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

//go:embed all:testdata/exams
var bundled embed.FS

// Root is where the exams live inside the embedded tree.
// (The pattern was written before the projects existed; the comment moves the build.)
const Root = "testdata/exams"

// Bundled is the tree this binary carries, for the tests that check what a build
// ships: an exam whose project was left out of the build is a question with no
// world to answer it in.
func Bundled() fs.FS { return bundled }

// TaskFile is the file an exam always carries: the prompt the agent is given.
// Every other file is the project the prompt is about, except the folders the
// exam uses to keep its own answer key, which start with a dot.
const TaskFile = "task.md"

// hiddenName is the suffix a fixture file carries when its real name would take
// the folder out of this module. A project's go.mod is embedded as go.mod.exam,
// because the go tool will not embed a subtree that looks like another module,
// and the run writes it back under the name the project needs.
const hiddenName = ".exam"

// File is one file of the skill an exam measures: the text the client loads, and
// the path it lives at inside the skill's folder.
type File struct {
	Path    string
	Content []byte
}

// Exam is one skill's exam: which skill it measures, the prompt, and the files of
// the project the prompt is worked in, keyed by their path under the project root
// with forward slashes.
type Exam struct {
	Skill string
	Task  string
	Files map[string][]byte
}

// Load answers the exam for one skill, or a refusal that names the ones that
// exist: a typo in the skill's name is answered with the list, not with an empty
// measurement.
func Load(skill string) (Exam, error) {
	names, err := Names()
	if err != nil {
		return Exam{}, err
	}
	found := false
	for _, name := range names {
		if name == skill {
			found = true
			break
		}
	}
	if !found {
		return Exam{}, fmt.Errorf("there is no exam for %q: the exams that ship are %s", skill, strings.Join(names, ", "))
	}

	root := path.Join(Root, skill)
	task, err := fs.ReadFile(bundled, path.Join(root, TaskFile))
	if err != nil {
		return Exam{}, fmt.Errorf("the exam for %q carries no %s: %w", skill, TaskFile, err)
	}

	exam := Exam{Skill: skill, Task: strings.TrimSpace(string(task)) + "\n", Files: map[string][]byte{}}
	err = fs.WalkDir(bundled, path.Join(root, "project"), func(one string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := fs.ReadFile(bundled, one)
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(one, path.Join(root, "project")), "/")
		rel = strings.TrimSuffix(rel, hiddenName)
		exam.Files[rel] = data
		return nil
	})
	if err != nil {
		return Exam{}, fmt.Errorf("could not read the project of %q: %w", skill, err)
	}
	if len(exam.Files) == 0 {
		return Exam{}, fmt.Errorf("the exam for %q carries no project", skill)
	}
	return exam, nil
}

// Names answers the skills that have an exam, sorted, so a message is stable.
func Names() ([]string, error) {
	entries, err := fs.ReadDir(bundled, Root)
	if err != nil {
		return nil, fmt.Errorf("the exams are not where this binary expects them: %w", err)
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}
