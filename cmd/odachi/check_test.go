package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/gate"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// checkBrief is a brief complete enough to pass the first two gates.
const checkBrief = `# turnos

## Qué tiene que existir
- reservar turnos

## Listo cuando
- reservo un turno y aparece en la agenda

## NO incluye (primera versión)
- pagos
`

// checkFence forbids anything under payments.
const checkFence = `[[rule]]
not   = "pagos"
paths = ["payment*", "**/payments/**"]
deps  = ["stripe"]
`

// writeFile writes one file, creating its directory.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", name, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// initRepo makes dir a git repository with no commit yet, which is the state a
// project is in right after `git init`. It copies the template TestMain built:
// same repository, without the process per test.
func initRepo(t *testing.T, dir string) {
	t.Helper()
	if gitTemplate != "" {
		if err := os.CopyFS(filepath.Join(dir, ".git"), os.DirFS(gitTemplate)); err != nil {
			t.Fatalf("copy the git template: %v", err)
		}
		return
	}

	cmd := exec.Command("git", "-C", dir, "init", "-q")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
}

// seedProject writes the three artifacts `odachi start` would have written,
// without running the interview.
func seedProject(t *testing.T, dir string) {
	t.Helper()
	writeFile(t, dir, "brief.md", checkBrief)
	writeFile(t, dir, "odd/fence.toml", checkFence)
	writeFile(t, dir, "odd/tasks/0001-turnos.md", taskdoc.TaskDoc{
		Title:        "turnos",
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "reservar turnos"}},
		DoneWhen:     []string{"reservo un turno"},
		History:      []string{`L1 · "reservar turnos"`},
	}.Render())
}

// TestCheckWithoutProofIsBlocked: no proof yet is the agent's work, exit 1, and
// the block names the command that produces it.
func TestCheckWithoutProofIsBlocked(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)

	code, out, errOut := runDir(t, dir, []string{"check"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(errOut, "the proof for this change is missing") {
		t.Errorf("stderr %q does not say the proof is missing", errOut)
	}
	if !strings.Contains(errOut, "odachi check --proof") {
		t.Errorf("stderr %q does not name the continuation", errOut)
	}
}

// TestCheckRunWritesTheProof: --run captures the command and its output, with
// the tree in the header, and the run is ready afterwards.
// TestCheckRunRecordsTheModelOfThisSession is the wiring, not the app: the
// host's answer only matters if the binary asks for it and puts it in the file.
// A proof written without the session's name leaves the identity gate with
// nothing to compare.
func TestCheckRunRecordsTheModelOfThisSession(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)
	t.Setenv("ODACHI_MODEL", "provider/model#variant")

	code, out, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}

	data, err := os.ReadFile(filepath.Join(dir, "proof", "0001-turnos.txt"))
	if err != nil {
		t.Fatalf("the proof was not written: %v", err)
	}
	if !strings.Contains(string(data), "\n# model provider/model#variant\n") {
		t.Errorf("proof =\n%s\nwant the model header naming this session", data)
	}
}

func TestCheckRunWritesTheProof(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)

	code, out, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(out, "ready: there is proof for this tree") {
		t.Errorf("stdout %q does not report the gate as ready", out)
	}
	if !strings.Contains(out, "proof: proof/0001-turnos.txt") {
		t.Errorf("stdout %q does not name the proof file", out)
	}

	data, err := os.ReadFile(filepath.Join(dir, "proof", "0001-turnos.txt"))
	if err != nil {
		t.Fatalf("the proof was not written: %v", err)
	}
	text := string(data)
	if !strings.HasPrefix(text, "# tree ") {
		t.Errorf("proof starts with %q, want the tree header", text)
	}
	if !strings.Contains(text, "\n# command echo ok\n") {
		t.Errorf("proof %q does not carry the command header", text)
	}
	if !strings.Contains(text, "ok") {
		t.Errorf("proof %q does not carry the output", text)
	}
}

// Two runs in a row on the same tree leave the same proof bytes.
func TestTwoCheckRunsProduceTheSameProofBytes(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)

	if code, _, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
		t.Fatalf("first run code = %d (stderr: %q)", code, errOut)
	}
	path := filepath.Join(dir, "proof", "0001-turnos.txt")
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read proof: %v", err)
	}

	if code, _, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
		t.Fatalf("second run code = %d (stderr: %q)", code, errOut)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read proof: %v", err)
	}

	if string(first) != string(second) {
		t.Errorf("proof changed between two runs on the same tree:\n%q\nvs\n%q", first, second)
	}
}

