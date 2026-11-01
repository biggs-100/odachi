package review

import (
	"fmt"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/layout"
)

// The project's review switch. It is a human decision, not an inference: a
// missing file means off, and the only way to turn the review on is to write the
// file with `odachi review mode on`. The gate reads the same file, so the switch
// is not a flag somebody has to remember on every commit.
const (
	// ModePath is where the switch lives, relative to the project root.
	ModePath = layout.Review
	// ModeOn makes `odachi check` require the verdict.
	ModeOn = "on"
	// ModeOff leaves the review optional, the same behavior as no file.
	ModeOff = "off"
)

// ProjectConfig is what odd/review.toml holds: the switch, optionally the model
// that judges each lens, the biggest prompt this project pays for, and how much
// evidence `odachi clean` keeps. A lens without an entry is judged by the
// session's model, which is what the file meant before this table existed.
type ProjectConfig struct {
	Mode   string
	Lenses map[string]string
	// MaxPromptKB is the largest prompt size, in kilobytes, this project will
	// spend a reviewer on. Zero means no limit: a file that never wrote the key
	// reviews as it always did.
	MaxPromptKB int
	// KeepVerdicts is how many review records survive `odachi clean --apply`, and
	// KeepPrompts how many prompts. They are the project's decision, so the file
	// carries the number in force: a project that never chose reads the defaults
	// odachi shipped, and `Render` writes them.
	KeepVerdicts int
	KeepPrompts  int
}

// The numbers odachi ships, and the reason they are not the human's: a project
// that never opened the file must still be cleanable, and the day it opens it the
// defaults are written there so nothing changes behind its back.
const (
	// DefaultKeepVerdicts keeps the reviews of the last ten trees: the gate reads
	// the newest one, and the rest are the week's evidence.
	DefaultKeepVerdicts = 10
	// DefaultKeepPrompts keeps none. A prompt is the question, not the evidence:
	// it is the biggest file in the store -59 to 114 KB measured on this
	// repository- and nothing in the gate reads it.
	DefaultKeepPrompts = 0
	// MaxKeep is the ceiling both lifetimes share. A number nobody reaches is a
	// typo wearing a decision's clothes, and a store that never empties is what
	// the lifetimes exist to prevent.
	MaxKeep = 1000
)

// NormalizeKeepVerdicts reads the review records' lifetime. Zero is the absence
// of a choice and means the default; a negative number or one above MaxKeep is a
// typo. There is no "keep none": the gate reads the newest record, so a store
// with no verdicts is a project that has to pay for its last review twice.
func NormalizeKeepVerdicts(value int) (int, error) {
	return normalizeKeep("keep_verdicts", value, DefaultKeepVerdicts)
}

// NormalizeKeepPrompts reads the prompts' lifetime. Zero is a real choice here -
// it is also the default- because a prompt nobody reads has no reason to survive.
func NormalizeKeepPrompts(value int) (int, error) {
	return normalizeKeep("keep_prompts", value, DefaultKeepPrompts)
}

// normalizeKeep is the one rule both lifetimes obey, so the two keys cannot
// disagree about what a typo is.
func normalizeKeep(key string, value, fallback int) (int, error) {
	switch {
	case value < 0:
		return 0, fmt.Errorf("%s is %d: a lifetime is zero or a positive number", key, value)
	case value == 0:
		return fallback, nil
	case value > MaxKeep:
		return 0, fmt.Errorf("%s is %d: the most a project keeps is %d", key, value, MaxKeep)
	}
	return value, nil
}

// NormalizeMaxPromptKB checks the size limit. It is a size a human measured for
// their reviewer, so a negative number is a typo and not a rule; zero is the
// absence of one.
func NormalizeMaxPromptKB(value int) (int, error) {
	if value < 0 {
		return 0, fmt.Errorf("the size limit is %d KB: a limit is zero (none) or a positive number of kilobytes", value)
	}
	return value, nil
}

// AllLenses is the lens vocabulary in its fixed order, for a writer that has to
// render the table without inventing an order of its own.
func AllLenses() []string {
	out := make([]string, len(order))
	copy(out, order)
	return out
}

// ModelIdentity is the model behind a model value: the provider and the name, with
// any #variant cut off. A variant is an effort or a setting of the same model, so
// two values that differ only there are the same judge -and calling them different
// is how a review gets accepted as a second opinion it never was.
func ModelIdentity(model string) string {
	trimmed := strings.ToLower(strings.TrimSpace(model))
	if cut, _, found := strings.Cut(trimmed, "#"); found {
		return strings.TrimSpace(cut)
	}
	return trimmed
}

// NormalizeModel checks a model identity. The value replaces {model} inside a
// reviewer command, so it is an identifier and not shell code: a space, a quote or
// a semicolon would change the command, and a repository's review.toml must not be
// able to run anything. The accepted alphabet is what model ids use:
// provider/name#variant, with dots, colons, plus and dashes.
func NormalizeModel(value string) (string, error) {
	model := strings.TrimSpace(value)
	for _, r := range model {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case strings.ContainsRune("._/#@:+-", r):
		default:
			return "", fmt.Errorf("model %q contains %q: a model is an identifier (provider/name#variant), and it replaces {model} inside a command", value, r)
		}
	}
	return model, nil
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
		if _, twice := out[name]; twice {
			// A map is iterated in no fixed order, so two spellings of one lens
			// would make the winner random: refuse instead of choosing.
			return nil, fmt.Errorf("the lens %q appears twice: write it once", name)
		}
		model = strings.TrimSpace(model)
		if model == "" {
			return nil, fmt.Errorf("the lens %q names no model: write one or remove the line", lens)
		}
		checked, err := NormalizeModel(model)
		if err != nil {
			return nil, fmt.Errorf("the lens %q: %w", lens, err)
		}
		out[name] = checked
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
