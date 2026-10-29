package check_test

import (
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/gate"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// reviewInstant is the instant a seeded verdict records.
var reviewInstant = time.Date(2026, 10, 21, 20, 31, 0, 0, time.UTC)

// The two blobs the git port can answer with: the one these bytes have now, and
// one a stale receipt would name.
const (
	checkBlob = "c1a0f9e8d7b6a5c4d3e2f1a0b9c8d7e6f5a4b3c2"
	otherBlob = "7d3c2b1a9e8f7a6b5c4d3e2f1a0b9c8d7e6f5a4b"
)

// verdictFile seals a verdict for tree and renders the bytes that live on disk.
func verdictFile(t *testing.T, tree string, findings []domainreview.Finding) string {
	t.Helper()
	return verdictFileWithRows(t, tree, findings, nil)
}

// verdictFileWithRows seals a verdict that also carries coverage rows (passes).
func verdictFileWithRows(t *testing.T, tree string, findings []domainreview.Finding, passes []domainreview.Pass) string {
	t.Helper()
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "claude",
		Lenses:   []string{domainreview.Reliability},
		Findings: findings,
		Passes:   passes,
	}, tree, "HEAD", reviewInstant)
	if err != nil {
		t.Fatalf("sealing the seeded verdict: %v", err)
	}
	return string(domainreview.Render(sealed))
}

// reviewPath is where a verdict for tree lives, in the fake's spelling: inside
// the clone's store, not the worktree.
func reviewPath(tree string) string {
	return filepath.Join(store, filepath.FromSlash(domainreview.Dir), tree+".json")
}

// proofPath is this feature's proof, in the fake's spelling.
func proofPath() string { return filepath.Join(dir, "proof", "0007-turnos.txt") }

// reviewableDoc is a task document that records its base and one step, open or
// done: the two states the review gate tells apart.
// reviewableDoc is a feature document; waivedRecord records the human's decision
// in the preamble, which is the only difference between the two halves of
// TestReviewOwedAnswersTheWaiver.
func reviewableDoc(base string, done bool) string {
	return reviewableDocWith(base, done, false)
}

// waivedReviewableDoc is reviewableDoc with "Review: waived" in the preamble.
func waivedReviewableDoc(base string, done bool) string {
	return reviewableDocWith(base, done, true)
}

func reviewableDocWith(base string, done, waived bool) string {
	step := taskdoc.Step{ID: "T1", Text: "reservar un turno", Requirement: "R1"}
	if done {
		step.Done = true
		step.Commit = "abc1234"
	}
	return taskdoc.TaskDoc{
		Title:        "turnos",
		Base:         base,
		ReviewWaived: waived,
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "reservar turnos"}},
		DoneWhen:     []string{"reservo un turno"},
		Steps:        []taskdoc.Step{step},
		History:      []string{`L1 · "reservar turnos"`},
	}.Render()
}

// planOnlyDoc is a feature before its first step: a base and no steps at all,
// which is what `start` writes and the plan commit carries.
func planOnlyDoc(base string) string {
	return taskdoc.TaskDoc{
		Title:        "turnos",
		Base:         base,
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "reservar turnos"}},
		DoneWhen:     []string{"reservo un turno"},
		History:      []string{`L1 · "reservar turnos"`},
	}.Render()
}

// verdictFileAtBase seals a verdict that says which diff it judged.
func verdictFileAtBase(t *testing.T, tree, base string) string {
	t.Helper()
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "claude",
		Lenses:   []string{domainreview.Reliability},
	}, tree, base, reviewInstant)
	if err != nil {
		t.Fatalf("sealing the seeded verdict: %v", err)
	}
	return string(domainreview.Render(sealed))
}

// codeChange is a candidate that touches ordinary code, so the evidence asks for
// the reliability lens and the review gate is on.
func codeChange() *fakeGit {
	return &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		blobs:   map[string]string{"internal/app/check/check.go": checkBlob},
	}
}

// verdictFileByModel seals a verdict that declares which model judged it.
func verdictFileByModel(t *testing.T, tree, model string) string {
	t.Helper()
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "reviewer",
		Model:    model,
		Lenses:   []string{domainreview.Reliability},
	}, tree, "HEAD", reviewInstant)
	if err != nil {
		t.Fatalf("sealing the seeded verdict: %v", err)
	}
	return string(domainreview.Render(sealed))
}

// proofFileByModel seeds this tree's proof with the model that ran it, in the
// format that carries one: the model line exists only before the terminator.
func proofFileByModel(tree, model string) string {
	return "# tree " + tree + "\n# command go test ./...\n# model " + model + "\n# output\nok\n"
}

