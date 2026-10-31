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
	"time"

	"github.com/biggs-100/odachi/internal/app/lock"
	"github.com/biggs-100/odachi/internal/domain/layout"
	"github.com/biggs-100/odachi/internal/domain/review"
	"github.com/biggs-100/odachi/internal/domain/store"
	"github.com/biggs-100/odachi/internal/domain/undo"
)

// Exit codes, the same shape as the other verbs.
const (
	// CodeOK means the report was printed, or the run removed what it named.
	CodeOK = 0
	// CodeBlocked means a path could not be read or removed, and this run will
	// not guess about what it cannot see.
	CodeBlocked = 1
)

// KeepBackups is how many versions of one artifact survive. It is not a project
// decision: a version is the way back from a mistake, and five of them are the
// same answer in every project.
const KeepBackups = 5

// verbName is the lock owner.
const verbName = "clean"

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

	// Apply removes what the report names. Without it the run only reads.
	Apply bool

	// Legacy also removes the notes an older agent wrote into the git directory.
	// It only means something with Apply, and it is never implied.
	Legacy bool

	// KeepVerdicts and KeepPrompts are the lifetimes the project decided, read
	// from odd/review.toml by the caller. Zero means the default odachi ships, so
	// a folder the caller could not read cleans the way it always did.
	KeepVerdicts int
	KeepPrompts  int
	// LimitsFrom names where the numbers came from, in the words a human reads:
	// "odd/review.toml" or "odachi's defaults". The report says it, so a number
	// nobody remembers choosing is never a mystery.
	LimitsFrom string
}

// keptVerdicts and keptPrompts are the numbers this run obeys: the project's when
// it decided, the shipped default when it did not.
func (o Options) keptVerdicts() int {
	if o.KeepVerdicts > 0 {
		return o.KeepVerdicts
	}
	return review.DefaultKeepVerdicts
}

func (o Options) keptPrompts() int {
	if o.KeepPrompts > 0 {
		return o.KeepPrompts
	}
	return review.DefaultKeepPrompts
}

// Files is the disk port: what the report reads, and what --apply removes.
type Files interface {
	Exists(path string) (bool, error)
	List(dir string) ([]string, error)
	IsDir(path string) (bool, error)
	Size(path string) (int64, error)
	ModTime(path string) (time.Time, error)
	Remove(path string) error
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
	// Action is what `odachi clean --apply` does to the class.
	Action string `json:"action"`
	// Doomed is how many files --apply would remove, and Dead is their size.
	Doomed int   `json:"doomed"`
	Dead   int64 `json:"doomedBytes"`
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
	// DoomedFiles and DoomedBytes are what --apply removed, or would remove.
	DoomedFiles int   `json:"doomedFiles"`
	DoomedBytes int64 `json:"doomedBytes"`
	// Applied is false for a report and true once the run removed the files.
	Applied bool `json:"applied"`
	// KeepVerdicts, KeepPrompts and LimitsFrom are the lifetimes this run obeyed
	// and where they came from, so the report answers "why ten?".
	KeepVerdicts int    `json:"keepVerdicts"`
	KeepPrompts  int    `json:"keepPrompts"`
	LimitsFrom   string `json:"limitsFrom,omitempty"`
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

// Lines is the block a human reads.
func (r Report) Lines() []string {
	lines := []string{"odachi clean · " + r.Store, ""}
	for _, class := range r.Classes {
		lines = append(lines, class.Line(r.Applied))
	}
	lines = append(lines, "", "  old agent notes are "+joinAnd(legacyPaths))
	if r.LimitsFrom != "" {
		lines = append(lines, "", "  retention: the newest "+strconv.Itoa(r.KeepVerdicts)+" trees and the prompts of the newest "+
			strconv.Itoa(r.KeepPrompts)+", from "+r.LimitsFrom)
	}
	if len(r.Problems) > 0 {
		lines = append(lines, "")
		for _, problem := range r.Problems {
			lines = append(lines, "  could not read "+problem)
		}
	}
	lines = append(lines, "", "  "+r.Message)
	return lines
}

// Line renders one class as the row a human reads, with what the run would do or
// did to it.
func (c Class) Line(applied bool) string {
	row := fmt.Sprintf("  %-20s %-18s %4d %-6s %9s  %s",
		c.Name, c.Owner, c.Files, plural(c.Files), sizeText(c.Bytes), c.Action)
	if c.Doomed == 0 {
		return row
	}
	verb := "would go"
	if applied {
		verb = "removed"
	}
	return row + " · " + strconv.Itoa(c.Doomed) + " " + plural(c.Doomed) + " " + verb
}

// Run performs one `odachi clean`.
func Run(opts Options, files Files, guard lock.Lock) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}
	if opts.Store == "" {
		return blocked("odachi clean reads the clone's store, and this folder is not a git repository: run it inside one")
	}
	if opts.Legacy && !opts.Apply {
		return blocked("--legacy names what --apply removes, and this run only reads: add --apply to remove the old agent notes, or leave --legacy out")
	}

	if !opts.Apply {
		report := scan(opts, files).report
		return Result{Code: CodeOK, Message: report.Message, Lines: report.Lines(), Report: report}
	}

	// Removing many files takes the lock: two runs deleting at once would each
	// count the other's work as their own.
	held := lock.Take(guard, opts.Store, verbName)
	if held.Refusal != "" {
		return blocked(held.Refusal)
	}
	defer held.Release()

	return withNotice(apply(opts, files), held.Notice)
}

