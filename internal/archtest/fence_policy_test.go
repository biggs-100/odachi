package archtest_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"

	"github.com/biggs-100/odachi/internal/domain/fence"
)

// TestTheFenceGuardsTheEvidenceAndNotTheWriter pins the boundary 0150 opened: this
// repository's own fence protects the stored proofs and lets the package that writes
// them change. The review of 0150 found that nothing said so, and a policy nobody
// asserts drifts back one commit at a time -invisibly, because a narrower fence still
// reports green.
func TestTheFenceGuardsTheEvidenceAndNotTheWriter(t *testing.T) {
	root := filepath.Join("..", "..")
	data, err := os.ReadFile(filepath.Join(root, "odd", "fence.toml"))
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

	// The behaviour, both halves: a stored proof is protected, the writer is not.
	if got := fence.Check(rules, []string{"proof/0001-x.txt"}, nil); len(got) == 0 {
		t.Errorf("the fence no longer protects a stored proof; rules: %+v", rules)
	}
	if got := fence.Check(rules, []string{"internal/domain/proof/proof.go"}, nil); len(got) != 0 {
		t.Errorf("the fence protects the package that writes the proofs again: %+v", got)
	}

	// And the shape: no rule reaches into the code that writes the evidence.
	for _, rule := range rules {
		for _, path := range rule.Paths {
			if strings.HasPrefix(path, "internal/") {
				t.Errorf("a fence rule names code again: %q in %q", path, rule.Not)
			}
		}
	}
}
