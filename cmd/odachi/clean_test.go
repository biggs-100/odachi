package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	if len(payload.Report.Classes) != 9 {
		t.Errorf("classes = %d, want the nine the map names", len(payload.Report.Classes))
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