// TestTheModeAsksOnceWhenTheFeatureCloses: the project mode is a gate per
// feature, not per tree. While a step is open, the code tree is ready with its
// proof; the tree that closes the feature owes the one verdict that judged the
// feature's diff, and the closing commit itself carries only the document.
func TestTheModeAsksOnceWhenTheFeatureCloses(t *testing.T) {
	const base = "8ee3e55f31ee50a088bb1f8aa8bca4afdbc40319"

	open := readyFiles()
	open["odd/tasks/0007-turnos.md"] = reviewableDoc(base, false)
	got := run(t, newFakeFiles(open), codeChange(), check.Options{ReviewMode: true})
	if got.Code != int(gate.Ready) {
		t.Fatalf("an open feature asked for a review: %s (%s)", got.Message, got.Next)
	}

	// The closing tree: the same document with its step done, and a docs-only
	// change. Its own lenses are none, and it still owes the feature's verdict.
	closing := readyFiles()
	closing["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)
	docs := &fakeGit{tree: currentTree, changed: []string{"odd/tasks/0007-turnos.md"}, between: []string{"internal/app/check/check.go", "odd/tasks/0007-turnos.md"}}

	blocked := run(t, newFakeFiles(closing), docs, check.Options{ReviewMode: true})
	if blocked.Code != int(gate.NeedsDecision) || blocked.Reason != check.ReasonReviewMissing {
		t.Fatalf("the closing tree without a verdict = %d/%q, want a missing review", blocked.Code, blocked.Reason)
	}
	if want := "odachi review --base " + base; blocked.Next != want {
		t.Errorf("Next = %q, want the runnable review of the feature %q", blocked.Next, want)
	}

	// A verdict that judged the closing commit instead of the feature does not
	// answer this tree.
	wrong := readyFiles()
	wrong["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)
	wrong[filepath.ToSlash(reviewPath(currentTree))] = verdictFileAtBase(t, currentTree, "HEAD")
	got = run(t, newFakeFiles(wrong), docs, check.Options{ReviewMode: true})
	if got.Code != int(gate.NeedsDecision) || got.Reason != check.ReasonReviewMissing {
		t.Fatalf("a verdict for another diff passed: %d/%q (%s)", got.Code, got.Reason, got.Message)
	}

	// The verdict that names the feature's base is the one it owes.
	right := readyFiles()
	right["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)
	right[filepath.ToSlash(reviewPath(currentTree))] = verdictFileAtBase(t, currentTree, base)
	got = run(t, newFakeFiles(right), docs, check.Options{ReviewMode: true})
	if got.Code != int(gate.Ready) {
		t.Fatalf("the feature's own verdict was refused: %s (%s)", got.Message, got.Next)
	}

	// The same commit, spelled short, answers the same diff. A base is a commit,
	// not a text: comparing the strings made `--base 6aeb877` pay a review that
	// judged exactly these bytes and that the gate then refused.
	short := readyFiles()
	short["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)
	short[filepath.ToSlash(reviewPath(currentTree))] = verdictFileAtBase(t, currentTree, base[:7])
	docs.commits = map[string]string{base[:7]: base, base: base}
	got = run(t, newFakeFiles(short), docs, check.Options{ReviewMode: true})
	if got.Code != int(gate.Ready) {
		t.Fatalf("a verdict naming the same commit short was refused: %s (%s)", got.Message, got.Next)
	}
}

// TestReviewOwedAsksAboutTheTreeThatStands: the question is about this worktree,
// so a caller that names no base gets HEAD, the same default `odachi check` uses.
// An empty base means "compare against nothing" for the fence, and here it would
// turn every file into code: the answer the report got wrong by wiring it so.
func TestReviewOwedAsksAboutTheTreeThatStands(t *testing.T) {
	const base = "c1398cef52c22969cba04a1cbfbf206c1068de2c"
	files := readyFiles()
	files["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)
	git := &fakeGit{
		tree:    currentTree,
		changed: []string{"CHANGELOG.md", "README.md"},
		filesAt: map[string]string{"HEAD:odd/tasks/0007-turnos.md": reviewableDoc(base, true)},
	}

	if _, _, err := check.ReviewOwed(check.Options{Dir: dir, Store: store, ReviewMode: true}, newFakeFiles(files), git); err != nil {
		t.Fatalf("ReviewOwed: %v", err)
	}

	for _, asked := range git.bases {
		if asked == check.EmptyTree {
			t.Errorf("the rule asked about the empty tree; want HEAD, the default check uses (%v)", git.bases)
		}
	}
}

// TestReviewOwedForcedAndItsErrors: the exported question has three ways to fail
// and one way to be forced. A silent "not owed" on any of them would be the report
// telling the human there is nothing to do while the gate is asking - so each one
// is driven here.
func TestReviewOwedForcedAndItsErrors(t *testing.T) {
	const base = "c1398cef52c22969cba04a1cbfbf206c1068de2c"

	t.Run("forced", func(t *testing.T) {
		tail := readyFiles()
		tail["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)
		after := &fakeGit{
			tree:    currentTree,
			changed: []string{"CHANGELOG.md", "README.md"},
			filesAt: map[string]string{"HEAD:odd/tasks/0007-turnos.md": reviewableDoc(base, true)},
		}

		owed, gotBase, err := check.ReviewOwed(check.Options{Dir: dir, Store: store, ReviewMode: true, RequireReview: true}, newFakeFiles(tail), after)
		if err != nil {
			t.Fatalf("ReviewOwed: %v", err)
		}
		if !owed || gotBase != base {
			t.Errorf("ReviewOwed(--require-review) = %v, %q; want the verdict forced at %q", owed, gotBase, base)
		}
	})

	t.Run("the project cannot be listed", func(t *testing.T) {
		files := newFakeFiles(readyFiles())
		files.err = errors.New("permiso denegado")

		if _, _, err := check.ReviewOwed(check.Options{Dir: dir, Store: store, ReviewMode: true}, files, &fakeGit{tree: currentTree}); err == nil {
			t.Error("an unreadable project answered without an error")
		}
	})

	t.Run("the worktree cannot be read", func(t *testing.T) {
		// A feature with no base is not "known": the rule falls through to the
		// lenses, and reading them is where git fails.
		unknown := readyFiles()
		unknown["odd/tasks/0007-turnos.md"] = reviewableDoc("", false)
		git := &fakeGit{tree: currentTree, diffErr: errors.New("git roto")}

		if _, _, err := check.ReviewOwed(check.Options{Dir: dir, Store: store, ReviewMode: true}, newFakeFiles(unknown), git); err == nil {
			t.Error("an unreadable worktree answered without an error")
		}
	})

	t.Run("the document at HEAD cannot be read", func(t *testing.T) {
		closed := readyFiles()
		closed["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)
		git := &fakeGit{
			tree:      currentTree,
			changed:   []string{"odd/tasks/0007-turnos.md"},
			between:   []string{"internal/app/check/check.go", "odd/tasks/0007-turnos.md"},
			fileAtErr: errors.New("git roto"),
		}

		if _, _, err := check.ReviewOwed(check.Options{Dir: dir, Store: store, ReviewMode: true}, newFakeFiles(closed), git); err == nil {
			t.Error("an unreadable HEAD answered without an error")
		}
	})
}

// TestReviewOwedAgreesWithTheGate: the exported question answers what the gate
// answers, for the same tree. That shared answer is the whole reason it is
// exported: a report that asks it cannot disagree with the gate about a tree.
func TestReviewOwedAgreesWithTheGate(t *testing.T) {
	const base = "c1398cef52c22969cba04a1cbfbf206c1068de2c"

	// The release tail: a documents-only change after the close, with no verdict.
	tail := readyFiles()
	tail["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)
	after := &fakeGit{
		tree:    currentTree,
		changed: []string{"CHANGELOG.md", "README.md"},
		filesAt: map[string]string{"HEAD:odd/tasks/0007-turnos.md": reviewableDoc(base, true)},
	}
	owed, gotBase, err := check.ReviewOwed(check.Options{Dir: dir, Store: store, ReviewMode: true}, newFakeFiles(tail), after)
	if err != nil {
		t.Fatalf("ReviewOwed on the release tail: %v", err)
	}
	if owed || gotBase != "" {
		t.Errorf("ReviewOwed(release tail) = %v, %q; want nothing owed", owed, gotBase)
	}
	if got := run(t, newFakeFiles(tail), after, check.Options{ReviewMode: true}); got.Code != int(gate.Ready) {
		t.Fatalf("the gate and the exported answer disagree: gate = %d %q", got.Code, got.Message)
	}

	// The close itself: the document that marked the last step done.
	closing := readyFiles()
	closing["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)
	atClose := &fakeGit{
		tree:    currentTree,
		changed: []string{"odd/tasks/0007-turnos.md"},
		between: []string{"internal/app/check/check.go", "odd/tasks/0007-turnos.md"},
		filesAt: map[string]string{"HEAD:odd/tasks/0007-turnos.md": reviewableDoc(base, false)},
	}
	owed, gotBase, err = check.ReviewOwed(check.Options{Dir: dir, Store: store, ReviewMode: true}, newFakeFiles(closing), atClose)
	if err != nil {
		t.Fatalf("ReviewOwed at the close: %v", err)
	}
	if !owed || gotBase != base {
		t.Errorf("ReviewOwed(close) = %v, %q; want owed at %q", owed, gotBase, base)
	}
	if got := run(t, newFakeFiles(closing), atClose, check.Options{ReviewMode: true}); got.Code == int(gate.Ready) {
		t.Fatalf("the gate asked for no verdict while the exported answer says it is owed")
	}
}

// TestReviewOwedAnswersTheWaiver: the exported question is what the report asks,
// and its waiver branch was new behavior no test drove. Both halves run on the
// SAME tree and the SAME git -a feature whose closing step is this change-, so the
// only difference is the line in the preamble: a waived close owes nothing, and the
// same document without the line is owed at its base. Wiring the halves differently
// would let the branch disappear and the test still pass, which is the defect this
// test exists to prevent.
//
// It is a characterization test: it was observed red by removing the branch before
// it was committed.
func TestReviewOwedAnswersTheWaiver(t *testing.T) {
	const base = "c1398cef52c22969cba04a1cbfbf206c1068de2c"

	// The close: the worktree document marks the last step done, and HEAD still has
	// it open. That is the tree where the gate owes the feature's verdict.
	atClose := &fakeGit{
		tree:    currentTree,
		changed: []string{"odd/tasks/0007-turnos.md"},
		between: []string{"internal/app/check/check.go", "odd/tasks/0007-turnos.md"},
		filesAt: map[string]string{"HEAD:odd/tasks/0007-turnos.md": reviewableDoc(base, false)},
	}

	waived := readyFiles()
	waived["odd/tasks/0007-turnos.md"] = waivedReviewableDoc(base, true)
	owed, gotBase, err := check.ReviewOwed(check.Options{Dir: dir, Store: store, ReviewMode: true}, newFakeFiles(waived), atClose)
	if err != nil {
		t.Fatalf("ReviewOwed on the waived close: %v", err)
	}
	if owed || gotBase != "" {
		t.Errorf("ReviewOwed(waived close) = %v, %q; want nothing owed", owed, gotBase)
	}

	plain := readyFiles()
	plain["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)
	owed, gotBase, err = check.ReviewOwed(check.Options{Dir: dir, Store: store, ReviewMode: true}, newFakeFiles(plain), atClose)
	if err != nil {
		t.Fatalf("ReviewOwed on the same close without the waiver: %v", err)
	}
	if !owed || gotBase != base {
		t.Errorf("ReviewOwed(close without the waiver) = %v, %q; want owed at %q", owed, gotBase, base)
	}
}

// TestThePlanCommitIsAFeatureThatHasNotStarted: a document with a base and no
// steps is a feature before its first step, and the plan commit is the tree that
// writes the fence. check called it a stranger and asked for a verdict before any
// behaviour existed - the fence is ordinary code on purpose, so the plan was a
// candidate - and the only way to land it was --no-verify. The same document with
// its one step done still owes the feature's verdict, so the pair stays whole.
func TestThePlanCommitIsAFeatureThatHasNotStarted(t *testing.T) {
	const base = "6f1c2d84a5e0b93c7d2f4a8b1e6c0d5f3a9b7c2e"

	plan := readyFiles()
	plan["odd/tasks/0007-turnos.md"] = planOnlyDoc(base)
	// The plan commit writes the fence, which is the path that made this a code
	// candidate in the first place.
	git := &fakeGit{tree: currentTree, changed: []string{"odd/fence.toml", "odd/tasks/0007-turnos.md"}}

	got := run(t, newFakeFiles(plan), git, check.Options{ReviewMode: true})
	if got.Code != int(gate.Ready) {
		t.Fatalf("the plan commit asked for a review: %s (%s)", got.Message, got.Next)
	}

	// The question the report asks answers the same thing the gate answers.
	owed, _, err := check.ReviewOwed(check.Options{Dir: dir, Store: store, ReviewMode: true}, newFakeFiles(plan), git)
	if err != nil {
		t.Fatalf("ReviewOwed: %v", err)
	}
	if owed {
		t.Error("ReviewOwed said the plan owes a verdict")
	}

	// The close of the same feature still owes it, at the document's base.
	closing := readyFiles()
	closing["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)
	docs := &fakeGit{tree: currentTree, changed: []string{"odd/tasks/0007-turnos.md"}, between: []string{"internal/app/check/check.go", "odd/tasks/0007-turnos.md"}}
	blocked := run(t, newFakeFiles(closing), docs, check.Options{ReviewMode: true})
	if blocked.Code != int(gate.NeedsDecision) || blocked.Next != "odachi review --base "+base {
		t.Fatalf("the close of the same feature = %d/%q (%q), want the verdict at %q",
			blocked.Code, blocked.Reason, blocked.Next, base)
	}
}

// TestTheReleaseTailDoesNotOweTheFeatureVerdict: after the close, a documents-
// only change -the release notes- is paperwork. The verdict it would ask for
// already judged the same code, so the gate leaves it with its proof.
func TestTheReleaseTailDoesNotOweTheFeatureVerdict(t *testing.T) {
	const base = "c1398cef52c22969cba04a1cbfbf206c1068de2c"
	files := readyFiles()
	files["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)
	docs := &fakeGit{
		tree:    currentTree,
		changed: []string{"CHANGELOG.md", "README.md"},
		filesAt: map[string]string{"HEAD:odd/tasks/0007-turnos.md": reviewableDoc(base, true)},
	}

	got := run(t, newFakeFiles(files), docs, check.Options{ReviewMode: true})

	if got.Code != int(gate.Ready) {
		t.Fatalf("the release tail asked for the feature's review: %s (%s)", got.Message, got.Next)
	}
}

// TestTheClosingCommitOwesTheFeatureVerdict: the change that marked the last step
// done is the close, even though it only carries the document. That is the tree
// the one verdict is owed by.
func TestTheClosingCommitOwesTheFeatureVerdict(t *testing.T) {
	const base = "c1398cef52c22969cba04a1cbfbf206c1068de2c"
	files := readyFiles()
	files["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)
	closing := &fakeGit{
		tree:    currentTree,
		changed: []string{"odd/tasks/0007-turnos.md"},
		between: []string{"internal/app/check/check.go", "odd/tasks/0007-turnos.md"},
		filesAt: map[string]string{"HEAD:odd/tasks/0007-turnos.md": reviewableDoc(base, false)},
	}

	got := run(t, newFakeFiles(files), closing, check.Options{ReviewMode: true})

	if got.Code != int(gate.NeedsDecision) || got.Reason != check.ReasonReviewMissing {
		t.Fatalf("the close = %d/%q, want the feature's review owed", got.Code, got.Reason)
	}
	if want := "odachi review --base " + base; got.Next != want {
		t.Errorf("Next = %q, want %q", got.Next, want)
	}
}

// TestCodeAfterTheCloseStillOwesAVerdict: a document written after the close is
// paperwork; code written after it is a change nobody reviewed. The gate must not
// go blind just because the feature is closed.
func TestCodeAfterTheCloseStillOwesAVerdict(t *testing.T) {
	const base = "c1398cef52c22969cba04a1cbfbf206c1068de2c"
	files := readyFiles()
	files["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)
	code := &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		filesAt: map[string]string{"HEAD:odd/tasks/0007-turnos.md": reviewableDoc(base, true)},
	}

	got := run(t, newFakeFiles(files), code, check.Options{ReviewMode: true})

	if got.Code != int(gate.NeedsDecision) || got.Reason != check.ReasonReviewMissing {
		t.Fatalf("code after the close = %d/%q, want a review owed", got.Code, got.Reason)
	}
}

// TestAQuestionAnsweredAfterTheCloseDoesNotReopenTheGate: the close already
// happened in this document's history; adding a line to it does not owe the
// verdict again, or every note after a feature would cost a review.
func TestAQuestionAnsweredAfterTheCloseDoesNotReopenTheGate(t *testing.T) {
	const base = "c1398cef52c22969cba04a1cbfbf206c1068de2c"
	closed := reviewableDoc(base, true)
	files := readyFiles()
	files["odd/tasks/0007-turnos.md"] = strings.Replace(closed, "L1 · \"reservar turnos\"", "L1 · \"reservar turnos\"\nL2 · OFICIO · \"una respuesta\"", 1)
	tail := &fakeGit{
		tree:    currentTree,
		changed: []string{"odd/tasks/0007-turnos.md"},
		between: []string{"internal/app/check/check.go", "odd/tasks/0007-turnos.md"},
		filesAt: map[string]string{"HEAD:odd/tasks/0007-turnos.md": closed},
	}

	got := run(t, newFakeFiles(files), tail, check.Options{ReviewMode: true})

	if got.Code != int(gate.Ready) {
		t.Fatalf("an answered question reopened the gate: %s (%s)", got.Message, got.Next)
	}
}

// TestTheModeAsksEveryTreeWithoutABase: a document that predates the base line
// has no unit to review once, so the mode keeps the old per-tree rule.
func TestTheModeAsksEveryTreeWithoutABase(t *testing.T) {
	old := readyFiles()
	old["odd/tasks/0007-turnos.md"] = reviewableDoc("", true)

	got := run(t, newFakeFiles(old), codeChange(), check.Options{ReviewMode: true})
	if got.Code != int(gate.NeedsDecision) || got.Reason != check.ReasonReviewMissing {
		t.Fatalf("a document with no base = %d/%q, want the old per-tree rule", got.Code, got.Reason)
	}
}

// TestNoReviewBeatsTheMode: --no-review turns the project mode off for one run,
// including the closed feature's verdict.
func TestNoReviewBeatsTheMode(t *testing.T) {
	const base = "8ee3e55f31ee50a088bb1f8aa8bca4afdbc40319"
	closing := readyFiles()
	closing["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)

	got := run(t, newFakeFiles(closing), &fakeGit{tree: currentTree, changed: []string{"a.go"}}, check.Options{ReviewMode: true, NoReview: true})
	if got.Code != int(gate.Ready) {
		t.Fatalf("--no-review did not turn the mode off: %s (%s)", got.Message, got.Next)
	}
}

// TestAReleaseNoteIsPaperwork: after a feature closed, the commit that only
// rewrites documents -the release note- is a new tree with no new behaviour, and
// the proof that judged the code still speaks for it; the suite runs again where
// the change is judged, which is the pull request. Every release of this project
// landed with --no-verify before this rule.
func TestAReleaseNoteIsPaperwork(t *testing.T) {
	const base = "8ee3e55f31ee50a088bb1f8aa8bca4afdbc40319"
	files := readyFiles()
	files["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)
	files["proof/0007-turnos.txt"] = proofFile(otherTree, "go test ./...", "ok\n")
	closedAtHead := map[string]string{"HEAD:odd/tasks/0007-turnos.md": reviewableDoc(base, true)}

	paper := &fakeGit{tree: currentTree, between: []string{"CHANGELOG.md", "docs/notes.md"}, filesAt: closedAtHead}
	got := run(t, newFakeFiles(files), paper, check.Options{})
	if got.Code != int(gate.Ready) {
		t.Fatalf("the release note asked for the suite: %s (%s)", got.Message, got.Next)
	}
	lines := strings.Join(got.Lines, "\n")
	if !strings.Contains(lines, otherTree) || !strings.Contains(lines, "CHANGELOG.md") {
		t.Errorf("lines = %v, want the tree the proof belongs to and what moved", got.Lines)
	}

	// A path that is not a document keeps the proof owed.
	code := &fakeGit{tree: currentTree, between: []string{"CHANGELOG.md", "internal/app/check/check.go"}, filesAt: closedAtHead}
	if got := run(t, newFakeFiles(files), code, check.Options{}); got.Reason != gate.ReasonProofStale {
		t.Fatalf("a code change passed as paperwork: %d/%q", got.Code, got.Reason)
	}

	// A Markdown compiled into the binary is shipped text, not paperwork: the suite
	// compares those bytes. One case per prefix the lane refuses, because a dropped
	// prefix is a silent skip of the suite.
	for _, shippedPath := range []string{"skills/drive/SKILL.md", "internal/assets/review/lens_readability.md", ".agents/skills/drive/SKILL.md"} {
		shipped := &fakeGit{tree: currentTree, between: []string{shippedPath}, filesAt: closedAtHead}
		if got := run(t, newFakeFiles(files), shipped, check.Options{}); got.Reason != gate.ReasonProofStale {
			t.Fatalf("a shipped Markdown passed as paperwork: %s (%d/%q)", shippedPath, got.Code, got.Reason)
		}
	}

	// The tree that closed the feature still owes the proof: the document at HEAD
	// was open, so this is the close itself, not the paperwork after it.
	closing := &fakeGit{tree: currentTree, between: []string{"CHANGELOG.md"},
		filesAt: map[string]string{"HEAD:odd/tasks/0007-turnos.md": reviewableDoc(base, false)}}
	if got := run(t, newFakeFiles(files), closing, check.Options{}); got.Reason != gate.ReasonProofStale {
		t.Fatalf("the closing tree passed as paperwork: %d/%q", got.Code, got.Reason)
	}
}

// TestAStaleProofNamesWhatChanged: "from another state of the code" is true and
// useless. A stray backup, an editor and a real edit look identical in one line,
// so the gate names the paths that moved between the proof's tree and this one.
func TestAStaleProofNamesWhatChanged(t *testing.T) {
	files := closedFiles()
	files["proof/0007-turnos.txt"] = proofFile(otherTree, "go test ./...", "ok\n")
	git := &fakeGit{tree: currentTree, between: []string{"stray.bak"}}

	got := run(t, newFakeFiles(files), git, check.Options{})

	if got.Code != int(gate.NeedsWork) || got.Reason != gate.ReasonProofStale {
		t.Fatalf("state = %d/%q, want a stale proof", got.Code, got.Reason)
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), "changed since the proof: stray.bak") {
		t.Errorf("lines = %v, want the path that moved", got.Lines)
	}
	if git.betweenCalls != 1 {
		t.Errorf("Between calls = %d, want one question", git.betweenCalls)
	}

	// A current proof does not ask: the question is only for the stale case.
	fresh := closedFiles()
	freshGit := &fakeGit{tree: currentTree, between: []string{"noise"}}
	if got := run(t, newFakeFiles(fresh), freshGit, check.Options{}); got.Code != int(gate.Ready) {
		t.Fatalf("a current proof = %d (%s), want ready", got.Code, got.Message)
	}
	if freshGit.betweenCalls != 0 {
		t.Errorf("Between calls = %d for a current proof, want none", freshGit.betweenCalls)
	}
}

// TestAStaleProofWithoutAPathKeepsTheGenericMessage: if git can no longer read
// the proof's tree, the message stays as it was: a second failure is noise.
func TestAStaleProofWithoutAPathKeepsTheGenericMessage(t *testing.T) {
	files := closedFiles()
	files["proof/0007-turnos.txt"] = proofFile(otherTree, "go test ./...", "ok\n")
	git := &fakeGit{tree: currentTree, betweenErr: errors.New("fatal: bad object")}

	got := run(t, newFakeFiles(files), git, check.Options{})

	if got.Code != int(gate.NeedsWork) || got.Reason != gate.ReasonProofStale {
		t.Fatalf("state = %d/%q, want a stale proof", got.Code, got.Reason)
	}
	if strings.Contains(strings.Join(got.Lines, "\n"), "changed since the proof") {
		t.Errorf("lines = %v, want no note when git cannot name the paths", got.Lines)
	}
}

// TestRequireReviewRefusesASelfReview: the proof and the verdict name the same
// model, so the tree is approving itself. The gate names the identity and the
// command that fixes it, because "find another reviewer" is not a next step.
func TestRequireReviewRefusesASelfReview(t *testing.T) {
	cases := []struct{ name, proof, verdict string }{
		{
			name:    "the same model",
			proof:   proofFileByModel(currentTree, "provider/model#variant"),
			verdict: verdictFileByModel(t, currentTree, "provider/model#variant"),
		},
		{
			// A self-review must not slip through by spelling the name differently:
			// model ids are lowercase in the wild, and case is not a second model.
			name:    "the same model in another case",
			proof:   proofFileByModel(currentTree, "Provider/Model"),
			verdict: verdictFileByModel(t, currentTree, "provider/model"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files := newFakeFiles(readyFiles())
			files.files[proofPath()] = []byte(tc.proof)
			files.files[reviewPath(currentTree)] = []byte(tc.verdict)

			got := check.Run(check.Options{Dir: dir, Store: store, Base: "HEAD", TimeoutSeconds: 120, RequireReview: true},
				files, codeChange(), &fakeShell{})

			if got.Reason != check.ReasonReviewSelf {
				t.Fatalf("Reason = %q, want %q (message: %s)", got.Reason, check.ReasonReviewSelf, got.Message)
			}
			lines := strings.Join(got.Lines, "\n")
			for _, want := range []string{"provider/model", "odachi review --model"} {
				if !strings.Contains(strings.ToLower(lines), want) {
					t.Errorf("the refusal does not carry %q:\n%s", want, lines)
				}
			}
		})
	}
}

// TestRequireReviewRefusesAVerdictThatSaysNobodyJudged: the identity gate compares
// two declarations, and a silence is the cheapest way past it. A verdict that names
// no model while the proof names one is refused, and the refusal offers the two
// honest roads: name the model that judged, or record the human's waiver.
func TestRequireReviewRefusesAVerdictThatSaysNobodyJudged(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[proofPath()] = []byte(proofFileByModel(currentTree, "provider/builder"))
	files.files[reviewPath(currentTree)] = []byte(verdictFileByModel(t, currentTree, ""))

	got := check.Run(check.Options{Dir: dir, Store: store, Base: "HEAD", TimeoutSeconds: 120, RequireReview: true},
		files, codeChange(), &fakeShell{})

	if got.Reason != check.ReasonReviewUnattributed {
		t.Fatalf("Reason = %q, want %q (message: %s)", got.Reason, check.ReasonReviewUnattributed, got.Message)
	}
	lines := strings.Join(got.Lines, "\n")
	// The tree, the lens, the model the proof carries and the two roads out: the
	// refusal has to name the file it is about, the way lensMissing does, or a
	// project with more than one tree cannot tell which review it read.
	for _, want := range []string{currentTree, "provider/builder", "odachi review --model", "waive-review"} {
		if !strings.Contains(strings.ToLower(lines), want) {
			t.Errorf("the refusal does not carry %q:\n%s", want, lines)
		}
	}
}

// TestRequireReviewRefusesOnlyTheLensesTheBuilderJudged: with a model per lens,
// the gate names the lenses that are self-reviewed and leaves the others alone.
func TestRequireReviewRefusesOnlyTheLensesTheBuilderJudged(t *testing.T) {
	cases := []struct {
		name       string
		models     map[string]string
		wantSelf   []string
		wantReason string
	}{
		{
			name:     "the builder judged one lens",
			models:   map[string]string{"reliability": "provider/builder", "risk": "provider/other"},
			wantSelf: []string{"reliability"},
		},
		{
			name:   "the builder judged none",
			models: map[string]string{"reliability": "provider/other", "risk": "provider/builder"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files := newFakeFiles(readyFiles())
			files.files[proofPath()] = []byte(proofFileByModel(currentTree, "provider/builder"))

			sealed, err := domainreview.Seal(domainreview.Verdict{
				Reviewer: "reviewer",
				Lenses:   []string{"reliability"},
				Models:   tc.models,
			}, currentTree, "HEAD", reviewInstant)
			if err != nil {
				t.Fatalf("sealing the seeded verdict: %v", err)
			}
			files.files[reviewPath(currentTree)] = []byte(domainreview.Render(sealed))

			got := check.Run(check.Options{Dir: dir, Store: store, Base: "HEAD", TimeoutSeconds: 120, RequireReview: true},
				files, codeChange(), &fakeShell{})

			self := got.Reason == check.ReasonReviewSelf
			if len(tc.wantSelf) > 0 {
				if !self {
					t.Fatalf("Reason = %q, want %q (message: %s)", got.Reason, check.ReasonReviewSelf, got.Message)
				}
				for _, lens := range tc.wantSelf {
					if !strings.Contains(got.Message, lens) {
						t.Errorf("the refusal does not name the lens %q: %s", lens, got.Message)
					}
				}
				return
			}
			if self {
				t.Fatalf("the gate refused a lens another model judged: %s", got.Message)
			}
		})
	}
}

// A review by another model, or one that cannot be attributed (a human wrote it,
// or the proof predates the field, is not a self-review.
func TestRequireReviewAcceptsAReviewByAnotherModelOrByNobody(t *testing.T) {
	cases := []struct {
		name    string
		proof   string
		verdict string
	}{
		{"another model", proofFileByModel(currentTree, "provider/builder"), verdictFileByModel(t, currentTree, "provider/reviewer")},
		{"proof without a model", "# tree " + currentTree + "\n# command go test ./...\nok\n", verdictFileByModel(t, currentTree, "provider/reviewer")},
		{"verdict without a model", proofFileByModel(currentTree, "provider/builder"), verdictFile(t, currentTree, nil)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files := newFakeFiles(readyFiles())
			files.files[proofPath()] = []byte(tc.proof)
			files.files[reviewPath(currentTree)] = []byte(tc.verdict)

			got := check.Run(check.Options{Dir: dir, Store: store, Base: "HEAD", TimeoutSeconds: 120, RequireReview: true},
				files, codeChange(), &fakeShell{})

			if got.Reason == check.ReasonReviewSelf {
				t.Fatalf("the gate refused a review it cannot attribute to the builder: %s", got.Message)
			}
		})
	}
}

// filesWithVerdict is a ready project plus one stored verdict.
func filesWithVerdict(t *testing.T, tree string, findings []domainreview.Finding) *fakeFiles {
	t.Helper()
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(tree)] = []byte(verdictFile(t, tree, findings))
	return files
}

// TestReviewDoesNotInfluenceTheGateWithoutTheFlag is the opt-in rule: a beginner
// who has never heard of a review must not be blocked by one.
func TestReviewDoesNotInfluenceTheGateWithoutTheFlag(t *testing.T) {
	files := filesWithVerdict(t, currentTree, []domainreview.Finding{
		{Path: "a.go", Severity: domainreview.Blocker, Causal: true, Why: "pierde el turno"},
	})

	result := run(t, files, &fakeGit{tree: currentTree}, check.Options{})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0: without the flag a review never blocks (%s)", result.Code, result.Message)
	}
	if result.Reason != "ready" {
		t.Errorf("reason = %q, want the gate's own ready", result.Reason)
	}
}

// concurrentCheckDiff is content the path class cannot see: a goroutine and a
// mutex in an otherwise ordinary file.
const concurrentCheckDiff = "diff --git a/internal/app/check/check.go b/internal/app/check/check.go\n--- a/internal/app/check/check.go\n+++ b/internal/app/check/check.go\n@@ -1 +1,2 @@\n package check\n+go func() { sync.Mutex{} }()\n"

// TestRequireReviewAsksTheRaisedClass: a changed file whose added lines spawn
// concurrency asks the four lenses. A verdict that covers only the path class is
// missing the rest, and the gate names it.
func TestRequireReviewAsksTheRaisedClass(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictFileWithRows(t, currentTree, nil, []domainreview.Pass{
		{Path: "internal/app/check/check.go", Blob: checkBlob, Lens: domainreview.Reliability},
	}))

	result := run(t, files, &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		diff:    concurrentCheckDiff,
		blobs:   map[string]string{"internal/app/check/check.go": checkBlob},
	}, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	if result.Reason != "review_lens_missing" {
		t.Errorf("reason = %q, want review_lens_missing", result.Reason)
	}
	if !strings.Contains(result.Message, "risk") {
		t.Errorf("message = %q, want the first raised lens named", result.Message)
	}
}

// TestRequireReviewNeedsTheContentRows: declaring the lens is not enough; the
// rows are per file. A verdict that claims the four lenses but answers only the
// path-class row of the concurrent file has not covered it.
func TestRequireReviewNeedsTheContentRows(t *testing.T) {
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "claude",
		Lenses:   []string{domainreview.Risk, domainreview.Readability, domainreview.Reliability, domainreview.Resilience},
		Findings: []domainreview.Finding{},
		Passes: []domainreview.Pass{
			{Path: "internal/app/check/check.go", Blob: checkBlob, Lens: domainreview.Reliability},
		},
	}, currentTree, "HEAD", reviewInstant)
	if err != nil {
		t.Fatalf("sealing the seeded verdict: %v", err)
	}

	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = domainreview.Render(sealed)

	result := run(t, files, &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		diff:    concurrentCheckDiff,
		blobs:   map[string]string{"internal/app/check/check.go": checkBlob},
	}, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	if result.Reason != "review_coverage" {
		t.Errorf("reason = %q, want review_coverage", result.Reason)
	}
	if !strings.Contains(result.Message, "risk") {
		t.Errorf("message = %q, want the missing raised lens named", result.Message)
	}
}

// TestRequireReviewBlocksWhenTheDiffCannotBeRead: the class cannot be measured
// without the diff, and a measurement nobody can make is not a pass.
func TestRequireReviewBlocksWhenTheDiffCannotBeRead(t *testing.T) {
	files := filesWithVerdict(t, currentTree, nil)

	result := run(t, files, &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		diffErr: errors.New("fatal: bad revision"),
	}, check.Options{RequireReview: true})

	if result.Code == 0 {
		t.Fatal("a diff failure must not read as an assessed class")
	}
	if !strings.Contains(result.Message, "could not read") {
		t.Errorf("message = %q, want the read failure named", result.Message)
	}
}

// TestCheckReadsTheWorktreeOnce: the fence, the lens class and the proof tree
// describe the same bytes. One read answers all three, however the gate ends.
func TestCheckReadsTheWorktreeOnce(t *testing.T) {
	files := filesWithVerdict(t, currentTree, nil)
	git := &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		diff:    "diff --git a/a.go b/a.go\n--- a/a.go\n+++ b/a.go\n@@ -1 +1 @@\n-viejo\n+nuevo\n",
	}

	run(t, files, git, check.Options{RequireReview: true})

	if git.candidateCalls != 1 {
		t.Errorf("candidate reads = %d, want 1: check asks the worktree once", git.candidateCalls)
	}
}

// TestADismissedFindingDoesNotBlockTheGate: a dismissal was judged with
// evidence in its verdict. The gate records it and lets the tree through — the
// receipt still covers the row, the severity no longer blocks.
func TestADismissedFindingDoesNotBlockTheGate(t *testing.T) {
	dismissed := domainreview.Finding{
		Path: "internal/app/check/check.go", Line: 12, Severity: domainreview.Critical, Causal: true,
		Why: "pierde el turno", Lens: domainreview.Reliability, Blob: checkBlob,
		Status: domainreview.Dismissed, Reason: "la transaccion lo cubre, la cita lo prueba",
	}
	dismissed.Fingerprint = domainreview.Fingerprint(dismissed)

	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictFileWithRows(t, currentTree, []domainreview.Finding{dismissed}, nil))

	result := run(t, files, &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		blobs:   map[string]string{"internal/app/check/check.go": checkBlob},
	}, check.Options{RequireReview: true})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0: a dismissal never blocks (%s)", result.Code, result.Message)
	}
}

// TestRequireReviewWithoutAVerdictBlocks: the flag is on, this tree has no
// verdict, so a human decision is missing: exit 2, and the continuation is the
// command that produces one.
func TestRequireReviewWithoutAVerdictBlocks(t *testing.T) {
	files := newFakeFiles(readyFiles())

	result := run(t, files, &fakeGit{tree: currentTree, changed: []string{"internal/app/check/check.go"}}, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	if result.Message != "no review for this tree" {
		t.Errorf("message = %q, want the documented sentence", result.Message)
	}
	if result.Next != "odachi review" {
		t.Errorf("next = %q, want odachi review", result.Next)
	}
	if len(files.writes) != 0 {
		t.Fatalf("writes = %v, want none: the gate writes nothing", files.writes)
	}
}

// TestRequireReviewWithApprovedContinues continues only when the verdict can
// serve coverage: a current answer — a pass row whose blob is exactly what git
// measures now — for every lens the candidate's files ask for.
func TestRequireReviewWithApprovedContinues(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictFileWithRows(t, currentTree, nil, []domainreview.Pass{
		{Path: "internal/app/check/check.go", Blob: checkBlob, Lens: domainreview.Reliability},
	}))

	result := run(t, files, &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		blobs:   map[string]string{"internal/app/check/check.go": checkBlob},
	}, check.Options{RequireReview: true})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if result.Reason != "ready" {
		t.Errorf("reason = %q, want ready", result.Reason)
	}
}

// TestRequireReviewRefusesReceiptsForOtherBytes: a pass whose blob describes
// other bytes is the potency boundary — the file changed and the old answer is
// dead. The gate says so instead of reusing it.
func TestRequireReviewRefusesReceiptsForOtherBytes(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictFileWithRows(t, currentTree, nil, []domainreview.Pass{
		{Path: "internal/app/check/check.go", Blob: otherBlob, Lens: domainreview.Reliability},
	}))

	result := run(t, files, &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		blobs:   map[string]string{"internal/app/check/check.go": checkBlob},
	}, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	if result.Reason != "review_coverage" {
		t.Errorf("reason = %q, want review_coverage", result.Reason)
	}
	if !strings.Contains(result.Message, "internal/app/check/check.go") {
		t.Errorf("message = %q, want the file whose answer is dead", result.Message)
	}
	if result.Next != "odachi review" {
		t.Errorf("next = %q, want odachi review", result.Next)
	}
	rel := filepath.ToSlash(reviewPath(currentTree))
	for _, want := range []string{"delete", rel} {
		if !slices.ContainsFunc(result.Lines, func(line string) bool { return strings.Contains(line, want) }) {
			t.Errorf("lines %v do not tell the human to delete %s", result.Lines, rel)
		}
	}
}

// TestRequireReviewNamesAnOldVerdictUncomputable: a verdict written before
// content receipts carries no coverage at all. The gate reports it as not
// computable — an approval the tool cannot prove is an approval it must not use.
func TestRequireReviewNamesAnOldVerdictUncomputable(t *testing.T) {
	files := filesWithVerdict(t, currentTree, nil)

	result := run(t, files, &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		blobs:   map[string]string{"internal/app/check/check.go": checkBlob},
	}, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	if result.Reason != "review_coverage" {
		t.Errorf("reason = %q, want review_coverage", result.Reason)
	}
	if !strings.Contains(result.Message, "before content receipts") {
		t.Errorf("message = %q, want the old-format diagnosis", result.Message)
	}
	if result.Next != "odachi review" {
		t.Errorf("next = %q, want odachi review", result.Next)
	}
}

// TestCoverageNeverAsksAboutAbsentBytes: a path git cannot hash has no bytes in
// the worktree — the candidate deleted it, and there is nothing to answer about.
// The gate demands coverage of changed content, not of absence.
func TestCoverageNeverAsksAboutAbsentBytes(t *testing.T) {
	files := filesWithVerdict(t, currentTree, nil)

	result := run(t, files, &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
	}, check.Options{RequireReview: true})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0: no bytes, no coverage to ask for (%s)", result.Code, result.Message)
	}
}

