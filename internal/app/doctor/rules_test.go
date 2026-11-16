package doctor

import (
	"strings"
	"testing"

	domaindoctor "github.com/biggs-100/odachi/internal/domain/doctor"
)

// rulesTable is a table whose three rows all resolve: a fence the project keeps,
// a check the review declares, and a test the suite runs.
const rulesTable = `[[rule]]
rule     = "the proof is this tree's real output"
skill    = "drive"
enforces = "fence:proof/**"

[[rule]]
rule     = "the gate runs the linter"
skill    = "commits"
enforces = "check:lint"

[[rule]]
rule     = "the red comes first"
skill    = "tests"
enforces = "test:TestSomething"
`

// reviewWithLint is a review table that declares one deterministic tool.
const reviewWithLint = `mode = "on"

[[check]]
name = "lint"
command = "golangci-lint run ./..."
severity = "CRITICAL"
remedy = "run it with --fix, then fix what it still names"
`

// fenceWithProof is a fence that keeps the proofs out.
const fenceWithProof = `[[rule]]
not = "the stored proofs are the evidence"
paths = ["proof/**"]
`

// TestCheckSixteenResolvesAGateReason: a rule held by the gate names the refusal
// the gate answers with, and doctor resolves it against the reasons the gate
// declares. A name it does not answer is a promise nobody keeps (0314).
func TestCheckSixteenResolvesAGateReason(t *testing.T) {
	rows := `[[rule]]
rule     = "a requirement is the human's words"
skill    = "drive"
enforces = "gate:requirement_not_the_human_words"
`
	files := configured()
	files.put("odd/rules.toml", rows)
	check := checkOf(t, run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi"), domaindoctor.CheckRules)
	if check.Status != domaindoctor.StatusOK {
		t.Fatalf("check 16 is %v with a gate reason the tool answers, want ok: %s", check.Status, domaindoctor.Line(check))
	}

	files = configured()
	files.put("odd/rules.toml", strings.Replace(rows, "requirement_not_the_human_words", "a_reason_nobody_declared", 1))
	check = checkOf(t, run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi"), domaindoctor.CheckRules)
	if check.Status != domaindoctor.StatusProblem {
		t.Fatalf("check 16 is %v with an invented reason, want a problem: %s", check.Status, domaindoctor.Line(check))
	}
	if !strings.Contains(check.Detail, "the gate does not have") {
		t.Errorf("the problem does not say where it looked: %s", check.Detail)
	}
}

// TestCheckSixteenSaysNoTableIsNoDebt: a project that keeps no table has nothing
// to count, and the line says so instead of inventing a zero.
func TestCheckSixteenSaysNoTableIsNoDebt(t *testing.T) {
	check := checkOf(t, run(configured(), cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi"), domaindoctor.CheckRules)

	if check.Status != domaindoctor.StatusOK {
		t.Fatalf("check 16 is %v without a table, want ok: %s", check.Status, domaindoctor.Line(check))
	}
	if !strings.Contains(check.Detail, "no debt to count") {
		t.Errorf("the line does not say why the count is absent: %s", check.Detail)
	}
}

// TestCheckSixteenCountsTheRulesNothingEnforces: debt is a number and a warning,
// and it names the rule with the skill that owns it, so the person can open the
// argument behind it.
func TestCheckSixteenCountsTheRulesNothingEnforces(t *testing.T) {
	files := configured()
	files.put("odd/rules.toml", `[[rule]]
rule     = "the voice is the profile's"
skill    = "drive"
enforces = "none"
because  = "no command reads a register"
`)

	check := checkOf(t, run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi"), domaindoctor.CheckRules)

	if check.Status != domaindoctor.StatusWarning {
		t.Fatalf("check 16 is %v with one debt row, want a warning: %s", check.Status, domaindoctor.Line(check))
	}
	for _, want := range []string{"1 rule without an enforcer", "the voice is the profile's", "drive"} {
		if !strings.Contains(check.Detail, want) {
			t.Errorf("check 16 does not say %q: %s", want, check.Detail)
		}
	}
}

// TestCheckSixteenResolvesEachMechanism: a row is a promise, and the three
// promises are kept by a fence path, a declared check and a test the suite runs.
// The table that names all three is clean.
func TestCheckSixteenResolvesEachMechanism(t *testing.T) {
	files := configured()
	files.put("odd/rules.toml", rulesTable)
	files.put("odd/fence.toml", fenceWithProof)
	files.put("odd/review.toml", reviewWithLint)
	files.put("internal/domain/x/x_test.go", "package x\n\nfunc TestSomething(t *testing.T) {}\n")

	check := checkOf(t, run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi"), domaindoctor.CheckRules)

	if check.Status != domaindoctor.StatusOK {
		t.Fatalf("check 16 is %v with three resolved rows, want ok: %s", check.Status, domaindoctor.Line(check))
	}
	if !strings.Contains(check.Detail, "3 rules, every one enforced") {
		t.Errorf("the line does not count the enforced rows: %s", check.Detail)
	}
}

// TestCheckSixteenRefusesATestNobodyRuns: a name the suite does not answer is the
// worst row in the table, because the project believes it is protected. It is a
// problem, not debt, and it names the row.
func TestCheckSixteenRefusesATestNobodyRuns(t *testing.T) {
	files := configured()
	files.put("odd/rules.toml", `[[rule]]
rule     = "the red comes first"
skill    = "tests"
enforces = "test:TestNobodyWrote"
`)

	check := checkOf(t, run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi"), domaindoctor.CheckRules)

	if check.Status != domaindoctor.StatusProblem {
		t.Fatalf("check 16 is %v with an unanswered test, want a problem: %s", check.Status, domaindoctor.Line(check))
	}
	for _, want := range []string{"test:TestNobodyWrote", "the suite does not have"} {
		if !strings.Contains(check.Detail, want) {
			t.Errorf("check 16 does not say %q: %s", want, check.Detail)
		}
	}
	if check.Fix == "" {
		t.Error("the problem names no repair")
	}
}

// TestCheckSixteenRefusesAFencePathNobodyKeeps: the fence is the file the row
// points at, so a path the fence does not carry is named with that file.
func TestCheckSixteenRefusesAFencePathNobodyKeeps(t *testing.T) {
	files := configured()
	files.put("odd/rules.toml", `[[rule]]
rule     = "the proofs are evidence"
skill    = "drive"
enforces = "fence:proofs/**"
`)
	files.put("odd/fence.toml", fenceWithProof)

	check := checkOf(t, run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi"), domaindoctor.CheckRules)

	if check.Status != domaindoctor.StatusProblem {
		t.Fatalf("check 16 is %v with a dangling fence path, want a problem: %s", check.Status, domaindoctor.Line(check))
	}
	if !strings.Contains(check.Detail, "odd/fence.toml") {
		t.Errorf("the problem does not name the fence: %s", check.Detail)
	}
}

// TestCheckSixteenRefusesACheckNobodyDeclares: a check is a row of review.toml,
// and a table that points at one the project does not declare is a promise the
// gate never runs.
func TestCheckSixteenRefusesACheckNobodyDeclares(t *testing.T) {
	files := configured()
	files.put("odd/rules.toml", `[[rule]]
rule     = "the gate runs the linter"
skill    = "commits"
enforces = "check:lint"
`)
	files.put("odd/review.toml", "mode = \"on\"\n")

	check := checkOf(t, run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi"), domaindoctor.CheckRules)

	if check.Status != domaindoctor.StatusProblem {
		t.Fatalf("check 16 is %v with a dangling check, want a problem: %s", check.Status, domaindoctor.Line(check))
	}
	if !strings.Contains(check.Detail, "odd/review.toml") {
		t.Errorf("the problem does not name the review table: %s", check.Detail)
	}
}

// TestCheckSixteenRefusesARowItCannotRead: a typo in the table is named by its
// row, and the repair is the file, not a quieter count.
func TestCheckSixteenRefusesARowItCannotRead(t *testing.T) {
	files := configured()
	files.put("odd/rules.toml", "[[rule]]\nrule = \"x\"\nskill = \"drive\"\nenforces = \"lint:yes\"\n")

	check := checkOf(t, run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi"), domaindoctor.CheckRules)

	if check.Status != domaindoctor.StatusProblem {
		t.Fatalf("check 16 is %v with an unreadable row, want a problem: %s", check.Status, domaindoctor.Line(check))
	}
	if !strings.Contains(check.Detail, "row 1") {
		t.Errorf("the problem does not name the row: %s", check.Detail)
	}
}
