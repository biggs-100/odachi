package review_test

import (
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// now is the instant a sealed verdict records. Fixing it is what makes the
// stored file byte-stable: a verdict that changes on every run cannot be
// compared, and the file name is the tree, not the clock.
var now = time.Date(2026, 10, 21, 20, 31, 0, 0, time.UTC)

// tree is a real-shaped git tree hash: hex only, which is what keeps the store's
// file names from ever describing a path.
const tree = "4a1c9f7d2e3b5a6c8d9e0f1a2b3c4d5e6f7a8b9c"

// otherTree is a different candidate.
const otherTree = "9f2b7c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b"

// TestDeriveIsTheWholeRule: the verdict is derived from the findings, never
// administered. Only a causal BLOCKER or CRITICAL requests changes.
func TestDeriveIsTheWholeRule(t *testing.T) {
	cases := []struct {
		name     string
		findings []review.Finding
		want     string
	}{
		{
			name:     "no findings approves",
			findings: nil,
			want:     review.Approved,
		},
		{
			name:     "causal BLOCKER requests changes",
			findings: []review.Finding{{Path: "a.go", Severity: review.Blocker, Causal: true, Why: "pierde datos"}},
			want:     review.ChangesRequested,
		},
		{
			name:     "causal CRITICAL requests changes",
			findings: []review.Finding{{Path: "a.go", Severity: review.Critical, Causal: true, Why: "rompe el login"}},
			want:     review.ChangesRequested,
		},
		{
			name:     "non-causal CRITICAL approves",
			findings: []review.Finding{{Path: "a.go", Severity: review.Critical, Causal: false, Why: "ya estaba roto"}},
			want:     review.Approved,
		},
		{
			name:     "causal WARNING approves",
			findings: []review.Finding{{Path: "a.go", Severity: review.Warning, Causal: true, Why: "queda lento"}},
			want:     review.Approved,
		},
		{
			name:     "causal SUGGESTION approves",
			findings: []review.Finding{{Path: "a.go", Severity: review.Suggestion, Causal: true, Why: "se puede simplificar"}},
			want:     review.Approved,
		},
		{
			name: "a non-causal finding does not hide the causal one that blocks",
			findings: []review.Finding{
				{Path: "a.go", Severity: review.Warning, Causal: false, Why: "estilo"},
				{Path: "b.go", Severity: review.Blocker, Causal: true, Why: "cobra dos veces"},
			},
			want: review.ChangesRequested,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := review.Derive(c.findings); got != c.want {
				t.Fatalf("Derive = %q, want %q", got, c.want)
			}
		})
	}
}

// TestBlockingIsTheFirstCausalFinding: the message names one finding, so the
// rule has to pick exactly one and pick the same one every time.
func TestBlockingIsTheFirstCausalFinding(t *testing.T) {
	findings := []review.Finding{
		{Path: "estilo.go", Severity: review.Warning, Causal: true, Why: "nombres"},
		{Path: "viejo.go", Severity: review.Critical, Causal: false, Why: "ya estaba"},
		{Path: "primero.go", Line: 12, Severity: review.Critical, Causal: true, Why: "pierde el turno"},
		{Path: "segundo.go", Severity: review.Blocker, Causal: true, Why: "otra cosa"},
	}

	got, ok := review.Blocking(findings)
	if !ok {
		t.Fatal("Blocking found nothing, want the first causal BLOCKER/CRITICAL")
	}
	if got.Path != "primero.go" || got.Line != 12 {
		t.Fatalf("Blocking = %s:%d, want primero.go:12", got.Path, got.Line)
	}

	if _, ok := review.Blocking([]review.Finding{{Path: "a.go", Severity: review.Critical, Causal: false}}); ok {
		t.Fatal("a non-causal finding blocked: pre-existing problems are follow-ups")
	}
}

// TestSeverityVocabularyIsClosed: four values, in the documented spelling. A
// lowercase severity is a typo that must not slip into a stored verdict.
func TestSeverityVocabularyIsClosed(t *testing.T) {
	for _, valid := range []review.Severity{review.Blocker, review.Critical, review.Warning, review.Suggestion} {
		if !valid.Valid() {
			t.Errorf("%q should be a valid severity", valid)
		}
	}
	for _, invalid := range []review.Severity{"", "blocker", "MAJOR", "INFO", "BLOCKER "} {
		if invalid.Valid() {
			t.Errorf("%q should not be a valid severity", invalid)
		}
	}
	if !review.Blocker.Blocks() || !review.Critical.Blocks() {
		t.Error("BLOCKER and CRITICAL must be able to block")
	}
	if review.Warning.Blocks() || review.Suggestion.Blocks() {
		t.Error("WARNING and SUGGESTION must never block")
	}
}

