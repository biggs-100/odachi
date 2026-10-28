package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// docsProject is a project whose only change is documentation: passive, no lens.
func docsProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)
	writeFile(t, dir, "docs/guia.md", "# guía\n\ncomo reservar un turno\n")
	commitAll(t, dir, "guía")

	writeFile(t, dir, "docs/guia.md", "# guía\n\ncomo reservar un turno, paso a paso\n")
	return dir
}

// TestReviewUsesTheLensTheEvidenceAsksFor: a code change is reviewed with the
// reliability lens, and only that one. The human does not have to know the word.
func TestReviewUsesTheLensTheEvidenceAsksFor(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)

	code, out, errOut := runDir(t, dir, []string{"review"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	for _, want := range []string{"lenses reliability", "### reliability", "## What to look at", "## What NOT to report", "WARNING and SUGGESTION never block"} {
		if !strings.Contains(out, want) {
			t.Errorf("the prompt does not contain %q", want)
		}
	}
	if strings.Contains(out, "### risk") {
		t.Errorf("a code-only change asked for the risk lens")
	}
	if strings.Contains(out, "name: reliability") {
		t.Errorf("the prompt carries the lens frontmatter")
	}
}

// TestReviewOfDocumentationAsksForNoLens: a docs-only change is passive, and the
// prompt it gets is the one this verb printed before lenses existed.
func TestReviewOfDocumentationAsksForNoLens(t *testing.T) {
	t.Parallel()
	dir := docsProject(t)

	code, out, errOut := runDir(t, dir, []string{"review"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if strings.Contains(out, "## The lenses") {
		t.Errorf("a documentation-only change got a lens section:\n%s", out)
	}
	if !strings.Contains(out, "## The six rules") {
		t.Errorf("the prompt lost the five rules")
	}
}

// TestReviewOfAnAuthChangeAsksForTheFour: the signals are the difference between
// one prompt and four.
func TestReviewOfAnAuthChangeAsksForTheFour(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)
	writeFile(t, dir, "internal/auth/session.go", "package auth\n\nfunc Validar() bool { return true }\n")

	code, out, errOut := runDir(t, dir, []string{"review"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	for _, lens := range []string{"risk", "readability", "reliability", "resilience"} {
		if !strings.Contains(out, "### "+lens) {
			t.Errorf("the prompt does not carry the %s lens", lens)
		}
	}
	if !strings.Contains(out, "lenses risk, readability, reliability, resilience") {
		t.Errorf("the header does not name the four lenses in order")
	}
}

// TestLensFlagAddsOne: an explicit lens adds to what the evidence asked for; it
// never removes it.
func TestLensFlagAddsOne(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)

	code, out, errOut := runDir(t, dir, []string{"review", "--lens", "risk"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	for _, want := range []string{"### risk", "### reliability"} {
		if !strings.Contains(out, want) {
			t.Errorf("the prompt does not contain %q", want)
		}
	}
}

// TestLensFlagAll: `--lens all` is the four, even where the evidence asked for
// nothing.
func TestLensFlagAll(t *testing.T) {
	t.Parallel()
	dir := docsProject(t)

	code, out, errOut := runDir(t, dir, []string{"review", "--lens", "all"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	for _, lens := range []string{"risk", "readability", "reliability", "resilience"} {
		if !strings.Contains(out, "### "+lens) {
			t.Errorf("--lens all does not carry the %s lens", lens)
		}
	}
}

// TestLensFlagWithoutAValueIsAccepted: the documented spelling is
// `odachi review --lens --show`, so a `--lens` with no value means "the lenses
// the evidence asks for" and must not eat the next flag.
func TestLensFlagWithoutAValueIsAccepted(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)

	code, out, errOut := runDir(t, dir, []string{"review", "--lens", "--show"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1: this tree has no verdict yet (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(errOut, "there is no review for this tree") {
		t.Errorf("stderr %q does not report the missing verdict: --lens ate --show", errOut)
	}
}

// TestShowListsRequiredAndMissingLenses: `--show` answers the question the gate
// will ask, before the gate asks it.
func TestShowListsRequiredAndMissingLenses(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)
	verdict := filepath.Join(t.TempDir(), "veredicto.json")
	writeAt(t, verdict, `{"reviewer":"humano","lenses":["reliability"],`+declared(t, dir)+`"findings":[],"passes":[{"path":"internal/app/turnos.go","lens":"reliability"}]}`)

	if code, out, errOut := runDir(t, dir, []string{"review", "--attach", verdict}, ""); code != 0 {
		t.Fatalf("attach: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	code, out, errOut := runDir(t, dir, []string{"review", "--lens", "--show"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	for _, want := range []string{"lenses: reliability", "required: reliability"} {
		if !strings.Contains(out, want) {
			t.Errorf("--show does not print %q:\n%s", want, out)
		}
	}
}

// TestAnInventedLensIsRefused: a typo must not print a prompt that reviewed
// nothing.
func TestAnInventedLensIsRefused(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)

	code, _, errOut := runDir(t, dir, []string{"review", "--lens", "seguridad"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "risk") {
		t.Errorf("stderr %q does not name the valid lenses", errOut)
	}
}

// TestRefutePrintsThePromptForTheStoredVerdict: the refuter's prompt names the
// blocking findings with their ids and carries the diff it has to cite.
func TestRefutePrintsThePromptForTheStoredVerdict(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)
	verdict := filepath.Join(t.TempDir(), "veredicto.json")
	writeAt(t, verdict, `{
  "reviewer": "claude",
`+declared(t, dir)+`  "findings": [
    {"path": "internal/app/turnos.go", "line": 3, "severity": "BLOCKER", "causal": true, "why": "devuelve nuevo sin guardar el turno"}
  ]
}`)
	if code, out, errOut := runDir(t, dir, []string{"review", "--attach", verdict}, ""); code != 0 {
		t.Fatalf("attach: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	code, out, errOut := runDir(t, dir, []string{"review", "--refute", verdict}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	for _, want := range []string{
		"# odachi refuter",
		"F1",
		"internal/app/turnos.go:3",
		"Is it causal of the diff, or was it already there?",
		"When in doubt",
		`"refutations"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the refuter prompt does not contain %q", want)
		}
	}
}

// TestRefutedBlockerStopsBlockingEndToEnd walks the whole loop the way a human
// would: a verdict blocks, the refuter is asked, its answer is attached, and the
// gate opens.
func TestRefutedBlockerStopsBlockingEndToEnd(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)
	proof := func(step string) {
		t.Helper()
		if code, out, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
			t.Fatalf("check --run (%s): code = %d (stdout: %q, stderr: %q)", step, code, out, errOut)
		}
	}

	proof("before the verdict")
	// The verdict is written OUTSIDE the project on purpose: a file left inside it
	// would become part of the tree the verdict is named after.
	verdict := filepath.Join(t.TempDir(), "veredicto.json")
	writeAt(t, verdict, `{
  "reviewer": "claude",
`+declared(t, dir)+`  "findings": [
    {"path": "internal/app/turnos.go", "line": 3, "severity": "BLOCKER", "causal": true, "why": "devuelve nuevo sin guardar el turno"}
  ]
}`)
	if code, out, errOut := runDir(t, dir, []string{"review", "--attach", verdict}, ""); code != 0 {
		t.Fatalf("attach: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	proof("after the verdict")

	// The gate blocks, naming the finding and the lens it needs.
	code, _, errOut := runDir(t, dir, []string{"check", "--require-review"}, "")
	if code != 2 {
		t.Fatalf("check --require-review: code = %d, want 2 (stderr: %q)", code, errOut)
	}
	if !strings.Contains(errOut, "internal/app/turnos.go:3") {
		t.Errorf("stderr %q does not name the blocking finding", errOut)
	}

	// The refuter is asked, and its answer is attached.
	code, out, refuteErr := runDir(t, dir, []string{"review", "--refute", verdict}, "")
	if code != 0 {
		t.Fatalf("--refute: code = %d (stderr: %q)", code, refuteErr)
	}
	if !strings.Contains(out, "F1") {
		t.Fatalf("the refuter prompt does not name F1")
	}

	refutation := filepath.Join(t.TempDir(), "refutacion.json")
	writeAt(t, refutation, `{"refuter":"humano","refutations":[{"id":"F1","verdict":"refuted","why":"el diff solo cambia el literal devuelto; el turno nunca se guardaba tampoco antes"}]}`)
	if code, out, errOut := runDir(t, dir, []string{"review", "--attach", refutation}, ""); code != 0 {
		t.Fatalf("attach refutation: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	// The store's ignore file already existed, so the tree did not move: the
	// proof is still valid and the gate is open.
	code, out, errOut = runDir(t, dir, []string{"check", "--require-review"}, "")
	if code != 0 {
		t.Fatalf("check --require-review after the refutation: code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}

	// And --show says so, naming the refutation file.
	code, out, _ = runDir(t, dir, []string{"review", "--show"}, "")
	if code != 0 {
		t.Fatalf("--show: code = %d", code)
	}
	if !strings.Contains(out, ".refuted.json") || !strings.Contains(out, "no longer blocks anything") {
		t.Errorf("--show does not report the refutation:\n%s", out)
	}

	// The refutation is written once.
	if code, _, errOut := runDir(t, dir, []string{"review", "--attach", refutation}, ""); code != 1 {
		t.Fatalf("second attach: code = %d, want 1", code)
	} else if !strings.Contains(errOut, "already has a refutation") {
		t.Errorf("second attach stderr %q does not say the refutation exists", errOut)
	}
}

// TestCheckRequireReviewNamesTheMissingLens: a verdict that does not cover what
// the candidate requires does not answer the gate, and the message says which
// lens is missing and which file to remove to redo it.
func TestCheckRequireReviewNamesTheMissingLens(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)
	proof := func(step string) {
		t.Helper()
		if code, out, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
			t.Fatalf("check --run (%s): code = %d (stdout: %q, stderr: %q)", step, code, out, errOut)
		}
	}

	proof("before the verdict")
	verdict := filepath.Join(t.TempDir(), "veredicto.json")
	writeAt(t, verdict, `{"reviewer":"humano",`+declared(t, dir)+`"findings":[],"passes":[{"path":"internal/app/turnos.go","lens":"reliability"}]}`)
	if code, out, errOut := runDir(t, dir, []string{"review", "--attach", verdict}, ""); code != 0 {
		t.Fatalf("attach: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	proof("after the verdict")

	// A verdict stored before lenses existed does not cover the lens this change
	// requires. It is the same artefact a human can write by hand, so the gate
	// has to name it and say how to redo it.
	blackLensVerdict(t, dir)

	code, _, errOut := runDir(t, dir, []string{"check", "--require-review"}, "")

	if code != 2 {
		t.Fatalf("code = %d, want 2 (stderr: %q)", code, errOut)
	}
	if !strings.Contains(errOut, "reliability") {
		t.Errorf("stderr %q does not name the missing lens", errOut)
	}
	if !strings.Contains(errOut, "delete") || !strings.Contains(errOut, ".json") {
		t.Errorf("stderr %q does not say how to redo the verdict", errOut)
	}
}

// blackLensVerdict rewrites this tree's verdict without lenses, which is what a
// verdict from before T21 looks like. It is the one way to reach the missing-lens
// branch, because the verb itself always stamps the lenses it asked for.
func blackLensVerdict(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(dir, ".git", "odachi", "reviews"))
	if err != nil {
		t.Fatalf("reading the store: %v", err)
	}

	written := false
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") || written {
			continue
		}
		content := `{"tree":"` + strings.TrimSuffix(entry.Name(), ".json") + `","base":"HEAD","reviewer":"humano","at":"2026-10-21T20:31:00Z","findings":[],"verdict":"approved"}`
		writeFile(t, dir, filepath.ToSlash(filepath.Join(".git", "odachi", "reviews", entry.Name())), content)
		written = true
	}
	if !written {
		t.Fatal("no verdict in the store to rewrite")
	}
}

// writeAt writes a file at an absolute path, outside any project: an attached
// document that lives inside the project would change the tree it is attached to.
func writeAt(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
