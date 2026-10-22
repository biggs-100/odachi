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
		put(".gitignore", "# odachi:begin\n/proof/\n/.odachi/reviews/\n# odachi:end\n").
		put("proof/0001-turnos.txt", "# tree abc\n").
		put(".odachi/reviews/abc.json", `{"verdict":"approved"}`)

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if result.Code != CodeOK {
		t.Fatalf("the repository's rules protect the evidence:\n%s", strings.Join(result.Lines, "\n"))
	}
	got := checkOf(t, result, domaindoctor.CheckIgnored)
	if got.Status != domaindoctor.StatusOK {
		t.Fatalf("check 10 = %s, want ok", domaindoctor.Line(got))
	}
	for _, want := range []string{"proof", ".odachi/reviews"} {
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

// TestOneOfTheTwoDirectoriesProtected is the half-done project: a rule for the
// store and none for the proof still moves the tree on the first run.
func TestOneOfTheTwoDirectoriesProtected(t *testing.T) {
	files := configured().
		put(".gitignore", "/.odachi/reviews/\n").
		put("proof/0001-turnos.txt", "# tree abc\n")

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	got := checkOf(t, result, domaindoctor.CheckIgnored)
	if got.Status != domaindoctor.StatusProblem {
		t.Fatalf("proof/ is not ignored: want a problem, got %s", domaindoctor.Line(got))
	}
	if strings.Contains(got.Detail, ".odachi/reviews no está ignorado") {
		t.Errorf("the store IS ignored, and the line says it is not: %s", domaindoctor.Line(got))
	}
	if !strings.Contains(got.Detail, "proof/") {
		t.Errorf("Detail = %q, want it to name proof/", got.Detail)
	}
}