// TestParseAcceptsTheCanonicalEmptyAnswer: the prose answer gentile-ai learned to
// demand is exactly "No findings.", and it means an empty findings array. A
// tool that rejects it would be arguing with its own prompt.
func TestParseAcceptsTheCanonicalEmptyAnswer(t *testing.T) {
	for _, text := range []string{"No findings.", "No findings.\n", "  No findings.  "} {
		parsed, err := review.Parse([]byte(text))
		if err != nil {
			t.Fatalf("Parse(%q) = %v, want the canonical empty verdict", text, err)
		}
		if len(parsed.Findings) != 0 {
			t.Errorf("Parse(%q) has %d findings, want none", text, len(parsed.Findings))
		}
	}
}

// TestParseAndValidateRefuseABrokenVerdict: the two halves the app always runs
// together — the shape, then the meaning — refuse an unknown field, a severity
// outside the vocabulary and trailing text. A verdict the tool half-understands
// is worse than none: it would be stored under a name that claims more than it
// can back.
func TestParseAndValidateRefuseABrokenVerdict(t *testing.T) {
	cases := map[string]string{
		"empty":              "",
		"not JSON":           "listo, aprobado",
		"unknown field":      `{"reviewer":"x","findings":[],"extra":1}`,
		"lowercase severity": `{"reviewer":"x","findings":[{"path":"a.go","severity":"blocker","why":"x"}]}`,
		"two objects":        `{"reviewer":"x","findings":[]}{"reviewer":"y","findings":[]}`,
	}

	for name, text := range cases {
		t.Run(name, func(t *testing.T) {
			parsed, err := review.Parse([]byte(text))
			if err == nil {
				err = parsed.Validate()
			}
			if err == nil {
				t.Fatalf("Parse+Validate(%q) = nil error, want a refusal", text)
			}
		})
	}
}

// TestParseReadsAFullVerdict: every documented field survives the round trip.
func TestParseReadsAFullVerdict(t *testing.T) {
	text := `{
  "tree": "` + tree + `",
  "base": "HEAD",
  "reviewer": "claude",
  "at": "2026-10-21T20:31:00Z",
  "findings": [
    {"path": "internal/app/check/check.go", "line": 42, "severity": "CRITICAL", "causal": true, "why": "un fallo de lectura se informa como listo"}
  ],
  "verdict": "changes_requested"
}`

	parsed, err := review.Parse([]byte(text))
	if err != nil {
		t.Fatalf("Parse = %v", err)
	}
	if parsed.Tree != tree || parsed.Base != "HEAD" || parsed.Reviewer != "claude" {
		t.Fatalf("identity fields = %+v", parsed)
	}
	if len(parsed.Findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(parsed.Findings))
	}
	found := parsed.Findings[0]
	if found.Path != "internal/app/check/check.go" || found.Line != 42 || found.Severity != review.Critical || !found.Causal {
		t.Fatalf("finding = %+v", found)
	}
	if parsed.Verdict != review.ChangesRequested {
		t.Fatalf("verdict = %q", parsed.Verdict)
	}
}

// TestRenderIsStableAndRoundTrips: two renders of the same verdict are the same
// bytes, and parsing them gives the verdict back. That is what makes the file
// comparable and its immutability checkable.
func TestRenderIsStableAndRoundTrips(t *testing.T) {
	verdict, err := review.Seal(review.Verdict{
		Reviewer: "humano",
		Findings: []review.Finding{{Path: "a.go", Line: 3, Severity: review.Warning, Causal: true, Why: "queda lento"}},
	}, tree, "HEAD", now)
	if err != nil {
		t.Fatalf("Seal = %v", err)
	}

	first := review.Render(verdict)
	second := review.Render(verdict)
	if string(first) != string(second) {
		t.Fatal("Render is not stable: two renders differ")
	}
	if !strings.HasSuffix(string(first), "\n") {
		t.Errorf("Render does not end with a newline: %q", string(first))
	}

	back, err := review.Parse(first)
	if err != nil {
		t.Fatalf("Parse(Render(v)) = %v", err)
	}
	if back.Tree != verdict.Tree || back.Verdict != verdict.Verdict || len(back.Findings) != 1 {
		t.Fatalf("round trip = %+v, want %+v", back, verdict)
	}
}

