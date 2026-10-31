// Package clean implements `odachi clean`: what a clone accumulates, who owns
// each class of it, and what the run would remove.
//
// Ownership is the rule, not a taste. The proof is the evidence the gate reads,
// the plan belongs to the team, and the rest of the git directory belongs to git
// and to other tools. A tool that deletes what it did not write is a tool nobody
// can trust with a clone, so this verb reports every class with its owner and its
// size, and touches only the classes odachi itself wrote - each with a declared
// lifetime, because a class with no lifetime is a class that grows forever.
package clean

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/layout"
	"github.com/biggs-100/odachi/internal/domain/review"
	"github.com/biggs-100/odachi/internal/domain/store"
)

// Exit codes, the same shape as the other verbs.
const (
	// CodeOK means the report was printed, or the run cleaned what it named.
	CodeOK = 0
	// CodeBlocked means a path could not be read, and this run will not act on
	// what it cannot see.
	CodeBlocked = 1
)

// The lifetimes this version ships. They are the default shape of a clone whose
// project has not decided otherwise: enough verdicts to answer a question about
// the last week's work, and enough versions of one file to undo a mistake twice.
const (
	// KeepRecords is how many review trees survive: a verdict, its prompt and
	// the records beside them are one record, and a record is the unit.
	KeepRecords = 10
	// KeepBackups is how many versions of one artifact survive.
	KeepBackups = 5
)

// maxDepth bounds the walk of a class. A real store is four levels deep; anything
// deeper is a symlink or a shape odachi does not write, and a report must not
// hang on it.
const maxDepth = 6

// Options is one `odachi clean` invocation.
type Options struct {
	// Dir is the project: every path is named relative to it when it can be.
	Dir string

	// Store is the absolute root of the clone's state, resolved by the caller.
	// An empty Store is a folder that is not a git repository, and there is
	// nothing to clean there.
	Store string
}

// Files is the disk port: what the report reads, and nothing else.
type Files interface {
	Exists(path string) (bool, error)
	List(dir string) ([]string, error)
	IsDir(path string) (bool, error)
	Size(path string) (int64, error)
}

// Class is one kind of file a clone accumulates.
type Class struct {
	// Name is what the class is, in the words a human reads.
	Name string `json:"name"`
	// Owner is who wrote the files, and who decides their lifetime.
	Owner string `json:"owner"`
	// Path is where the class lives, named from the project.
	Path string `json:"path"`
	// Files and Bytes are what this run could read.
	Files int   `json:"files"`
	Bytes int64 `json:"bytes"`
	// Action is what `odachi clean --apply` would do to the class.
	Action string `json:"action"`
	// Cleanable marks the classes odachi wrote and this verb maintains. The
	// totals count these; nothing else is ever removed.
	Cleanable bool `json:"cleanable"`
	// Problems names the paths this run could not read. A failed read is not an
	// empty directory, and a class that lies about its size is worse than a
	// class that says it could not see.
	Problems []string `json:"unreadable,omitempty"`
}

// Report is the map a human reads and a machine consumes.
type Report struct {
	Store   string  `json:"store"`
	Classes []Class `json:"classes"`
	// TotalFiles and TotalBytes are odachi's own cleanable records.
	TotalFiles int   `json:"totalFiles"`
	TotalBytes int64 `json:"totalBytes"`
	// Problems is every unreadable path, in the order the classes found them.
	Problems []string `json:"unreadable,omitempty"`
	// Message is the one line the run answers with.
	Message string `json:"message"`
}

// Result is what the CLI prints and exits with.
type Result struct {
	Code    int
	Message string
	Lines   []string
	Report  Report
}

func (r Report) Lines() []string {
	lines := []string{"odachi clean · " + r.Store, ""}
	for _, class := range r.Classes {
		lines = append(lines, class.Line())
	}
	lines = append(lines, "", "  old agent notes are "+joinAnd(legacyPaths))
	if len(r.Problems) > 0 {
		lines = append(lines, "")
		for _, problem := range r.Problems {
			lines = append(lines, "  could not read "+problem)
		}
	}
	lines = append(lines, "", "  "+r.Message)
	return lines
}

// Line renders one class as the row a human reads.
func (c Class) Line() string {
	return fmt.Sprintf("  %-20s %-18s %4d %-6s %9s  %s",
		c.Name, c.Owner, c.Files, plural(c.Files), sizeText(c.Bytes), c.Action)
}

// Run performs one `odachi clean`.
func Run(opts Options, files Files) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}
	if opts.Store == "" {
		return blocked("odachi clean reads the clone's store, and this folder is not a git repository: run it inside one")
	}

	report := build(opts, files)
	return Result{
		Code:    CodeOK,
		Message: report.Message,
		Lines:   report.Lines(),
		Report:  report,
	}
}

// The owners. They are words a human reads, so they are spelled once.
const (
	ownerOdachi = "odachi"
	ownerAgent  = "an agent"
	ownerGit    = "git, other tools"
	ownerGate   = "the gate"
	ownerTeam   = "the team"
)

