package review

import (
	"fmt"
	"path/filepath"

	domain "github.com/biggs-100/odachi/internal/domain/review"
)

// ConfigOptions is one write of the review file from the wizard: the switch and
// the model of each lens. A field left zero keeps what the file already has.
type ConfigOptions struct {
	Dir    string
	Mode   string            // empty: keep the switch
	Models map[string]string // nil: keep the lens table; non-nil: replace it
}

// SetConfig writes the switch and the lens table, preserving the fields the
// wizard does not touch: the prompt limit and the two retention lifetimes. The
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
