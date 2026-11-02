// Package undo implements `odachi undo`: putting back the last artifact odachi
// wrote, without ever losing the state it replaces.
//
// The rule that makes it safe is the order. The current file is backed up
// first, and only then is the chosen backup copied over it. So an undo is
// itself undoable, and a hand edit made after the agent's last write survives as
// a backup instead of disappearing.
//
// Backups live in the clone's store; `discover` also reads the old project
// store and the versions left beside their artifacts, and `--collect` moves
// both into the store without rewriting them.
//
// undo never touches git. The code is the human's and git's; a tool that starts
// choosing commits to revert becomes exactly what this one criticises, so the
// message names `git revert <sha>` and runs nothing.
package undo

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/biggs-100/odachi/internal/app/lock"
	undodomain "github.com/biggs-100/odachi/internal/domain/undo"
)

// Exit codes, the same shape as the other verbs.
const (
	// CodeOK means something was restored, or listed.
	CodeOK = 0
	// CodeBlocked means there is nothing to undo, or the flag names an artifact
	// odachi does not write.
	CodeBlocked = 1
)

// dateFormat is how a backup's date reads in the list: to the minute, because
// two writes in the same hour are the normal case.
const dateFormat = "2006-01-02 15:04"

// codeHint is the one thing undo does not do. Naming it costs a line; executing
// git would cost the tool its reason to exist.
const codeHint = "I did not touch git: undoing code is yours, with git revert <sha>"

// Options is one `odachi undo` invocation.
type Options struct {
	Dir string

	// Store is the absolute root of the clone's state, resolved by the caller
	// from the repository's git common directory. The lock lives there. Empty
	// means there is no repository to name one, which the caller refuses before
	// this function runs.
	Store string

	// List shows what can be undone instead of restoring anything.
	List bool
	// Collect moves the backups that still live beside their artifacts into the
	// project store. It never deletes a version that is not already in the
	// store.
	Collect bool
	// File keeps one artifact: brief | profile | fence | task.
	File string
	// Artifact names one document: the path `--list` prints, or its bare file
	// name. It is what makes a project with several task documents answerable.
	Artifact string
}

// Files is the disk port: enough to find the backups, learn when they were
// taken, save the current state aside and write the restored bytes.
type Files interface {
	Exists(path string) (bool, error)
	List(dir string) ([]string, error)
	Read(path string) ([]byte, error)
	Write(path string, data []byte) error
	Remove(path string) error
	Rename(from, to string) error
	Backup(path string) (string, error)
	// ModTime is how the domain chooses: by what the filesystem recorded, not
	// by a counter odachi would have to keep in sync.
	ModTime(path string) (time.Time, error)
}

// Result is what the CLI prints and exits with. It has the same shape as
// start.Result, so the CLI reports every verb the same way.
type Result struct {
	Code    int
	Message string
	Lines   []string
	Files   []string
}

// verbName is the lock owner: what the next run reads when this one is still
// writing.
const verbName = "undo"

// Run performs one `odachi undo`.
func Run(opts Options, files Files, guard lock.Lock) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}

	// undo replaces one artifact and backs up the one it replaces, so it writes
	// more than one file and takes the lock like start does. It comes before the
	// first listing: a refusal must not read a single backup.
	//
	// A project that is not a git repository has no store and no place for a
	// lock. start keeps working there and saves its versions beside the
	// artifacts, so undo has to keep working there too: refusing would strand a
	// project the tool itself created.
	held := lock.Held{}
	if opts.Store != "" {
		held = lock.Take(guard, opts.Store, verbName)
	}
	if held.Refusal != "" {
		return blocked(held.Refusal)
	}
	defer held.Release()

	return withNotice(run(opts, files), held.Notice)
}

// Collect moves every version that still sits beside an artifact, or in the old
// project store, into the clone's store. The caller holds the lock: `undo
// --collect` takes it, and `odachi store move` already has it for the rest of the
// migration. It never overwrites a version with different bytes and never loses
// one it cannot move.
func Collect(project, store string, files Files) Result {
	return collect(project, store, files)
}