// TestRequireReviewBlocksWhenTheBlobCannotBeMeasured: git failing for real is
// not a coverage answer — the measurement is load-bearing, so a failure blocks
// instead of pretending the rows are current.
func TestRequireReviewBlocksWhenTheBlobCannotBeMeasured(t *testing.T) {
	files := filesWithVerdict(t, currentTree, nil)

	result := run(t, files, &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/app/check/check.go"},
		blobs:   map[string]string{"internal/app/check/check.go": checkBlob},
		blobErr: errors.New("fatal: bad object"),
	}, check.Options{RequireReview: true})

	if result.Code == 0 {
		t.Fatal("a git failure must not read as a current receipt")
	}
	if !strings.Contains(result.Message, "could not measure") {
		t.Errorf("message = %q, want the measurement failure", result.Message)
	}
}

// TestRequireReviewWithChangesRequestedNamesTheFirstFinding: exit 2, and the
// sentence names the finding that blocks, so the human knows what to fix.
func TestRequireReviewWithChangesRequestedNamesTheFirstFinding(t *testing.T) {
	files := filesWithVerdict(t, currentTree, []domainreview.Finding{
		{Path: "estilo.go", Severity: domainreview.Warning, Causal: true, Why: "nombres largos"},
		{Path: "internal/app/check/check.go", Line: 42, Severity: domainreview.Critical, Causal: true, Why: "un fallo de lectura se informa como listo"},
		{Path: "otro.go", Severity: domainreview.Blocker, Causal: true, Why: "cobra dos veces"},
	})

	result := run(t, files, &fakeGit{tree: currentTree, changed: []string{"internal/app/check/check.go"}}, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	for _, want := range []string{"internal/app/check/check.go:42", "CRITICAL", "un fallo de lectura se informa como listo"} {
		if !strings.Contains(result.Message, want) {
			t.Errorf("message %q does not name %q", result.Message, want)
		}
	}
	if strings.Contains(result.Message, "cobra dos veces") {
		t.Errorf("message %q names the second blocking finding, want the first", result.Message)
	}
	if !strings.Contains(result.Next, "odachi review") {
		t.Errorf("next = %q, want a command that shows the review", result.Next)
	}
}

// TestRequireReviewIsPerTree: a verdict belongs to one tree. A different tree is
// a different candidate, and approval does not travel.
func TestRequireReviewIsPerTree(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(otherTree)] = []byte(verdictFile(t, otherTree, nil))

	result := run(t, files, &fakeGit{tree: currentTree, changed: []string{"internal/app/check/check.go"}}, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2: another tree's approval must not count", result.Code)
	}
	if result.Message != "no review for this tree" {
		t.Errorf("message = %q, want the documented sentence", result.Message)
	}
}

