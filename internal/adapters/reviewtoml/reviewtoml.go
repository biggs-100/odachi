// Package reviewtoml decodes .odachi/review.toml into the project's review mode.
//
// The file has exactly one key, `mode`, and the adapter is strict about it: a
// typo must be named instead of silently leaving the review off when the human
// asked for it on.
package reviewtoml

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/biggs-100/odachi/internal/domain/review"
)

// file mirrors the one field a human writes.
type file struct {
	Mode string `toml:"mode"`
}

// Decode parses the switch file. Unknown keys are an error naming them; an
// empty or misspelled mode is refused by the domain, which owns the vocabulary.
func Decode(data []byte) (string, error) {
	var f file
	md, err := toml.Decode(string(data), &f)
	if err != nil {
		return "", fmt.Errorf("review: TOML inválido: %w", err)
	}

	if unknown := md.Undecoded(); len(unknown) > 0 {
		keys := make([]string, 0, len(unknown))
		for _, key := range unknown {
			keys = append(keys, key.String())
		}
		return "", fmt.Errorf("review: claves desconocidas: %s; el archivo solo configura mode", strings.Join(keys, ", "))
	}

	mode, err := review.NormalizeMode(f.Mode)
	if err != nil {
		return "", fmt.Errorf("review: %w", err)
	}
	return mode, nil
}

// Render writes the canonical switch file. The comment is part of the artifact:
// the human who opens it must see that the switch is theirs, not the agent's.
func Render(mode string) ([]byte, error) {
	normalized, err := review.NormalizeMode(mode)
	if err != nil {
		return nil, fmt.Errorf("review: %w", err)
	}
	return []byte("# odachi: el review es una decisión del humano; on lo exige en check, off lo deja opcional\nmode = \"" + normalized + "\"\n"), nil
}
