package main

import (
	"strings"
	"testing"
)

// TestSkillsEvalPrintsThePromptAndRunsNobody: without --exec the verb asks the
// question and spends nothing, the same contract the probe keeps.
func TestSkillsEvalPrintsThePromptAndRunsNobody(t *testing.T) {
	t.Parallel()

	code, stdout, stderr := runDir(t, t.TempDir(), []string{"skills", "eval", "tests"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", code, stderr)
	}
	if !strings.Contains(stdout, "config.Parse") {
		t.Errorf("the prompt is not printed:\n%s", stdout)
	}
	if !strings.Contains(stdout, "--exec") {
		t.Errorf("the prompt does not say what would run it:\n%s", stdout)
	}
}

// TestSkillsEvalTakesTheSkillBeforeTheFlags: the shape the page and the verb's own
// next line print -the skill first, the flags after it- has to run, because the
// Go flag parser stops at the first word that is not a flag.
func TestSkillsEvalTakesTheSkillBeforeTheFlags(t *testing.T) {
	t.Parallel()

	code, stdout, stderr := runDir(t, t.TempDir(), []string{"skills", "eval", "tests", "--runs", "2"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", code, stderr)
	}
	if !strings.Contains(stdout, "config.Parse") {
		t.Errorf("the documented shape printed no prompt:\n%s", stdout)
	}
}

// TestSkillsEvalNamesTheExamsThatExist: a skill with no exam is answered with the
// list, because a typo and a missing measurement are different answers.
func TestSkillsEvalNamesTheExamsThatExist(t *testing.T) {
	t.Parallel()

	code, _, stderr := runDir(t, t.TempDir(), []string{"skills", "eval", "absent"}, "")
	if code == 0 {
		t.Fatal("an exam that does not exist ran")
	}
	if !strings.Contains(stderr, "tests") {
		t.Errorf("the refusal does not name the exams that ship: %s", stderr)
	}
}
