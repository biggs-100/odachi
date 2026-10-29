package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// TestTheHumansDismissalRetiresACarriedRow: a warning the reviewer got wrong is
// potent about bytes nobody touches again, so every later verdict re-reported it and
// no verb could retire it. The human's judgment, stored beside the verdict it
// answers, travels with the row: the carried row is filed dismissed with the human's
// reason, and the prompt tells the next reviewer not to raise it.
func TestTheHumansDismissalRetiresACarriedRow(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.git.blobs = map[string]string{"a.go": gitBlob}
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]},"passes":[{"path":"a.go","lens":"reliability"}]}`

	// A previous verdict, about another tree, holding one open warning whose
	// receipt still describes a.go's bytes.
	stored, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "primero",
		Lenses:   []string{"reliability"},
		Findings: []domainreview.Finding{{
			Path:     "a.go",
			Line:     1,
			Severity: domainreview.Warning,
			Causal:   true,
			Why:      "the propagation is pinned by no test",
			Lens:     "reliability",
			Blob:     gitBlob,
		}},
	}, otherTree, "HEAD", now)
	if err != nil {
		t.Fatalf("seeding the verdict: %v", err)
	}
	p.files.files[storePath(otherTree)] = domainreview.Render(stored)

	// The human retires it, with evidence.
	sidecar, err := domainreview.DismissedPath(otherTree)
	if err != nil {
		t.Fatalf("DismissedPath: %v", err)
	}
	p.files.files[storeFile(sidecar)] = domainreview.RenderDismissals(domainreview.Dismissals{
		Rows: []domainreview.Dismissal{{
			Fingerprint: domainreview.Fingerprint(stored.Findings[0]),
			Path:        "a.go",
			Line:        1,
			Reason:      "la prueba existe en feature_test.go, línea 139",
			By:          "humano",
		}},
	})

	result := p.run(review.Options{Exec: "revisor --model {model}"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	merged, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the verdict does not parse: %v", err)
	}
	var carried *domainreview.Finding
	for i, finding := range merged.Findings {
		if finding.Path == "a.go" {
			carried = &merged.Findings[i]
		}
	}
	if carried == nil {
		t.Fatal("the potent row did not travel")
	}
	if !domainreview.IsDismissed(*carried) {
		t.Errorf("the row travels open: %+v", *carried)
	}
	if !strings.Contains(carried.Reason, "feature_test.go") {
		t.Errorf("reason = %q, want the human's evidence", carried.Reason)
	}

	// The next question carries the judgment: a reviewer that reads it must not
	// re-raise a row a person already judged.
	printed := p.run(review.Options{})
	if printed.Code != review.CodeOK {
		t.Fatalf("the printed form: code = %d (%s)", printed.Code, printed.Message)
	}
	if !strings.Contains(printed.Prompt, "dismissed:") || !strings.Contains(printed.Prompt, "feature_test.go") {
		t.Errorf("the prompt does not tell the next reviewer the row was judged:\n%s", printed.Prompt)
	}
}

// TestTheHumanRetiresARowWithEvidence: the act is a flag, the record is a file
// beside the verdict, and the verdict itself is not rewritten. Without a reason the
// act is refused, because that is the silent edit the file exists to replace.
func TestTheHumanRetiresARowWithEvidence(t *testing.T) {
	p := newProject()
	p.files.files[storePath(currentTree)] = domainreview.Render(warningVerdict(t, currentTree))

	result := p.run(review.Options{Dismiss: "a.go:1", Because: "la prueba existe en feature_test.go, línea 139", By: "humano"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	sidecar, err := domainreview.DismissedPath(currentTree)
	if err != nil {
		t.Fatalf("DismissedPath: %v", err)
	}
	data, ok := p.files.files[storeFile(sidecar)]
	if !ok {
		t.Fatalf("no dismissals written at %s", sidecar)
	}
	parsed, err := domainreview.ParseDismissals(data)
	if err != nil {
		t.Fatalf("the written file does not parse: %v", err)
	}
	if len(parsed.Rows) != 1 || parsed.Rows[0].By != "humano" || !strings.Contains(parsed.Rows[0].Reason, "feature_test.go") {
		t.Errorf("rows = %+v, want the human's judgment", parsed.Rows)
	}
	// The verdict is untouched.
	stored, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the verdict no longer parses: %v", err)
	}
	if domainreview.IsDismissed(stored.Findings[0]) {
		t.Error("the verdict was rewritten by a dismissal")
	}

	// And the act needs its evidence.
	refused := p.run(review.Options{Dismiss: "a.go:1"})
	if refused.Code != review.CodeBlocked || !strings.Contains(refused.Message, "--because") {
		t.Errorf("a dismissal without a reason = %d %q, want it refused naming --because", refused.Code, refused.Message)
	}
}

// TestTheHumanCannotRetireWhatTheVerdictDoesNotHold: a judgment about a row nobody
// can see is a judgment about nothing, and a causal blocker is answered by the
// refuter or a fix - letting a person retire it would be the gate opening by
// omission.
func TestTheHumanCannotRetireWhatTheVerdictDoesNotHold(t *testing.T) {
	p := newProject()
	p.files.files[storePath(currentTree)] = domainreview.Render(warningVerdict(t, currentTree))

	missing := p.run(review.Options{Dismiss: "otro.go:9", Because: "no está"})
	if missing.Code != review.CodeBlocked || !strings.Contains(missing.Message, "a.go:1") {
		t.Errorf("a row that does not exist = %d %q, want the rows it holds named", missing.Code, missing.Message)
	}

	blocking := p.run(review.Options{Dismiss: "a.go", Because: "sin línea"})
	if blocking.Code != review.CodeBlocked {
		t.Errorf("a target without a line = %d, want it refused", blocking.Code)
	}
}

// TestACausalBlockerIsRefutedNotDismissed: the score is the gate's, not a person's.
func TestACausalBlockerIsRefutedNotDismissed(t *testing.T) {
	p := newProject()
	p.files.files[storePath(currentTree)] = domainreview.Render(blockingVerdict(t, currentTree))

	result := p.run(review.Options{Dismiss: "a.go:1", Because: "no me parece"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	lines := strings.Join(result.Lines, "\n")
	for _, want := range []string{"causal BLOCKER", "--refute"} {
		if !strings.Contains(lines, want) {
			t.Errorf("the refusal does not carry %q:\n%s", want, lines)
		}
	}
	if _, written := p.files.files[storeFile(mustDismissedPath(t, currentTree))]; written {
		t.Error("a blocker was retired")
	}
}

// warningVerdict is a sealed verdict holding the warning the tests retire.
func warningVerdict(t *testing.T, tree string) domainreview.Verdict {
	t.Helper()
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "revisor",
		Lenses:   []string{"reliability"},
		Findings: []domainreview.Finding{{
			Path:     "a.go",
			Line:     1,
			Severity: domainreview.Warning,
			Causal:   true,
			Why:      "la propagación no tiene prueba",
			Lens:     "reliability",
		}},
	}, tree, "HEAD", now)
	if err != nil {
		t.Fatalf("sealing the seeded warning: %v", err)
	}
	return sealed
}

// blockingVerdict is the same row at BLOCKER severity.
func blockingVerdict(t *testing.T, tree string) domainreview.Verdict {
	t.Helper()
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "revisor",
		Lenses:   []string{"reliability"},
		Findings: []domainreview.Finding{{
			Path:     "a.go",
			Line:     1,
			Severity: domainreview.Blocker,
			Causal:   true,
			Why:      "pierde el turno",
			Lens:     "reliability",
		}},
	}, tree, "HEAD", now)
	if err != nil {
		t.Fatalf("sealing the seeded blocker: %v", err)
	}
	return sealed
}

func mustDismissedPath(t *testing.T, tree string) string {
	t.Helper()
	rel, err := domainreview.DismissedPath(tree)
	if err != nil {
		t.Fatalf("DismissedPath: %v", err)
	}
	return rel
}

// TestShowNamesTheRowsAPersonRetired: a retired row is not a row to act on, and the
// question the gate asks is answered by --show. The judgment is listed with its
// evidence, and a sidecar nobody can read is a note instead of a silence.
func TestShowNamesTheRowsAPersonRetired(t *testing.T) {
	p := newProject()
	p.files.files[storePath(currentTree)] = domainreview.Render(warningVerdict(t, currentTree))
	sidecar := mustDismissedPath(t, currentTree)
	p.files.files[storeFile(sidecar)] = domainreview.RenderDismissals(domainreview.Dismissals{
		Rows: []domainreview.Dismissal{{
			Fingerprint: domainreview.Fingerprint(warningVerdict(t, currentTree).Findings[0]),
			Path:        "a.go",
			Line:        1,
			Reason:      "la prueba existe, línea 139",
			By:          "humano",
		}},
	})

	shown := p.run(review.Options{Show: true})
	if shown.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", shown.Code, shown.Message)
	}
	lines := strings.Join(shown.Lines, "\n")
	if !strings.Contains(lines, "dismissed by a person") || !strings.Contains(lines, "línea 139") {
		t.Errorf("--show does not name the retired row:\n%s", lines)
	}

	// A sidecar that cannot be read is reported, never silently ignored.
	p.files.files[storeFile(sidecar)] = []byte("not json")
	broken := p.run(review.Options{Show: true})
	if !strings.Contains(strings.Join(broken.Lines, "\n"), "cannot be read") {
		t.Errorf("--show hides an unreadable sidecar:\n%s", strings.Join(broken.Lines, "\n"))
	}
}