// apply removes what the report named, and refuses before the first file when a
// class it maintains could not be read: a run that cannot see does not delete.
func apply(opts Options, files Files) Result {
	found := scan(opts, files)
	report := found.report

	for _, class := range report.Classes {
		if !class.Cleanable || len(class.Problems) == 0 {
			continue
		}
		return blocked(fmt.Sprintf(
			"clean removes nothing while it cannot read %s: %s; fix the permissions and run it again",
			class.Name, class.Problems[0]))
	}

	removed := 0
	freed := int64(0)
	for _, at := range found.doomed {
		if err := files.Remove(at); err != nil {
			report.Applied = removed > 0
			report.Message = fmt.Sprintf("removed %d of %d files before %s could not be removed: %s",
				removed, len(found.doomed), label(opts.Dir, at), oneLine(err.Error()))
			return Result{Code: CodeBlocked, Message: report.Message, Lines: report.Lines(), Report: report}
		}
		removed++
		freed += found.sizes[at]
	}

	report.Applied = removed > 0
	report.Message = summary(report.TotalBytes, report.TotalFiles)
	if removed > 0 {
		report.Message += fmt.Sprintf(" · removed %d %s, %s freed", removed, plural(removed), sizeText(freed))
	} else {
		report.Message += " · nothing to remove"
	}
	return Result{Code: CodeOK, Message: report.Message, Lines: report.Lines(), Report: report}
}

