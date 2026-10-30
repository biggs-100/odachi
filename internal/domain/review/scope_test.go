package review_test

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
	"github.com/biggs-100/odachi/internal/domain/schema"
)

// TestRaisedNamesOnlyWhatTheEvidenceDidNotAsk: --lens adds, and the record says
// which lenses came from the person and not from the evidence. The order is the
// canonical one, so two runs that asked for the same scope write the same bytes.
func TestRaisedNamesOnlyWhatTheEvidenceDidNotAsk(t *testing.T) {
	cases := []struct {
		name     string
		required []string
		extra    []string
		want     []string
	}{
		{"nothing was asked for", []string{review.Reliability}, nil, nil},
		{"the evidence already asked for it", []string{review.Reliability}, []string{review.Reliability}, nil},
		{"one lens was added", []string{review.Reliability}, []string{review.Risk}, []string{review.Risk}},
		{"a passive tree forced to one lens", nil, []string{review.Risk}, []string{review.Risk}},
		{"two added, in the fixed order", nil, []string{review.Resilience, review.Risk}, []string{review.Risk, review.Resilience}},
	}

	for _, c := range cases {
		if got := review.Raised(c.required, c.extra); !slices.Equal(got, c.want) {
			t.Errorf("%s: Raised(%v, %v) = %v, want %v", c.name, c.required, c.extra, got, c.want)
		}
	}
}

// TestTheScopeLivesBesideTheVerdict: the record is a sibling of the verdict, named
// by the tree it judged, so reading one never depends on the other and a verdict is
// never rewritten to hold a decision somebody else made.
func TestTheScopeLivesBesideTheVerdict(t *testing.T) {
	path, err := review.ScopePath("5ed52095cfd95f668453f730ffe25832fe81bfd3")
	if err != nil {
		t.Fatalf("ScopePath: %v", err)
	}
	if want := "reviews/5ed52095cfd95f668453f730ffe25832fe81bfd3.scope.json"; path != want {
		t.Errorf("ScopePath = %q, want %q", path, want)
	}

	// A tree names the file: anything that is not a hash cannot be one, and a
	// record filed under a path would be a record nobody can find.
	if _, err := review.ScopePath("../etc"); err == nil {
		t.Error("a tree that is not a hash was accepted as a file name")
	}
	if _, err := review.ScopePath(""); err == nil {
		t.Error("an empty tree was accepted as a file name")
	}
}

// TestTheScopeFileRoundTrips: the canonical file is what a person reads, so it is
// indented, ordered, stamped with the schema, and a parse of it answers the same
// document. Strict on purpose: a key this binary does not know is a file written by
// a newer odachi, and the schema answers that before the strict pass names it.
func TestTheScopeFileRoundTrips(t *testing.T) {
	scope := review.Scope{
		Required: []string{review.Reliability},
		Added:    []string{review.Risk},
		Asked:    "risk",
		Because:  "the change touches the gate itself",
		By:       "biggs-100",
		At:       "2026-10-29T16:00:00Z",
	}

	data := review.RenderScope(scope)
	text := string(data)
	if !strings.HasPrefix(text, "{\n  \"schema\": "+strconv.Itoa(schema.Current)+",\n") {
		t.Errorf("the file does not open with the schema:\n%s", text)
	}
	if !strings.HasSuffix(text, "}\n") {
		t.Errorf("the file does not end with a newline:\n%q", text)
	}

	parsed, err := review.ParseScope(data)
	if err != nil {
		t.Fatalf("ParseScope: %v", err)
	}
	if !slices.Equal(parsed.Required, scope.Required) || !slices.Equal(parsed.Added, scope.Added) {
		t.Errorf("the lenses did not round trip: %+v", parsed)
	}
	if parsed.Asked != scope.Asked || parsed.Because != scope.Because || parsed.By != scope.By || parsed.At != scope.At {
		t.Errorf("the scope did not round trip: %+v", parsed)
	}
}

// TestAScopeThatRaisedNothingIsRefused: the file is evidence, not ceremony. A
// record with no added lens says nothing happened, and a file that says nothing is
// the noise this feature exists to avoid - so it is refused instead of written.
func TestAScopeThatRaisedNothingIsRefused(t *testing.T) {
	cases := map[string]string{
		"nothing raised": `{"schema":1,"required":["reliability"],"added":[],"asked":"","at":"2026-10-29T16:00:00Z"}`,
		"no flag":        `{"schema":1,"required":["reliability"],"added":["risk"],"asked":"","at":"2026-10-29T16:00:00Z"}`,
	}

	for name, doc := range cases {
		if _, err := review.ParseScope([]byte(doc)); err == nil {
			t.Errorf("%s: a scope that says nothing was accepted", name)
		}
	}
}

// TestAScopeFromANewerOdachiAsksForAnUpgrade: the schema is read before the strict
// pass, so a field this binary has never heard of is answered as an old binary, not
// as a broken file.
func TestAScopeFromANewerOdachiAsksForAnUpgrade(t *testing.T) {
	_, err := review.ParseScope([]byte(`{"schema":99,"added":["risk"],"asked":"risk","future":"x"}`))
	if err == nil {
		t.Fatal("a scope from a newer odachi was accepted")
	}
	if !strings.Contains(err.Error(), "newer odachi") {
		t.Errorf("the refusal does not name the way out: %v", err)
	}

	if _, err := review.ParseScope([]byte(`{"schema":1,"added":["risk"],"asked":"risk","future":"x"}`)); err == nil {
		t.Error("an unknown key was accepted")
	}
}
