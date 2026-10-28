package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/hostmodel"
	"github.com/biggs-100/odachi/internal/domain/layout"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// headOf is the commit the repository is on: the base a feature starts from.
func headOf(t *testing.T, dir string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatalf("git rev-parse HEAD: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// commitAll commits everything in dir, so the next change is a real diff against
// a base. A review needs a base to compare against, and a review of nothing is
// not a review.
func commitAll(t *testing.T, dir, message string) {
	t.Helper()
	commands := [][]string{
		{"add", "--all"},
		{"-c", "user.email=odachi@example.com", "-c", "user.name=odachi", "commit", "-q", "-m", message},
	}
	for _, args := range commands {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, out)
		}
	}
}

// reviewsIn lists the verdict files inside the clone's store.
func reviewsIn(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(dir, ".git", "odachi", "reviews"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("reading the store: %v", err)
	}

	var verdicts []string
	for _, entry := range entries {
		verdicts = append(verdicts, entry.Name())
	}
	return verdicts
}

// changedProject is a project with a real commit and one real change on top of
// it: the state a candidate is in when someone asks for a review.
func changedProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)
	writeFile(t, dir, "internal/app/turnos.go", "package app\n\nfunc Reservar() string { return \"viejo\" }\n")
	commitAll(t, dir, "primer turno")

	writeFile(t, dir, "internal/app/turnos.go", "package app\n\nfunc Reservar() string { return \"nuevo\" }\n")
	return dir
}

// TestReviewPrintsThePromptForTheCandidate: no flags prints the review prompt,
// built from the brief and the diff, with the six rules and the schema.
func TestReviewPrintsThePromptForTheCandidate(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)

	code, out, errOut := runDir(t, dir, []string{"review"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	for _, want := range []string{
		"# odachi review",
		"## The six rules",
		"reservar turnos",                            // the brief
		"internal/app/turnos.go",                     // the changed path
		"not an operator",                            // rule 1: the reviewer reads, it does not run
		"When in doubt, say nothing.",                // rule 2
		"BLOCKER, CRITICAL, WARNING, SUGGESTION",     // rule 3
		"No findings.",                               // rule 4
		"introduced it, activated it or worsened it", // rule 5
		"One pass per lens",                          // rule 6
		"Answer with the JSON only",                  // the answer shape
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the prompt does not contain %q\n%s", want, out)
		}
	}
	if verdicts := reviewsIn(t, dir); len(verdicts) != 0 {
		t.Errorf("printing the prompt stored %v, want nothing", verdicts)
	}
}

