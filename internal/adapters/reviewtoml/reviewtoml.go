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
)

// file mirrors the two things a human writes.
type file struct {
	Mode        string            `toml:"mode"`
	Lenses      map[string]string `toml:"lenses"`
	MaxPromptKB int               `toml:"max_prompt_kb"`
}

// Decode parses the file. Unknown keys are an error naming them; an empty or
// misspelled mode is refused by the domain, which owns the vocabulary.
func Decode(data []byte) (review.ProjectConfig, error) {
	var f file
	md, err := toml.Decode(string(data), &f)
	if err != nil {
		return review.ProjectConfig{}, fmt.Errorf("review: invalid TOML: %w", err)
	}

	if unknown := md.Undecoded(); len(unknown) > 0 {
		keys := make([]string, 0, len(unknown))
		for _, key := range unknown {
			keys = append(keys, key.String())
		}
		return review.ProjectConfig{}, fmt.Errorf("review: unknown keys: %s; this file configures mode, the model of each lens and the biggest prompt you pay for", strings.Join(keys, ", "))
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
	return review.ProjectConfig{Mode: mode, Lenses: lenses, MaxPromptKB: limit}, nil
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
	out.WriteString("# odachi: the review is the human's decision; on requires it in check, off leaves it optional\n")
	out.WriteString("mode = \"" + normalized + "\"\n")
	limit, err := review.NormalizeMaxPromptKB(config.MaxPromptKB)
	if err != nil {
		return nil, fmt.Errorf("review: %w", err)
	}
	if limit > 0 {
		out.WriteString("# the biggest prompt this project pays a reviewer for; a reviewer that stalls at a size does not answer late\n")
		out.WriteString("max_prompt_kb = " + strconv.Itoa(limit) + "\n")
	}
	if len(config.Lenses) > 0 {
		out.WriteString("\n[lenses]\n")
		for _, lens := range review.AllLenses() {
			if model, ok := config.Lenses[lens]; ok {
				out.WriteString(lens + " = \"" + model + "\"\n")
			}
		}
	}
	return []byte(out.String()), nil
}