// TestRenderOfNoFindingsIsAnEmptyArray: the stored JSON never says null. An
// empty review is an empty list, not a missing one.
func TestRenderOfNoFindingsIsAnEmptyArray(t *testing.T) {
	verdict, err := review.Seal(review.Verdict{Reviewer: "humano"}, tree, "HEAD", now)
	if err != nil {
		t.Fatalf("Seal = %v", err)
	}

	rendered := string(review.Render(verdict))
	if !strings.Contains(rendered, `"findings": []`) {
		t.Errorf("rendered verdict %q does not carry an empty findings array", rendered)
	}
	if strings.Contains(rendered, "null") {
		t.Errorf("rendered verdict %q carries null", rendered)
	}
}

// TestValidateRefusesIncompleteFindings: what only the author can fix is
// refused before anything is written, with a line that says which field.
func TestValidateRefusesIncompleteFindings(t *testing.T) {
	cases := map[string]struct {
		verdict review.Verdict
		mention string
	}{
		"no reviewer": {
			verdict: review.Verdict{Findings: []review.Finding{}},
			mention: "reviewer",
		},
		"finding without path": {
			verdict: review.Verdict{Reviewer: "x", Findings: []review.Finding{{Severity: review.Warning, Why: "algo"}}},
			mention: "path",
		},
		"finding without why": {
			verdict: review.Verdict{Reviewer: "x", Findings: []review.Finding{{Path: "a.go", Severity: review.Warning}}},
			mention: "why",
		},
		"finding with an invented severity": {
			verdict: review.Verdict{Reviewer: "x", Findings: []review.Finding{{Path: "a.go", Severity: "MAJOR", Why: "algo"}}},
			mention: "severity",
		},
		"finding with a negative line": {
			verdict: review.Verdict{Reviewer: "x", Findings: []review.Finding{{Path: "a.go", Line: -1, Severity: review.Warning, Why: "algo"}}},
			mention: "line",
		},
		"at that is not RFC3339": {
			verdict: review.Verdict{Reviewer: "x", At: "21/10/2026"},
			mention: "RFC3339",
		},
		"invented verdict": {
			verdict: review.Verdict{Reviewer: "x", Verdict: "approved_by_me"},
			mention: "verdict",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			err := c.verdict.Validate()
			if err == nil {
				t.Fatalf("Validate = nil, want a refusal naming %q", c.mention)
			}
			if !strings.Contains(err.Error(), c.mention) {
				t.Errorf("error %q does not name %q", err.Error(), c.mention)
			}
		})
	}
}

// TestSealStampsWhatOnlyTheToolKnows: the tree, the base, the instant and the
// derived verdict are the tool's, not the author's.
func TestSealStampsWhatOnlyTheToolKnows(t *testing.T) {
	sealed, err := review.Seal(review.Verdict{
		Reviewer: "claude",
		Findings: []review.Finding{{Path: "a.go", Severity: review.Blocker, Causal: true, Why: "cobra dos veces"}},
	}, tree, "HEAD", now)
	if err != nil {
		t.Fatalf("Seal = %v", err)
	}

	if sealed.Tree != tree {
		t.Errorf("Tree = %q, want %q", sealed.Tree, tree)
	}
	if sealed.Base != "HEAD" {
		t.Errorf("Base = %q, want HEAD", sealed.Base)
	}
	if sealed.At != "2026-10-21T20:31:00Z" {
		t.Errorf("At = %q, want the fixed instant in RFC3339", sealed.At)
	}
	if sealed.Verdict != review.ChangesRequested {
		t.Errorf("Verdict = %q, want the derived changes_requested", sealed.Verdict)
	}
	if sealed.Reviewer != "claude" {
		t.Errorf("Reviewer = %q, the author's reviewer was lost", sealed.Reviewer)
	}
}

// TestSealNormalizesTheInstant: an author may write an offset; the stored file
// always carries the same shape, so two verdicts can be compared.
func TestSealNormalizesTheInstant(t *testing.T) {
	sealed, err := review.Seal(review.Verdict{Reviewer: "x", At: "2026-10-21T17:31:00-03:00"}, tree, "HEAD", now)
	if err != nil {
		t.Fatalf("Seal = %v", err)
	}
	if sealed.At != "2026-10-21T20:31:00Z" {
		t.Errorf("At = %q, want the same instant in UTC", sealed.At)
	}
}

