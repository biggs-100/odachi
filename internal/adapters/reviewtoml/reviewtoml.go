// Package reviewtoml decodes and writes odd/review.toml: the project's review
// switch and, optionally, the model that judges each lens.
//
// The adapter is strict about both: a typo must be named instead of silently
// leaving the review off, or the session's model judging a lens the human paid
// another model for.
package reviewtoml

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/biggs-100/odachi/internal/domain/review"
	"github.com/biggs-100/odachi/internal/domain/schema"
)

// file mirrors the two things a human writes.
type file struct {
	// Schema is the version that wrote this file. It is not a human choice: a
	// file written before the key existed carries none and reads as the first
	// schema, and a value above this binary's is refused as newer, not broken.
	Schema      int               `toml:"schema"`
	Mode        string            `toml:"mode"`
	Lenses      map[string]string `toml:"lenses"`
	MaxPromptKB int               `toml:"max_prompt_kb"`
	// The two lifetimes `odachi clean` obeys. Absent -or a zero- means the
	// defaults odachi shipped, so a file written before they existed cleans the
	// same way it always did.
	KeepVerdicts int `toml:"keep_verdicts"`
	KeepPrompts  int `toml:"keep_prompts"`
	// Checks are the deterministic tools the review runs before any lens, and
	// CheckTimeout how long one may take. A file with neither declares no tools:
	// the review runs nothing on its behalf, exactly as it did before this key
	// existed.
	CheckTimeout int     `toml:"check_timeout_s"`
	Checks       []check `toml:"check"`
}

// check is one declared tool as the file writes it. The domain decides what a
// usable check is; this struct only names the keys.
type check struct {
	Name     string `toml:"name"`
	Command  string `toml:"command"`
	Severity string `toml:"severity"`
}

// Decode parses the file. Unknown keys are an error naming them; an empty or
// misspelled mode is refused by the domain, which owns the vocabulary.
func Decode(data []byte) (review.ProjectConfig, error) {
	var f file
	md, err := toml.Decode(string(data), &f)
	if err != nil {
		return review.ProjectConfig{}, fmt.Errorf("review: invalid TOML: %w", err)
	}

	// The version is asked before the unknown keys are named: a file from a newer
	// odachi carries both, and the key that makes it newer is not a typo for the
	// human to fix.
	if _, err := schema.Read(f.Schema); err != nil {
		return review.ProjectConfig{}, fmt.Errorf("review: %w", err)
	}

	if unknown := md.Undecoded(); len(unknown) > 0 {
		keys := make([]string, 0, len(unknown))
		for _, key := range unknown {
			keys = append(keys, key.String())
		}
		return review.ProjectConfig{}, fmt.Errorf("review: unknown keys: %s; this file configures mode, the model of each lens, the biggest prompt you pay for, how many records and prompts you keep, and the deterministic tools the review runs", strings.Join(keys, ", "))
	}

	mode, err := review.NormalizeMode(f.Mode)
	if err != nil {
		return review.ProjectConfig{}, fmt.Errorf("review: %w", err)
	}
	lenses, err := review.NormalizeLensModels(f.Lenses)
	if err != nil {
		return review.ProjectConfig{}, fmt.Errorf("review: %w", err)
	}
	limit, err := review.NormalizeMaxPromptKB(f.MaxPromptKB)
	if err != nil {
		return review.ProjectConfig{}, fmt.Errorf("review: %w", err)
	}
	verdicts, err := review.NormalizeKeepVerdicts(f.KeepVerdicts)
	if err != nil {
		return review.ProjectConfig{}, fmt.Errorf("review: %w", err)
	}
	prompts, err := review.NormalizeKeepPrompts(f.KeepPrompts)
	if err != nil {
		return review.ProjectConfig{}, fmt.Errorf("review: %w", err)
	}
	checks, err := review.NormalizeChecks(declaredChecks(f.Checks))
	if err != nil {
		return review.ProjectConfig{}, fmt.Errorf("review: %w", err)
	}
	timeout, err := review.NormalizeCheckTimeout(f.CheckTimeout)
	if err != nil {
		return review.ProjectConfig{}, fmt.Errorf("review: %w", err)
	}
	return review.ProjectConfig{
		Mode:         mode,
		Lenses:       lenses,
		MaxPromptKB:  limit,
		KeepVerdicts: verdicts,
		KeepPrompts:  prompts,
		Checks:       checks,
		CheckTimeout: timeout,
	}, nil
}

