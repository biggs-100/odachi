// Package rulebooktoml reads and writes odd/rules.toml: one [[rule]] per rule the
// catalog states, each carrying the skill that owns it and the mechanism that
// makes breaking it impossible, or the reason nothing does.
//
// Decode and Render live together for the same reason the fence's do: a writer
// whose grammar drifts from the reader's is a table that stops counting in
// silence.
package rulebooktoml

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/biggs-100/odachi/internal/domain/rulebook"
	"github.com/biggs-100/odachi/internal/domain/schema"
)

// file mirrors the two things a human writes: the version and the rows.
type file struct {
	// Schema is the version that wrote this file. A file written before the key
	// existed reads as the first schema, and a newer one is refused as newer
	// instead of being called broken.
	Schema int    `toml:"schema"`
	Rules  []rule `toml:"rule"`
}

// rule is one row as the file writes it. The domain decides what a usable row is;
// this struct only names the keys.
type rule struct {
	Rule     string `toml:"rule"`
	Skill    string `toml:"skill"`
	Enforces string `toml:"enforces"`
	Because  string `toml:"because"`
}

// Decode parses the file and refuses what cannot be counted: unknown keys by
// name, a newer schema in the words that name the upgrade, and a row the domain
// cannot read by its number.
func Decode(data []byte) ([]rulebook.Rule, error) {
	var decoded file
	md, err := toml.Decode(string(data), &decoded)
	if err != nil {
		return nil, fmt.Errorf("rules: invalid TOML: %w", err)
	}

	// The version is asked before the unknown keys: a file from a newer odachi
	// carries both, and the key that makes it newer is not a typo.
	if _, err := schema.Read(decoded.Schema); err != nil {
		return nil, fmt.Errorf("rules: %w", err)
	}

	if unknown := md.Undecoded(); len(unknown) > 0 {
		keys := make([]string, 0, len(unknown))
		for _, key := range unknown {
			keys = append(keys, key.String())
		}
		return nil, fmt.Errorf("rules: unknown keys: %s; a row carries rule, skill, enforces and, when nothing enforces it, because", strings.Join(keys, ", "))
	}

	rules := make([]rulebook.Rule, 0, len(decoded.Rules))
	for _, one := range decoded.Rules {
		rules = append(rules, rulebook.Rule{
			Statement: one.Rule,
			Skill:     one.Skill,
			Enforces:  one.Enforces,
			Because:   one.Because,
		})
	}

	if problems := rulebook.Validate(rules); len(problems) > 0 {
		return nil, fmt.Errorf("rules: %s", problemsLine(problems))
	}
	return rules, nil
}

// Render writes the table in its canonical shape: one comment naming what the
// file is for, the schema, and one block per row with its keys aligned.
func Render(rules []rulebook.Rule) ([]byte, error) {
	if problems := rulebook.Validate(rules); len(problems) > 0 {
		return nil, fmt.Errorf("rules: %s", problemsLine(problems))
	}

	var out strings.Builder
	out.WriteString("# odachi: the schema that wrote this file; an older binary says upgrade instead of calling it broken\n")
	out.WriteString("schema = " + strconv.Itoa(schema.Current) + "\n")
	out.WriteString("# odachi: one row per rule the catalog states. enforces is fence:<glob>, check:<name> or test:<TestName>;\n")
	out.WriteString("# a rule nothing enforces says so with enforces = \"none\" and its because, and doctor counts it as debt\n")

	for _, one := range rules {
		out.WriteString("\n[[rule]]\n")
		fmt.Fprintf(&out, "%-8s = %s\n", "rule", strconv.Quote(one.Statement))
		fmt.Fprintf(&out, "%-8s = %s\n", "skill", strconv.Quote(one.Skill))
		fmt.Fprintf(&out, "%-8s = %s\n", "enforces", strconv.Quote(one.Enforces))
		if strings.TrimSpace(one.Because) != "" {
			fmt.Fprintf(&out, "%-8s = %s\n", "because", strconv.Quote(one.Because))
		}
	}
	return []byte(out.String()), nil
}

// problemsLine renders every problem with its row, so one run names all of them
// and a human fixes the file once.
func problemsLine(problems []rulebook.Problem) string {
	lines := make([]string, 0, len(problems))
	for _, problem := range problems {
		lines = append(lines, fmt.Sprintf("row %d: %s", problem.Row, problem.What))
	}
	return strings.Join(lines, "; ")
}
