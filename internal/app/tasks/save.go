package tasks

// The write guard for task documents lives here, and only here. `ask` and
// `plan` both grow a document, so both must answer the same question the same
// way before touching the disk: would the canonical render keep every line the
// human wrote? When the answer is no, the write does not happen and the message
// says where the notes belong.

import (
	"fmt"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// Writer is the disk port a rewrite needs.
type Writer interface {
	Write(path string, data []byte) error
}

// BackupWriter can also copy the document aside before replacing it.
type BackupWriter interface {
	Writer
	Backup(path string) (string, error)
}

// Save rewrites the task document with the grown version, through the canonical
// renderer. It refuses when that render would drop a line the document has.
//
// Every error is one English line naming the file, because the caller only has
// to print it.
func Save(files Writer, current Document, grown taskdoc.TaskDoc) error {
	rendered, err := guardedRender(current, grown)
	if err != nil {
		return err
	}
	return write(files, current, rendered)
}

// SaveWithBackup is Save plus a backup of the file before it is replaced. The
// guard runs first, so a refused write leaves no backup behind: nothing was
// about to be replaced.
func SaveWithBackup(files BackupWriter, current Document, grown taskdoc.TaskDoc) error {
	rendered, err := guardedRender(current, grown)
	if err != nil {
		return err
	}
	if _, err := files.Backup(current.Path); err != nil {
		return fmt.Errorf("could not back up %s before rewriting it: %s; check the permissions and try again",
			current.Rel, oneLine(err.Error()))
	}
	return write(files, current, rendered)
}

// guardedRender is the single place that decides whether a task document may be
// written: the canonical render, compared line by line with what is on disk.
func guardedRender(current Document, grown taskdoc.TaskDoc) ([]byte, error) {
	rendered := grown.Render()
	if dropped := firstDroppedLine(current.Raw, rendered); dropped != "" {
		return nil, fmt.Errorf(
			"recorded nothing in %s: rewriting it would lose the line %q, which is not in the canonical format; move those notes to ## History and try again",
			current.Rel, dropped)
	}
	return []byte(rendered), nil
}

// write replaces the document and translates an adapter failure into a line the
// human can act on.
func write(files Writer, current Document, rendered []byte) error {
	if err := files.Write(current.Path, rendered); err != nil {
		return fmt.Errorf("could not write %s: %s; check the permissions and try again",
			current.Rel, oneLine(err.Error()))
	}
	return nil
}

// firstDroppedLine returns the first line of before that the canonical render
// cannot carry over and that is absent from after, or "" when the rewrite
// loses nothing. Canonical lines may legitimately change — a step being marked
// done rewrites its own line — so only unaccounted content is a loss. Rendering
// is canonical, and anything outside the canonical sections is not carried over,
// so the caller must not throw it away silently.
func firstDroppedLine(before, after string) string {
	present := map[string]bool{}
	for _, line := range strings.Split(after, "\n") {
		present[strings.TrimSpace(line)] = true
	}
	for _, line := range taskdoc.Unaccounted(before) {
		if !present[line] {
			return line
		}
	}
	return ""
}

// oneLine keeps every message to a single line, whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
