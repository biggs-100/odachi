package check

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// stagedFiles makes project reads describe the commit a message lane is judging.
// The clone store, Git's own files, clone-local proof evidence, the explicit
// proof file and the commit-message file stay on the real filesystem: they are
// evidence about the candidate, not bytes inside it.
type stagedFiles struct {
	real    Files
	git     Git
	root    string
	store   string
	message string
	proof   string
}

func newStagedFiles(opts Options, files Files, git Git) Files {
	return stagedFiles{
		real:    files,
		git:     git,
		root:    filepath.Clean(opts.Dir),
		store:   absolutePath(opts.Store),
		message: absolutePath(explicitPath(opts.Dir, opts.Message)),
		proof:   absolutePath(explicitPath(opts.Dir, opts.Proof)),
	}
}

func (f stagedFiles) Exists(path string) (bool, error) {
	rel, staged := f.projectPath(path)
	if !staged {
		return f.real.Exists(path)
	}
	_, found, err := f.git.StagedFile(f.root, rel)
	return found, err
}

func (f stagedFiles) Read(path string) ([]byte, error) {
	rel, staged := f.projectPath(path)
	if !staged {
		return f.real.Read(path)
	}
	data, found, err := f.git.StagedFile(f.root, rel)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	}
	return data, nil
}

func (f stagedFiles) List(dir string) ([]string, error) {
	rel, staged := f.projectPath(dir)
	if !staged {
		return f.real.List(dir)
	}
	return f.git.StagedList(f.root, rel)
}

func (f stagedFiles) Write(path string, data []byte) error {
	if _, staged := f.projectPath(path); staged {
		return fmt.Errorf("cannot write %s during a commit-message check: project files come from the staged index", path)
	}
	return f.real.Write(path, data)
}

func (f stagedFiles) ModTime(path string) (time.Time, error) {
	if _, staged := f.projectPath(path); staged {
		return time.Time{}, fmt.Errorf("cannot read the worktree timestamp of %s during a commit-message check", path)
	}
	return f.real.ModTime(path)
}

func (f stagedFiles) Remove(path string) error {
	if _, staged := f.projectPath(path); staged {
		return fmt.Errorf("cannot remove %s during a commit-message check: project files come from the staged index", path)
	}
	return f.real.Remove(path)
}

// projectPath names a path relative to the staged tree, or reports that it
// belongs to the real filesystem. The comparisons are lexical and volume-aware:
// a sibling such as /project-other is outside /project, not inside it.
func (f stagedFiles) projectPath(path string) (string, bool) {
	abs := absolutePath(path)
	if withinPath(f.store, abs) || samePath(f.message, abs) || samePath(f.proof, abs) {
		return "", false
	}
	rel, inside := relativePath(absolutePath(f.root), abs)
	if !inside {
		return "", false
	}
	slashed := filepath.ToSlash(rel)
	if slashed == proofDir || strings.HasPrefix(slashed, proofDir+"/") {
		return "", false
	}
	for part := range strings.SplitSeq(slashed, "/") {
		if strings.EqualFold(part, ".git") {
			return "", false
		}
	}
	return slashed, true
}

func absolutePath(path string) string {
	if path == "" {
		return ""
	}
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return filepath.Clean(path)
}

func samePath(left, right string) bool {
	if left == "" || right == "" {
		return false
	}
	rel, err := filepath.Rel(left, right)
	return err == nil && rel == "."
}

func relativePath(root, path string) (string, bool) {
	if root == "" {
		return "", false
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return rel, true
}

func withinPath(root, path string) bool {
	_, inside := relativePath(root, path)
	return inside
}
