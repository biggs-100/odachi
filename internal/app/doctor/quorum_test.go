package doctor

import (
	"strings"
	"testing"

	domaindoctor "github.com/biggs-100/odachi/internal/domain/doctor"
)

// lensTable is a review file the human wrote: the switch and the model per lens.
func lensTable(models map[string]string) string {
	var b strings.Builder
	b.WriteString("mode = \"on\"\n\n[lenses]\n")
	for _, lens := range []string{"risk", "readability", "reliability", "resilience"} {
		if model, ok := models[lens]; ok {
			b.WriteString(lens + " = \"" + model + "\"\n")
		}
	}
	return b.String()
}

// TestOneModelIsAWarningThatNamesItAndTheFix: four lenses answered by one model are
// one opinion in four costumes. The line says the number, the model and the fix -
// one model per lens in odd/review.toml - because a human cannot act on "the quorum
// is poor" (0316).
func TestOneModelIsAWarningThatNamesItAndTheFix(t *testing.T) {
	files := configured()
	files.put("odd/review.toml", lensTable(map[string]string{
		"risk":        "opencode-go/deepseek-v4.1-flash",
		"readability": "opencode-go/deepseek-v4.1-flash",
		"reliability": "opencode-go/deepseek-v4.1-flash",
		"resilience":  "opencode-go/deepseek-v4.1-flash",
	}))

	check := checkOf(t, run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi"), domaindoctor.CheckQuorum)

	if check.Status != domaindoctor.StatusWarning {
		t.Fatalf("check 17 is %v with one model, want a warning: %s", check.Status, domaindoctor.Line(check))
	}
	for _, want := range []string{"opencode-go/deepseek-v4.1-flash", "1 of 4", "odd/review.toml"} {
		if !strings.Contains(check.Detail, want) {
			t.Errorf("the line does not name %q: %s", want, check.Detail)
		}
	}
}

// TestTheWarningSaysOneModelIsEnough: the quorum is a warning about independence,
// not a requirement to own a second vendor. A person with one model can still
// review, and the line has to say so (0358).
func TestTheWarningSaysOneModelIsEnough(t *testing.T) {
	files := configured()
	files.put("odd/review.toml", lensTable(map[string]string{
		"risk":        "opencode-go/deepseek-v4.1-flash",
		"readability": "opencode-go/deepseek-v4.1-flash",
		"reliability": "opencode-go/deepseek-v4.1-flash",
		"resilience":  "opencode-go/deepseek-v4.1-flash",
	}))

	check := checkOf(t, run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi"), domaindoctor.CheckQuorum)
	for _, want := range []string{"one model is enough to review", "a second is optional"} {
		if !strings.Contains(check.Detail, want) {
			t.Errorf("the warning does not say %q: %s", want, check.Detail)
		}
	}
}

// TestTwoModelsAreNoWarning: the day a second vendor judges one lens, the number
// moves and the line stops warning.
func TestTwoModelsAreNoWarning(t *testing.T) {
	files := configured()
	files.put("odd/review.toml", lensTable(map[string]string{
		"risk":        "vendor/one",
		"readability": "vendor/two",
		"reliability": "vendor/one",
		"resilience":  "vendor/two",
	}))

	check := checkOf(t, run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi"), domaindoctor.CheckQuorum)

	if check.Status != domaindoctor.StatusOK {
		t.Fatalf("check 17 is %v with two models, want ok: %s", check.Status, domaindoctor.Line(check))
	}
	if !strings.Contains(check.Detail, "2 of 4") {
		t.Errorf("the line does not say how many models judge: %s", check.Detail)
	}
}

// TestNoReviewTableHasNoQuorumToCount: a project that runs no review has no lens
// table, and this check does not invent a warning about a quorum nobody declared.
func TestNoReviewTableHasNoQuorumToCount(t *testing.T) {
	check := checkOf(t, run(configured(), cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi"), domaindoctor.CheckQuorum)

	if check.Status != domaindoctor.StatusOK {
		t.Fatalf("check 17 is %v with no review file, want ok: %s", check.Status, domaindoctor.Line(check))
	}
	if !strings.Contains(check.Detail, "no "+"odd/review.toml") {
		t.Errorf("the line does not say the file is missing: %s", check.Detail)
	}
}

// TestABrokenTableIsAProblem: a file the decoder refuses is named with its fix, and
// the review that reads it would fail the same way.
func TestABrokenTableIsAProblem(t *testing.T) {
	files := configured()
	files.put("odd/review.toml", "mode = \"on\"\n[lenses]\nperformance = \"m\"\n")

	check := checkOf(t, run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi"), domaindoctor.CheckQuorum)

	if check.Status != domaindoctor.StatusProblem {
		t.Fatalf("check 17 is %v with a broken table, want a problem: %s", check.Status, domaindoctor.Line(check))
	}
	if !strings.Contains(check.Fix, "odd/review.toml") {
		t.Errorf("the fix does not name the file: %q", check.Fix)
	}
}

// TestTheQuorumNeverFailsTheReport: the line is information, and a project whose
// lenses share one model still exits 0. A warning that moved the exit code would be
// the tool deciding the project has to buy a second vendor (0316).
func TestTheQuorumNeverFailsTheReport(t *testing.T) {
	files := configured()
	files.put("odd/review.toml", lensTable(map[string]string{
		"risk":        "one/model",
		"readability": "one/model",
		"reliability": "one/model",
		"resilience":  "one/model",
	}))

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if result.Code != 0 {
		t.Fatalf("doctor exits %d with a quorum warning, want 0: %s", result.Code, result.Message)
	}
	if result.Warnings == 0 {
		t.Error("the report counts no warning, and the quorum is one")
	}
}