// TestRequireReviewReportsAMangledVerdict: the store is a plain file, so a human
// can break one. A verdict that cannot be read is not an approval.
func TestRequireReviewReportsAMangledVerdict(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte("basura")

	result := run(t, files, &fakeGit{tree: currentTree, changed: []string{"internal/app/check/check.go"}}, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, currentTree) {
		t.Errorf("message = %q, want the tree named", result.Message)
	}
	if !strings.Contains(result.Message, "odachi review --attach") {
		t.Errorf("message = %q, want the command that rewrites it", result.Message)
	}
}

// TestRequireReviewComesAfterTheGatesTheAgentCanActOn: a missing proof is
// actionable, so it is reported before a missing review. Sending an agent to
// `odachi review` while the gate is failing for another reason is a dead end.
func TestRequireReviewComesAfterTheGatesTheAgentCanActOn(t *testing.T) {
	files := newFakeFiles(closedFiles())
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.txt"))

	result := run(t, files, &fakeGit{tree: currentTree}, check.Options{RequireReview: true})

	if result.Code != 1 {
		t.Fatalf("code = %d, want 1: the proof gate comes first", result.Code)
	}
	if result.Reason != "proof_missing" {
		t.Errorf("reason = %q, want proof_missing", result.Reason)
	}
}

// TestAnOpenFeatureOwesTheProofAtItsClose: the wiring, not the rule. An open
// feature's tree is ready without a proof - the plan and every step commit land
// cheap - and the tree that closes it owes one again. Without this test the rule
// could exist in the domain and never reach the gate.
func TestAnOpenFeatureOwesTheProofAtItsClose(t *testing.T) {
	open := readyFiles()
	delete(open, "proof/0007-turnos.txt")

	got := run(t, newFakeFiles(open), &fakeGit{tree: currentTree}, check.Options{})
	if got.Code != int(gate.Ready) {
		t.Fatalf("an open feature without a proof = %d (%s), want ready", got.Code, got.Message)
	}
	if !strings.Contains(got.Message, "owed when this feature closes") {
		t.Errorf("message = %q, want the close named", got.Message)
	}

	closed := closedFiles()
	delete(closed, "proof/0007-turnos.txt")

	got = run(t, newFakeFiles(closed), &fakeGit{tree: currentTree}, check.Options{})
	if got.Code != int(gate.NeedsWork) || got.Reason != gate.ReasonProofMissing {
		t.Fatalf("the closing tree without a proof = %d/%q, want proof_missing", got.Code, got.Reason)
	}
}

