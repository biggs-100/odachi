package review

import (
	"fmt"
	"path/filepath"
	"strings"

	domain "github.com/biggs-100/odachi/internal/domain/review"
)

// The three actions `odachi review mode` understands. Status is the default,
// because asking the question without an action should answer it, not change it.
const (
	ModeStatus = "status"
	ModeOn     = "on"
	ModeOff    = "off"
)

// ModeOptions is one `odachi review mode` invocation.
type ModeOptions struct {
	Dir    string
	Action string
}

// ModeFiles is the disk port for the switch: read it, back it up before
// replacing it, and write it. It is narrower than Files on purpose; the mode
// command never touches the verdict store.
type ModeFiles interface {
	Exists(path string) (bool, error)
	Read(path string) ([]byte, error)
	Write(path string, data []byte) error
	Backup(path string) (string, error)
}

// ModeCodec is the format port: the app knows the actions, the adapter knows
// TOML. A malformed file must be an error, never a silent "off".
type ModeCodec interface {
	Decode(data []byte) (string, error)
	Render(mode string) ([]byte, error)
}

// Mode answers or changes the project's review switch. A missing file is off:
// the review is opt-in until the human writes the switch.
func Mode(opts ModeOptions, files ModeFiles, codec ModeCodec) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}

	path := filepath.Join(opts.Dir, filepath.FromSlash(domain.ModePath))
	current, exists, err := readMode(path, files, codec)
	if err != nil {
		return modeBlocked(err)
	}

	action := strings.ToLower(strings.TrimSpace(opts.Action))
	if action == "" {
		action = ModeStatus
	}

	switch action {
	case ModeStatus:
		return modeResult(fmt.Sprintf("review: %s", current), exists, nil)
	case ModeOn, ModeOff:
		if exists && current == action {
			return modeResult(fmt.Sprintf("review: %s (sin cambios)", action), true, nil)
		}

		data, err := codec.Render(action)
		if err != nil {
			return modeBlocked(err)
		}
		if exists {
			if _, err := files.Backup(path); err != nil {
				return modeBlocked(fmt.Errorf("no pude respaldar %s antes de cambiarlo: %w", domain.ModePath, err))
			}
		}
		if err := files.Write(path, data); err != nil {
			return modeBlocked(fmt.Errorf("no pude escribir %s: %w", domain.ModePath, err))
		}
		return modeResult(fmt.Sprintf("review: %s", action), true, []string{domain.ModePath})
	default:
		return modeBlocked(fmt.Errorf("modo de review desconocido: %q; usá status, on u off", opts.Action))
	}
}

// Enabled reports whether the project's switch requires review. A missing file
// is off; a malformed one blocks instead of quietly turning the gate off.
func Enabled(dir string, files ModeFiles, codec ModeCodec) (bool, error) {
	if dir == "" {
		dir = "."
	}

	path := filepath.Join(dir, filepath.FromSlash(domain.ModePath))
	mode, _, err := readMode(path, files, codec)
	if err != nil {
		return false, err
	}
	return mode == domain.ModeOn, nil
}

// readMode reads and validates the switch. The boolean says whether the file
// exists, so status can name the default instead of pretending the human wrote
// it.
func readMode(path string, files ModeFiles, codec ModeCodec) (string, bool, error) {
	exists, err := files.Exists(path)
	if err != nil {
		return "", false, fmt.Errorf("no pude revisar %s: %w", domain.ModePath, err)
	}
	if !exists {
		return domain.ModeOff, false, nil
	}

	data, err := files.Read(path)
	if err != nil {
		return "", true, fmt.Errorf("no pude leer %s: %w", domain.ModePath, err)
	}

	mode, err := codec.Decode(data)
	if err != nil {
		return "", true, fmt.Errorf("%s: %w; corregí el archivo o volvé a escribir el modo con odachi review mode on|off", domain.ModePath, err)
	}
	return mode, true, nil
}

// modeResult is the one-line answer the command prints. A switch change also
// names the file it wrote; a status writes nothing.
func modeResult(message string, exists bool, files []string) Result {
	lines := []string{message}
	if !exists {
		lines[0] = message + " (sin " + domain.ModePath + ": el default es off)"
	}
	return Result{Code: CodeOK, Message: lines[0], Lines: lines, Files: files}
}

// modeBlocked is a switch that cannot be read or written. It is an actionable
// error, not a gate result: the human has to fix the file or the permissions.
func modeBlocked(err error) Result {
	return Result{Code: CodeBlocked, Message: err.Error()}
}
