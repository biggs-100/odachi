package review_test

import (
	"slices"
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

// The model that judged survives the file, and a verdict written before the
// field existed still parses: an absent model means "nobody named one", which is
// not the same as the builder's.
func TestTheVerdictKeepsTheModelThatJudged(t *testing.T) {
	sealed, err := review.Seal(review.Verdict{
		Reviewer: "revisor --model provider/model",
		Model:    "provider/model",
	}, tree, "HEAD", now)
	if err != nil {
		t.Fatalf("Seal = %v", err)
	}

	back, err := review.Parse(review.Render(sealed))
	if err != nil {
		t.Fatalf("Parse = %v", err)
	}
	if back.Model != "provider/model" {
		t.Errorf("Model = %q, want the identity that judged", back.Model)
	}

	old, err := review.Parse([]byte(`{"tree":"` + tree + `","base":"HEAD","reviewer":"humano","at":"` + now.Format(time.RFC3339) + `","findings":[],"verdict":"approved"}`))
	if err != nil {
		t.Fatalf("Parse of a verdict without a model: %v", err)
	}
	if old.Model != "" {
		t.Errorf("Model = %q, want empty for an older verdict", old.Model)
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
// must never write outside the store's reviews/.
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

// TestStoreNamesTheVerdictsAndNothingElse: the store's directory is named once,
// and it holds no ignore file. A verdict lives outside the worktree, so the tool
// does not need a rule to keep it out of the tree it describes — that is where
// the immutability comes from now, not from a file somebody remembers to write.
func TestStoreNamesTheVerdictsAndNothingElse(t *testing.T) {
	if review.Dir != "reviews" {
		t.Errorf("Dir = %q, want reviews", review.Dir)
	}
	path, err := review.Path(tree)
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	if path != "reviews/"+tree+".json" {
		t.Errorf("Path = %q, want reviews/%s.json", path, tree)
	}
}

// blobContent is a finding's content receipt: the blob id git owns for the
// reviewed bytes. Hex, like the tree, because both name content.
const blobContent = "9c1185a5c5e9fc54612808977ee8f548b2258d31"

// TestAFindingCanCarryItsContentReceipt: a finding knows which bytes it judged
// (blob) and which lens produced it. The round trip is the point: a verdict
// stored with the receipt parses back with the same identity, and sealing keeps
// what the tool did not compute — the receipt survives the write.
func TestAFindingCanCarryItsContentReceipt(t *testing.T) {
	verdict := review.Verdict{
		Tree:     tree,
		Reviewer: "agent",
		Verdict:  review.Approved,
		Findings: []review.Finding{{
			Path:     "internal/app/review/review.go",
			Line:     421,
			Severity: review.Warning,
			Causal:   true,
			Why:      "el recibo se pierde al sellar",
			Blob:     blobContent,
			Lens:     "reliability",
		}},
	}

	if err := verdict.Validate(); err != nil {
		t.Fatalf("a complete content receipt must validate, got: %s", err)
	}

	parsed, err := review.Parse(review.Render(verdict))
	if err != nil {
		t.Fatalf("render then parse must round trip, got: %s", err)
	}
	got := parsed.Findings[0]
	if got.Blob != blobContent {
		t.Errorf("Blob = %q, want %q: a verdict that loses the blob loses its identity", got.Blob, blobContent)
	}
	if got.Lens != "reliability" {
		t.Errorf("Lens = %q, want reliability", got.Lens)
	}
}

// TestABlobThatIsNotAHashIsRefused: the blob is content identity, and a blob
// that is not hex names nothing. An invented hash under an honest name would be
// worse than no receipt at all.
func TestABlobThatIsNotAHashIsRefused(t *testing.T) {
	verdict := review.Verdict{
		Tree:     tree,
		Reviewer: "agent",
		Verdict:  review.Approved,
		Findings: []review.Finding{{
			Path: "internal/app/review/review.go", Line: 1,
			Severity: review.Warning, Causal: true, Why: "defect",
			Blob: "not-a-blob", Lens: "reliability",
		}},
	}

	err := verdict.Validate()
	if err == nil {
		t.Fatal("a finding with a blob that is not a hash must be refused")
	}
	if !strings.Contains(err.Error(), "blob") {
		t.Errorf("the error must name the field it refuses, got: %s", err)
	}
}

// TestALensTheToolDoesNotKnowIsRefused: a finding says which lens produced it,
// and the lens vocabulary is closed (like the severity vocabulary). An unknown
// lens would silently fail the coverage arithmetic later, so it is refused at
// the door.
func TestALensTheToolDoesNotKnowIsRefused(t *testing.T) {
	verdict := review.Verdict{
		Tree:     tree,
		Reviewer: "agent",
		Verdict:  review.Approved,
		Findings: []review.Finding{{
			Path: "internal/app/review/review.go", Line: 1,
			Severity: review.Warning, Causal: true, Why: "defect",
			Blob: blobContent, Lens: "elegancia",
		}},
	}

	err := verdict.Validate()
	if err == nil {
		t.Fatal("a finding that names a lens the tool does not know must be refused")
	}
	if !strings.Contains(err.Error(), "lens") {
		t.Errorf("the error must name the lens, got: %s", err)
	}
}

// TestFindingsWithoutBlobOrLensStillParse: every verdict written before this
// decision parses as it always did. The old rows carry no receipt and the new
// reader must not demand one — one-time cost, not a migration.
func TestFindingsWithoutBlobOrLensStillParse(t *testing.T) {
	old := []byte(`{
  "tree": "` + tree + `",
  "base": "HEAD",
  "reviewer": "agent",
  "at": "2026-10-21T20:31:00Z",
  "findings": [
    {
      "path": "internal/app/review/review.go",
      "line": 0,
      "severity": "WARNING",
      "causal": false,
      "why": "ya estaba"
    }
  ],
  "verdict": "approved"
}`)

	parsed, err := review.Parse(old)
	if err != nil {
		t.Fatalf("a verdict written before blob keys must still parse, got: %s", err)
	}
	if len(parsed.Findings) != 1 || parsed.Findings[0].Blob != "" || parsed.Findings[0].Lens != "" {
		t.Errorf("an old finding parses as an unkeyed finding, got %+v", parsed.Findings)
	}
}

// TestAPassIsOneCleanLensAnswer: a pass is the coverage row for a blob a lens
// audited and found clean. The round trip keeps it, and Validate refuses one
// without path, blob or a lens the tool knows.
func TestAPassIsOneCleanLensAnswer(t *testing.T) {
	verdict := review.Verdict{
		Tree:     tree,
		Reviewer: "agent",
		Verdict:  review.Approved,
		Findings: []review.Finding{},
		Passes: []review.Pass{
			{Path: "internal/app/review/review.go", Blob: blobContent, Lens: "reliability"},
		},
	}

	if err := verdict.Validate(); err != nil {
		t.Fatalf("a complete pass must validate, got: %s", err)
	}

	parsed, err := review.Parse(review.Render(verdict))
	if err != nil {
		t.Fatalf("render then parse must round trip, got: %s", err)
	}
	if len(parsed.Passes) != 1 {
		t.Fatalf("passes = %d, want 1", len(parsed.Passes))
	}
	got := parsed.Passes[0]
	if got.Path != "internal/app/review/review.go" || got.Blob != blobContent || got.Lens != "reliability" {
		t.Errorf("pass = %+v, want the audited (path, blob, lens)", got)
	}
}

// TestAIncompletePassIsRefused: a pass that names a blob that is not a hash
// names bytes nobody measured, and one with an invented lens answers a question
// nobody asked. A pass WITHOUT a blob validates: it is a fresh answer waiting
// for the tool's stamp, and it is inert until stamped.
func TestAIncompletePassIsRefused(t *testing.T) {
	cases := []struct {
		name string
		pass review.Pass
		want string
	}{
		{name: "blob that is not a hash", pass: review.Pass{Path: "a.go", Blob: "not-a-blob", Lens: "reliability"}, want: "blob"},
		{name: "no lens", pass: review.Pass{Path: "a.go", Blob: blobContent}, want: "lens"},
		{name: "unknown lens", pass: review.Pass{Path: "a.go", Blob: blobContent, Lens: "elegancia"}, want: "lens"},
		{name: "no path", pass: review.Pass{Blob: blobContent, Lens: "reliability"}, want: "path"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			verdict := review.Verdict{
				Tree:     tree,
				Reviewer: "agent",
				Verdict:  review.Approved,
				Findings: []review.Finding{},
				Passes:   []review.Pass{testCase.pass},
			}

			err := verdict.Validate()
			if err == nil {
				t.Fatal("an incomplete pass must be refused")
			}
			if !strings.Contains(err.Error(), testCase.want) {
				t.Errorf("the error must name %q, got: %s", testCase.want, err)
			}
		})
	}
}

// TestAUnstampedPassIsTheReviewerSAnswer: a pass without a blob is valid — it
// is how a fresh answer looks before the tool stamps it — and the round trip
// keeps it empty.
func TestAUnstampedPassIsTheReviewerSAnswer(t *testing.T) {
	verdict := review.Verdict{
		Tree:     tree,
		Reviewer: "agent",
		Verdict:  review.Approved,
		Findings: []review.Finding{},
		Passes:   []review.Pass{{Path: "a.go", Lens: "reliability"}},
	}

	if err := verdict.Validate(); err != nil {
		t.Fatalf("an unstamped pass is a fresh answer, not an invalid verdict: %s", err)
	}

	parsed, err := review.Parse(review.Render(verdict))
	if err != nil {
		t.Fatalf("render then parse must round trip, got: %s", err)
	}
	if len(parsed.Passes) != 1 || parsed.Passes[0].Blob != "" {
		t.Errorf("pass = %+v, want the unstamped row kept empty", parsed.Passes)
	}
}

// TestUncoveredIsTheCoverageArithmetic: for every file that must answer and
// each lens its path class asks for, the verdict must carry a current answer — a
// finding with a receipt or a pass. Stale blobs and missing lens attributions
// are gaps; a verdict written before content receipts covers nothing.
func TestUncoveredIsTheCoverageArithmetic(t *testing.T) {
	blobs := map[string]string{
		"init.go":     "aaaa00000a0b0c0d0e0f1e2d3c4b5a6978695a4b",
		"review.go":   "bbbb11111a0b0c0d0e0f1e2d3c4b5a6978695a4b",
		"contract.go": "cccc22222a0b0c0d0e0f1e2d3c4b5a6978695a4b",
	}
	need := map[string][]string{
		"init.go":     {"reliability"},
		"review.go":   {"reliability"},
		"contract.go": {"risk", "readability", "reliability", "resilience"},
	}

	t.Run("full coverage has no gaps", func(t *testing.T) {
		verdict := review.Verdict{
			Findings: []review.Finding{},
			Passes: []review.Pass{
				{Path: "init.go", Blob: blobs["init.go"], Lens: "reliability"},
				{Path: "review.go", Blob: blobs["review.go"], Lens: "reliability"},
				{Path: "contract.go", Blob: blobs["contract.go"], Lens: "risk"},
				{Path: "contract.go", Blob: blobs["contract.go"], Lens: "readability"},
				{Path: "contract.go", Blob: blobs["contract.go"], Lens: "reliability"},
				{Path: "contract.go", Blob: blobs["contract.go"], Lens: "resilience"},
			},
		}

		gaps := verdict.Uncovered(need, blobs)
		if len(gaps) != 0 {
			t.Errorf("gaps = %v, want none", gaps)
		}
	})

	t.Run("a finding covers its own row", func(t *testing.T) {
		verdict := review.Verdict{
			Findings: []review.Finding{
				{Path: "review.go", Line: 3, Severity: review.Warning, Causal: true,
					Why: "defect", Blob: blobs["review.go"], Lens: "reliability"},
			},
			Passes: []review.Pass{
				{Path: "init.go", Blob: blobs["init.go"], Lens: "reliability"},
				{Path: "contract.go", Blob: blobs["contract.go"], Lens: "risk"},
				{Path: "contract.go", Blob: blobs["contract.go"], Lens: "readability"},
				{Path: "contract.go", Blob: blobs["contract.go"], Lens: "reliability"},
				{Path: "contract.go", Blob: blobs["contract.go"], Lens: "resilience"},
			},
		}

		gaps := verdict.Uncovered(need, blobs)
		if len(gaps) != 0 {
			t.Errorf("gaps = %v, want none: the finding is the answer for its row", gaps)
		}
	})

	t.Run("a stale blob is a gap", func(t *testing.T) {
		verdict := review.Verdict{
			Findings: []review.Finding{},
			Passes: []review.Pass{
				{Path: "init.go", Blob: "dddd33333a0b0c0d0e0f1e2d3c4b5a6978695a4b", Lens: "reliability"},
				{Path: "review.go", Blob: blobs["review.go"], Lens: "reliability"},
				{Path: "contract.go", Blob: blobs["contract.go"], Lens: "risk"},
				{Path: "contract.go", Blob: blobs["contract.go"], Lens: "readability"},
				{Path: "contract.go", Blob: blobs["contract.go"], Lens: "reliability"},
				{Path: "contract.go", Blob: blobs["contract.go"], Lens: "resilience"},
			},
		}

		gaps := verdict.Uncovered(need, blobs)
		if len(gaps) != 1 || gaps[0].Path != "init.go" || gaps[0].Lens != "reliability" {
			t.Errorf("gaps = %v, want exactly init.go/reliability: the receipt does not describe these bytes", gaps)
		}
	})

	t.Run("an old verdict covers nothing", func(t *testing.T) {
		verdict := review.Verdict{Findings: []review.Finding{
			{Path: "review.go", Line: 0, Severity: review.Warning, Causal: false, Why: "already there"},
		}}

		gaps := verdict.Uncovered(need, blobs)
		if len(gaps) == 0 {
			t.Fatal("a verdict without content receipts must leave every row uncovered")
		}
		for _, gap := range gaps {
			if !slices.Contains(need[gap.Path], gap.Lens) {
				t.Errorf("gap %s/%s is not one of %v", gap.Path, gap.Lens, need[gap.Path])
			}
		}
	})

	t.Run("a file with unneeded lenses never asks twice", func(t *testing.T) {
		verdict := review.Verdict{
			Findings: []review.Finding{},
			Passes: []review.Pass{
				{Path: "init.go", Blob: blobs["init.go"], Lens: "reliability"},
				{Path: "review.go", Blob: blobs["review.go"], Lens: "reliability"},
			},
		}

		gaps := verdict.Uncovered(need, blobs)
		if len(gaps) != 4 {
			t.Fatalf("gaps = %d, want the four contract.go rows", len(gaps))
		}
		for _, gap := range gaps {
			if gap.Path != "contract.go" {
				t.Errorf("unexpected gap %s/%s", gap.Path, gap.Lens)
			}
		}
	})
}