// found is what a run sees and what it would remove.
type found struct {
	report Report
	doomed []string
	sizes  map[string]int64
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

// entry is one file a class holds.
type entry struct {
	rel  string // relative to the class root, forward slashes
	size int64
	at   time.Time
}

// scan reads every class, applies each lifetime and assembles the report.
func scan(opts Options, files Files) found {
	// The store hangs from the git common directory, so its parent is the
	// directory git owns.
	gitDir := filepath.Dir(opts.Store)
	seen := found{sizes: map[string]int64{}}

	records, prompts := seen.reviews(files, opts)
	classes := []Class{
		records,
		prompts,
		seen.everything(files, opts, Class{
			Name:      "review scaffolding",
			Owner:     ownerOdachi,
			Action:    "emptied by --apply",
			Cleanable: true,
		}, filepath.Join(opts.Store, review.AnswerDir)),
		seen.backups(files, opts),
		seen.everything(files, opts, Class{
			Name:      "agent scratch",
			Owner:     ownerOdachi,
			Action:    "emptied by --apply",
			Cleanable: true,
		}, filepath.Join(opts.Store, store.Scratch)),
		seen.legacy(files, opts, gitDir),
		state(files, opts.Dir, opts.Store),
		seen.read(files, opts.Dir, Class{
			Name:   "the rest of .git/",
			Owner:  ownerGit,
			Action: "never touched",
		}, gitDir, scanRule{
			// The store lists its own classes, and the old agent notes are a
			// class of their own: counting them here would count them twice.
			dirs:  []string{store.DirName},
			files: topLevelLegacy,
		}),
		seen.read(files, opts.Dir, Class{
			Name:   "proof/",
			Owner:  ownerGate,
			Action: "never touched: the proof the gate reads",
		}, filepath.Join(opts.Dir, "proof"), scanRule{}),
		seen.read(files, opts.Dir, Class{
			Name:   "odd/tasks",
			Owner:  ownerTeam,
			Action: "never touched: the plan",
		}, filepath.Join(opts.Dir, filepath.FromSlash(layout.Tasks)), scanRule{}),
	}

	seen.report = Report{
		Store:        label(opts.Dir, opts.Store),
		Classes:      classes,
		KeepVerdicts: opts.keptVerdicts(),
		KeepPrompts:  opts.keptPrompts(),
		LimitsFrom:   opts.LimitsFrom,
	}
	for _, class := range classes {
		if class.Cleanable {
			seen.report.TotalFiles += class.Files
			seen.report.TotalBytes += class.Bytes
		}
		seen.report.DoomedFiles += class.Doomed
		seen.report.DoomedBytes += class.Dead
		seen.report.Problems = append(seen.report.Problems, class.Problems...)
	}

	seen.report.Message = summary(seen.report.TotalBytes, seen.report.TotalFiles)
	switch {
	case seen.report.DoomedFiles > 0:
		seen.report.Message += fmt.Sprintf(" · --apply would free %s in %d %s",
			sizeText(seen.report.DoomedBytes), seen.report.DoomedFiles, plural(seen.report.DoomedFiles))
	case len(seen.report.Problems) > 0:
		seen.report.Message += fmt.Sprintf("; %d %s it could not read",
			len(seen.report.Problems), plural2(len(seen.report.Problems), "path"))
	}
	return seen
}

// reviews reads the store's reviews directory once and parts it in two classes,
// because the two kinds of file are two decisions: a record is the evidence the
// gate reads, and a prompt is the question that was sent, which nothing reads.
// The project decides how many of each survive, counted in trees.
func (f *found) reviews(files Files, opts Options) (Class, Class) {
	root := filepath.Join(opts.Store, store.Reviews)
	records := Class{
		Name:      "review records",
		Owner:     ownerOdachi,
		Path:      label(opts.Dir, root),
		Action:    "keeps the newest " + strconv.Itoa(opts.keptVerdicts()) + " " + plural2(opts.keptVerdicts(), "tree"),
		Cleanable: true,
	}
	prompts := Class{
		Name:      "review prompts",
		Owner:     ownerOdachi,
		Path:      label(opts.Dir, root),
		Action:    "keeps the prompts of the newest " + strconv.Itoa(opts.keptPrompts()) + " " + plural2(opts.keptPrompts(), "tree"),
		Cleanable: true,
	}

	class, entries := gather(files, opts.Dir, Class{}, root, scanRule{})
	records.Problems = class.Problems
	prompts.Problems = class.Problems

	byTree := map[string][]entry{}
	for _, item := range entries {
		tree := treeOf(item.rel)
		byTree[tree] = append(byTree[tree], item)
	}

	// The trees are ordered once, by the newest file of each, and both lifetimes
	// read that order: a prompt and a verdict about the same tree are the same
	// day's evidence.
	for index, tree := range treesNewestFirst(byTree) {
		keepRecords := index < opts.keptVerdicts()
		keepPrompt := index < opts.keptPrompts()
		for _, item := range byTree[tree] {
			isPromptFile := isPrompt(item.rel)
			if isPromptFile {
				prompts.Files++
				prompts.Bytes += item.size
				if keepPrompt {
					continue
				}
				prompts = f.doom(prompts, root, item)
				continue
			}
			records.Files++
			records.Bytes += item.size
			if keepRecords {
				continue
			}
			records = f.doom(records, root, item)
		}
	}
	return records, prompts
}

// isPrompt says whether a store file is the question that was sent for a tree:
// `<tree>.prompt.md` or one per lens group, `<tree>.prompt-risk+readability.md`.
func isPrompt(rel string) bool {
	rest := strings.TrimPrefix(rel, treeOf(rel))
	return strings.HasPrefix(rest, ".prompt")
}

// backups is the class with a lifetime counted in artifacts: the newest
// KeepBackups versions of each file survive, and a name that is not a version
// odachi wrote never joins a group.
func (f *found) backups(files Files, opts Options) Class {
	root := filepath.Join(opts.Store, store.Backups)
	class, entries := gather(files, opts.Dir, Class{
		Name:      "undo backups",
		Owner:     ownerOdachi,
		Action:    "keeps the newest " + strconv.Itoa(KeepBackups) + " of each file",
		Cleanable: true,
	}, root, scanRule{})

	byArtifact := map[string][]entry{}
	for _, item := range entries {
		artifact, ok := artifactOf(item.rel)
		if !ok {
			continue
		}
		byArtifact[artifact] = append(byArtifact[artifact], item)
	}
	return f.keepVersions(class, root, byArtifact, KeepBackups)
}

// everything is a class with no lifetime at all: --apply empties it.
func (f *found) everything(files Files, opts Options, class Class, root string) Class {
	class, entries := gather(files, opts.Dir, class, root, scanRule{})
	for _, item := range entries {
		class = f.doom(class, root, item)
	}
	return class
}

// read is a class nobody may remove: it is walked, measured and named, and no
// file of it ever enters the plan.
func (f *found) read(files Files, dir string, class Class, root string, rule scanRule) Class {
	class, _ = gather(files, dir, class, root, rule)
	return class
}

// legacy names the notes an older agent left in the git directory, and plans them
// only when the run asked for them.
func (f *found) legacy(files Files, opts Options, gitDir string) Class {
	class := Class{
		Name:   "old agent notes",
		Owner:  ownerAgent,
		Path:   label(opts.Dir, gitDir) + "/",
		Action: "removed only by --apply --legacy",
	}
	names, err := files.List(gitDir)
	if err != nil {
		class.Problems = append(class.Problems, label(opts.Dir, gitDir)+": "+oneLine(err.Error()))
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
			class.Problems = append(class.Problems, label(opts.Dir, full)+": "+oneLine(err.Error()))
			continue
		}
		if isDir {
			continue
		}
		size, err := files.Size(full)
		if err != nil {
			class.Problems = append(class.Problems, label(opts.Dir, full)+": "+oneLine(err.Error()))
			continue
		}
		class.Files++
		class.Bytes += size
		if opts.Apply && opts.Legacy {
			class = f.doom(class, gitDir, entry{rel: name, size: size})
		}
	}
	return class
}

