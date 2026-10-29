package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// scopeOf reads the record beside this tree's verdict, failing the test with the
// store's contents when it is not there.
func scopeOf(t *testing.T, p *project) domainreview.Scope {
	t.Helper()

	rel, err := domainreview.ScopePath(currentTree)
	if err != nil {
		t.Fatalf("ScopePath: %v", err)
	}
	data, ok := p.files.files[storeFile(rel)]
	if !ok {
		t.Fatalf("no scope record beside the verdict: %s was not written (wrote %v)", rel, p.files.written())
	}
	scope, err := domainreview.ParseScope(data)
	if err != nil {
		t.Fatalf("the scope record does not parse: %v\n%s", err, data)
	}
	return scope
}

// TestARunThatAddsALensFilesTheScopeBesideTheVerdict: --lens adds a lens the
// evidence did not ask for, and a later reader has to see who asked and what was
// asked for. The record names the evidence's answer, the addition and the flag as
// the caller wrote it.
func TestARunThatAddsALensFilesTheScopeBesideTheVerdict(t *testing.T) {
	p := newLensProject()
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"humano","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}`)

	result := p.run(review.Options{Attach: "veredicto.json", Lens: "risk", Because: "the change touches the gate"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	scope := scopeOf(t, p)
	if strings.Join(scope.Required, ",") != domainreview.Reliability {
		t.Errorf("required = %v, want [reliability]: the evidence asked for one lens", scope.Required)
	}
	if strings.Join(scope.Added, ",") != domainreview.Risk {
		t.Errorf("added = %v, want [risk]", scope.Added)
	}
	if scope.Asked != "risk" {
		t.Errorf("asked = %q, want the flag as it was given", scope.Asked)
	}
	if scope.Because != "the change touches the gate" {
		t.Errorf("because = %q, want the caller's words", scope.Because)
	}
}

// TestAScopeWithoutWordsSaysNothingInsteadOfInventing: the tool keeps the caller's
// reason only when they gave one. A record that wrote a reason in the tool's words
// would be the tool talking about a decision it did not make.
func TestAScopeWithoutWordsSaysNothingInsteadOfInventing(t *testing.T) {
	p := newLensProject()
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"humano","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}`)

	if result := p.run(review.Options{Attach: "veredicto.json", Lens: "risk"}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	if scope := scopeOf(t, p); scope.Because != "" {
		t.Errorf("because = %q, want no words", scope.Because)
	}
}

// TestARunThatAddedNothingWritesNoScope: the file is evidence, not ceremony. A
// verdict that covers exactly what the evidence asked for needs no explanation, and
// a file that says "nothing was raised" would be noise in the store.
func TestARunThatAddedNothingWritesNoScope(t *testing.T) {
	p := newLensProject()
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"humano","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}`)

	if result := p.run(review.Options{Attach: "veredicto.json"}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	rel, err := domainreview.ScopePath(currentTree)
	if err != nil {
		t.Fatalf("ScopePath: %v", err)
	}
	if _, written := p.files.files[storeFile(rel)]; written {
		t.Errorf("a run that added no lens wrote %s", rel)
	}
}

// TestASecondRunKeepsTheFirstScope: the store is evidence about a tree and is never
// rewritten. The run that asked first is the one whose words travel; a later run
// that asks the same thing has nothing to add.
func TestASecondRunKeepsTheFirstScope(t *testing.T) {
	p := newLensProject()
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"humano","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}`)

	rel, err := domainreview.ScopePath(currentTree)
	if err != nil {
		t.Fatalf("ScopePath: %v", err)
	}
	first := domainreview.RenderScope(domainreview.Scope{
		Required: []string{domainreview.Reliability},
		Added:    []string{domainreview.Risk},
		Asked:    "risk",
		Because:  "the first run's words",
	})
	p.files.files[storeFile(rel)] = first

	if result := p.run(review.Options{Attach: "veredicto.json", Lens: "risk", Because: "a second run's words"}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	if got := string(p.files.files[storeFile(rel)]); got != string(first) {
		t.Errorf("the second run rewrote the scope record:\n%s\nwant:\n%s", got, first)
	}
}

// TestThePrintedPromptSaysTheScopeWasRaised: the prompt the run prints is the one
// a reviewer reads, so it carries the decision too. The lenses section cannot claim
// the evidence chose a lens somebody added.
func TestThePrintedPromptSaysTheScopeWasRaised(t *testing.T) {
	p := newLensProject()

	result := p.run(review.Options{Lens: "risk", Because: "the change touches the gate"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Prompt, "risk was added with --lens risk") {
		t.Errorf("the prompt does not say the scope was raised:\n%s", result.Prompt)
	}
	if !strings.Contains(result.Prompt, "the change touches the gate") {
		t.Errorf("the prompt does not carry the caller's words:\n%s", result.Prompt)
	}
}

// TestShowSaysTheScopeWasRaised: a later reader runs --show and sees why the verdict
// asks for more lenses than the evidence did, in the words of whoever asked.
func TestShowSaysTheScopeWasRaised(t *testing.T) {
	p := newLensProject()
	stored, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "humano",
		Inspection: &domainreview.Inspection{
			Status: domainreview.InspectionCompleted,
			Paths:  []string{"a.go"},
		},
		Lenses: []string{domainreview.Reliability, domainreview.Risk},
	}, currentTree, "HEAD", now)
	if err != nil {
		t.Fatalf("seeding the verdict: %v", err)
	}
	p.files.files[storePath(currentTree)] = domainreview.Render(stored)

	rel, err := domainreview.ScopePath(currentTree)
	if err != nil {
		t.Fatalf("ScopePath: %v", err)
	}
	p.files.files[storeFile(rel)] = domainreview.RenderScope(domainreview.Scope{
		Required: []string{domainreview.Reliability},
		Added:    []string{domainreview.Risk},
		Asked:    "risk",
		Because:  "the change touches the gate",
	})

	result := p.run(review.Options{Show: true})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	lines := strings.Join(result.Lines, "\n")
	if !strings.Contains(lines, "scope raised: risk was added with --lens risk") {
		t.Errorf("--show does not say the scope was raised:\n%s", lines)
	}
	if !strings.Contains(lines, "the change touches the gate") {
		t.Errorf("--show does not carry the words of whoever asked:\n%s", lines)
	}
}