// run does the work once the lock is held.
func run(opts Options, files Files) Result {
	if opts.Collect {
		if opts.List || opts.File != "" || strings.TrimSpace(opts.Artifact) != "" {
			return blocked("--collect moves every old backup at once: it does not go with --list, --file or --artifact")
		}
		if opts.Store == "" {
			return blocked("--collect moves versions into the clone's store, and this folder is not a git repository; run it inside one")
		}
		return Collect(opts.Dir, opts.Store, files)
	}

	kind, problem := parseFile(opts.File)
	if problem != "" {
		return blocked(problem)
	}

	artifact := normalizeArtifact(opts.Artifact)
	// --file and --artifact are two ways of naming one document: asking for the
	// brief and naming a task is a contradiction, not a filter that finds nothing.
	if artifact != "" && kind != "" {
		if named, ok := undodomain.KindOf(artifact); ok && named != kind {
			return blocked(fmt.Sprintf("--artifact %s is a %s, and this run asked for %s: drop --file or name a %s",
				artifact, named, kind, kind))
		}
	}

	kinds := undodomain.Kinds()
	if opts.File != "" {
		// A filtered run never lists the directories of the other artifacts, so
		// the answer cannot depend on what they hold.
		kinds = []undodomain.Kind{kind}
	}

	found, err := discover(opts.Dir, opts.Store, kinds, files)
	if err != nil {
		return blocked(err.Error())
	}
	if len(found) == 0 {
		return blocked(nothingToUndo(kind))
	}
	if artifact != "" {
		named := ofArtifact(found, artifact)
		if len(named) == 0 {
			return blocked(noVersionOf(artifact, found))
		}
		found = named
	}
	if opts.List {
		return listing(found)
	}
	return restore(opts.Dir, found, files)
}

// normalizeArtifact reads --artifact the way a human writes it: a path relative
// to the project, with either separator, or the bare file name.
func normalizeArtifact(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return path.Clean(filepath.ToSlash(value))
}

// ofArtifact keeps the backups of the one document the run names.
func ofArtifact(found []undodomain.Backup, artifact string) []undodomain.Backup {
	kept := make([]undodomain.Backup, 0, 1)
	for _, candidate := range found {
		if candidate.Artifact == artifact || path.Base(candidate.Artifact) == artifact {
			kept = append(kept, candidate)
		}
	}
	return kept
}

// noVersionOf names the documents that do have a backup, so naming the wrong one
// is one editable line and not a mystery.
func noVersionOf(artifact string, found []undodomain.Backup) string {
	groups := undodomain.Groups(found)
	named := make([]string, 0, len(groups))
	for _, group := range groups {
		named = append(named, group.Artifact)
	}
	return fmt.Sprintf("there is no backup of %s; the documents I can put back are %s",
		artifact, joined(named))
}

// parseFile reads --file. An empty value means "any artifact"; a value outside
// the allow-list is refused before anything is read.
func parseFile(value string) (undodomain.Kind, string) {
	if strings.TrimSpace(value) == "" {
		return "", ""
	}
	kind, ok := undodomain.ParseKind(value)
	if ok {
		return kind, ""
	}
	return "", fmt.Sprintf(
		"--file %s is not an artifact odachi writes; the ones I can undo are %s",
		strings.TrimSpace(value), allowedFlags())
}

// allowedFlags lists the four artifacts with their paths, taken from the domain
// so the allow-list has one source of truth.
func allowedFlags() string {
	names := make([]string, 0, len(undodomain.Kinds()))
	for _, kind := range undodomain.Kinds() {
		names = append(names, string(kind)+" ("+kind.Target()+")")
	}
	return joined(names)
}

// allowedTargets lists only the paths: the sentence a human reads when there is
// nothing to undo and has to know what odachi writes.
func allowedTargets() string {
	targets := make([]string, 0, len(undodomain.Kinds()))
	for _, kind := range undodomain.Kinds() {
		targets = append(targets, kind.Target())
	}
	return joined(targets)
}

// discover finds the backups of the given artifacts. Everything else in those
// directories — a note, a temporary, a backup of a file odachi never writes — is
// left alone.
//
// Three places can hold a version, and all three are read so a project can move
// at its own pace: the clone's store, where new backups go; the old project
// store a previous release used; and beside the artifact, before any store
// existed. Nothing is rewritten by reading.
func discover(project, store string, kinds []undodomain.Kind, files Files) ([]undodomain.Backup, error) {
	type location struct {
		root string
		dir  string
	}

	found := make([]undodomain.Backup, 0, len(kinds))

	for _, kind := range kinds {
		locations := []location{
			{store, path.Join(undodomain.StoreBackups, kind.BackupSubdir())},
			{project, path.Join(undodomain.LegacyBackups, kind.LegacySubdir())},
			{project, kind.Dir()},
		}
		for _, at := range locations {
			if at.root == "" {
				// No store resolved: there is nowhere to look, and reading the
				// project twice would list the same file as two versions.
				continue
			}
			backups, err := scan(at.root, at.dir, kind, files)
			if err != nil {
				return nil, err
			}
			found = append(found, backups...)
		}
	}
	return found, nil
}

