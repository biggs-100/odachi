package skilleval_test

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/skilleval"
)

// TestTheTestsExamIsWhole: the exam ships with the binary, so the first thing the
// tool can say about it is that the prompt and the project it is worked in are
// both there.
func TestTheTestsExamIsWhole(t *testing.T) {
	t.Parallel()

	exam, err := skilleval.Load("tests")
	if err != nil {
		t.Fatalf("Load(tests): %v", err)
	}
	if !strings.Contains(exam.Task, "config.Parse") {
		t.Errorf("the task does not name the function it asks for:\n%s", exam.Task)
	}
	for _, want := range []string{"go.mod", "brief.md", "odd/tasks/0001-config.md"} {
		if _, ok := exam.Files[want]; !ok {
			t.Errorf("the project carries no %s: %v", want, keys(exam.Files))
		}
	}
}

// TestAnExamThatDoesNotExistNamesTheOnesThatDo: a typo is answered with the list,
// because "no exam" and "the wrong name" are different answers.
func TestAnExamThatDoesNotExistNamesTheOnesThatDo(t *testing.T) {
	t.Parallel()

	_, err := skilleval.Load("absent")
	if err == nil {
		t.Fatal("Load(absent) = nil error, want a refusal")
	}
	if !strings.Contains(err.Error(), "tests") {
		t.Errorf("the refusal does not name the exams that exist: %v", err)
	}
}

// TestTheBuildCarriesTheWholeProject: a nested go.mod takes its folder out of the
// embed without a word, and an exam without its project is a prompt with no world
// to answer it in. This reads the tree the binary carries, not the disk.
func TestTheBuildCarriesTheWholeProject(t *testing.T) {
	t.Parallel()

	for _, want := range []string{
		"testdata/exams/tests/project/odd/tasks/0001-config.md",
		"testdata/exams/tests/project/brief.md",
		"testdata/exams/tests/project/go.mod.exam",
	} {
		if _, err := fs.Stat(skilleval.Bundled(), want); err != nil {
			t.Errorf("the embedded exam has no %s: %v", want, err)
		}
	}
}

func keys(files map[string][]byte) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	return names
}