// TestADocumentsOnlyCloseOwesNoVerdict: the review twin of the proof's paperwork
// rule. A feature whose own diff -base to the tree that closes it- asks for no lens
// has nothing for a verdict to judge, so the gate that would demand one answers
// "none is due" the way the review itself does. Without it a measurement feature pays
// a forced four-lens review of a paragraph, or has to be waived before it starts.
func TestADocumentsOnlyCloseOwesNoVerdict(t *testing.T) {
	const base = "c1398cef52c22969cba04a1cbfbf206c1068de2c"
	files := readyFiles()
	files["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)
	docs := &fakeGit{
		tree:    currentTree,
		changed: []string{"odd/tasks/0007-turnos.md", "docs/review-debt.md"},
		between: []string{"CHANGELOG.md", "docs/review-debt.md", "odd/tasks/0007-turnos.md"},
		filesAt: map[string]string{"HEAD:odd/tasks/0007-turnos.md": reviewableDoc(base, false)},
	}

	got := run(t, newFakeFiles(files), docs, check.Options{ReviewMode: true})

	if got.Code != int(gate.Ready) {
		t.Fatalf("a documents-only close asked for a verdict: %s (%s)", got.Message, got.Next)
	}
}

// TestACodeFeatureStillOwesTheVerdictAtItsClose: the rule reads the feature's own
// diff, so a close that only marks a step still owes the verdict when the feature
// wrote code. Measuring the closing tree alone would go blind here.
func TestACodeFeatureStillOwesTheVerdictAtItsClose(t *testing.T) {
	const base = "c1398cef52c22969cba04a1cbfbf206c1068de2c"
	files := readyFiles()
	files["odd/tasks/0007-turnos.md"] = reviewableDoc(base, true)
	code := &fakeGit{
		tree:    currentTree,
		changed: []string{"odd/tasks/0007-turnos.md"},
		between: []string{"internal/app/check/check.go", "odd/tasks/0007-turnos.md"},
		filesAt: map[string]string{"HEAD:odd/tasks/0007-turnos.md": reviewableDoc(base, false)},
	}

	got := run(t, newFakeFiles(files), code, check.Options{ReviewMode: true})

	if got.Code != int(gate.NeedsDecision) || got.Reason != check.ReasonReviewMissing {
		t.Fatalf("a code feature closed without a verdict: %d/%q", got.Code, got.Reason)
	}
	if want := "odachi review --base " + base; got.Next != want {
		t.Errorf("Next = %q, want %q", got.Next, want)
	}
}
