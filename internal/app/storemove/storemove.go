// Package storemove implements `odachi store move`: the one migration this tool
// has.
//
// It splits the old folder by owner, which is what the layout did when `.odachi/`
// stopped being a thing: the evidence that never travelled — verdicts, versions
// and the client choice — is adopted into the clone's store, byte for byte; the
// rules the team wrote — the profile, the fence, the review table, the roles —
// are tracked files, and a tracked file is moved by git and committed by the
// human. This verb prints those commands and never runs them.
//
// Running it twice changes nothing the second time: a version already in the
// store is compared, and the same bytes are reported instead of copied again.
package storemove

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/app/lock"
	"github.com/biggs-100/odachi/internal/app/undo"
	"github.com/biggs-100/odachi/internal/domain/layout"
	"github.com/biggs-100/odachi/internal/domain/store"
)

// Exit codes, the same shape as the other verbs.
const (
	// CodeOK means the migration ran, or there was nothing to migrate.
	CodeOK = 0
	// CodeBlocked means a file could not be moved, or a target holds other bytes
	// and the run refuses to overwrite it.
	CodeBlocked = 1
)

// verbName is the lock owner.
const verbName = "store move"

// Options is one `odachi store move` invocation.
type Options struct {
	Dir string

	// Store is the absolute root of the clone's state, resolved by the caller.
	Store string

	// DryRun reports what would move and touches nothing, not even the lock.
	DryRun bool
}

// Files is the disk port. It is undo's port because half of the migration is
// undo's: the versions.
type Files = undo.Files

// Result is what the CLI prints and exits with.
type Result struct {
	Code    int
	Message string
	Lines   []string
	Files   []string
}

// Run performs one `odachi store move`.
func Run(opts Options, files Files, guard lock.Lock) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}

	// A dry run reads; it takes no lock, because a lock is a write announcing
	// itself to the next writer.
	var held lock.Held
	if !opts.DryRun {
		held = lock.Take(guard, opts.Store, verbName)
		if held.Refusal != "" {
			return blocked(held.Refusal)
		}
		defer held.Release()
	}

	return withNotice(run(opts, files), held.Notice)
}

// run does the work once the lock is held.
func run(o Options, files Files) Result {
	moved := []string{}
	lines := []string{}

	// 1. The verdicts: one file per tree, in a flat directory.
	verdicts, err := o.adopt(files, store.Reviews, filepath.Join(layout.LegacyDir, store.Reviews), &lines)
	if err != nil {
		return blocked(err.Error())
	}
	moved = append(moved, verdicts...)

	// 2. The client choice: a tracked file that becomes per-clone. It is copied,
	// never deleted: the copy in git is the human's to remove.
	choice, err := o.adoptChoice(files, &lines)
	if err != nil {
		return blocked(err.Error())
	}
	if choice != "" {
		moved = append(moved, choice)
	}

	// 3. The versions: undo owns the rule that never overwrites a different
	// backup, so the migration calls it instead of repeating it.
	if o.has(files, filepath.Join(layout.LegacyDir, store.Backups)) {
		if o.DryRun {
			lines = append(lines, "the versions in "+layout.LegacyDir+"/"+store.Backups+" would be adopted by the same rule as odachi undo --collect")
		} else {
			collected := undo.Collect(o.Dir, o.Store, files)
			if collected.Code != undo.CodeOK {
				return blocked(collected.Message)
			}
			if len(collected.Files) > 0 {
				moved = append(moved, collected.Files...)
			}
			lines = append(lines, collected.Message)
		}
	}

	// 4. The rules git has to move.
	moves := o.moves(files)
	if len(moves) > 0 {
		lines = append(lines, "what git has to move (I never touch a tracked file):")
		lines = append(lines, moves...)
	}

	switch {
	case len(lines) == 0:
		return Result{
			Code:    CodeOK,
			Message: "there is nothing to move: this project has no " + layout.LegacyDir + "/",
			Lines:   []string{"there is nothing to move: this project has no " + layout.LegacyDir + "/"},
		}
	}

	verb := "moved"
	if o.DryRun {
		verb = "would move"
	}
	if len(moved) > 0 {
		lines = append([]string{fmt.Sprintf("%s %d file(s) into the clone's store", verb, len(moved))}, lines...)
	}
	return Result{Code: CodeOK, Message: strings.Join(lines, "; "), Lines: lines, Files: moved}
}

