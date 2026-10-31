package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestCleanReportsEveryClassAndRemovesNothing: the first half of the verb. A
// clone with evidence in it gets a map -what each class is, who owns it and what
// it would do- and the report itself writes nothing.
func TestCleanReportsEveryClassAndRemovesNothing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)

	storeDir := filepath.Join(dir, ".git", "odachi")
	writeFile(t, dir, ".git/odachi/reviews/aaaa.json", strings.Repeat("v", 100))
	writeFile(t, dir, ".git/odachi/backups/brief.md.bak-1111", strings.Repeat("b", 20))
	writeFile(t, dir, ".git/PR100.md", strings.Repeat("p", 10))
	writeFile(t, dir, "proof/0001-a-feature.txt", strings.Repeat("g", 5))
	writeFile(t, dir, "odd/tasks/0001-a-feature.md", strings.Repeat("t", 5))

	code, stdout, stderr := runDir(t, dir, []string{"clean"}, "")
	if code != 0 {
		t.Fatalf("clean: exit %d (stdout: %q, stderr: %q)", code, stdout, stderr)
	}

	for _, want := range []string{
		"odachi clean · .git/odachi",
		"review records",
		"undo backups",
		"old agent notes",
		"the rest of .git/",
		"proof/",
		"odd/tasks",
		"odachi's records:",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("the report does not name %q:\n%s", want, stdout)
		}
	}

	// Nothing was removed: the report is the verb.
	for _, kept := range []string{
		filepath.Join(storeDir, "reviews", "aaaa.json"),
		filepath.Join(storeDir, "backups", "brief.md.bak-1111"),
		filepath.Join(dir, ".git", "PR100.md"),
		filepath.Join(dir, "proof", "0001-a-feature.txt"),
		filepath.Join(dir, "odd", "tasks", "0001-a-feature.md"),
	} {
		if _, err := os.Stat(kept); err != nil {
			t.Errorf("clean removed %s: %v", kept, err)
		}
	}
}

