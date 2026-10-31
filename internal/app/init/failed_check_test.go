package init_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	initapp "github.com/biggs-100/odachi/internal/app/init"
)

// adopting perms one `odachi init` in a repository that already lives: the one
// state where the footprint is an inference, read from what is or is not on disk.
func adopting(files *fakeFiles) initapp.Result {
	return initapp.Run(initapp.Options{Dir: dir, Existing: true}, files, &fakeSkills{})
}

// TestACraftThatCannotBeCheckedIsNotAbsent: a stat that fails is not the answer
// "this project carries no craft". Under the adoption inference that answer puts
// the whole curriculum in a project that keeps it in the machine, so the run
// stops and names the path it could not read.
func TestACraftThatCannotBeCheckedIsNotAbsent(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	files.existsErr = map[string]error{
		filepath.Join(dir, ".agents", "skills"): errors.New("permission denied"),
	}

	got := adopting(files)

	if got.Code != initapp.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message %q)", got.Code, initapp.CodeBlocked, got.Message)
	}
	for _, want := range []string{"could not check", ".agents/skills", "permission denied"} {
		if !strings.Contains(got.Message, want) {
			t.Errorf("Message = %q, want it to carry %q", got.Message, want)
		}
	}
	if len(files.events) != 0 {
		t.Errorf("the run wrote %v from a check it could not make", files.events)
	}
}

// TestAnAgentThatCannotBeCheckedIsNotAbsent: the agent half has its own
// inference, and reporting the failure there is not adoption. A failed Exists is
// not "this project has no agent of ours", so the run stops instead of deciding
// where the agent lives from a question it did not get an answer to.
func TestAnAgentThatCannotBeCheckedIsNotAbsent(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	files.existsErr = map[string]error{
		filepath.Join(dir, ".opencode", "agents", "odachi.md"): errors.New("permission denied"),
	}

	got := adopting(files)

	if got.Code != initapp.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message %q)", got.Code, initapp.CodeBlocked, got.Message)
	}
	for _, want := range []string{"could not check", ".opencode/agents/odachi.md", "permission denied"} {
		if !strings.Contains(got.Message, want) {
			t.Errorf("Message = %q, want it to carry %q", got.Message, want)
		}
	}
	if len(files.events) != 0 {
		t.Errorf("the run wrote %v from a check it could not make", files.events)
	}
}