// A failing command is a failed proof, never the command's exit code: `check`
// exits 1 whatever the command exited with.
func TestFailingRunExitsOneAndQuotesTheLastLine(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)

	code, _, errOut := runDir(t, dir, []string{"check", "--run", "echo boom && exit 3"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1: check returns the gate's code, not the command's (stderr: %q)", code, errOut)
	}
	if !strings.Contains(errOut, "the proof failed: boom") {
		t.Errorf("stderr %q does not quote the last line", errOut)
	}
	if !strings.Contains(errOut, `odachi check --run "echo boom && exit 3"`) {
		t.Errorf("stderr %q does not name the same command", errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, "proof", "0001-turnos.txt")); !os.IsNotExist(err) {
		t.Errorf("a failed run wrote a proof: %v", err)
	}
}

// A proof written by hand has no header and proves nothing.
func TestProofWithoutHeaderIsRefused(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)
	writeFile(t, dir, "manual.txt", "ok\n")

	code, _, errOut := runDir(t, dir, []string{"check", "--proof", "manual.txt"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1 (stderr: %q)", code, errOut)
	}
	if !strings.Contains(errOut, "--run") {
		t.Errorf("stderr %q does not name --run", errOut)
	}
}

func TestCheckProofUsesAGivenProof(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)

	if code, _, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
		t.Fatalf("run a proof first: code = %d (stderr: %q)", code, errOut)
	}
	// The copy lives outside the project: a file inside the worktree would
	// change the very tree the proof describes.
	data, err := os.ReadFile(filepath.Join(dir, "proof", "0001-turnos.txt"))
	if err != nil {
		t.Fatalf("read proof: %v", err)
	}
	elsewhere := filepath.Join(t.TempDir(), "proof-copy.txt")
	if err := os.WriteFile(elsewhere, data, 0o644); err != nil {
		t.Fatalf("write proof copy: %v", err)
	}

	code, out, errOut := runDir(t, dir, []string{"check", "--proof", elsewhere}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}
}

// --quiet is what the pre-commit hook depends on: nothing when ready...
func TestQuietPrintsNothingWhenReady(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)
	if code, _, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
		t.Fatalf("run a proof first: code = %d (stderr: %q)", code, errOut)
	}

	code, out, errOut := runDir(t, dir, []string{"check", "--quiet"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stderr: %q)", code, errOut)
	}
	if out != "" || errOut != "" {
		t.Errorf("--quiet printed stdout %q and stderr %q, want nothing", out, errOut)
	}
}

// ...and the block when it is not.
func TestQuietStillPrintsTheBlock(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)

	code, _, errOut := runDir(t, dir, []string{"check", "--quiet"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1 (stderr: %q)", code, errOut)
	}
	if !strings.Contains(errOut, "the proof for this change is missing") {
		t.Errorf("stderr %q is silent about the block", errOut)
	}
}

// The fence runs before the proof: a path the human excluded is a decision only
// they can take, exit 2, even when the proof is now older than the tree.
func TestForbiddenPathNeedsADecision(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)
	if code, _, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
		t.Fatalf("run a proof first: code = %d (stderr: %q)", code, errOut)
	}

	writeFile(t, dir, "internal/payments/stripe.go", "package payments\n")

	code, _, errOut := runDir(t, dir, []string{"check"}, "")

	if code != 2 {
		t.Fatalf("code = %d, want 2 (stderr: %q)", code, errOut)
	}
	if !strings.Contains(errOut, "out of scope") || !strings.Contains(errOut, "pagos") {
		t.Errorf("stderr %q does not report the fence violation", errOut)
	}
}

// A repository with no commit must not fail, and neither must --base "": both
// mean "compare against nothing committed".
func TestNoCommitAndEmptyBaseDoNotFail(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)

	code, _, errOut := runDir(t, dir, []string{"check"}, "")
	if code != 1 {
		t.Fatalf("no commit: code = %d, want 1 (stderr: %q)", code, errOut)
	}
	if strings.Contains(errOut, "could not read the git tree") || strings.Contains(errOut, "could not compare the fence") {
		t.Errorf("no commit: stderr %q reports a git failure", errOut)
	}
	if !strings.Contains(errOut, "the proof for this change is missing") {
		t.Errorf("no commit: stderr %q does not reach the proof gate", errOut)
	}

	code, _, errOut = runDir(t, dir, []string{"check", "--base", ""}, "")
	if code != 1 {
		t.Fatalf(`--base "": code = %d, want 1 (stderr: %q)`, code, errOut)
	}
	if strings.Contains(errOut, "could not compare the fence") {
		t.Errorf(`--base "": stderr %q reports a git failure`, errOut)
	}
}

