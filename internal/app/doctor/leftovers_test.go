package doctor

import (
	"strings"
	"testing"

	domaindoctor "github.com/biggs-100/odachi/internal/domain/doctor"
)

// TestALegacyFenceIsAProblemNotANotice: after the rules moved to odd/, a project
// that still keeps its fence in .odachi/ has a fence nobody enforces. That is
// worse than no fence — the human believes it is there — so doctor fails and
// names the command that prints the move.
func TestALegacyFenceIsAProblemNotANotice(t *testing.T) {
	files := configured().put(".odachi/fence.toml", "[[rule]]\nnot = \"pagos\"\n")

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if result.Code != CodeProblems {
		t.Fatalf("a fence nobody reads must fail the run:\n%s", strings.Join(result.Lines, "\n"))
	}
	got := checkOf(t, result, domaindoctor.CheckLeftovers)
	if got.Status != domaindoctor.StatusProblem {
		t.Fatalf("check %d = %s, want a problem", domaindoctor.CheckLeftovers, domaindoctor.Line(got))
	}
	if !strings.Contains(got.Detail, ".odachi/fence.toml") || !strings.Contains(got.Detail, "odd/fence.toml") {
		t.Errorf("Detail = %q, want it to name both paths", got.Detail)
	}
	if !strings.Contains(got.Fix, "odachi store move") {
		t.Errorf("Fix = %q, want it to name the migration", got.Fix)
	}
}

// TestTheOldFolderIsAWarningOnceTheRulesMoved: a project halfway through the
// migration — the rules in odd/, the old copies still there — is not broken. The
// line says the leftovers are dead weight and how to finish.
func TestTheOldFolderIsAWarningOnceTheRulesMoved(t *testing.T) {
	files := configured().
		put("odd/fence.toml", "[[rule]]\nnot = \"pagos\"\n").
		put(".odachi/fence.toml", "[[rule]]\nnot = \"pagos\"\n").
		put(".odachi/reviews/abc.json", "{}").
		mkdir(".odachi/backups")

	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	if result.Code != CodeOK {
		t.Fatalf("a leftover is not a problem:\n%s", strings.Join(result.Lines, "\n"))
	}
	got := checkOf(t, result, domaindoctor.CheckLeftovers)
	if got.Status != domaindoctor.StatusWarning {
		t.Fatalf("check %d = %s, want a warning", domaindoctor.CheckLeftovers, domaindoctor.Line(got))
	}
	for _, want := range []string{".odachi/reviews", ".odachi/backups"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("Detail = %q, want it to name %q", got.Detail, want)
		}
	}
}

// TestAProjectWithNoOldFolderIsClean: the check has something to say even when
// there is nothing to migrate, and it says it as an answer, not a silence.
func TestAProjectWithNoOldFolderIsClean(t *testing.T) {
	result := run(configured(), cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	got := checkOf(t, result, domaindoctor.CheckLeftovers)
	if got.Status != domaindoctor.StatusOK {
		t.Fatalf("check %d = %s, want ok", domaindoctor.CheckLeftovers, domaindoctor.Line(got))
	}
	if !strings.Contains(got.Detail, "no .odachi/") {
		t.Errorf("Detail = %q, want it to say there is no old folder", got.Detail)
	}
}