// scan reads one directory and keeps only the files that are backups of this
// kind. Anything else is left alone.
func scan(root, dirRel string, kind undodomain.Kind, files Files) ([]undodomain.Backup, error) {
	names, err := files.List(filepath.Join(root, filepath.FromSlash(dirRel)))
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %s; check the permissions and try again",
			dirRel, oneLine(err.Error()))
	}

	found := make([]undodomain.Backup, 0, len(names))
	for _, name := range names {
		artifactName, ok := undodomain.SplitBackup(name)
		if !ok {
			continue
		}
		artifact, ok := kind.Artifact(artifactName)
		if !ok {
			continue
		}

		backupRel := path.Join(dirRel, name)
		when, err := files.ModTime(filepath.Join(root, filepath.FromSlash(backupRel)))
		if err != nil {
			return nil, fmt.Errorf("could not read the date of %s: %s; check the permissions and try again",
				backupRel, oneLine(err.Error()))
		}

		found = append(found, undodomain.Backup{
			Kind:     kind,
			Artifact: artifact,
			Root:     root,
			Path:     backupRel,
			Entry:    path.Join(kind.BackupSubdir(), name),
			ModTime:  when,
		})
	}
	return found, nil
}

// restore puts back the most recent backup, saving the current state first.
func restore(dir string, found []undodomain.Backup, files Files) Result {
	// Guessing is not a restore. The newest backup belongs to whichever artifact
	// the tool happened to write last — inside start that is fence.toml — so with
	// more than one candidate "the most recent" is an accident of internal write
	// order, never what the human meant. One candidate restores; several ask.
	if groups := undodomain.Groups(found); len(groups) > 1 {
		names := make([]string, 0, len(groups))
		for _, group := range groups {
			names = append(names, group.Artifact)
		}
		return blocked(fmt.Sprintf("there are %d artifacts with backups (%s): tell me which one you want with --file brief|profile|fence|task",
			len(groups), joined(names)))
	}

	chosen, ok := undodomain.Latest(found)
	if !ok {
		return blocked(nothingToUndo(""))
	}

	artifactPath := filepath.Join(dir, filepath.FromSlash(chosen.Artifact))
	backupPath := filepath.Join(chosen.Root, filepath.FromSlash(chosen.Path))
	label := relativeBackup(dir, backupPath)

	// The bytes are read before anything moves: the restore cannot depend on the
	// file it is about to replace.
	data, err := files.Read(backupPath)
	if err != nil {
		return blocked(fmt.Sprintf("could not read %s: %s; check the permissions and try again",
			relativeBackup(dir, backupPath), oneLine(err.Error())))
	}

	// The rule the whole safety story rests on: what is in the artifact now is
	// saved before it is replaced.
	exists, err := files.Exists(artifactPath)
	if err != nil {
		return blocked(fmt.Sprintf("could not check %s: %s; check the permissions and try again",
			chosen.Artifact, oneLine(err.Error())))
	}

	previous := ""
	if exists {
		current, err := files.Read(artifactPath)
		if err != nil {
			return blocked(fmt.Sprintf("could not read %s: %s; check the permissions and try again",
				chosen.Artifact, oneLine(err.Error())))
		}
		// Restoring what is already there is not a change, and saying it was
		// would print the same backup name as both the source and the result.
		if bytes.Equal(current, data) {
			message := fmt.Sprintf("no changes: %s was already the same as %s", chosen.Artifact, label)
			return Result{Code: CodeOK, Message: message, Lines: []string{message, codeHint}}
		}

		made, err := files.Backup(artifactPath)
		if err != nil {
			return blocked(fmt.Sprintf("could not back up %s before restoring it: %s; check the permissions and try again",
				chosen.Artifact, oneLine(err.Error())))
		}
		previous = relativeBackup(dir, made)
	}

	if err := files.Write(artifactPath, data); err != nil {
		return blocked(fmt.Sprintf("could not write %s: %s; check the permissions and try again",
			chosen.Artifact, oneLine(err.Error())))
	}

	message := fmt.Sprintf("restored %s from %s; %s did not exist, so there was nothing to back up",
		chosen.Artifact, label, chosen.Artifact)
	affected := []string{chosen.Artifact}
	if previous != "" {
		message = fmt.Sprintf("restored %s from %s; what was there was left in %s",
			chosen.Artifact, label, previous)
		affected = append(affected, previous)
	}

	return Result{
		Code:    CodeOK,
		Message: message,
		Lines:   []string{message, codeHint},
		Files:   affected,
	}
}