// declaredChecks copies the file's tables into the domain's vocabulary. The
// severity arrives as the text the human wrote and the domain decides whether it
// is one of the four.
func declaredChecks(raw []check) []review.Check {
	if len(raw) == 0 {
		return nil
	}
	out := make([]review.Check, 0, len(raw))
	for _, one := range raw {
		out = append(out, review.Check{
			Name:     one.Name,
			Command:  one.Command,
			Severity: review.Severity(one.Severity),
		})
	}
	return out
}

// Render writes the canonical file: the comment, the switch, and the lens table
// when there is one. The comment is part of the artifact: the human who opens it
// must see that the switch is theirs, not the agent's.
func Render(config review.ProjectConfig) ([]byte, error) {
	normalized, err := review.NormalizeMode(config.Mode)
	if err != nil {
		return nil, fmt.Errorf("review: %w", err)
	}

	var out strings.Builder
	out.WriteString("# odachi: the schema that wrote this file; an older binary says upgrade instead of calling it broken\n")
	out.WriteString("schema = " + strconv.Itoa(schema.Current) + "\n")
	out.WriteString("# odachi: the review is the human's decision; on asks for it at the close and check says so, off leaves it silent, and --require-review refuses\n")
	out.WriteString("mode = \"" + normalized + "\"\n")
	limit, err := review.NormalizeMaxPromptKB(config.MaxPromptKB)
	if err != nil {
		return nil, fmt.Errorf("review: %w", err)
	}
	if limit > 0 {
		out.WriteString("# the biggest prompt this project pays a reviewer for; a reviewer that stalls at a size does not answer late\n")
		out.WriteString("max_prompt_kb = " + strconv.Itoa(limit) + "\n")
	}
	verdicts, err := review.NormalizeKeepVerdicts(config.KeepVerdicts)
	if err != nil {
		return nil, fmt.Errorf("review: %w", err)
	}
	prompts, err := review.NormalizeKeepPrompts(config.KeepPrompts)
	if err != nil {
		return nil, fmt.Errorf("review: %w", err)
	}
	out.WriteString("# odachi clean --apply keeps this many review records -the newest trees- and this many prompts\n")
	out.WriteString("keep_verdicts = " + strconv.Itoa(verdicts) + "\n")
	out.WriteString("keep_prompts = " + strconv.Itoa(prompts) + "\n")
	checks, err := review.NormalizeChecks(config.Checks)
	if err != nil {
		return nil, fmt.Errorf("review: %w", err)
	}
	if len(checks) > 0 {
		timeout, err := review.NormalizeCheckTimeout(config.CheckTimeout)
		if err != nil {
			return nil, fmt.Errorf("review: %w", err)
		}
		out.WriteString("# how long one declared tool may run before the review stops waiting for it\n")
		out.WriteString("check_timeout_s = " + strconv.Itoa(timeout) + "\n")
	}
	if len(config.Lenses) > 0 {
		out.WriteString("# odachi: the model that judges each lens; when a provider retires one, the run names the model and the lens it served, and this line is the fix\n")
		out.WriteString("\n[lenses]\n")
		for _, lens := range review.AllLenses() {
			if model, ok := config.Lenses[lens]; ok {
				out.WriteString(lens + " = \"" + model + "\"\n")
			}
		}
	}
	if len(checks) > 0 {
		out.WriteString("# odachi: the deterministic tools a review runs on the candidate before any lens; a failure is filed at this severity, and the project that declares none runs none\n")
		for _, check := range checks {
			out.WriteString("\n[[check]]\n")
			out.WriteString("name = \"" + tomlString(check.Name) + "\"\n")
			out.WriteString("command = \"" + tomlString(check.Command) + "\"\n")
			out.WriteString("severity = \"" + string(check.Severity) + "\"\n")
		}
	}
	return []byte(out.String()), nil
}

// tomlString escapes the two characters a TOML basic string cannot carry as
// themselves. A command is a human's shell line and travels as written, so the
// writer escapes instead of refusing what the reader would accept.
func tomlString(value string) string {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(value)
}
