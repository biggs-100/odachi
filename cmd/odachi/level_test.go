package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestAskThenLevelCountsTheOpenQuestionAndTheAnswer is the CLI contract of
// Contract IV: level reads the history the agent wrote and reports it.
func TestAskThenLevelCountsTheOpenQuestionAndTheAnswer(t *testing.T) {
	dir, _ := startedProject(t)

	if code, _, stderr := runDir(t, dir, []string{"ask", "¿Los turnos se pueden cancelar?"}, ""); code != 0 {
		t.Fatalf("ask: exit %d, stderr=%s", code, stderr)
	}

	code, stdout, stderr := runDir(t, dir, []string{"level"}, "")
	if code != 0 {
		t.Fatalf("level: exit %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "open questions: 1") {
		t.Errorf("level does not see the open question:\n%s", stdout)
	}

	if code, _, stderr := runDir(t, dir, []string{"ask", "--answer", "No en la primera versión"}, ""); code != 0 {
		t.Fatalf("ask --answer: exit %d, stderr=%s", code, stderr)
	}

	code, stdout, stderr = runDir(t, dir, []string{"level"}, "")
	if code != 0 {
		t.Fatalf("level: exit %d, stderr=%s", code, stderr)
	}
	for _, want := range []string{"answered questions: 1", "open questions: 0"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("level does not report %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "--to") {
		t.Errorf("level proposed a graduation with one feature and one answer:\n%s", stdout)
	}
}

// TestLevelShowsTheProfileInForce: the profile is read from the human's file,
// never re-inferred over it.
func TestLevelShowsTheProfileInForce(t *testing.T) {
	dir, _ := startedProject(t)

	code, stdout, stderr := runDir(t, dir, []string{"level"}, "")

	if code != 0 {
		t.Fatalf("exit = %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "profile: ci") {
		t.Errorf("stdout %q does not show the profile start wrote", stdout)
	}
}

// TestLevelToAppliesTheProfileAndKeepsTheOldFile: the human accepts, the tool
// backs up and writes the eight fields.
func TestLevelToAppliesTheProfileAndKeepsTheOldFile(t *testing.T) {
	dir, _ := startedProject(t)

	code, stdout, stderr := runDir(t, dir, []string{"level", "--to", "novato"}, "")
	if code != 0 {
		t.Fatalf("exit = %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "novato") {
		t.Errorf("stdout %q does not report the applied profile", stdout)
	}
	for _, want := range []string{"questions", "proof", "explain"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout does not report the changed field %q:\n%s", want, stdout)
		}
	}

	files := tree(t, dir)
	profile := files[".odachi/profile.toml"]
	if !strings.Contains(profile, `"novato"`) {
		t.Errorf("profile.toml does not hold the new profile:\n%s", profile)
	}
	if !strings.Contains(profile, `"one-by-one"`) || !strings.Contains(profile, `"why"`) {
		t.Errorf("profile.toml lost its fields:\n%s", profile)
	}

	var backups []string
	for path := range files {
		if strings.HasPrefix(filepath.Base(path), "profile.toml.bak-") {
			backups = append(backups, path)
		}
	}
	if len(backups) != 1 {
		t.Fatalf("backups = %v, want exactly one: the previous profile is never lost", backups)
	}
	if !strings.Contains(files[backups[0]], `"ci"`) {
		t.Errorf("the backup does not hold the previous profile:\n%s", files[backups[0]])
	}
}

// TestLevelToUnknownProfileListsTheAvailableOnes: a typo is a blocked run with
// a way out.
func TestLevelToUnknownProfileListsTheAvailableOnes(t *testing.T) {
	dir, _ := startedProject(t)
	before := tree(t, dir)

	code, stdout, stderr := runDir(t, dir, []string{"level", "--to", "sinior"}, "")

	if code != 1 {
		t.Fatalf("exit = %d, want 1 (stdout=%s stderr=%s)", code, stdout, stderr)
	}
	for _, want := range []string{"novato", "senior", "ci"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr %q does not name the available profile %q", stderr, want)
		}
	}
	after := tree(t, dir)
	if after[".odachi/profile.toml"] != before[".odachi/profile.toml"] {
		t.Errorf("a blocked level rewrote the profile")
	}
}

// TestLevelNeverWritesOnItsOwn: showing the facts is read-only.
func TestLevelNeverWritesOnItsOwn(t *testing.T) {
	dir, _ := startedProject(t)
	before := tree(t, dir)

	if code, _, stderr := runDir(t, dir, []string{"level"}, ""); code != 0 {
		t.Fatalf("exit = %d, stderr=%s", code, stderr)
	}

	after := tree(t, dir)
	if len(after) != len(before) {
		t.Fatalf("level changed the tree: before=%v after=%v", len(before), len(after))
	}
	for path, content := range before {
		if after[path] != content {
			t.Errorf("level rewrote %s", path)
		}
	}
}

// TestSummaryListsBothFeatures is the end of the T6 story: two features started
// with two different outcomes, both reported, in order, with totals.
func TestSummaryListsBothFeatures(t *testing.T) {
	dir := t.TempDir()

	if code, _, stderr := runDir(t, dir, allFlags(), ""); code != 0 {
		t.Fatalf("first start: exit %d, stderr=%s", code, stderr)
	}
	other := allFlags()
	for i := 0; i < len(other); i++ {
		if other[i] == "--outcome" {
			other[i+1] = "Quiero facturar a mis obras sociales"
		}
	}
	if code, _, stderr := runDir(t, dir, append(other, "--force"), ""); code != 0 {
		t.Fatalf("second start: exit %d, stderr=%s", code, stderr)
	}

	code, stdout, stderr := runDir(t, dir, []string{"summary"}, "")
	if code != 0 {
		t.Fatalf("summary: exit %d, stderr=%s", code, stderr)
	}

	for _, want := range []string{
		"0001-quiero-reservar-turnos",
		"0002-quiero-facturar",
		`R1 · "Quiero reservar turnos desde el teléfono"`,
		`R1 · "Quiero facturar a mis obras sociales"`,
		"2 features",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("summary does not contain %q:\n%s", want, stdout)
		}
	}
	if strings.Index(stdout, "0001-") > strings.Index(stdout, "0002-") {
		t.Errorf("features are out of order:\n%s", stdout)
	}
}

// TestSummaryOnAnEmptyProjectSaysSo: one English line, exit 0, no error.
func TestSummaryOnAnEmptyProjectSaysSo(t *testing.T) {
	dir := t.TempDir()

	code, stdout, stderr := runDir(t, dir, []string{"summary"}, "")

	if code != 0 {
		t.Fatalf("exit = %d, want 0 (stderr=%s)", code, stderr)
	}
	lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("summary printed %d lines, want one:\n%s", len(lines), stdout)
	}
	if !strings.Contains(lines[0], "odachi start") {
		t.Errorf("the line %q does not say how to create the first feature", lines[0])
	}
}

// TestUnknownCommandListsEveryVerb keeps the discovery path honest now that
// there are four commands.
func TestUnknownCommandListsEveryVerb(t *testing.T) {
	dir := t.TempDir()

	code, _, stderr := runDir(t, dir, []string{"deploy"}, "")

	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	for _, want := range []string{"start", "ask", "level", "summary"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr %q does not list %q", stderr, want)
		}
	}
}