// TestCleanJSONCarriesTheSameMap: a machine reads the report as data, with the
// classes and the totals the block prints.
func TestCleanJSONCarriesTheSameMap(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, dir, ".git/odachi/reviews/aaaa.json", strings.Repeat("v", 100))
	writeFile(t, dir, ".git/odachi/reviews/bbbb.json", strings.Repeat("v", 50))
	writeFile(t, dir, ".git/PR100.md", strings.Repeat("p", 10))

	code, stdout, stderr := runDir(t, dir, []string{"clean", "--json"}, "")
	if code != 0 {
		t.Fatalf("clean --json: exit %d (stdout: %q, stderr: %q)", code, stdout, stderr)
	}

	var payload struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Report  struct {
			Store   string `json:"store"`
			Classes []struct {
				Name      string `json:"name"`
				Owner     string `json:"owner"`
				Files     int    `json:"files"`
				Bytes     int64  `json:"bytes"`
				Action    string `json:"action"`
				Cleanable bool   `json:"cleanable"`
			} `json:"classes"`
			TotalFiles int   `json:"totalFiles"`
			TotalBytes int64 `json:"totalBytes"`
		} `json:"report"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("the JSON does not parse: %v\n%s", err, stdout)
	}

	if payload.Code != 0 {
		t.Errorf("code = %d, want 0", payload.Code)
	}
	if len(payload.Report.Classes) != 10 {
		t.Errorf("classes = %d, want the ten the map names", len(payload.Report.Classes))
	}
	if payload.Report.TotalFiles != 2 || payload.Report.TotalBytes != 150 {
		t.Errorf("totals = %d files, %d bytes; want the two verdicts", payload.Report.TotalFiles, payload.Report.TotalBytes)
	}
	if payload.Report.Store != ".git/odachi" {
		t.Errorf("store = %q, want it named from the project", payload.Report.Store)
	}
	for _, class := range payload.Report.Classes {
		if class.Name == "old agent notes" && class.Files != 1 {
			t.Errorf("old agent notes = %d files, want the one PR body", class.Files)
		}
	}
}

// TestCleanApplyKeepsTheNewestRecordsAndRemovesTheRest drives the lifetime
// end to end: the run removes the old trees and leaves proof/, the plan and
// git's files exactly as they were.
func TestCleanApplyKeepsTheNewestRecordsAndRemovesTheRest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)

	storeDir := filepath.Join(dir, ".git", "odachi")
	writeFile(t, dir, "proof/0001-a-feature.txt", "the proof")
	writeFile(t, dir, "odd/tasks/0001-a-feature.md", "the plan")
	writeFile(t, dir, ".git/COMMIT_EDITMSG", "a commit")

	for i := 0; i < 12; i++ {
		at := filepath.Join(storeDir, "reviews", fmt.Sprintf("tree%02d.json", i))
		writeFile(t, dir, filepath.ToSlash(filepath.Join(".git", "odachi", "reviews", fmt.Sprintf("tree%02d.json", i))), "verdict")
		when := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i)
		if err := os.Chtimes(at, when, when); err != nil {
			t.Fatalf("could not date %s: %v", at, err)
		}
	}

	code, stdout, stderr := runDir(t, dir, []string{"clean", "--apply"}, "")
	if code != 0 {
		t.Fatalf("clean --apply: exit %d (stdout: %q, stderr: %q)", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "removed 2 files") {
		t.Errorf("the report does not say what it removed:\n%s", stdout)
	}

	for _, gone := range []string{"tree00.json", "tree01.json"} {
		if _, err := os.Stat(filepath.Join(storeDir, "reviews", gone)); !os.IsNotExist(err) {
			t.Errorf("%s survived although it is one of the two oldest", gone)
		}
	}
	for _, kept := range []string{
		filepath.Join(storeDir, "reviews", "tree11.json"),
		filepath.Join(dir, "proof", "0001-a-feature.txt"),
		filepath.Join(dir, "odd", "tasks", "0001-a-feature.md"),
		filepath.Join(dir, ".git", "COMMIT_EDITMSG"),
	} {
		if _, err := os.Stat(kept); err != nil {
			t.Errorf("%s was removed: %v", kept, err)
		}
	}
}

// TestCleanApplyLegacyRemovesTheOldNotes: the notes an older agent left in the
// git directory go only when both flags ask for them, and git's own files stay.
func TestCleanApplyLegacyRemovesTheOldNotes(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)

	writeFile(t, dir, ".git/PR100.md", "a pull request body")
	writeFile(t, dir, ".git/CLOSE99.txt", "a close message")
	writeFile(t, dir, ".git/COMMIT_EDITMSG", "a commit")

	if code, _, stderr := runDir(t, dir, []string{"clean", "--apply"}, ""); code != 0 {
		t.Fatalf("clean --apply: exit %d: %s", code, stderr)
	}
	if _, err := os.Stat(filepath.Join(dir, ".git", "PR100.md")); err != nil {
		t.Fatalf("a note was removed without --legacy: %v", err)
	}

	if code, _, stderr := runDir(t, dir, []string{"clean", "--legacy"}, ""); code != 1 {
		t.Fatalf("clean --legacy alone: exit %d, want 1: %s", code, stderr)
	}

	if code, _, stderr := runDir(t, dir, []string{"clean", "--apply", "--legacy"}, ""); code != 0 {
		t.Fatalf("clean --apply --legacy: exit %d: %s", code, stderr)
	}
	for _, gone := range []string{"PR100.md", "CLOSE99.txt"} {
		if _, err := os.Stat(filepath.Join(dir, ".git", gone)); !os.IsNotExist(err) {
			t.Errorf("%s survived --apply --legacy", gone)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".git", "COMMIT_EDITMSG")); err != nil {
		t.Errorf("git's own file was removed: %v", err)
	}
}

// TestCleanObeysTheProjectRetention: the numbers are the project's, read from
// odd/review.toml, and the report says where they came from so a lifetime nobody
// remembers choosing is never a mystery.
func TestCleanObeysTheProjectRetention(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, dir, "odd/review.toml", "mode = \"on\"\nkeep_verdicts = 2\nkeep_prompts = 1\n")
	writeFile(t, dir, ".git/odachi/reviews/aaaa.json", "verdict")
	writeFile(t, dir, ".git/odachi/reviews/aaaa.prompt.md", "prompt")

	code, stdout, stderr := runDir(t, dir, []string{"clean"}, "")
	if code != 0 {
		t.Fatalf("clean: exit %d (stdout: %q, stderr: %q)", code, stdout, stderr)
	}
	for _, want := range []string{
		"keeps the newest 2 trees",
		"keeps the prompts of the newest 1 tree",
		"retention: the newest 2 trees and the prompts of the newest 1, from odd/review.toml",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("the report does not say %q:\n%s", want, stdout)
		}
	}
}

// TestCleanRefusesABrokenReviewToml: the numbers are a decision, so a file the
// tool cannot read is named instead of falling back to a default nobody chose.
func TestCleanRefusesABrokenReviewToml(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, dir, "odd/review.toml", "mode = \"on\"\nkeep_verdicts = -1\n")

	code, _, stderr := runDir(t, dir, []string{"clean"}, "")
	if code != 1 {
		t.Fatalf("clean with a broken review.toml: exit %d, want 1 (stderr: %q)", code, stderr)
	}
	for _, want := range []string{"odd/review.toml", "keep_verdicts"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("the refusal does not name %q: %q", want, stderr)
		}
	}
}

// TestCleanRefusesAFolderThatIsNotARepository: there is no clone, so there is no
// store to report on, and the run says which folder it looked at.
func TestCleanRefusesAFolderThatIsNotARepository(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	code, _, stderr := runDir(t, dir, []string{"clean"}, "")
	if code != 1 {
		t.Fatalf("clean outside a repository: exit %d, want 1 (stderr: %q)", code, stderr)
	}
	if !strings.Contains(stderr, "not a git repository") {
		t.Errorf("the refusal does not name the missing repository: %q", stderr)
	}
}
