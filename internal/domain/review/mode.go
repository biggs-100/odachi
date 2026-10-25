package review

import (
	"fmt"
	"strings"
)

// The project's review switch. It is a human decision, not an inference: a
// missing file means off, and the only way to turn the review on is to write the
// file with `odachi review mode on`. The gate reads the same file, so the switch
// is not a flag somebody has to remember on every commit.
const (
	// ModePath is where the switch lives, relative to the project root.
	ModePath = ".odachi/review.toml"
	// ModeOn makes `odachi check` require the verdict.
	ModeOn = "on"
	// ModeOff leaves the review optional, the same behavior as no file.
	ModeOff = "off"
)

// ProjectConfig is what .odachi/review.toml holds: the switch, and optionally
// the model that judges each lens. A lens without an entry is judged by the
// session's model, which is what the file meant before this table existed.
type ProjectConfig struct {
	Mode   string
	Lenses map[string]string
}

// AllLenses is the lens vocabulary in its fixed order, for a writer that has to
// render the table without inventing an order of its own.
func AllLenses() []string {
	out := make([]string, len(order))
	copy(out, order)
	return out
}

// NormalizeLensModels checks a lens table and lowercases its keys. An unknown
// lens is refused naming the four: a model assigned to a lens that does not exist
// would be a plan nobody follows. An empty model is a human who meant to remove
// the line, and a table with no entries is no table.
func NormalizeLensModels(raw map[string]string) (map[string]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	out := make(map[string]string, len(raw))
	for lens, model := range raw {
		name := strings.ToLower(strings.TrimSpace(lens))
		if !IsLens(name) {
			return nil, fmt.Errorf("%q is not a lens: the lenses are %s", lens, strings.Join(AllLenses(), ", "))
		}
		model = strings.TrimSpace(model)
		if model == "" {
			return nil, fmt.Errorf("the lens %q names no model: write one or remove the line", lens)
		}
		out[name] = model
	}
	return out, nil
}

// NormalizeMode accepts the two values the switch knows. Matching is case
// insensitive and ignores surrounding space because a human writes this file;
// anything else is refused, naming the two values, instead of being guessed.
func NormalizeMode(value string) (string, error) {
	switch normalized := strings.ToLower(strings.TrimSpace(value)); normalized {
	case ModeOn:
		return ModeOn, nil
	case ModeOff:
		return ModeOff, nil
	default:
		return "", fmt.Errorf("invalid review mode %q: use on or off", strings.TrimSpace(value))
	}
}
