package archtest_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"

	"github.com/biggs-100/odachi/internal/domain/fence"
)

// thisRepositoryFence reads odd/fence.toml the way the gate does: the file, the
// rules, and nothing else.
func thisRepositoryFence(t *testing.T) []fence.Rule {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("..", "..", "odd", "fence.toml"))
	if err != nil {
		t.Fatalf("read odd/fence.toml: %v", err)
	}

	var doc struct {
		Rules []struct {
			Not   string   `toml:"not"`
			Paths []string `toml:"paths"`
			Deps  []string `toml:"deps"`
		} `toml:"rule"`
	}
	if _, err := toml.Decode(string(data), &doc); err != nil {
		t.Fatalf("decode odd/fence.toml: %v", err)
	}

	rules := make([]fence.Rule, 0, len(doc.Rules))
	for _, rule := range doc.Rules {
		rules = append(rules, fence.Rule{Not: rule.Not, Paths: rule.Paths, Deps: rule.Deps})
	}
	if len(rules) == 0 {
		t.Fatal("odd/fence.toml has no rule: this repository protects nothing")
	}
	return rules
}

// TestTheFenceGuardsTheEvidenceAndNotTheWriter pins the boundary 0150 opened: this
// repository's own fence protects the stored proofs and lets the package that writes
// them change. The review of 0150 found that nothing said so, and a policy nobody
// asserts drifts back one commit at a time -invisibly, because a narrower fence still
// reports green.
func TestTheFenceGuardsTheEvidenceAndNotTheWriter(t *testing.T) {
	rules := thisRepositoryFence(t)

	// The behaviour, both halves: a stored proof is protected, the writer is not.
	if got := fence.Check(rules, []string{"proof/0001-x.txt"}, nil); len(got) == 0 {
		t.Errorf("the fence no longer protects a stored proof; rules: %+v", rules)
	}
	if got := fence.Check(rules, []string{"internal/domain/proof/proof.go"}, nil); len(got) != 0 {
		t.Errorf("the fence protects the package that writes the proofs again: %+v", got)
	}

	// And the shape: no rule of this fence is over the package that writes proofs.
	if problem := boundaryProblem(rules); problem != "" {
		t.Errorf("%s", problem)
	}
}

// TestTheBoundaryIsTheProofPackageAndNotEveryNameOfCode: the shape rule names one
// package. A fence over another internal package is ordinary scoping, and a test
// wide enough to refuse it would blame the proof boundary for a rule that has
// nothing to do with evidence.
func TestTheBoundaryIsTheProofPackageAndNotEveryNameOfCode(t *testing.T) {
	cases := []struct {
		name  string
		paths []string
		want  string
	}{
		{name: "the package that writes proofs", paths: []string{"internal/domain/proof/**"}, want: "the proof package"},
		{name: "the package itself", paths: []string{"internal/domain/proof"}, want: "the proof package"},
		{name: "another internal package", paths: []string{"internal/domain/gate/**"}, want: ""},
		{name: "a package whose name starts like it", paths: []string{"internal/domain/proofs/**"}, want: ""},
		{name: "the stored proofs", paths: []string{"proof/**"}, want: ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			problem := boundaryProblem([]fence.Rule{{Not: "a rule", Paths: c.paths}})
			if c.want == "" {
				if problem != "" {
					t.Fatalf("a rule over %v was refused: %s", c.paths, problem)
				}
				return
			}
			if !strings.Contains(problem, c.want) {
				t.Fatalf("a rule over %v was not refused with %q: %q", c.paths, c.want, problem)
			}
		})
	}
}

// boundaryProblem names the rule that reaches into the package that writes the
// proofs, or returns "" when none does. It is a function and not a loop inside the
// test so the test can ask about fences this repository does not have.
func boundaryProblem(rules []fence.Rule) string {
	for _, rule := range rules {
		for _, path := range rule.Paths {
			if overTheProofWriter(path) {
				return "a fence rule is over the proof package again: " + path + " in " + rule.Not
			}
		}
	}
	return ""
}

// overTheProofWriter reports whether the path is inside the package that writes the
// proofs. The prefix alone is not enough: internal/domain/proofs is another package
// whose name only starts with the same letters, and refusing it would be the same
// too-wide rule this test was written to remove.
func overTheProofWriter(path string) bool {
	const writer = "internal/domain/proof"
	return path == writer || strings.HasPrefix(path, writer+"/")
}
