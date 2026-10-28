// Package undo decides which backup of which artifact `odachi undo` restores.
//
// odachi writes four kinds of artifact and keeps a backup before rewriting it.
// New versions live in the clone's store, outside the worktree; the old project
// store and the versions left beside the artifact are still read so a project can
// migrate at its own pace. Undoing is putting one of those backups back — never
// a commit, never a file of somebody else's: the code belongs to the human and to
// git, and a tool that starts choosing commits to revert becomes exactly what
// this one criticises.
//
// The domain answers "which artifact" and "which backup, out of which set". The
// app layer is the one that touches the disk. Stdlib only, like the rest of the
// domain.
package undo

import (
	"path"
	"sort"
	"strings"
	"time"
)

// backupMark separates an artifact from the digest of the bytes it held when
// the backup was taken. fsx.Backup writes exactly this shape.
const backupMark = ".bak-"

// StoreBackups is the directory inside the clone's store that holds the
// versions. A backup keeps its artifact's path inside it, so one directory holds
// every version without a second naming scheme.
const StoreBackups = "backups"

// LegacyBackups is the project-relative directory older versions used. It is
// read and adopted by --collect, never written again: the worktree keeps the
// rules and the plan, and the versions of them belong to the clone.
const LegacyBackups = ".odachi/backups"

// digestLen is how many hex characters fsx.Backup keeps of the content digest.
const digestLen = 8

// Kind is one of the four artifacts odachi writes. Nothing else can be undone.
type Kind string

const (
	// Brief is the human's own words: brief.md.
	Brief Kind = "brief"
	// Profile is the ceremony knob: .odachi/profile.toml.
	Profile Kind = "profile"
	// Fence is what the project refuses to do: .odachi/fence.toml.
	Fence Kind = "fence"
	// Task is one feature's document: odd/tasks/NNNN-<slug>.md.
	Task Kind = "task"
)

// Kinds lists the four artifacts in the order the messages name them. This is
// the allow-list: nothing outside it is read or touched.
func Kinds() []Kind { return []Kind{Brief, Profile, Fence, Task} }

// ParseKind reads the value of --file. The flag takes the short name, and the
// error that follows a miss lists these four.
func ParseKind(value string) (Kind, bool) {
	kind := Kind(strings.TrimSpace(value))
	for _, known := range Kinds() {
		if kind == known {
			return kind, true
		}
	}
	return "", false
}

// Dir is the directory that holds the artifact, relative to the project root.
// The project root itself is "." so joining never has to special-case it.
func (k Kind) Dir() string {
	switch k {
	case Profile, Fence:
		return ".odachi"
	case Task:
		return "odd/tasks"
	default:
		return "."
	}
}

// BackupSubdir is where this kind's versions live inside the store's backups
// directory. The artifact's own directory is preserved underneath, so
// `.odachi/profile.toml` keeps its path and a task keeps its place in
// `odd/tasks`.
func (k Kind) BackupSubdir() string {
	return k.Dir()
}

// Target is the artifact as a human reads it in a project: the list a blocked
// message repeats, so the human knows exactly what undo can put back.
func (k Kind) Target() string {
	switch k {
	case Brief:
		return "brief.md"
	case Profile:
		return ".odachi/profile.toml"
	case Fence:
		return ".odachi/fence.toml"
	default:
		return "odd/tasks/*.md"
	}
}

// Artifact reports whether a bare file name inside Dir is an artifact of this
// kind, and returns its path relative to the project root when it is. Anything
// else in those directories — a note, a temporary, a file the human owns — is
// not undo's business.
func (k Kind) Artifact(name string) (string, bool) {
	switch k {
	case Brief:
		if name == "brief.md" {
			return "brief.md", true
		}
	case Profile:
		if name == "profile.toml" {
			return path.Join(k.Dir(), name), true
		}
	case Fence:
		if name == "fence.toml" {
			return path.Join(k.Dir(), name), true
		}
	case Task:
		if strings.HasSuffix(name, ".md") && !strings.HasPrefix(name, ".") {
			return path.Join(k.Dir(), name), true
		}
	}
	return "", false
}

// SplitBackup reports whether a file name is the name of a backup and returns
// the artifact's name when it is. Strict on purpose: eight hex characters after
// `.bak-`, nothing after them. A backup of some other file is not a backup for
// odachi, and guessing would restore the wrong thing.
func SplitBackup(name string) (string, bool) {
	at := strings.LastIndex(name, backupMark)
	if at <= 0 {
		return "", false
	}

	artifact := name[:at]
	digest := name[at+len(backupMark):]
	if len(digest) != digestLen {
		return "", false
	}
	for _, c := range digest {
		if !isHex(c) {
			return "", false
		}
	}
	return artifact, true
}

// isHex is one hex digit, either case: the digest is written lowercase, and
// reading an uppercase one costs nothing.
func isHex(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// Backup is one backup file of one artifact. Root is the directory Path is
// relative to: the clone's store for a version this tool wrote, the project for
// a version that still sits where older releases left it. Entry is the same file
// relative to the store's backups directory, which is what --collect needs to
// move it without recomputing the artifact's path.
type Backup struct {
	Kind     Kind
	Artifact string // artifact path relative to the project root, slash-separated
	Root     string // absolute directory Path is relative to
	Path     string // the backup's own path, relative to Root
	Entry    string // the backup's path inside the store's backups directory
	ModTime  time.Time
}

// OfKind keeps the backups of one artifact: what `--file` asks for.
func OfKind(backups []Backup, kind Kind) []Backup {
	kept := make([]Backup, 0, len(backups))
	for _, candidate := range backups {
		if candidate.Kind == kind {
			kept = append(kept, candidate)
		}
	}
	return kept
}

// Sorted returns the backups newest first. Two backups can share a timestamp on
// a filesystem with coarse dates, so ties break on the path: the answer must
// never depend on the order the directory happened to list.
func Sorted(backups []Backup) []Backup {
	ordered := append([]Backup(nil), backups...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if !ordered[i].ModTime.Equal(ordered[j].ModTime) {
			return ordered[i].ModTime.After(ordered[j].ModTime)
		}
		return ordered[i].Path < ordered[j].Path
	})
	return ordered
}

// Latest is the backup the next `undo` restores: the most recent by mtime,
// across every artifact it was given.
func Latest(backups []Backup) (Backup, bool) {
	ordered := Sorted(backups)
	if len(ordered) == 0 {
		return Backup{}, false
	}
	return ordered[0], true
}

// Group is one artifact with its backups, newest first. It is what `--list`
// prints one line of.
type Group struct {
	Kind     Kind
	Artifact string
	Backups  []Backup
}

// Last is when the artifact was last backed up: the date `--list` shows.
func (g Group) Last() time.Time {
	if len(g.Backups) == 0 {
		return time.Time{}
	}
	return g.Backups[0].ModTime
}

// Groups collects the backups per artifact, newest artifact first, each group
// newest first. Only artifacts that have backups appear: a line saying zero
// would be noise in a list of what can be undone.
func Groups(backups []Backup) []Group {
	ordered := Sorted(backups)

	at := make(map[string]int, len(ordered))
	groups := make([]Group, 0, len(ordered))

	for _, candidate := range ordered {
		if index, seen := at[candidate.Artifact]; seen {
			groups[index].Backups = append(groups[index].Backups, candidate)
			continue
		}
		at[candidate.Artifact] = len(groups)
		groups = append(groups, Group{
			Kind:     candidate.Kind,
			Artifact: candidate.Artifact,
			Backups:  []Backup{candidate},
		})
	}
	return groups
}