// Machine output carries the gate's reason and the continuation.
func TestCheckJSONCarriesTheReason(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)

	code, out, errOut := runDir(t, dir, []string{"check", "--json"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1 (stderr: %q)", code, errOut)
	}

	var payload struct {
		Code    int    `json:"code"`
		Reason  string `json:"reason"`
		Message string `json:"message"`
		Next    string `json:"next"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out)
	}
	if payload.Code != 1 || payload.Reason != "proof_missing" {
		t.Errorf("JSON code/reason = %d/%q, want 1/proof_missing", payload.Code, payload.Reason)
	}
	if !strings.HasPrefix(payload.Next, "odachi ") {
		t.Errorf("JSON next = %q, want a command", payload.Next)
	}
}

// TestTheGateJSONContract: the names another program reads -the reasons and the
// keys- are the contract, so they are pinned here by literals. Before this, only
// proof_missing had one and a renamed key or reason kept every suite green.
func TestTheGateJSONContract(t *testing.T) {
	t.Parallel()

	reasons := map[string]string{
		"ready":                        gate.ReasonReady,
		"no_brief":                     gate.ReasonNoBrief,
		"brief_incomplete":             gate.ReasonBriefIncomplete,
		"out_of_scope":                 gate.ReasonOutOfScope,
		"fence_loosened":               gate.ReasonFenceLoosened,
		"open_question":                gate.ReasonOpenQuestion,
		"proof_missing":                gate.ReasonProofMissing,
		"proof_stale":                  gate.ReasonProofStale,
		"proof_failed":                 gate.ReasonProofFailed,
		"review_missing":               check.ReasonReviewMissing,
		"review_changes_requested":     check.ReasonReviewChanges,
		"review_lens_missing":          check.ReasonReviewLensMissing,
		"review_refutation_unreadable": check.ReasonReviewRefutedUnreadable,
		"review_coverage":              check.ReasonReviewCoverage,
		"review_self":                  check.ReasonReviewSelf,
	}
	for want, got := range reasons {
		if got != want {
			t.Errorf("a reason changed: the contract says %q, the code says %q", want, got)
		}
	}

	// A blocked tree: the keys the gate prints, and the continuation.
	blocked := t.TempDir()
	initRepo(t, blocked)
	seedProject(t, blocked)
	code, out, errOut := runDir(t, blocked, []string{"check", "--json"}, "")
	if code != 1 {
		t.Fatalf("check --json on a tree without proof = %d, want 1 (stderr: %q)", code, errOut)
	}
	blockedKeys := gateJSONKeys(t, out)
	for _, key := range []string{"code", "reason", "message", "next", "lines", "files"} {
		if _, ok := blockedKeys[key]; !ok {
			t.Errorf("a blocked answer does not carry %q: %s", key, out)
		}
	}
	if _, ok := blockedKeys["proof"]; ok {
		t.Errorf("a blocked answer carries a proof: %s", out)
	}

	// A ready tree: the same answer as data, with the proof and no continuation.
	ready := t.TempDir()
	initRepo(t, ready)
	seedProject(t, ready)
	if code, out, errOut := runDir(t, ready, []string{"check", "--run", "echo ok"}, ""); code != 0 {
		t.Fatalf("check --run: %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	code, out, errOut = runDir(t, ready, []string{"check", "--json"}, "")
	if code != 0 {
		t.Fatalf("check --json on a ready tree = %d, want 0 (stderr: %q)", code, errOut)
	}
	readyKeys := gateJSONKeys(t, out)
	for _, key := range []string{"code", "reason", "message", "lines", "proof", "files"} {
		if _, ok := readyKeys[key]; !ok {
			t.Errorf("a ready answer does not carry %q: %s", key, out)
		}
	}
	if _, ok := readyKeys["next"]; ok {
		t.Errorf("a ready answer carries a continuation: %s", out)
	}
	var machine struct {
		Code   int    `json:"code"`
		Reason string `json:"reason"`
		Proof  string `json:"proof"`
	}
	if err := json.Unmarshal([]byte(out), &machine); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out)
	}
	if machine.Reason != "ready" || machine.Proof == "" {
		t.Errorf("ready = %+v, want the ready reason and the proof", machine)
	}
}

// gateJSONKeys is the key set of one JSON object, so a vanished field is visible.
func gateJSONKeys(t *testing.T, out string) map[string]json.RawMessage {
	t.Helper()
	var keys map[string]json.RawMessage
	if err := json.Unmarshal([]byte(out), &keys); err != nil {
		t.Fatalf("stdout is not a JSON object: %v\n%s", err, out)
	}
	return keys
}
