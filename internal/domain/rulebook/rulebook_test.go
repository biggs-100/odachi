package rulebook

import (
	"strings"
	"testing"
)

// TestParseNamesTheMechanism: the value after the colon is what the app layer
// resolves, so a parse that swallows the colon would resolve "fence:proof/**" as
// a path named "fence:proof/**" and find nothing.
func TestParseNamesTheMechanism(t *testing.T) {
	cases := []struct {
		value string
		kind  Kind
		name  string
	}{
		{"fence:proof/**", KindFence, "proof/**"},
		{"check:lint", KindCheck, "lint"},
		{"test:TestProofIsNamedByTree", KindTest, "TestProofIsNamedByTree"},
		{"none", KindNone, ""},
		{"  fence:internal/assets/review/*.md  ", KindFence, "internal/assets/review/*.md"},
		{"", KindUnknown, ""},
		{"lint", KindUnknown, "lint"},
		{"fence:", KindFence, ""},
	}
	for _, c := range cases {
		kind, name := Parse(c.value)
		if kind != c.kind || name != c.name {
			t.Errorf("Parse(%q) = %v, %q; want %v, %q", c.value, kind, name, c.kind, c.name)
		}
	}
}

// TestDebtIsTheRowsNothingEnforces: debt is the number a session watches, so a
// row with a fence, a check or a test must never be counted as debt, and a rule
// with no enforcer must never disappear from the count.
func TestDebtIsTheRowsNothingEnforces(t *testing.T) {
	rules := []Rule{
		{Statement: "a", Skill: "drive", Enforces: "fence:proof/**"},
		{Statement: "b", Skill: "drive", Enforces: None, Because: "no command can hold it"},
		{Statement: "c", Skill: "tests", Enforces: "test:TestRedFirst"},
		{Statement: "d", Skill: "measure", Enforces: None, Because: "the decision is the human's"},
	}

	debt := Debt(rules)
	if len(debt) != 2 {
		t.Fatalf("Debt = %d rows, want 2: %+v", len(debt), debt)
	}
	if debt[0].Statement != "b" || debt[1].Statement != "d" {
		t.Errorf("Debt kept %q and %q; want b and d in file order", debt[0].Statement, debt[1].Statement)
	}
}

// TestValidateRefusesWhatCannotBeTrusted: the table is refused, not counted, when
// a row cannot be read. A row with no owner is a rule nobody maintains, and a row
// that names a mechanism this binary does not know is a typo that would otherwise
// read as enforcement.
func TestValidateRefusesWhatCannotBeTrusted(t *testing.T) {
	cases := []struct {
		name  string
		rule  Rule
		wants string
	}{
		{
			name:  "no statement",
			rule:  Rule{Skill: "drive", Enforces: "check:lint"},
			wants: "there is no rule",
		},
		{
			name:  "no skill",
			rule:  Rule{Statement: "x", Enforces: "check:lint"},
			wants: "there is no skill",
		},
		{
			name:  "no enforcer",
			rule:  Rule{Statement: "x", Skill: "drive"},
			wants: "there is no enforcer",
		},
		{
			name:  "a mechanism it does not know",
			rule:  Rule{Statement: "x", Skill: "drive", Enforces: "lint:yes"},
			wants: "is not an enforcer",
		},
		{
			name:  "a mechanism with no name",
			rule:  Rule{Statement: "x", Skill: "drive", Enforces: "check:"},
			wants: "names no check",
		},
		{
			name:  "a reason where the enforcer is the reason",
			rule:  Rule{Statement: "x", Skill: "drive", Enforces: "check:lint", Because: "it is nice"},
			wants: "a because on an enforced rule",
		},
		{
			name:  "a debt with no reason",
			rule:  Rule{Statement: "x", Skill: "drive", Enforces: None},
			wants: "none needs its because",
		},
	}
	for _, c := range cases {
		problems := Validate([]Rule{c.rule})
		if len(problems) != 1 {
			t.Errorf("%s: Validate = %d problems, want 1: %+v", c.name, len(problems), problems)
			continue
		}
		if !strings.Contains(problems[0].What, c.wants) {
			t.Errorf("%s: problem = %q, want it to name %q", c.name, problems[0].What, c.wants)
		}
		if problems[0].Row != 1 {
			t.Errorf("%s: row = %d, want 1", c.name, problems[0].Row)
		}
	}
}

// TestValidateRefusesOneRuleTwice: two rows for one rule count it twice, and the
// debt number would be a lie in the direction that matters least visibly.
func TestValidateRefusesOneRuleTwice(t *testing.T) {
	rules := []Rule{
		{Statement: "the red comes first", Skill: "tests", Enforces: "test:TestRed"},
		{Statement: "the red comes first", Skill: "drive", Enforces: "test:TestRed"},
	}
	problems := Validate(rules)
	if len(problems) != 1 {
		t.Fatalf("Validate = %d problems, want 1: %+v", len(problems), problems)
	}
	if problems[0].Row != 2 || !strings.Contains(problems[0].What, "already row 1") {
		t.Errorf("problem = %+v, want row 2 naming row 1", problems[0])
	}
}

// TestValidateAcceptsTheThreeMechanisms: the table exists to be green, so every
// legal shape validates with no problem.
func TestValidateAcceptsTheThreeMechanisms(t *testing.T) {
	rules := []Rule{
		{Statement: "a", Skill: "drive", Enforces: "fence:proof/**"},
		{Statement: "b", Skill: "drive", Enforces: "check:lint"},
		{Statement: "c", Skill: "drive", Enforces: "test:TestSomething"},
		{Statement: "d", Skill: "drive", Enforces: None, Because: "a judgment call cannot be held by a command"},
	}
	if problems := Validate(rules); len(problems) != 0 {
		t.Fatalf("Validate = %+v, want no problems", problems)
	}
}

// TestDebtLineStaysOneLine: doctor prints one line per check, so the debt list is
// capped and a sentence that arrived wrapped is folded instead of breaking the
// report.
func TestDebtLineStaysOneLine(t *testing.T) {
	debt := []Rule{
		{Statement: "the red comes\nfirst", Skill: "tests"},
		{Statement: "the voice is a rule", Skill: "writing"},
		{Statement: "the third", Skill: "slices"},
	}
	line := DebtLine(debt, 2)
	if strings.Contains(line, "\n") {
		t.Errorf("DebtLine kept a newline: %q", line)
	}
	if !strings.Contains(line, "tests") || !strings.Contains(line, "the red comes first") {
		t.Errorf("DebtLine = %q, want the folded statement and its skill", line)
	}
	if !strings.Contains(line, "and 1 more") {
		t.Errorf("DebtLine = %q, want it to name the rows the cap hid", line)
	}
	if got := DebtLine(nil, 2); got != "" {
		t.Errorf("DebtLine(nil) = %q, want empty", got)
	}
}
