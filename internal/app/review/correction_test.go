package review_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// correctedTree is the tree a verdict judged before the fix: the tree --since
// names, and the bytes the correction measures its delta from.
const correctedTree = "1111111111111111111111111111111111111111"

// verdictFor seals one verdict the way a finished review leaves it: a reviewer, a
// base and the rows it answered. The store is keyed by tree, so a test seeds the
// predecessor by its tree.
func verdictFor(t *testing.T, tree, base string, lenses []string, passes ...domainreview.Pass) string {
	t.Helper()
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "before",
		Model:    "other/model",
		Lenses:   lenses,
		Passes:   passes,
	}, tree, base, now)
	if err != nil {
		t.Fatalf("sealing the predecessor: %v", err)
	}
	return string(domainreview.Render(sealed))
}

// TestACorrectionReviewsOnlyWhatMoved: a fix after a verdict is a delta. The
// prompt is built from the paths and the diff since the tree the verdict named,
// the answer judges only those, and the stored verdict carries the predecessor's
// rows -so every path stays answered while the run paid for the fix alone.
func TestACorrectionReviewsOnlyWhatMoved(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.git.changed = []string{"b.go"}
	p.git.diff = "diff --git a/b.go b/b.go\n--- a/b.go\n+++ b/b.go\n@@ -1 +1 @@\n-viejo\n+nuevo\n"
	p.git.blobs = map[string]string{"a.go": gitBlob, "b.go": gitBlob}
	p.shell.output = `{"findings":[],"passes":[{"path":"b.go","lens":"reliability"}],"inspection":{"status":"completed","paths":["b.go"]}}`
	p.files.files[storePath(correctedTree)] = []byte(verdictFor(t, correctedTree, "abc1234", []string{"reliability"},
		domainreview.Pass{Path: "a.go", Lens: "reliability", Blob: gitBlob}))

	result := p.run(review.Options{Exec: "revisor --model {model}", Since: correctedTree})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	if !slices.Equal(p.git.corrections, []string{correctedTree}) {
		t.Errorf("corrections = %v, want one measured from %s", p.git.corrections, correctedTree)
	}

	data, ok := p.files.files[storePath(currentTree)]
	if !ok {
		t.Fatalf("no verdict stored for the corrected tree")
	}
	stored, err := domainreview.Parse(data)
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if !slices.Contains(stored.Lenses, "reliability") {
		t.Errorf("lenses = %v, want the predecessor's lens carried", stored.Lenses)
	}
	if stored.Base != "abc1234" {
		t.Errorf("base = %q, want the predecessor's base abc1234", stored.Base)
	}
	for _, path := range []string{"a.go", "b.go"} {
		answered := slices.ContainsFunc(stored.Passes, func(pass domainreview.Pass) bool {
			return pass.Path == path && strings.EqualFold(pass.Lens, "reliability")
		})
		if !answered {
			t.Errorf("the correction's verdict does not answer %s: %+v", path, stored.Passes)
		}
	}
}

// TestACorrectionWithNothingToCorrectRefuses: the bytes the verdict judged are the
// bytes that stand, and paying a reviewer for the same tree twice is the cost this
// flag exists to remove. The refusal says so and names the read that already has
// the answer.
func TestACorrectionWithNothingToCorrectRefuses(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.git.changed = nil
	p.git.diff = ""
	p.shell.output = `{"findings":[]}`
	p.files.files[storePath(correctedTree)] = []byte(verdictFor(t, correctedTree, "HEAD", []string{"reliability"}))

	result := p.run(review.Options{Exec: "revisor --model {model}", Since: correctedTree})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "nothing changed since the verdict") {
		t.Errorf("message = %q, want the emptiness named", result.Message)
	}
	if len(p.shell.commands) != 0 {
		t.Errorf("the reviewer ran for a tree that did not move: %v", p.shell.commands)
	}
}

// TestACorrectionNeedsTheVerdictItCompletes: a correction completes an answer this
// clone holds. Without that file there is nothing to complete, and the refusal says
// which tree it looked for instead of reviewing the whole candidate in silence.
func TestACorrectionNeedsTheVerdictItCompletes(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = `{"findings":[]}`

	result := p.run(review.Options{Exec: "revisor --model {model}", Since: correctedTree})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, correctedTree) {
		t.Errorf("message = %q, want the tree --since named", result.Message)
	}
	if len(p.shell.commands) != 0 {
		t.Errorf("the reviewer ran without a verdict to complete: %v", p.shell.commands)
	}
}

