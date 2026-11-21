package review

import (
	"fmt"
	"path/filepath"
	"strings"

	domain "github.com/biggs-100/odachi/internal/domain/review"
)

// ConfigOptions is one write of the review file from the wizard: the switch, the
// model of each lens, the model the verifier runs with and the four scalars the
// limits section edits. A field left at its zero value keeps what the file has.
type ConfigOptions struct {
	Dir    string
	Mode   string            // empty: keep the switch
	Models map[string]string // nil: keep the lens table; non-nil: replace it
	// Verify is the model `odachi verify` substitutes for {model}. Nil keeps the
	// key; a value replaces it, and an empty value removes the line, because a
	// project that names no verifier pays for the session's model (0337).
	Verify *string
	// The four scalars are pointers because three of them have a meaningful zero:
	// no prompt limit, keep no prompts, and the shipped check timeout. Nil means
	// the caller never named the field, so the file's value survives (0342).
	MaxPromptKB  *int
	KeepVerdicts *int
	KeepPrompts  *int
	CheckTimeout *int
}

// SetConfig writes the switch, the lens table and any of the four scalars the
// caller names, preserving the fields the wizard does not touch: the schema, the
// mode, the lenses, the verifier and the whole [[check]] array. The array is not
// an option here on purpose: it carries the project's gate severity, and a
// picker that could lower it would change policy without a decision (0342). The
// file is read once, changed in memory and rendered by the same codec the mode
// command uses, so the two verbs cannot disagree about the format.
func SetConfig(opts ConfigOptions, files ModeFiles, codec ModeCodec) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}

	path := filepath.Join(opts.Dir, filepath.FromSlash(domain.ModePath))
	current, exists, err := readConfig(path, files, codec)
	if err != nil {
		return modeBlocked(err)
	}

	if opts.Mode != "" {
		mode, err := domain.NormalizeMode(opts.Mode)
		if err != nil {
			return modeBlocked(err)
		}
		current.Mode = mode
	}
	if opts.Models != nil {
		models, err := domain.NormalizeLensModels(opts.Models)
		if err != nil {
			return modeBlocked(err)
		}
		current.Lenses = models
	}
	// The verifier is one key of the same file: it follows the lens table's rule
	// -the value replaces {model} inside a command- and an empty value is the
	// project writing that it names no verifier, which Render omits (0337).
	if opts.Verify != nil {
		value := strings.TrimSpace(*opts.Verify)
		if value != "" {
			checked, err := domain.NormalizeModel(value)
			if err != nil {
				return modeBlocked(err)
			}
			value = checked
		}
		current.Verify = value
	}
	// The four scalars follow the rule the verifier set: a caller that names one
	// has the domain check it before the codec writes it, and a caller that names
	// none keeps the file's value. Nothing here offers the checks array -that is
	// the project's gate severity, owned by the file (0342).
	if opts.MaxPromptKB != nil {
		limit, err := domain.NormalizeMaxPromptKB(*opts.MaxPromptKB)
		if err != nil {
			return modeBlocked(err)
		}
		current.MaxPromptKB = limit
	}
	if opts.KeepVerdicts != nil {
		verdicts, err := domain.NormalizeKeepVerdicts(*opts.KeepVerdicts)
		if err != nil {
			return modeBlocked(err)
		}
		current.KeepVerdicts = verdicts
	}
	if opts.KeepPrompts != nil {
		prompts, err := domain.NormalizeKeepPrompts(*opts.KeepPrompts)
		if err != nil {
			return modeBlocked(err)
		}
		current.KeepPrompts = prompts
	}
	if opts.CheckTimeout != nil {
		timeout, err := domain.NormalizeCheckTimeout(*opts.CheckTimeout)
		if err != nil {
			return modeBlocked(err)
		}
		current.CheckTimeout = timeout
	}

	data, err := codec.Render(current)
	if err != nil {
		return modeBlocked(err)
	}
	if exists {
		if _, err := files.Backup(path); err != nil {
			return modeBlocked(fmt.Errorf("could not back up %s before changing it: %w", domain.ModePath, err))
		}
	}
	if err := files.Write(path, data); err != nil {
		return modeBlocked(fmt.Errorf("could not write %s: %w", domain.ModePath, err))
	}
	return modeResult(fmt.Sprintf("review: %s", current.Mode), true, []string{domain.ModePath})
}
