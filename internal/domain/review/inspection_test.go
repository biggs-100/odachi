package review_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// inspectionOf is a reviewer's declaration of the paths it read, the way a
// complete answer writes it.
func inspectionOf(paths ...string) *review.Inspection {
	return &review.Inspection{Status: review.InspectionCompleted, Paths: paths}
}

// TestInspectionRefusesAStatusOutsideTheVocabulary: two values, and the refusal
// names both. A status the tool does not understand is not a completed
// inspection, and guessing on the reviewer's behalf is what this field exists to
// end.
func TestInspectionRefusesAStatusOutsideTheVocabulary(t *testing.T) {
	verdict := review.Verdict{
		Reviewer:   "x",
		Inspection: &review.Inspection{Status: "done", Paths: []string{"a.go"}},
	}

	err := verdict.Validate()
	if err == nil {
		t.Fatal("a declaration with an invented status passed")
	}
	for _, want := range []string{"completed", "unavailable"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want the vocabulary to name %q", err, want)
		}
	}
}

// TestInspectionStatusDecidesWhatItMustSay: "completed" is a claim about every
// path it lists, so it cannot list none; "unavailable" is a claim that nothing
// was read, so it needs the reason and no path.
func TestInspectionStatusDecidesWhatItMustSay(t *testing.T) {
	cases := []struct {
		name       string
		inspection *review.Inspection
		want       string
	}{
		{
			name:       "completed without a path",
			inspection: &review.Inspection{Status: review.InspectionCompleted},
			want:       "path",
		},
		{
			name:       "unavailable without a reason",
			inspection: &review.Inspection{Status: review.InspectionUnavailable},
			want:       "reason",
		},
		{
			name:       "unavailable that still names a path",
			inspection: &review.Inspection{Status: review.InspectionUnavailable, Paths: []string{"a.go"}, Reason: "the reviewer timed out"},
			want:       "path",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			verdict := review.Verdict{Reviewer: "x", Inspection: testCase.inspection}

			err := verdict.Validate()
			if err == nil {
				t.Fatalf("%s passed", testCase.name)
			}
			if !strings.Contains(err.Error(), testCase.want) {
				t.Errorf("error = %q, want it to name %q", err, testCase.want)
			}
		})
	}
}

// TestInspectionNamesEachPathOnce: the list is a set, and a path written twice
// makes a count that says nothing.
func TestInspectionNamesEachPathOnce(t *testing.T) {
	verdict := review.Verdict{
		Reviewer:   "x",
		Inspection: inspectionOf("a.go", "A.go"),
	}

	err := verdict.Validate()
	if err == nil {
		t.Fatal("a declaration that names the same path twice passed")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "a.go") {
		t.Errorf("error = %q, want the repeated path", err)
	}
}

// TestCoverageNamesWhatTheDeclarationLeftOutAndWhatNobodyChanged: the two
// directions of the comparison, named. A caller reading the refusal learns which
// path the reviewer never declared and which one it invented.
func TestCoverageNamesWhatTheDeclarationLeftOutAndWhatNobodyChanged(t *testing.T) {
	verdict := review.Verdict{Inspection: inspectionOf("a.go", "notes/x.md")}

	missing, foreign := verdict.InspectionCoverage([]string{"a.go", "b.go"})

	if !slices.Equal(missing, []string{"b.go"}) {
		t.Errorf("missing = %v, want [b.go]", missing)
	}
	if !slices.Equal(foreign, []string{"notes/x.md"}) {
		t.Errorf("foreign = %v, want [notes/x.md]", foreign)
	}
}

// TestCoverageComparesPathsTheWayGitNamesThem: the declaration is written by a
// model, so a backslash and a capital letter are the same file, not a foreign
// path.
func TestCoverageComparesPathsTheWayGitNamesThem(t *testing.T) {
	verdict := review.Verdict{Inspection: inspectionOf(`Internal\App\A.go`, " b.go ")}

	missing, foreign := verdict.InspectionCoverage([]string{"internal/app/a.go", "b.go"})

	if len(missing) != 0 || len(foreign) != 0 {
		t.Errorf("missing = %v, foreign = %v, want both empty", missing, foreign)
	}
}

// TestAVerdictWrittenBeforeInspectionStillParses: the store keeps verdicts about
// trees nobody can rewrite. A verdict without the field is not incomplete, it is
// older, and reading it must not start a refusal.
func TestAVerdictWrittenBeforeInspectionStillParses(t *testing.T) {
	parsed, err := review.Parse([]byte(`{"reviewer":"x","findings":[]}`))
	if err != nil {
		t.Fatalf("an old verdict does not parse: %v", err)
	}
	if parsed.Inspection != nil {
		t.Errorf("inspection = %+v, want none", parsed.Inspection)
	}
	if err := parsed.Validate(); err != nil {
		t.Errorf("an old verdict is refused: %v", err)
	}
}

// TestParseReadsTheDeclaration: the field is part of the answer now, and the
// strict decoder is what proves it.
func TestParseReadsTheDeclaration(t *testing.T) {
	parsed, err := review.Parse([]byte(`{"reviewer":"x","findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`))
	if err != nil {
		t.Fatalf("the declaration does not parse: %v", err)
	}
	if parsed.Inspection == nil {
		t.Fatal("the declaration was dropped")
	}
	if parsed.Inspection.Status != review.InspectionCompleted || !slices.Equal(parsed.Inspection.Paths, []string{"a.go"}) {
		t.Errorf("inspection = %+v, want completed over a.go", *parsed.Inspection)
	}
}

// TestMergeKeepsOneDeclarationForTheVerdict: a verdict that grows by lens group
// carries one account of the candidate, not one per group. Both groups were
// handed the same paths, so the union is that set; a group that could not read
// the tree keeps its claim, because the verdict cannot say it read what one of
// its reviewers did not.
func TestMergeKeepsOneDeclarationForTheVerdict(t *testing.T) {
	first := inspectionOf("a.go", "b.go")
	second := inspectionOf("b.go", "c.go")

	merged := review.MergeInspection(first, second)
	if merged == nil || merged.Status != review.InspectionCompleted {
		t.Fatalf("merged = %+v, want a completed declaration", merged)
	}
	if !slices.Equal(merged.Paths, []string{"a.go", "b.go", "c.go"}) {
		t.Errorf("paths = %v, want the union in first-seen order", merged.Paths)
	}

	couldNotRead := &review.Inspection{Status: review.InspectionUnavailable, Reason: "the reviewer timed out"}
	if got := review.MergeInspection(first, couldNotRead); got != couldNotRead {
		t.Errorf("merged = %+v, want the unavailable claim to survive", got)
	}
	if got := review.MergeInspection(nil, first); got != first {
		t.Errorf("merged = %+v, want the only declaration there is", got)
	}
}
