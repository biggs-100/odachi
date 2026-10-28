package doctor

import (
	"strings"
	"testing"

	domaindoctor "github.com/biggs-100/odachi/internal/domain/doctor"
)

// TestTheRepositoryRulesAreEnough: after `odachi init` the evidence is ignored by
// the project's own .gitignore, and the per-directory files are not there at all.
// Doctor must call that clean, because it is the state init produces.
func TestTheRepositoryRulesAreEnough(t *testing.T) {
	files := configured().
		put(".gitignore", "# odachi:begin\n/proof/\n*.bak-*\n# odachi:end\n").
		put("proof/0001-turnos.txt", "# tree abc\n")

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if result.Code != CodeOK {
		t.Fatalf("the repository's rules protect the evidence:\n%s", strings.Join(result.Lines, "\n"))
	}
	got := checkOf(t, result, domaindoctor.CheckIgnored)
	if got.Status != domaindoctor.StatusOK {
		t.Fatalf("check 10 = %s, want ok", domaindoctor.Line(got))
	}
	for _, want := range []string{"proof"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("the line must say what it found protected, missing %q: %s", want, domaindoctor.Line(got))
		}
	}
}

// TestUnprotectedEvidenceNamesInit: when nothing ignores the evidence, the fix is
// the command that writes the rule. `check` and `review` are what produce evidence,
// not what protects it, and sending a human there would resolve nothing.
func TestUnprotectedEvidenceNamesInit(t *testing.T) {
	files := configured().
		put(".gitignore", "node_modules/\n").
		put("proof/0001-turnos.txt", "# tree abc\n")

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	got := checkOf(t, result, domaindoctor.CheckIgnored)
	if got.Status != domaindoctor.StatusProblem {
		t.Fatalf("an unprotected proof/ is a problem: %s", domaindoctor.Line(got))
	}
	if !strings.Contains(got.Fix, "odachi init") {
		t.Errorf("Fix = %q, want it to name odachi init", got.Fix)
	}
}

// TestAProjectThatStillCarriesTheOldReviewRule: the verdicts left the worktree,
// so the old rule says nothing about anything. What still matters is proof/.
func TestAProjectThatStillCarriesTheOldReviewRule(t *testing.T) {
	files := configured().
		put(".gitignore", "/the store's reviews/\n").
		put("proof/0001-turnos.txt", "# tree abc\n")

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	got := checkOf(t, result, domaindoctor.CheckIgnored)
	if got.Status != domaindoctor.StatusProblem {
		t.Fatalf("proof/ is not ignored: want a problem, got %s", domaindoctor.Line(got))
	}
	if !strings.Contains(got.Detail, "proof/") {
		t.Errorf("Detail = %q, want it to name proof/", got.Detail)
	}
	if strings.Contains(got.Detail, "reviews") {
		t.Errorf("the old rule is gone from the tool, and the line still mentions it: %s", domaindoctor.Line(got))
	}
}
