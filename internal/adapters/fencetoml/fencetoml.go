// Package fencetoml reads and writes odd/fence.toml, the project's fence: one
// [[rule]] per non-goal, each carrying the sentence it is named after and the
// paths and dependencies it keeps out of the tree.
//
// Decode and Render live together on purpose. A run writes the file the gate
// will read, so a writer whose grammar drifts from the reader's is a fence that
// stops protecting anything in silence (0266).
package fencetoml

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/biggs-100/odachi/internal/domain/fence"
)

// document is the TOML shape of the file: one [[rule]] per non-goal, each with
// a name and the paths and dependencies it forbids.
type document struct {
	Rules []rule `toml:"rule"`
}

type rule struct {
	Not   string   `toml:"not"`
	Paths []string `toml:"paths"`
	Deps  []string `toml:"deps"`
}

// Decode turns the file into the project's rules. A file with no rules means
// nothing was excluded; a malformed one is an error, because silently running
// with an empty fence would mean the gate stopped protecting anything without
// saying so.
func Decode(data []byte) ([]fence.Rule, error) {
	var decoded document
	if _, err := toml.Decode(string(data), &decoded); err != nil {
		return nil, err
	}

	rules := make([]fence.Rule, 0, len(decoded.Rules))
	for _, r := range decoded.Rules {
		rules = append(rules, fence.Rule{Not: r.Not, Paths: r.Paths, Deps: r.Deps})
	}
	return rules, nil
}

// Render writes the fence in the canonical TOML shape: one [[rule]] per rule,
// one key per line, and, for every non-goal that arrived as a phrase, the
// commented block that turns it into one when the human knows the paths.
func Render(rules []fence.Rule, loose []string) string {
	if len(rules) == 0 && len(loose) == 0 {
		// An empty file teaches nothing. The fence is the mechanism that keeps
		// scope small, so when the human chose no exclusion the file explains
		// what it is for and what a rule looks like.
		return "# Empty on purpose: you chose no \"NOT included\".\n" +
			"# The fence is what stops scope from growing: without rules, nothing\n" +
			"# warns you when the work goes past what you asked for.\n" +
			"#\n" +
			"# When you decide something that does not belong in this first version, add it here:\n" +
			"#\n" +
			"# [[rule]]\n" +
			"# not   = \"pagos\"\n" +
			"# paths = [\"payment*\", \"**/payments/**\"]\n" +
			"# deps  = [\"stripe\", \"mercadopago\"]\n"
	}

	var out strings.Builder

	for i, rule := range rules {
		if i > 0 {
			out.WriteString("\n")
		}
		out.WriteString("[[rule]]\n")
		fmt.Fprintf(&out, "%-5s = %s\n", "not", strconv.Quote(rule.Not))
		fmt.Fprintf(&out, "%-5s = %s\n", "paths", quoteList(rule.Paths))
		fmt.Fprintf(&out, "%-5s = %s\n", "deps", quoteList(rule.Deps))
	}

	for _, text := range loose {
		if out.Len() > 0 {
			out.WriteString("\n")
		}
		// A comment, not a rule: this fence holds only what somebody gave paths
		// for, and the absence of a path is a question, not a policy.
		out.WriteString("# This non-goal has no paths yet:\n")
		out.WriteString("#   " + text + "\n")
		out.WriteString("# A rule protects files: write what it must keep out of the tree, then\n")
		out.WriteString("# remove the '#' in front of the four lines below.\n")
		out.WriteString("#\n")
		out.WriteString("# [[rule]]\n")
		fmt.Fprintf(&out, "# %-5s = %s\n", "not", strconv.Quote(text))
		out.WriteString("# paths = []\n")
		out.WriteString("# deps  = []\n")
	}

	return out.String()
}

// quoteList renders a TOML array of strings.
func quoteList(items []string) string {
	quoted := make([]string, 0, len(items))
	for _, item := range items {
		quoted = append(quoted, strconv.Quote(item))
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}
