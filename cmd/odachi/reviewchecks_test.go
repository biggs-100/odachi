package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/layout"
	"github.com/biggs-100/odachi/internal/domain/review"
	"github.com/biggs-100/odachi/internal/domain/store"
)

// declareChecks writes the project's review file with one declared tool.
func declareChecks(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "odd"), 0o755); err != nil {
		t.Fatalf("mkdir odd: %v", err)
	}
	file := `mode = "on"

[[check]]
name = "lint"
command = "golangci-lint run ./..."
severity = "WARNING"
`
	if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(layout.Review)), []byte(file), 0o644); err != nil {
		t.Fatalf("write review.toml: %v", err)
	}
}

// TestReviewChecksApprovesTheDeclaredList: the declaration travels with the
// repository and the agreement to run it does not, so the person who cloned it
// says yes once and the record lives in the clone.
func TestReviewChecksApprovesTheDeclaredList(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)
	declareChecks(t, dir)

	code, out, errOut := runDir(t, dir, []string{"review", "checks"}, "")
	if code != 0 {
		t.Fatalf("checks status: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	for _, want := range []string{"not approved:", "lint", "golangci-lint run ./...", "odachi review checks approve"} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout %q does not carry %q", out, want)
		}
	}

	code, out, errOut = runDir(t, dir, []string{"review", "checks", "approve"}, "")
	if code != 0 {
		t.Fatalf("checks approve: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(out, "approved 1 check") {
		t.Errorf("stdout %q does not say what was approved", out)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".git", "odachi", store.Checks))
	if err != nil {
		t.Fatalf("the clone has no approval record: %v", err)
	}
	approval, err := review.ParseCheckApproval(data)
	if err != nil {
		t.Fatalf("the approval record does not parse: %v", err)
	}
	if approval.Count != 1 {
		t.Errorf("the record says %d tools, want 1", approval.Count)
	}

	code, out, errOut = runDir(t, dir, []string{"review", "checks"}, "")
	if code != 0 {
		t.Fatalf("checks status after approving: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(out, "approved:") {
		t.Errorf("stdout %q does not report the approval", out)
	}
}

// TestReviewChecksApproveRefusesAProjectThatDeclaresNothing: there is nothing to
// agree to, and a record that says otherwise would be a lie about what runs.
func TestReviewChecksApproveRefusesAProjectThatDeclaresNothing(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)

	code, out, errOut := runDir(t, dir, []string{"review", "checks", "approve"}, "")
	if code == 0 {
		t.Fatalf("approving an empty list: code = 0 (stdout: %q), want a refusal", out)
	}
	if !strings.Contains(errOut+out, "declares no checks") {
		t.Errorf("the refusal %q does not say the project declares none", errOut+out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".git", "odachi", store.Checks)); err == nil {
		t.Error("a refusal wrote the approval record")
	}
}

// TestReviewChecksRefusesAWordItDoesNotKnow: the grammar is two words, and a third
// is named instead of guessed.
func TestReviewChecksRefusesAWordItDoesNotKnow(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)

	code, _, errOut := runDir(t, dir, []string{"review", "checks", "maybe"}, "")
	if code == 0 {
		t.Fatal("review checks maybe: code = 0, want a refusal")
	}
	if !strings.Contains(errOut, "maybe") {
		t.Errorf("the refusal %q does not name the word", errOut)
	}
}