// TestACorrectionDropsThePredecessorRowsAboutWhatMoved: the rows the predecessor
// earned about the paths the fix changed describe bytes nobody judged any more. They
// do not travel -the rule 0133 wrote for inheritance- because a receipt for old bytes
// counted as an answer is the gate vouching for a tree nobody read. The rows about
// the paths that did not move travel, and the fresh answer is the only word about the
// ones that did.
func TestACorrectionDropsThePredecessorRowsAboutWhatMoved(t *testing.T) {
	const oldBlob, newBlob = "1111111111111111111111111111111111111111", "2222222222222222222222222222222222222222"

	p := newProject()
	p.host.model = "host/default"
	p.git.changed = []string{"b.go"}
	p.git.diff = "diff --git a/b.go b/b.go\n--- a/b.go\n+++ b/b.go\n@@ -1 +1 @@\n-viejo\n+nuevo\n"
	p.git.blobs = map[string]string{"a.go": gitBlob, "b.go": newBlob}
	p.shell.output = `{"findings":[],"passes":[{"path":"b.go","lens":"reliability"}],"inspection":{"status":"completed","paths":["b.go"]}}`
	p.files.files[storePath(correctedTree)] = []byte(verdictFor(t, correctedTree, "HEAD", []string{"reliability"},
		domainreview.Pass{Path: "a.go", Lens: "reliability", Blob: gitBlob},
		domainreview.Pass{Path: "b.go", Lens: "reliability", Blob: oldBlob}))

	result := p.run(review.Options{Exec: "revisor --model {model}", Since: correctedTree})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	stored, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	for _, pass := range stored.Passes {
		if pass.Path == "b.go" && pass.Blob == oldBlob {
			t.Errorf("a clean row about the bytes the fix changed travelled: %+v", pass)
		}
	}
	if !slices.ContainsFunc(stored.Passes, func(pass domainreview.Pass) bool {
		return pass.Path == "b.go" && pass.Blob == newBlob
	}) {
		t.Errorf("the correction's own answer about b.go is missing: %+v", stored.Passes)
	}
	if !slices.ContainsFunc(stored.Passes, func(pass domainreview.Pass) bool {
		return pass.Path == "a.go" && pass.Blob == gitBlob
	}) {
		t.Errorf("the row about the path that did not move did not travel: %+v", stored.Passes)
	}
}

// TestACorrectionTurnsABlockerAboutMovedBytesIntoAFollowUp: the correction exists to
// clear a blocker. A finding that cites bytes the fix moved can neither keep blocking
// -nobody can read those bytes again- nor vanish, because dropping it would let a fix
// launder a reason not to ship. It travels as a follow-up: in the verdict, and not a
// verdict of its own.
func TestACorrectionTurnsABlockerAboutMovedBytesIntoAFollowUp(t *testing.T) {
	const oldBlob, newBlob = "3333333333333333333333333333333333333333", "4444444444444444444444444444444444444444"

	p := newProject()
	p.host.model = "host/default"
	p.git.changed = []string{"b.go"}
	p.git.diff = "diff --git a/b.go b/b.go\n--- a/b.go\n+++ b/b.go\n@@ -1 +1 @@\n-viejo\n+nuevo con el chequeo\n"
	p.git.blobs = map[string]string{"b.go": newBlob}
	p.shell.output = `{"findings":[],"passes":[{"path":"b.go","lens":"reliability"}],"inspection":{"status":"completed","paths":["b.go"]}}`

	before, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "before",
		Model:    "other/model",
		Lenses:   []string{"reliability"},
		Findings: []domainreview.Finding{{
			Path:     "b.go",
			Line:     3,
			Severity: domainreview.Blocker,
			Lens:     "reliability",
			Causal:   true,
			Why:      "the bytes dropped the error check",
			Blob:     oldBlob,
		}},
	}, correctedTree, "HEAD", now)
	if err != nil {
		t.Fatalf("sealing the predecessor: %v", err)
	}
	p.files.files[storePath(correctedTree)] = []byte(domainreview.Render(before))

	result := p.run(review.Options{Exec: "revisor --model {model}", Since: correctedTree})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	stored, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if stored.Verdict != domainreview.Approved {
		t.Errorf("verdict = %q, want %q: a blocker about bytes the fix moved still blocks", stored.Verdict, domainreview.Approved)
	}
	var kept *domainreview.Finding
	for i, finding := range stored.Findings {
		if finding.Path == "b.go" {
			kept = &stored.Findings[i]
		}
	}
	if kept == nil {
		t.Fatalf("the blocker vanished instead of becoming a follow-up: %+v", stored.Findings)
	}
	if kept.Causal {
		t.Errorf("the blocker still calls itself causal after the fix: %+v", *kept)
	}
}