// treesNewestFirst orders the groups by the newest file of each, and breaks a tie
// on the key, so the answer never depends on the order the directory was listed
// in.
func treesNewestFirst(groups map[string][]entry) []string {
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left, right := latest(groups[keys[i]]), latest(groups[keys[j]])
		if !left.Equal(right) {
			return left.After(right)
		}
		return keys[i] > keys[j]
	})
	return keys
}

// keepVersions keeps the newest files of every group: a backup is one version
// of one artifact, and each artifact carries its own lifetime. A tie breaks on
// the path, so the answer never depends on the order the directory was listed in.
func (f *found) keepVersions(class Class, root string, groups map[string][]entry, keep int) Class {
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		items := append([]entry(nil), groups[key]...)
		sort.Slice(items, func(i, j int) bool {
			if !items[i].at.Equal(items[j].at) {
				return items[i].at.After(items[j].at)
			}
			return items[i].rel > items[j].rel
		})
		for index, item := range items {
			if index < keep {
				continue
			}
			class = f.doom(class, root, item)
		}
	}
	return class
}

// doom adds one file to the plan and to the class that owns it.
func (f *found) doom(class Class, root string, item entry) Class {
	full := filepath.Join(root, filepath.FromSlash(item.rel))
	class.Doomed++
	class.Dead += item.size
	f.doomed = append(f.doomed, full)
	f.sizes[full] = item.size
	return class
}

// gather walks a class and returns the row with every file it read, and every
// file it could not.
func gather(files Files, dir string, class Class, root string, rule scanRule) (Class, []entry) {
	class.Path = label(dir, root)
	names, problems := walk(files, dir, root, rule)
	class.Problems = problems

	entries := make([]entry, 0, len(names))
	for _, rel := range names {
		full := filepath.Join(root, filepath.FromSlash(rel))
		size, err := files.Size(full)
		if err != nil {
			class.Problems = append(class.Problems, label(dir, full)+": "+oneLine(err.Error()))
			continue
		}
		// A date nobody could read is not "old": without it the file cannot be
		// ordered, so it leaves the class's count and the run says why.
		at, err := files.ModTime(full)
		if err != nil {
			class.Problems = append(class.Problems, label(dir, full)+": "+oneLine(err.Error()))
			continue
		}
		class.Files++
		class.Bytes += size
		entries = append(entries, entry{rel: rel, size: size, at: at})
	}
	return class, entries
}

// treeOf is the record a file in the reviews directory belongs to: the tree hash
// every name starts with.
func treeOf(rel string) string {
	if at := strings.Index(rel, "."); at > 0 {
		return rel[:at]
	}
	return rel
}

// artifactOf is the file a backup saves, or false when the name is not a version
// odachi wrote. The guess is refused on purpose: deleting a file somebody else
// named would be exactly the mistake this verb exists to avoid.
func artifactOf(rel string) (string, bool) {
	dir, name := path.Split(rel)
	artifact, ok := undo.SplitBackup(name)
	if !ok {
		return "", false
	}
	return path.Join(dir, artifact), true
}

// latest is when the most recent file of a group was written.
func latest(items []entry) time.Time {
	newest := time.Time{}
	for _, item := range items {
		if item.at.After(newest) {
			newest = item.at
		}
	}
	return newest
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
	for _, name := range []string{store.Agents, lock.File} {
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

// leavesOutDir reports whether a name at the first level of the walk is not part
// of this class.
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

// summary is the one line about odachi's own records.
func summary(bytes int64, files int) string {
	return fmt.Sprintf("odachi's records: %s in %d %s", sizeText(bytes), files, plural(files))
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

// withNotice puts the lock's announcement in front of the run's own answer.
func withNotice(result Result, notice string) Result {
	if notice == "" {
		return result
	}
	result.Lines = append([]string{notice}, result.Lines...)
	result.Message = notice + " " + result.Message
	return result
}

// blocked is a Result with exit code 1.
func blocked(message string) Result {
	return Result{Code: CodeBlocked, Message: oneLine(message), Lines: []string{oneLine(message)}}
}

// oneLine keeps every message on one line, whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
