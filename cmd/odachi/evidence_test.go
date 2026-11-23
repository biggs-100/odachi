package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// findIgnores lists every .gitignore under dir, as slash paths relative to it.
func findIgnores(t *testing.T, dir string) []string {
	t.Helper()
	var found []string
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Name() != ".gitignore" {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		found = append(found, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", dir, err)
	}
	return found
}

// TestInitKeepsTheProofValidAcrossAReview is the whole point of the block: the
// evidence is ignored by the project itself, so taking a review does not move the
// tree, the proof stays valid and the gate does not ask for a retake.
//
// The verdict is attached from OUTSIDE the worktree, so the only thing that can
// move the tree is what the tool writes.
func TestInitKeepsTheProofValidAcrossAReview(t *testing.T) {
	dir := changedProject(t)

	if code, out, errOut := runDir(t, dir, []string{"init"}, ""); code != 0 {
		t.Fatalf("init: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	ignore, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf(".gitignore was not written: %v", err)
	}
	for _, want := range []string{"# odachi:begin", "/proof/", "*.bak-*", "# odachi:end"} {
		if !strings.Contains(string(ignore), want) {
			t.Errorf(".gitignore does not carry %q:\n%s", want, ignore)
		}
	}
	// The verdicts left the worktree, so the rule that used to keep them out of
	// the tree has no reason to exist. A stale rule is a lie about where the store
	// is, and this is the test that would let it back in.
	if strings.Contains(string(ignore), ".odachi/reviews") {
		t.Errorf(".gitignore still carries the review store rule:\n%s", ignore)
	}

	outside := t.TempDir()
	writeFile(t, outside, "veredicto.json", `{"reviewer":"humano",`+declared(t, dir)+`"findings":[],"passes":[{"path":"internal/app/turnos.go","lens":"reliability"},{"path":".gitignore","lens":"reliability"},{"path":"odd/review.toml","lens":"reliability"}]}`)
	attach := filepath.Join(outside, "veredicto.json")

	if code, out, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
		t.Fatalf("check --run: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	proofPath := filepath.Join(dir, "proof", "0001-turnos.txt")
	before, err := os.ReadFile(proofPath)
	if err != nil {
		t.Fatalf("the proof was not written: %v", err)
	}

	if code, out, errOut := runDir(t, dir, []string{"review", "--attach", attach}, ""); code != 0 {
		t.Fatalf("review --attach: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	after, err := os.ReadFile(proofPath)
	if err != nil {
		t.Fatalf("the proof disappeared after the review: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("the review rewrote the proof:\n%q\nvs\n%q", after, before)
	}

	code, out, errOut := runDir(t, dir, []string{"check", "--require-review"}, "")
	if code != 0 {
		t.Fatalf("check --require-review: code = %d, want 0 — the proof did not move (stdout: %q, stderr: %q)", code, out, errOut)
	}

	// No retake, and no lazy ignore file either: init's block is the only one.
	if got := findIgnores(t, dir); len(got) != 1 || got[0] != ".gitignore" {
		t.Errorf(".gitignore files = %v, want only the project's own", got)
	}
	if last, err := os.ReadFile(proofPath); err != nil || string(last) != string(before) {
		t.Errorf("the proof changed across the gate: %q vs %q (err %v)", last, before, err)
	}
}

// TestWithoutInitTheFallbackKeepsTheProofValid is the other half: a project that
// never ran init still has its evidence out of the tree, through the interior
// ignore file the verb writes once.
func TestWithoutInitTheFallbackKeepsTheProofValid(t *testing.T) {
	t.Setenv("ODACHI_MODEL", "")
	t.Setenv("OPENCODE_SESSION_ID", "")
	dir := changedProject(t)

	if code, out, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
		t.Fatalf("check --run: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	proofPath := filepath.Join(dir, "proof", "0001-turnos.txt")
	before, err := os.ReadFile(proofPath)
	if err != nil {
		t.Fatalf("the proof was not written: %v", err)
	}

	if code, out, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
		t.Fatalf("second check --run: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	after, err := os.ReadFile(proofPath)
	if err != nil {
		t.Fatalf("read the proof: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("a repeated run rewrote the proof:\n%q\nvs\n%q", after, before)
	}

	ignores := findIgnores(t, dir)
	if len(ignores) != 1 || ignores[0] != "proof/.gitignore" {
		t.Errorf(".gitignore files = %v, want exactly the fallback inside proof/", ignores)
	}
	if _, err := os.Stat(filepath.Join(dir, ".gitignore")); !os.IsNotExist(err) {
		t.Errorf("check created a project .gitignore (err %v), want init to be the only writer of it", err)
	}
}

const cliTree = "0123456789abcdef0123456789abcdef01234567"

// TestEvidenceExportRejectsWrongArity pins the two words and the single tree.
func TestEvidenceExportRejectsWrongArity(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{
		{"evidence"},
		{"evidence", "import"},
		{"evidence", "export"},
		{"evidence", "export", cliTree, cliTree},
	} {
		code, _, stderr := runDir(t, dir, args, "")
		if code == 0 {
			t.Errorf("%v: exit 0, want a refusal", args)
		}
		if !strings.Contains(stderr, "usage: odachi evidence export") {
			t.Errorf("%v: stderr %q does not carry the usage line", args, stderr)
		}
	}
}

// TestEvidenceExportNeedsOut refuses before touching disk when the human gave
// no place for the bundle.
func TestEvidenceExportNeedsOut(t *testing.T) {
	dir := t.TempDir()
	code, _, stderr := runDir(t, dir, []string{"evidence", "export", cliTree}, "")
	if code == 0 {
		t.Fatal("evidence export without --out = 0, want a refusal")
	}
	if !strings.Contains(stderr, "--out") {
		t.Errorf("stderr %q does not name --out", stderr)
	}
}

// TestEvidenceExportWritesAParseableBundle is the verb's acceptance: the tree
// passed on the command line becomes a portable file with the three raw parts
// and their hashes, readable without odachi.
func TestEvidenceExportWritesAParseableBundle(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	verdict := `{"findings":[],"passes":[],"reviewer":"test","tree":"` + cliTree + `"}` + "\n"
	verification := `{"at":"2026-01-01T00:00:00Z","by":"test","notes":[],"schema":1,"tree":"` + cliTree + `","verdict":"verified"}` + "\n"
	proofText := "# tree " + cliTree + "\n# command echo ok\n# output\nok\n"
	writeFile(t, dir, ".git/odachi/reviews/"+cliTree+".json", verdict)
	writeFile(t, dir, ".git/odachi/verifications/"+cliTree+".json", verification)
	writeFile(t, dir, "proof/0001-close.txt", proofText)

	out := filepath.Join(dir, "exports", "close.json")
	code, stdout, stderr := runDir(t, dir, []string{"evidence", "export", cliTree, "--out", out}, "")
	if code != 0 {
		t.Fatalf("evidence export: exit %d (stdout: %q, stderr: %q)", code, stdout, stderr)
	}
	if !strings.Contains(stdout, cliTree) {
		t.Errorf("stdout %q does not name the tree", stdout)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("the bundle was not written: %v", err)
	}
	var b struct {
		Schema       int               `json:"schema"`
		Tree         string            `json:"tree"`
		Verdict      *string           `json:"verdict"`
		Verification *string           `json:"verification"`
		Proof        *string           `json:"proof"`
		SHA256       map[string]string `json:"sha256"`
	}
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatalf("the bundle is not JSON: %v: %s", err, data)
	}
	if b.Schema != 1 || b.Tree != cliTree {
		t.Errorf("schema/tree = %d/%q, want 1/%s", b.Schema, b.Tree, cliTree)
	}
	if b.Verdict == nil || *b.Verdict != verdict || b.Verification == nil || *b.Verification != verification || b.Proof == nil || *b.Proof != proofText {
		t.Errorf("the raw parts did not survive the export: %+v", b)
	}
	for name, want := range map[string]string{
		"verdict":      "eb98c8f8c638a2a94bde159065c8bbd68ec0c122b014d7c5488838d00b83e71a",
		"verification": "cc2eec198efd4ca34a34fd8e9cd415269a56cbd8a82195b55b3496b15645a869",
		"proof":        "3fa3ab272919d01a8d1c7451ebdeecbedcbd54c57474f1235dd0777b243c514b",
	} {
		if b.SHA256[name] != want {
			t.Errorf("sha256[%s] = %q, want %s", name, b.SHA256[name], want)
		}
	}
}

// TestEvidenceExportRefusesOverwrite names the file it protected and leaves
// its bytes alone.
func TestEvidenceExportRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, dir, "proof/0001-close.txt", "# tree "+cliTree+"\n# command echo ok\n# output\nok\n")

	out := filepath.Join(dir, "close.json")
	writeFile(t, dir, "close.json", "keep me")
	code, _, stderr := runDir(t, dir, []string{"evidence", "export", cliTree, "--out", out}, "")
	if code == 0 {
		t.Fatal("evidence export over an existing file = 0, want a refusal")
	}
	if !strings.Contains(stderr, out) {
		t.Errorf("stderr %q does not name the output %s", stderr, out)
	}
	got, err := os.ReadFile(out)
	if err != nil || string(got) != "keep me" {
		t.Errorf("the existing output was changed: %q (%v)", got, err)
	}
}