// collect moves the backups that still live beside their artifacts, or in the
// old project store, into the clone's store. It is a move, not a copy: the
// target keeps the bytes and the date, so the next `undo` still restores the
// same version. A target that already exists is compared first; a different file
// is never overwritten.
func collect(dir, store string, files Files) Result {
	found, err := discover(dir, store, undodomain.Kinds(), files)
	if err != nil {
		return blocked(err.Error())
	}

	label := storeLabel(dir, store)

	moved := 0
	targets := make([]string, 0, len(found))
	for _, backup := range found {
		if backup.Root == store {
			continue
		}

		source := filepath.Join(backup.Root, filepath.FromSlash(backup.Path))
		target := filepath.Join(store, filepath.FromSlash(path.Join(undodomain.StoreBackups, backup.Entry)))
		targetRel := relativeBackup(dir, target)

		data, err := files.Read(source)
		if err != nil {
			return blocked(fmt.Sprintf("could not read %s to move it: %s; check the permissions and try again",
				relativeBackup(dir, source), oneLine(err.Error())))
		}

		exists, err := files.Exists(target)
		if err != nil {
			return blocked(fmt.Sprintf("could not check %s: %s; check the permissions and try again",
				targetRel, oneLine(err.Error())))
		}

		if exists {
			current, err := files.Read(target)
			if err != nil {
				return blocked(fmt.Sprintf("could not read %s: %s; check the permissions and try again",
					targetRel, oneLine(err.Error())))
			}
			if !bytes.Equal(current, data) {
				return blocked(fmt.Sprintf("did not move %s: %s already exists with other contents; I do not overwrite a backup",
					relativeBackup(dir, source), targetRel))
			}
			if err := files.Remove(source); err != nil {
				return blocked(fmt.Sprintf("could not delete the old backup %s: %s", relativeBackup(dir, source), oneLine(err.Error())))
			}
		} else {
			if err := files.Rename(source, target); err != nil {
				return blocked(fmt.Sprintf("could not move %s to %s: %s; check the permissions and try again",
					relativeBackup(dir, source), targetRel, oneLine(err.Error())))
			}
		}

		moved++
		targets = append(targets, targetRel)
	}

	if moved == 0 {
		message := "there are no old backups to move: they are all in " + label
		return Result{Code: CodeOK, Message: message, Lines: []string{message}}
	}

	message := fmt.Sprintf("moved %d backups to %s", moved, label)
	return Result{Code: CodeOK, Message: message, Lines: []string{message}, Files: targets}
}

// storeLabel is how a message names the clone's store: relative to the project
// when it can be, so the usual line reads `.git/odachi`, and absolute when the
// store lives outside it (an external git directory).
func storeLabel(dir, store string) string {
	return relativeBackup(dir, store)
}

// listing is `--list`: one line per artifact with backups, newest first.
func listing(found []undodomain.Backup) Result {
	groups := undodomain.Groups(found)

	lines := make([]string, 0, len(groups))
	for _, group := range groups {
		lines = append(lines, fmt.Sprintf("%s · %s · last %s",
			group.Artifact, count(len(group.Backups)), group.Last().Format(dateFormat)))
	}

	// The message is the same content on one string, so a machine that only
	// reads `message` gets the whole answer.
	return Result{Code: CodeOK, Message: strings.Join(lines, "; "), Lines: lines}
}

// nothingToUndo says what odachi writes and which command writes it: with no
// backup there is nothing to restore, and the human needs to know why.
func nothingToUndo(kind undodomain.Kind) string {
	scope := "there is nothing to undo"
	if kind != "" {
		scope = "there is no backup of " + kind.Target()
	}
	return fmt.Sprintf("%s: odachi writes %s, and there are no backups yet; run odachi start to write the first one",
		scope, allowedTargets())
}

// withNotice puts the lock's announcement in front of the run's own answer: a
// run that took over an abandoned lock says so, on its own line when it
// succeeded and at the front of the message when it did not.
func withNotice(result Result, notice string) Result {
	if notice == "" {
		return result
	}

	if result.Code == CodeOK {
		result.Lines = append([]string{notice}, result.Lines...)
	}
	if result.Message == "" {
		result.Message = notice
		return result
	}
	result.Message = notice + " " + result.Message
	return result
}

// relativeBackup names a backup the way the message reads it: relative to the
// project, so a version in the clone's store says `.git/odachi/backups/...` and
// an old one keeps its place beside the artifact.
func relativeBackup(root, backupPath string) string {
	rel, err := filepath.Rel(root, backupPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return filepath.ToSlash(backupPath)
	}
	return filepath.ToSlash(rel)
}

// count is how many backups an artifact has, as the line reads it.
func count(n int) string {
	if n == 1 {
		return "1 backup"
	}
	return fmt.Sprintf("%d backups", n)
}

// joined lists names the way English does: "a, b, c and d".
func joined(names []string) string {
	if len(names) <= 1 {
		return strings.Join(names, "")
	}
	return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
}

// blocked is a Result with exit code 1.
func blocked(message string) Result {
	return Result{Code: CodeBlocked, Message: oneLine(message)}
}

// oneLine keeps every message to a single line, whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