// TestReviewWithoutChangesHasNothingToReview: no diff, no prompt, exit 1, and not
// a single file written.
func TestReviewWithoutChangesHasNothingToReview(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)
	commitAll(t, dir, "everything committed")

	code, out, errOut := runDir(t, dir, []string{"review"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(errOut, "there is nothing to review in this tree") {
		t.Errorf("stderr %q does not say there is nothing to review", errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".git", "odachi", "reviews")); !os.IsNotExist(err) {
		t.Errorf("the run created the store without a diff: %v", err)
	}
}

// TestReviewAttachStoresAndShowReads: the file is named after the tree, a second
// verdict for that tree is reported, and --show reads what was stored.
func TestReviewAttachStoresAndShowReads(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)
	// The answer lives outside the project: a document inside it would be part of
	// the change it judges.
	verdict := filepath.Join(t.TempDir(), "veredicto.json")
	writeAt(t, verdict, `{"reviewer":"humano","findings":[],"passes":[{"path":"internal/app/turnos.go","lens":"reliability"}]}`)

	code, out, errOut := runDir(t, dir, []string{"review", "--attach", verdict}, "")
	if code != 0 {
		t.Fatalf("attach: code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}

	verdicts := reviewsIn(t, dir)
	if len(verdicts) != 1 {
		t.Fatalf("stored verdicts = %v, want exactly one", verdicts)
	}
	if !strings.HasSuffix(verdicts[0], ".json") {
		t.Errorf("stored file %q is not a verdict", verdicts[0])
	}

	code, out, errOut = runDir(t, dir, []string{"review", "--show"}, "")
	if code != 0 {
		t.Fatalf("show: code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	for _, want := range []string{"approved", "humano", "no findings"} {
		if !strings.Contains(out, want) {
			t.Errorf("shown review does not contain %q:\n%s", want, out)
		}
	}

	code, _, errOut = runDir(t, dir, []string{"review", "--attach", verdict}, "")
	if code != 1 {
		t.Fatalf("second attach: code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "this tree already has a verdict") {
		t.Errorf("stderr %q does not say the tree already has a verdict", errOut)
	}
}

// TestReviewShowAnotherTreeSaysThereIsNone: approval does not travel between
// trees, and the store's file name is the evidence.
func TestReviewShowAnotherTreeSaysThereIsNone(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)
	verdict := filepath.Join(t.TempDir(), "veredicto.json")
	writeAt(t, verdict, `{"reviewer":"humano","findings":[],"passes":[{"path":"internal/app/turnos.go","lens":"reliability"}]}`)
	if code, out, errOut := runDir(t, dir, []string{"review", "--attach", verdict}, ""); code != 0 {
		t.Fatalf("attach: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	// One real change later, this is a different tree with no verdict.
	writeFile(t, dir, "internal/app/turnos.go", "package app\n\nfunc Reservar() string { return \"otra vez\" }\n")

	code, _, errOut := runDir(t, dir, []string{"review", "--show"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "there is no review for this tree") {
		t.Errorf("stderr %q does not say there is no review", errOut)
	}
}

// TestReviewAttachRefusesGarbage: one Spanish line, exit 1, and no verdict file.
func TestReviewAttachRefusesGarbage(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)
	writeFile(t, dir, "veredicto.json", `{"reviewer":"humano","findings":[{"path":"a.go","severity":"MAJOR","why":"algo"}]}`)

	code, _, errOut := runDir(t, dir, []string{"review", "--attach", "veredicto.json"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "is no good") || !strings.Contains(errOut, "severity") {
		t.Errorf("stderr %q does not name what is wrong with the verdict", errOut)
	}
	if verdicts := reviewsIn(t, dir); len(verdicts) != 0 {
		t.Errorf("stored %v from an invalid verdict, want nothing", verdicts)
	}
}

// TestReviewExecFailureShowsTheOutputAndWritesNothing: a reviewer that fails is
// not a review.
func TestReviewExecFailureShowsTheOutputAndWritesNothing(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)

	code, out, errOut := runDir(t, dir, []string{"review", "--exec", "echo boom-chico && exit 1"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(out+errOut, "boom-chico") {
		t.Errorf("the command's output is not shown (stdout: %q, stderr: %q)", out, errOut)
	}
	if verdicts := reviewsIn(t, dir); len(verdicts) != 0 {
		t.Errorf("stored %v after a failed review, want nothing", verdicts)
	}
}

// TestReviewRefusesAPlaceholderWithoutAModel: {model} with no model behind it
// is a reviewer nobody can name, so the verb stops before it starts one.
func TestReviewRefusesAPlaceholderWithoutAModel(t *testing.T) {
	// The host is this process: cleared so the test does not depend on the
	// agent that runs the suite. It also means this test cannot be parallel.
	t.Setenv(hostmodel.EnvModel, "")
	t.Setenv(hostmodel.SessionEnv, "")

	dir := changedProject(t)
	code, _, errOut := runDir(t, dir, []string{"review", "--exec", "echo {model}"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1 (stderr: %q)", code, errOut)
	}
	for _, want := range []string{"{model}", "--model"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr %q does not name %q", errOut, want)
		}
	}
	if verdicts := reviewsIn(t, dir); len(verdicts) != 0 {
		t.Errorf("stored %v without a model, want nothing", verdicts)
	}
}

// TestReviewModelFlagReplacesThePlaceholder: --model is the override the human
// asked for, and the command that runs is the one carrying the identity.
func TestReviewModelFlagReplacesThePlaceholder(t *testing.T) {
	t.Setenv(hostmodel.EnvModel, "")
	t.Setenv(hostmodel.SessionEnv, "")

	dir := changedProject(t)
	code, out, errOut := runDir(t, dir, []string{"review", "--exec", "echo {model}", "--model", "human/model"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1: an echo is not a verdict (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(out+errOut, "human/model") {
		t.Errorf("the resolved command is not shown (stdout: %q, stderr: %q)", out, errOut)
	}
	if strings.Contains(out+errOut, "{model}") {
		t.Errorf("the placeholder reached the reviewer (stdout: %q, stderr: %q)", out, errOut)
	}
}

// TestReviewFlagsAreExclusive: the three forms answer different questions, and
// asking for two at once is a mistake the CLI names.
func TestReviewFlagsAreExclusive(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)

	code, _, errOut := runDir(t, dir, []string{"review", "--exec", "revisor", "--show"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "do not go together") {
		t.Errorf("stderr %q does not say the flags are exclusive", errOut)
	}
}

// TestCheckRequireReviewInTheThreeSituations walks the optional gate end to end:
// without the flag the review never counts, with the flag a missing verdict is a
// human decision, and an approved one lets the gate through.
func TestCheckRequireReviewInTheThreeSituations(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)

	// A proof for this tree first, so the other gates are satisfied.
	if code, out, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
		t.Fatalf("check --run: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	// 1. Without the flag, a missing verdict does not matter at all.
	if code, out, errOut := runDir(t, dir, []string{"check"}, ""); code != 0 {
		t.Errorf("check without the flag: code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}

	// 2. With the flag and no verdict for this tree: exit 2, next `odachi review`.
	code, _, errOut := runDir(t, dir, []string{"check", "--require-review"}, "")
	if code != 2 {
		t.Fatalf("check --require-review without a verdict: code = %d, want 2 (stderr: %q)", code, errOut)
	}
	if !strings.Contains(errOut, "no review for this tree") || !strings.Contains(errOut, "odachi review") {
		t.Errorf("stderr %q does not name the missing review and the command", errOut)
	}

	// Storing a verdict creates the store's ignore file, so the tree moves once
	// and the proof has to be retaken. The message says exactly that. The verdict
	// says what it looked at: one pass row for the changed file, whose content
	// receipt the tool stamps.
	writeFile(t, dir, "veredicto.json", `{"reviewer":"humano","findings":[],"passes":[{"path":"internal/app/turnos.go","lens":"reliability"},{"path":"veredicto.json","lens":"reliability"}]}`)
	if code, out, errOut := runDir(t, dir, []string{"review", "--attach", "veredicto.json"}, ""); code != 0 {
		t.Fatalf("attach: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if code, out, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
		t.Fatalf("check --run after the store appeared: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	// 3. An approved verdict with a current pass row lets the gate continue.
	code, out, errOut := runDir(t, dir, []string{"check", "--require-review"}, "")
	if code != 0 {
		t.Fatalf("check --require-review with an approved verdict: code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(out, "ready") {
		t.Errorf("stdout %q does not report the gate as ready", out)
	}
}

// TestReviewModeOnMakesCheckRequireTheVerdict: the switch is a project file, so
// a plain `check` reads it; the flags override it for one run.
func TestReviewModeOnMakesCheckRequireTheVerdict(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)

	if code, out, errOut := runDir(t, dir, []string{"review", "mode", "status"}, ""); code != 0 {
		t.Fatalf("mode status: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	} else if !strings.Contains(out, "review: off") {
		t.Errorf("status = %q, want the default off", out)
	}

	if code, out, errOut := runDir(t, dir, []string{"review", "mode", "on"}, ""); code != 0 {
		t.Fatalf("mode on: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(layout.Review)))
	if err != nil {
		t.Fatalf("the switch was not written: %v", err)
	}
	if !strings.Contains(string(data), `mode = "on"`) {
		t.Errorf("switch = %q, want mode = \"on\"", data)
	}

	// The switch is part of the tree, so the proof taken before it is stale; the
	// run that retakes it also asks for the verdict the switch now requires.
	if code, _, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 2 {
		t.Fatalf("check under mode on: code = %d, want 2 (stderr: %q)", code, errOut)
	}
	if code, out, errOut := runDir(t, dir, []string{"check", "--no-review"}, ""); code != 0 {
		t.Fatalf("check --no-review: code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if code, _, errOut := runDir(t, dir, []string{"check", "--require-review"}, ""); code != 2 {
		t.Fatalf("check --require-review: code = %d, want 2 (stderr: %q)", code, errOut)
	}

	if code, out, errOut := runDir(t, dir, []string{"review", "mode", "off"}, ""); code != 0 {
		t.Fatalf("mode off: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if code, out, errOut := runDir(t, dir, []string{"review", "mode", "status"}, ""); code != 0 {
		t.Fatalf("mode status after off: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	} else if !strings.Contains(out, "review: off") {
		t.Errorf("status = %q, want off", out)
	}
}

// TestReviewModeAsksOnceWhenTheFeatureCloses: the CLI reads the switch into the
// mode, and the feature's document decides the unit. An open feature is ready
// with its proof; the closed one owes the verdict of its base, and the message
// names the runnable review.
func TestReviewModeAsksOnceWhenTheFeatureCloses(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)

	// The base the feature starts from is the commit the tree is on now.
	base := headOf(t, dir)
	doc := taskdoc.TaskDoc{
		Title:        "turnos",
		Base:         base,
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "reservar turnos"}},
		DoneWhen:     []string{"reservo un turno"},
		Steps:        []taskdoc.Step{{ID: "T1", Text: "reservar un turno", Requirement: "R1"}},
		History:      []string{`L1 · "reservar turnos"`},
	}
	writeFile(t, dir, "odd/tasks/0001-turnos.md", doc.Render())
	if code, out, errOut := runDir(t, dir, []string{"review", "mode", "on"}, ""); code != 0 {
		t.Fatalf("mode on: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	// An open step: the mode asks nothing, the proof is the gate.
	if code, out, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
		t.Fatalf("an open feature asked for a review: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	// The closed feature owes the verdict that judged its diff, not the commit
	// that closes it: the proof is taken, and the review gate names the base.
	doc.Steps[0].Done = true
	doc.Steps[0].Commit = base
	writeFile(t, dir, "odd/tasks/0001-turnos.md", doc.Render())

	if code, _, _ := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 2 {
		t.Fatalf("check --run on a closed feature = %d, want 2: the proof lands and the review is owed", code)
	}
	code, _, errOut := runDir(t, dir, []string{"check"}, "")
	if code != 2 {
		t.Fatalf("check on a closed feature = %d, want 2 (stderr: %q)", code, errOut)
	}
	if want := "odachi review --base " + base; !strings.Contains(errOut, want) {
		t.Errorf("stderr %q does not name the runnable review %q", errOut, want)
	}
}

// TestCheckAcceptsAVerdictThatNamesTheBaseShort: a base is a commit, not a text.
// A receipt that spells it with seven characters judged exactly the same diff,
// and the gate used to refuse it - after the review was paid for.
func TestCheckAcceptsAVerdictThatNamesTheBaseShort(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)
	base := headOf(t, dir)

	doc := taskdoc.TaskDoc{
		Title:        "turnos",
		Base:         base,
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "reservar turnos"}},
		DoneWhen:     []string{"reservo un turno"},
		Steps:        []taskdoc.Step{{ID: "T1", Text: "reservar un turno", Requirement: "R1", Done: true, Commit: base}},
		History:      []string{`L1 · "reservar turnos"`},
	}
	writeFile(t, dir, "odd/tasks/0001-turnos.md", doc.Render())
	// The same two lines `odachi init` writes: the store and the proof are
	// evidence, not content, so taking a proof cannot change the tree it proves.
	writeFile(t, dir, ".gitignore", "/proof/\n*.bak-*\n")

	// The store arrives with the attach; the proof is taken after it, because a
	// proof is evidence about one tree and the store's own ignore file is part of
	// the project. A short spelling of the base is the same diff.
	verdict := filepath.Join(t.TempDir(), "veredicto.json")
	writeAt(t, verdict, `{"reviewer":"humano","findings":[],"passes":[{"path":"internal/app/turnos.go","lens":"reliability"},{"path":".gitignore","lens":"reliability"}]}`)
	if code, out, errOut := runDir(t, dir, []string{"review", "--attach", verdict, "--base", base[:7]}, ""); code != 0 {
		t.Fatalf("attach with a short base: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	code, out, errOut := runDir(t, dir, []string{"check", "--run", "echo ok", "--require-review"}, "")
	if code != 0 {
		t.Fatalf("a verdict naming the base short was refused: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
}

// TestStartWritesTheFeatureBaseEndToEnd: the wiring that reads HEAD before the
// task doc is written is the whole of "the document records where its feature
// started". Without it every document ships an empty base and the gate silently
// falls back to a verdict per tree, which is the regression the once-per-feature
// gate exists to prevent.
func TestStartWritesTheFeatureBaseEndToEnd(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, dir, "README.md", "# demo\n")
	commitAll(t, dir, "base")
	base := headOf(t, dir)

	code, out, errOut := runDir(t, dir, []string{
		"start", "--force",
		"--outcome", "the base line is recorded",
		"--audience", "cpu",
		"--done", "the document names the commit the feature started from",
		"--exclude", "no UI",
		"--sensitive", "none",
		"--deadline", "today",
		"--learn", "how the base is read",
	}, "")
	if code != 0 {
		t.Fatalf("start: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	path := filepath.Join(dir, "odd", "tasks", "0001-the-base-line.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("the task doc was not written: %v", err)
	}
	doc := taskdoc.Parse(string(data))
	if doc.Base != base {
		t.Errorf("Base = %q, want the commit the feature started from %q", doc.Base, base)
	}
}

// TestReviewModeRejectsAMangledSwitch: a typo in a human file blocks instead of
// silently meaning off.
func TestReviewModeRejectsAMangledSwitch(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, dir, "odd/review.toml", `mode = "maybe"`+"\n")

	code, _, errOut := runDir(t, dir, []string{"check"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1 (stderr: %q)", code, errOut)
	}
	if !strings.Contains(errOut, "review.toml") || !strings.Contains(errOut, "on or off") {
		t.Errorf("stderr %q does not name the file and the valid values", errOut)
	}
}

// TestCheckRequireReviewReportsChangesRequested: a verdict that blocks is a human
// decision, and the sentence names the first finding that causes it.
func TestCheckRequireReviewReportsChangesRequested(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)

	// A proof first, then the verdict, then the proof again: the store's ignore
	// file moves the tree exactly once, when the store first appears. From then on
	// every run agrees, and that is the state this test asserts.
	if code, out, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
		t.Fatalf("check --run: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	// The answer lives outside the project: a document inside it would be part of
	// the change it judges.
	verdict := filepath.Join(t.TempDir(), "veredicto.json")
	writeAt(t, verdict, `{
  "reviewer": "claude",
  "findings": [
    {"path": "internal/app/turnos.go", "line": 3, "severity": "CRITICAL", "causal": true, "lens": "reliability", "why": "devuelve nuevo sin guardar el turno"}
  ]
}`)
	if code, out, errOut := runDir(t, dir, []string{"review", "--attach", verdict}, ""); code != 0 {
		t.Fatalf("attach: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if code, out, errOut := runDir(t, dir, []string{"check", "--run", "echo ok"}, ""); code != 0 {
		t.Fatalf("check --run: code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	code, _, errOut := runDir(t, dir, []string{"check", "--require-review"}, "")

	if code != 2 {
		t.Fatalf("code = %d, want 2 (stderr: %q)", code, errOut)
	}
	for _, want := range []string{"the review asks for changes", "internal/app/turnos.go:3", "CRITICAL", "devuelve nuevo sin guardar el turno"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr %q does not name %q", errOut, want)
		}
	}
}
