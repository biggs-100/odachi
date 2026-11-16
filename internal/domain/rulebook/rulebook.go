// Package rulebook owns odd/rules.toml: the table that pairs a rule the catalog
// states with the mechanism that makes breaking it impossible, or says in words
// that nothing does.
//
// It exists because a rule written in prose is a rule the next session may skip,
// and a rule that repeats after it was already written is not a missing sentence
// but a missing check. The table is the place where that is counted: a row whose
// enforcer is `none` is debt a human can read, and a row that names a check or a
// test that does not exist is worse than no row at all, because the project
// believes it is protected.
//
// It is pure: no filesystem, no os. The adapter reads the file, the app layer
// resolves each enforcer against the tree, and this package only classifies what
// the human wrote.
package rulebook

import (
	"fmt"
	"strings"
)

const (
	// None is the enforcer a rule carries when nothing makes breaking it
	// impossible. It is not an empty value: the file says so out loud, and the
	// row carries the reason, because a debt nobody can explain is a silence with
	// a key.
	None = "none"

	// The three mechanisms a rule can name. Each one is resolved by the app layer
	// against something that already exists on disk: a fence path the gate
	// refuses, a check the review runs, a test the suite runs.
	PrefixFence = "fence:"
	PrefixCheck = "check:"
	PrefixTest  = "test:"
)

// Rule is one row of the table: the rule as a person reads it, the skill that
// states it, and the mechanism that enforces it.
type Rule struct {
	// Statement is the rule, one line, close to the words the skill uses. A rule
	// that needs a paragraph is two rules.
	Statement string
	// Skill is the folder under skills/ that states it: the owner a reader opens
	// to see the whole argument.
	Skill string
	// Enforces is one of fence:<glob>, check:<name>, test:<TestName>, or none.
	Enforces string
	// Because is the reason a rule has no enforcer. It is required when Enforces
	// is none and refused otherwise: a justification on an enforced rule is a
	// sentence nobody reads.
	Because string
}

// Kind is the mechanism an enforcer names.
type Kind int

const (
	// KindNone means nothing enforces the rule.
	KindNone Kind = iota
	// KindFence means a path or a dependency the gate refuses.
	KindFence
	// KindCheck means a deterministic tool the review runs.
	KindCheck
	// KindTest means a test the suite runs.
	KindTest
	// KindUnknown means the value names no mechanism this binary knows, which is
	// a typo or a newer grammar.
	KindUnknown
)

// String is the word a report uses for the kind.
func (k Kind) String() string {
	switch k {
	case KindFence:
		return "fence"
	case KindCheck:
		return "check"
	case KindTest:
		return "test"
	case KindNone:
		return None
	default:
		return "unknown"
	}
}

// Parse splits an enforcer into its mechanism and its name. A value that names no
// mechanism returns KindUnknown with the value as the name, so the message can
// quote what the human wrote.
func Parse(enforces string) (Kind, string) {
	value := strings.TrimSpace(enforces)
	for _, one := range []struct {
		prefix string
		kind   Kind
	}{
		{PrefixFence, KindFence},
		{PrefixCheck, KindCheck},
		{PrefixTest, KindTest},
	} {
		if strings.HasPrefix(value, one.prefix) {
			return one.kind, strings.TrimSpace(strings.TrimPrefix(value, one.prefix))
		}
	}
	if value == None {
		return KindNone, ""
	}
	return KindUnknown, value
}

// Debt returns the rules nothing enforces, in the order the file lists them. The
// count is the number a session watches: it goes down when a rule becomes a
// check, and up when a feature writes a rule nobody can hold.
func Debt(rules []Rule) []Rule {
	var out []Rule
	for _, rule := range rules {
		if kind, _ := Parse(rule.Enforces); kind == KindNone {
			out = append(out, rule)
		}
	}
	return out
}

// Problem is one defect in the table, named by the row a human counts.
type Problem struct {
	// Row is the 1-based position of the row in the file.
	Row int
	// What says what is wrong in one line, with the fix when it is short.
	What string
}

// Validate reports every row that cannot be trusted: a rule with no sentence, no
// owner, an enforcer this binary does not know, a reason missing where it is owed
// and a reason where it is not. It is the reader's gate: a table that does not
// validate is refused instead of being counted, because a count of rows odachi
// cannot read would be a number about nothing.
func Validate(rules []Rule) []Problem {
	var problems []Problem
	seen := make(map[string]int, len(rules))
	for i, rule := range rules {
		row := i + 1
		statement := strings.TrimSpace(rule.Statement)
		skill := strings.TrimSpace(rule.Skill)
		enforces := strings.TrimSpace(rule.Enforces)

		if statement == "" {
			problems = append(problems, Problem{row, "there is no rule: a row that names nothing pairs nothing; write the rule as one line"})
		}
		if skill == "" {
			problems = append(problems, Problem{row, "there is no skill: a rule with no owner is a sentence nobody maintains; name the folder under skills/"})
		}
		if enforces == "" {
			// One problem per defect: an empty value is not also a mechanism this
			// binary does not know.
			problems = append(problems, Problem{row, "there is no enforcer: name fence:<glob>, check:<name> or test:<TestName>, or write enforces = \"none\" and its because"})
		} else {
			problems = append(problems, enforcerProblems(row, rule, enforces)...)
		}

		if statement != "" {
			if first, ok := seen[statement]; ok {
				problems = append(problems, Problem{row, fmt.Sprintf("the same rule is already row %d: two rows for one rule count it twice", first)})
			} else {
				seen[statement] = row
			}
		}
	}
	return problems
}

// enforcerProblems judges the value a row carries once the row has one.
func enforcerProblems(row int, rule Rule, enforces string) []Problem {
	var problems []Problem
	kind, name := Parse(enforces)
	switch kind {
	case KindUnknown:
		problems = append(problems, Problem{row, fmt.Sprintf("%q is not an enforcer: an enforcer is fence:<glob>, check:<name>, test:<TestName>, or none", enforces)})
	case KindFence, KindCheck, KindTest:
		if name == "" {
			problems = append(problems, Problem{row, fmt.Sprintf("%q names no %s: write the name after the colon", enforces, kind)})
		}
		if strings.TrimSpace(rule.Because) != "" {
			problems = append(problems, Problem{row, "a because on an enforced rule is a sentence nobody reads: the enforcer is the reason"})
		}
	case KindNone:
		if strings.TrimSpace(rule.Because) == "" {
			problems = append(problems, Problem{row, "none needs its because: a debt nobody can explain is a silence with a key"})
		}
	}
	return problems
}

// DebtLine renders the list a person reads on doctor's line: each rule's skill
// and its sentence, capped so one line stays one line.
func DebtLine(debt []Rule, max int) string {
	if len(debt) == 0 {
		return ""
	}
	const reason = "nothing enforces it"
	parts := make([]string, 0, len(debt))
	for i, rule := range debt {
		if max > 0 && i == max {
			parts = append(parts, fmt.Sprintf("and %d more", len(debt)-max))
			break
		}
		parts = append(parts, fmt.Sprintf("%s (%s, %s)", oneLine(rule.Statement), strings.TrimSpace(rule.Skill), reason))
	}
	return strings.Join(parts, "; ")
}

// oneLine folds a sentence onto one line: a newline in a doctor line breaks the
// report, and the file's rule is one line anyway.
func oneLine(text string) string {
	fields := strings.Fields(text)
	return strings.Join(fields, " ")
}