// TestSealRefusesAnotherTree: approval cannot float. A verdict written for one
// tree is not a verdict for another, and the refusal names both.
func TestSealRefusesAnotherTree(t *testing.T) {
	_, err := review.Seal(review.Verdict{Tree: otherTree, Reviewer: "x"}, tree, "HEAD", now)
	if err == nil {
		t.Fatal("Seal accepted a verdict for another tree")
	}
	if !strings.Contains(err.Error(), otherTree) || !strings.Contains(err.Error(), tree) {
		t.Errorf("error %q does not name both trees", err.Error())
	}
}

// TestSealRefusesAVerdictThatContradictsTheRule: the verdict is derived, so an
// author who declares "approved" next to a causal BLOCKER is contradicting the
// one rule that decides it.
func TestSealRefusesAVerdictThatContradictsTheRule(t *testing.T) {
	_, err := review.Seal(review.Verdict{
		Reviewer: "x",
		Verdict:  review.Approved,
		Findings: []review.Finding{{Path: "a.go", Severity: review.Critical, Causal: true, Why: "pierde el turno"}},
	}, tree, "HEAD", now)
	if err == nil {
		t.Fatal("Seal stored an approved verdict next to a causal CRITICAL")
	}
	if !strings.Contains(err.Error(), review.ChangesRequested) {
		t.Errorf("error %q does not say what the rule derives", err.Error())
	}
}

// TestSealNeedsATree: a verdict that does not say what it reviewed is not
// content-addressed, and cannot be stored.
func TestSealNeedsATree(t *testing.T) {
	if _, err := review.Seal(review.Verdict{Reviewer: "x"}, "", "HEAD", now); err == nil {
		t.Fatal("Seal accepted an empty tree")
	}
}

// TestStoreNameIsTheTree: one file per tree, and the name can only ever be a
// hash. A tree hash that could describe a path is refused, because the store
// must never write outside .odachi/reviews/.
func TestStoreNameIsTheTree(t *testing.T) {
	name, err := review.Name(tree)
	if err != nil {
		t.Fatalf("Name = %v", err)
	}
	if name != tree+".json" {
		t.Errorf("Name = %q, want %q", name, tree+".json")
	}

	path, err := review.Path(tree)
	if err != nil {
		t.Fatalf("Path = %v", err)
	}
	if path != review.Dir+"/"+tree+".json" {
		t.Errorf("Path = %q, want %q", path, review.Dir+"/"+tree+".json")
	}

	for _, bad := range []string{"", "../afuera", "a/b", `a\b`, "arbol.json", "..", "."} {
		if _, err := review.Name(bad); err == nil {
			t.Errorf("Name(%q) = nil error, want a refusal", bad)
		}
	}
}

// TestDifferentTreesAreDifferentFiles: the whole point of the store. Two trees
// never share a verdict, so no approval can float.
func TestDifferentTreesAreDifferentFiles(t *testing.T) {
	one, err := review.Path(tree)
	if err != nil {
		t.Fatalf("Path = %v", err)
	}
	two, err := review.Path(otherTree)
	if err != nil {
		t.Fatalf("Path = %v", err)
	}
	if one == two {
		t.Fatalf("two trees share the file %q", one)
	}
}

// TestStoreKeepsVerdictsOutOfTheTree: the verdict lives inside the candidate it
// describes, so the store's own directory has to be ignored or writing a verdict
// would change the tree that names it.
func TestStoreKeepsVerdictsOutOfTheTree(t *testing.T) {
	if !strings.Contains(review.IgnoreContent, "*") {
		t.Errorf("ignore content %q does not ignore the verdicts", review.IgnoreContent)
	}
	if !strings.Contains(review.IgnoreContent, "!.gitignore") {
		t.Errorf("ignore content %q does not keep itself versioned", review.IgnoreContent)
	}
	if review.IgnoreName != ".gitignore" {
		t.Errorf("IgnoreName = %q, want .gitignore", review.IgnoreName)
	}
	if review.Dir != ".odachi/reviews" {
		t.Errorf("Dir = %q, want .odachi/reviews", review.Dir)
	}
}