// legacyPaths are the names an agent gave its own notes before the store had a
// folder for them: the commit messages, the pull request bodies and the fences it
// generated, written into the git directory because the tree stayed clean there.
// They are another writer's files, so only --legacy names them as removable.
var legacyPaths = []string{
	".git/PR*.md",
	".git/PLAN*.txt",
	".git/CLOSE*.txt",
	".git/REL*.txt",
	".git/*MSG.txt",
	".git/T[0-9]_*.txt",
	".git/ROWBASE.txt",
	".git/HOLE.txt",
	".git/fence.generated",
}

// legacyPatterns is the same list as names inside the git directory, which is
// what the scan matches.
var legacyPatterns = []string{
	"PR*.md",
	"PLAN*.txt",
	"CLOSE*.txt",
	"REL*.txt",
	"*MSG.txt",
	"T[0-9]_*.txt",
	"ROWBASE.txt",
	"HOLE.txt",
	"fence.generated",
}

// build reads every class and assembles the report.
func build(opts Options, files Files) Report {
	// The store hangs from the git common directory, so its parent is the
	// directory git owns.
	gitDir := filepath.Dir(opts.Store)

	report := Report{Store: label(opts.Dir, opts.Store)}
	report.Classes = append(report.Classes,
		scan(files, opts.Dir, Class{
			Name:      "review records",
			Owner:     ownerOdachi,
			Action:    "keeps the newest " + strconv.Itoa(KeepRecords) + " trees",
			Cleanable: true,
		}, filepath.Join(opts.Store, store.Reviews), scanRule{}),
		scan(files, opts.Dir, Class{
			Name:      "review scaffolding",
			Owner:     ownerOdachi,
			Action:    "emptied by --apply",
			Cleanable: true,
		}, filepath.Join(opts.Store, review.AnswerDir), scanRule{}),
		scan(files, opts.Dir, Class{
			Name:      "undo backups",
			Owner:     ownerOdachi,
			Action:    "keeps the newest " + strconv.Itoa(KeepBackups) + " of each file",
			Cleanable: true,
		}, filepath.Join(opts.Store, store.Backups), scanRule{}),
		scan(files, opts.Dir, Class{
			Name:      "agent scratch",
			Owner:     ownerOdachi,
			Action:    "emptied by --apply",
			Cleanable: true,
		}, filepath.Join(opts.Store, store.Scratch), scanRule{}),
		legacy(files, opts.Dir, gitDir),
		state(files, opts.Dir, opts.Store),
		scan(files, opts.Dir, Class{
			Name:   "the rest of .git/",
			Owner:  ownerGit,
			Action: "never touched",
		}, gitDir, scanRule{
			// The store lists its own classes, and the old agent notes are a
			// class of their own: counting them here would count them twice.
			dirs:  []string{store.DirName},
			files: topLevelLegacy,
		}),
		scan(files, opts.Dir, Class{
			Name:   "proof/",
			Owner:  ownerGate,
			Action: "never touched: the proof the gate reads",
		}, filepath.Join(opts.Dir, "proof"), scanRule{}),
		scan(files, opts.Dir, Class{
			Name:   "odd/tasks",
			Owner:  ownerTeam,
			Action: "never touched: the plan",
		}, filepath.Join(opts.Dir, filepath.FromSlash(layout.Tasks)), scanRule{}),
	)

	for _, class := range report.Classes {
		if class.Cleanable {
			report.TotalFiles += class.Files
			report.TotalBytes += class.Bytes
		}
		report.Problems = append(report.Problems, class.Problems...)
	}

	report.Message = fmt.Sprintf("odachi's records: %s in %d %s",
		sizeText(report.TotalBytes), report.TotalFiles, plural(report.TotalFiles))
	if len(report.Problems) > 0 {
		report.Message += fmt.Sprintf("; %d %s it could not read",
			len(report.Problems), plural2(len(report.Problems), "path"))
	}
	return report
}

// scan builds a class from every file under root, leaving out what the rule
// names.
func scan(files Files, dir string, class Class, root string, rule scanRule) Class {
	class.Path = label(dir, root)
	names, problems := walk(files, dir, root, rule)
	class.Problems = problems
	for _, rel := range names {
		size, err := files.Size(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			class.Problems = append(class.Problems, label(dir, filepath.Join(root, filepath.FromSlash(rel)))+": "+oneLine(err.Error()))
			continue
		}
		class.Files++
		class.Bytes += size
	}
	return class
}