// adopt moves a flat directory of the old folder into the store, byte for byte.
// A target that already exists is compared: the same bytes end the old copy, and
// different bytes are named instead of overwritten.
func (o Options) adopt(files Files, entry, oldRel string, lines *[]string) ([]string, error) {
	names, err := files.List(filepath.Join(o.Dir, filepath.FromSlash(oldRel)))
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %s; check the permissions and run odachi store move again", oldRel, oneLine(err.Error()))
	}

	moved := make([]string, 0, len(names))
	for _, name := range names {
		if strings.HasPrefix(name, ".") {
			// The old store carried its own .gitignore; it has no meaning now.
			continue
		}

		source := filepath.Join(o.Dir, filepath.FromSlash(path.Join(oldRel, name)))
		target := filepath.Join(o.Store, filepath.FromSlash(path.Join(entry, name)))

		exists, err := files.Exists(target)
		if err != nil {
			return nil, fmt.Errorf("could not check %s: %s", storeLabel(o.Dir, target), oneLine(err.Error()))
		}

		if exists {
			same, err := sameBytes(files, source, target)
			if err != nil {
				return nil, err
			}
			if !same {
				*lines = append(*lines, "did not move "+filepath.ToSlash(path.Join(oldRel, name))+": "+storeLabel(o.Dir, target)+" already holds other bytes")
				continue
			}
			if !o.DryRun {
				if err := files.Remove(source); err != nil {
					return nil, fmt.Errorf("could not remove the old copy %s: %s", path.Join(oldRel, name), oneLine(err.Error()))
				}
			}
			*lines = append(*lines, filepath.ToSlash(path.Join(oldRel, name))+": already in the store, the old copy is gone")
			continue
		}

		if !o.DryRun {
			if err := files.Rename(source, target); err != nil {
				return nil, fmt.Errorf("could not move %s to %s: %s; check the permissions and run odachi store move again",
					path.Join(oldRel, name), storeLabel(o.Dir, target), oneLine(err.Error()))
			}
		}
		*lines = append(*lines, filepath.ToSlash(path.Join(oldRel, name))+" -> "+storeLabel(o.Dir, target))
		moved = append(moved, storeLabel(o.Dir, target))
	}
	return moved, nil
}

// adoptChoice brings the old client selection into the store. A choice already
// recorded is never replaced: which client receives the agent is the person's
// decision, and the migration is not the place to make it.
func (o Options) adoptChoice(files Files, lines *[]string) (string, error) {
	source := filepath.Join(o.Dir, filepath.FromSlash(layout.LegacyDir), store.Agents)
	exists, err := files.Exists(source)
	if err != nil {
		return "", fmt.Errorf("could not check %s: %s", source, oneLine(err.Error()))
	}
	if !exists {
		return "", nil
	}

	target := filepath.Join(o.Store, store.Agents)
	recorded, err := files.Exists(target)
	if err != nil {
		return "", fmt.Errorf("could not check %s: %s", storeLabel(o.Dir, target), oneLine(err.Error()))
	}
	if recorded {
		same, err := sameBytes(files, source, target)
		if err != nil {
			return "", err
		}
		if !same {
			*lines = append(*lines, "the store already records another client choice ("+storeLabel(o.Dir, target)+"); the old one is not copied")
			return "", nil
		}
		*lines = append(*lines, "the client choice is already in the store; remove "+layout.LegacyDir+"/"+store.Agents+" from git when you can")
		return "", nil
	}

	if !o.DryRun {
		data, err := files.Read(source)
		if err != nil {
			return "", fmt.Errorf("could not read %s: %s", source, oneLine(err.Error()))
		}
		if err := files.Write(target, data); err != nil {
			return "", fmt.Errorf("could not write %s: %s", storeLabel(o.Dir, target), oneLine(err.Error()))
		}
	}
	*lines = append(*lines, layout.LegacyDir+"/"+store.Agents+" -> "+storeLabel(o.Dir, target))
	return storeLabel(o.Dir, target), nil
}

// moves is the tracked half: what git has to do, in the order a human reads it.
func (o Options) moves(files Files) []string {
	present := []string{}
	for _, to := range layout.LegacyRules {
		rel := filepath.ToSlash(filepath.Join(layout.LegacyDir, filepath.Base(to)))
		if o.has(files, rel) {
			present = append(present, rel)
		}
	}
	moves := layout.LegacyMoves(present)
	if o.has(files, filepath.Join(layout.LegacyDir, store.Agents)) {
		moves = append(moves, "git rm --cached "+layout.LegacyDir+"/"+store.Agents)
	}
	return moves
}

// has reports whether a path of the project exists.
func (o Options) has(files Files, rel string) bool {
	found, err := files.Exists(filepath.Join(o.Dir, filepath.FromSlash(rel)))
	return err == nil && found
}

// sameBytes compares two files.
func sameBytes(files Files, left, right string) (bool, error) {
	a, err := files.Read(left)
	if err != nil {
		return false, fmt.Errorf("could not read %s: %s", left, oneLine(err.Error()))
	}
	b, err := files.Read(right)
	if err != nil {
		return false, fmt.Errorf("could not read %s: %s", right, oneLine(err.Error()))
	}
	return bytes.Equal(a, b), nil
}

// storeLabel names a file of the store the way a human reads it.
func storeLabel(dir, path string) string {
	if out, err := filepath.Rel(dir, path); err == nil && out != ".." && !strings.HasPrefix(out, ".."+string(filepath.Separator)) {
		return filepath.ToSlash(out)
	}
	return filepath.ToSlash(path)
}

// withNotice puts the lock's announcement in front of the run's own answer.
func withNotice(result Result, notice string) Result {
	if notice == "" {
		return result
	}
	result.Lines = append([]string{notice}, result.Lines...)
	if result.Message == "" {
		return Result{Code: result.Code, Message: notice, Lines: result.Lines, Files: result.Files}
	}
	result.Message = notice + " " + result.Message
	return result
}

// blocked is a Result with exit code 1.
func blocked(message string) Result {
	return Result{Code: CodeBlocked, Message: oneLine(message), Lines: []string{oneLine(message)}}
}

// oneLine keeps every message to a single line, whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
