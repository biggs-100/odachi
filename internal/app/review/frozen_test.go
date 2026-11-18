package review_test

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
	"github.com/biggs-100/odachi/internal/domain/scope"
)

// TestTheReviewerRunsInsideItsOwnCandidate: the command runs in the snapshot, not
// in the project. A reviewer whose working directory is the live worktree reads
// the bytes that may move under the verdict, and the configuration that confines
// it governs the directory the run starts in, so the two have to be the same one.
// The rule is written beside the candidate, where the client discovers it.
func TestTheReviewerRunsInsideItsOwnCandidate(t *testing.T) {
	p := newProject()
	p.git.run = filepath.FromSlash("C:/temp/odachi-review-abc")
	p.git.candidateDir = filepath.Join(p.git.run, "candidate")
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{Exec: "revisor"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(p.shell.dirs) == 0 || p.shell.dirs[0] != p.git.candidateDir {
		t.Errorf("the reviewer ran in %v, want the frozen candidate %q", p.shell.dirs, p.git.candidateDir)
	}
	rule, ok := p.files.files[filepath.Join(p.git.run, scope.FileName)]
	if !ok {
		t.Fatalf("the run has no %s beside the candidate, so nothing confines the reviewer; it wrote %v", scope.FileName, keys(p.files.files))
	}
	if !bytes.Contains(rule, []byte("external_directory")) {
		t.Errorf("the rule beside the candidate does not deny an external directory:\n%s", rule)
	}
	if got := p.git.discarded; len(got) != 1 || got[0] != p.git.run {
		t.Errorf("discarded = %v, want the whole run %q so the rule goes with it", got, p.git.run)
	}
}

// keys is the names of the files a run wrote, in the assertion's words: a map of
// bytes prints as a page nobody can read.
func keys(files map[string][]byte) []string {
	out := make([]string, 0, len(files))
	for path := range files {
		out = append(out, path)
	}
	return out
}

// TestTheReviewerGetsTheFrozenTree: exec freezes the tree the verdict names,
// gives that directory to the reviewer in the prompt, and discards it when the
// run ends. A prompt naming no snapshot leaves the reviewer blind; one naming a
// different tree approves bytes it never judged.
func TestTheReviewerGetsTheFrozenTree(t *testing.T) {
	p := newProject()
	p.git.run = filepath.FromSlash("C:/temp/odachi-review-abc")
	p.git.candidateDir = filepath.Join(p.git.run, "candidate")
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{Exec: "revisor"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if got := p.git.archived; len(got) != 1 || got[0] != currentTree {
		t.Errorf("archived = %v, want the verdict's tree %q", got, currentTree)
	}
	if !strings.Contains(p.shell.input, p.git.candidateDir) {
		t.Errorf("the reviewer's prompt does not name the frozen directory:\n%s", p.shell.input)
	}
	if got := p.git.discarded; len(got) != 1 || got[0] != p.git.run {
		t.Errorf("discarded = %v, want %q", got, p.git.run)
	}
}

// TestAFrozenCandidateIsRequired: a repository that cannot export the tree must
// not start a reviewer that would then read the live worktree instead.
func TestAFrozenCandidateIsRequired(t *testing.T) {
	p := newProject()
	p.git.archiveErr = errors.New("git archive failed")

	result := p.run(review.Options{Exec: "revisor"})
	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if len(p.shell.commands) != 0 {
		t.Error("the reviewer ran without a frozen candidate")
	}
	if len(p.git.discarded) != 0 {
		t.Error("a failed freeze discarded a directory it never made")
	}
}

// TestACandidateThatCarriesItsOwnRuleIsRefused: the client reads the nearest
// configuration, so a file inside the candidate is discovered after the rule
// odachi writes beside it and wins. A reviewer started over such a candidate
// reads whatever the candidate allows while the verdict still claims a boundary,
// which is the one answer this run may never give: it refuses and names the file.
func TestACandidateThatCarriesItsOwnRuleIsRefused(t *testing.T) {
	p := newProject()
	p.git.run = filepath.FromSlash("C:/temp/odachi-review-abc")
	p.git.candidateDir = filepath.Join(p.git.run, "candidate")
	carried := filepath.Join(p.git.candidateDir, scope.FileName)
	p.files.files[carried] = []byte(`{"permissions": []}`)

	result := p.run(review.Options{Exec: "revisor"})
	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, scope.FileName) {
		t.Errorf("the refusal does not name the file that voids the boundary: %q", result.Message)
	}
	if len(p.shell.commands) != 0 {
		t.Errorf("the reviewer ran over a candidate that voids the boundary: %v", p.shell.commands)
	}
	if got := p.git.discarded; len(got) != 1 || got[0] != p.git.run {
		t.Errorf("discarded = %v, want the whole run %q", got, p.git.run)
	}
}