// legacy builds the class of notes an agent wrote into the git directory before
// the store had a folder for them. Only files are matched: a directory whose name
// fits a pattern belongs to somebody else.
func legacy(files Files, dir, gitDir string) Class {
	class := Class{
		Name:   "old agent notes",
		Owner:  ownerAgent,
		Path:   label(dir, gitDir) + "/",
		Action: "removed only by --apply --legacy",
	}
	names, err := files.List(gitDir)
	if err != nil {
		class.Problems = append(class.Problems, label(dir, gitDir)+": "+oneLine(err.Error()))
		return class
	}
	sort.Strings(names)
	for _, name := range names {
		if !matchesLegacy(name) {
			continue
		}
		full := filepath.Join(gitDir, name)
		isDir, err := files.IsDir(full)
		if err != nil {
			class.Problems = append(class.Problems, label(dir, full)+": "+oneLine(err.Error()))
			continue
		}
		if isDir {
			continue
		}
		size, err := files.Size(full)
		if err != nil {
			class.Problems = append(class.Problems, label(dir, full)+": "+oneLine(err.Error()))
			continue
		}
		class.Files++
		class.Bytes += size
	}
	return class
}

// matchesLegacy reports whether a name in the git directory is one of the notes
// an older agent left there.
func matchesLegacy(name string) bool {
	for _, pattern := range legacyPatterns {
		if ok, err := path.Match(pattern, name); err == nil && ok {
			return true
		}
	}
	return false
}

// topLevelLegacy says whether a path of the git directory is an old agent note:
// the patterns match names, so only the first level is asked.
func topLevelLegacy(rel string) bool {
	if strings.Contains(rel, "/") {
		return false
	}
	return matchesLegacy(rel)
}

// state is the store's own state -the client choice and the lock- which is never
// an accumulation and never has a lifetime.
func state(files Files, dir, storeRoot string) Class {
	class := Class{
		Name:   "the store's state",
		Owner:  ownerOdachi,
		Path:   label(dir, storeRoot),
		Action: "never touched: the client choice and the lock",
	}
	for _, name := range []string{store.Agents, "lock"} {
		full := filepath.Join(storeRoot, name)
		exists, err := files.Exists(full)
		if err != nil {
			class.Problems = append(class.Problems, label(dir, full)+": "+oneLine(err.Error()))
			continue
		}
		if !exists {
			continue
		}
		size, err := files.Size(full)
		if err != nil {
			class.Problems = append(class.Problems, label(dir, full)+": "+oneLine(err.Error()))
			continue
		}
		class.Files++
		class.Bytes += size
	}
	return class
}

// scanRule is what a class leaves out of its walk: a directory at the class's
// first level, or a file that already belongs to another class.
type scanRule struct {
	dirs  []string
	files func(rel string) bool
}

// leavesOut reports whether a name at the first level of the walk is not part of
// this class.
func (r scanRule) leavesOutDir(name string) bool {
	for _, dir := range r.dirs {
		if name == dir {
			return true
		}
	}
	return false
}

// walk lists every file under root, relative to it and with forward slashes. A
// directory that cannot be read is a problem, never an empty answer.
func walk(files Files, dir, root string, rule scanRule) ([]string, []string) {
	var (
		names    []string
		problems []string
	)

	var visit func(rel string, depth int)
	visit = func(rel string, depth int) {
		at := root
		if rel != "" {
			at = filepath.Join(root, filepath.FromSlash(rel))
		}
		entries, err := files.List(at)
		if err != nil {
			problems = append(problems, label(dir, at)+": "+oneLine(err.Error()))
			return
		}
		sort.Strings(entries)

		for _, name := range entries {
			if depth == 0 && rule.leavesOutDir(name) {
				continue
			}
			child := name
			if rel != "" {
				child = path.Join(rel, name)
			}
			full := filepath.Join(root, filepath.FromSlash(child))
			isDir, err := files.IsDir(full)
			if err != nil {
				problems = append(problems, label(dir, full)+": "+oneLine(err.Error()))
				continue
			}
			if isDir {
				if depth+1 >= maxDepth {
					problems = append(problems, label(dir, full)+": deeper than "+strconv.Itoa(maxDepth)+" levels, not read")
					continue
				}
				visit(child, depth+1)
				continue
			}
			if rule.files != nil && rule.files(child) {
				continue
			}
			names = append(names, child)
		}
	}
	visit("", 0)
	return names, problems
}

// label names a path the way the human reads it: relative to the project when it
// lives inside one, absolute when it does not.
func label(dir, at string) string {
	rel, err := filepath.Rel(dir, at)
	if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(at)
}

// joinAnd lists names the way a sentence does: commas between them and "and"
// before the last one.
func joinAnd(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	}
	return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
}

// plural says one file or many.
func plural(n int) string {
	if n == 1 {
		return "file"
	}
	return "files"
}

// plural2 says one word or many.
func plural2(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}

// sizeText is how a human reads a size: bytes below a kilobyte, one decimal
// below ten of them, whole kilobytes below a megabyte, one decimal above it.
func sizeText(bytes int64) string {
	switch {
	case bytes < 1024:
		return fmt.Sprintf("%d B", bytes)
	case bytes < 10*1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	case bytes < 1024*1024:
		return fmt.Sprintf("%.0f KB", float64(bytes)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
}

// blocked is a Result with exit code 1.
func blocked(message string) Result {
	return Result{Code: CodeBlocked, Message: oneLine(message), Lines: []string{oneLine(message)}}
}

// oneLine keeps every message on one line, whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
