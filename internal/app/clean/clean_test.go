package clean_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/clean"
)

const (
	dir   = "/proyecto"
	store = "/proyecto/.git/odachi"
)

// fakeFiles is the disk in memory: every path the report reads, when it was
// written, and what refuses to be read or removed.
type fakeFiles struct {
	files   map[string]string
	dates   map[string]time.Time
	broken  map[string]error
	refuse  map[string]error
	removed []string
}

func newFake(seed map[string]string) *fakeFiles {
	f := &fakeFiles{
		files:  map[string]string{},
		dates:  map[string]time.Time{},
		broken: map[string]error{},
		refuse: map[string]error{},
	}
	for at, content := range seed {
		f.files[filepath.FromSlash(filepath.Join(dir, at))] = content
	}
	return f
}

// put writes one file with the date it was written, so a lifetime that orders by
// time can be driven without a clock.
func (f *fakeFiles) put(at, content string, when time.Time) *fakeFiles {
	full := filepath.FromSlash(filepath.Join(dir, at))
	f.files[full] = content
	f.dates[full] = when
	return f
}

// brokenDir makes one directory unreadable, so a test can prove the report says
// so instead of counting it as empty.
func (f *fakeFiles) brokenDir(at string, err error) *fakeFiles {
	f.broken[filepath.FromSlash(filepath.Join(dir, at))] = err
	return f
}

// refuseRemove makes one file impossible to remove.
func (f *fakeFiles) refuseRemove(at string, err error) *fakeFiles {
	f.refuse[filepath.FromSlash(filepath.Join(dir, at))] = err
	return f
}

func (f *fakeFiles) Exists(path string) (bool, error) {
	if _, ok := f.files[path]; ok {
		return true, nil
	}
	return f.isDir(path), nil
}

func (f *fakeFiles) List(at string) ([]string, error) {
	if err, ok := f.broken[at]; ok {
		return nil, err
	}
	// A directory exists when something lives under it, so the listing is the
	// first segment of every path below at.
	seen := map[string]bool{}
	names := []string{}
	prefix := at + string(filepath.Separator)
	for path := range f.files {
		if !strings.HasPrefix(path, prefix) {
			continue
		}
		rest := strings.TrimPrefix(path, prefix)
		if i := strings.IndexRune(rest, filepath.Separator); i >= 0 {
			rest = rest[:i]
		}
		if seen[rest] {
			continue
		}
		seen[rest] = true
		names = append(names, rest)
	}
	return names, nil
}

func (f *fakeFiles) IsDir(path string) (bool, error) {
	return f.isDir(path), nil
}

func (f *fakeFiles) isDir(path string) bool {
	prefix := path + string(filepath.Separator)
	for key := range f.files {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

func (f *fakeFiles) Size(path string) (int64, error) {
	content, ok := f.files[path]
	if !ok {
		return 0, errors.New("missing file: " + path)
	}
	return int64(len(content)), nil
}

func (f *fakeFiles) ModTime(path string) (time.Time, error) {
	if _, ok := f.files[path]; !ok {
		return time.Time{}, errors.New("missing file: " + path)
	}
	return f.dates[path], nil
}

func (f *fakeFiles) Remove(path string) error {
	if err, ok := f.refuse[path]; ok {
		return err
	}
	if _, ok := f.files[path]; !ok {
		return errors.New("missing file: " + path)
	}
	delete(f.files, path)
	f.removed = append(f.removed, path)
	return nil
}

// has reports whether a path of the project survived the run.
func (f *fakeFiles) has(at string) bool {
	_, ok := f.files[filepath.FromSlash(filepath.Join(dir, at))]
	return ok
}

func run(f *fakeFiles) clean.Result {
	return clean.Run(clean.Options{Dir: dir, Store: store}, f, nil)
}

func apply(f *fakeFiles, opts clean.Options) clean.Result {
	opts.Dir = dir
	opts.Store = store
	return clean.Run(opts, f, nil)
}

// day is a fixed date, so a lifetime ordered by time never depends on the clock.
func day(n int) time.Time {
	return time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, n)
}

// A clone that has one of everything: the report names each class, its owner,
// its files and its size, and the block a human reads is pinned whole.
func TestTheReportNamesEveryClassWithItsOwner(t *testing.T) {
	f := newFake(map[string]string{
		".git/odachi/reviews/aaaa.json":               strings.Repeat("v", 100),
		".git/odachi/reviews/aaaa.prompt.md":          strings.Repeat("p", 50),
		".git/odachi/reviews/bbbb.json":               strings.Repeat("v", 20),
		".git/odachi/answers/cccc.json":               strings.Repeat("a", 10),
		".git/odachi/backups/brief.md.bak-1111":       strings.Repeat("b", 20),
		".git/odachi/backups/odd/fence.toml.bak-2222": strings.Repeat("f", 7),
		".git/odachi/scratch/commit.md":               strings.Repeat("s", 5),
		".git/odachi/agents.toml":                     strings.Repeat("c", 30),
		".git/PR100.md":                               strings.Repeat("r", 10),
		".git/CLOSE99.txt":                            strings.Repeat("c", 4),
		".git/COMMIT_EDITMSG":                         strings.Repeat("m", 6),
		".git/objects/ab/cdef":                        strings.Repeat("o", 3),
		"proof/0132-a-feature.txt":                    strings.Repeat("g", 9),
		"odd/tasks/0132-a-feature.md":                 strings.Repeat("t", 12),
	})

	result := run(f)
	if result.Code != clean.CodeOK {
		t.Fatalf("code = %d, want %d: %s", result.Code, clean.CodeOK, result.Message)
	}

	want := []string{
		"odachi clean · .git/odachi",
		"",
		"  review records       odachi                3 files      170 B  keeps the newest 10 trees",
		"  review scaffolding   odachi                1 file        10 B  emptied by --apply · 1 file would go",
		"  undo backups         odachi                2 files       27 B  keeps the newest 5 of each file",
		"  agent scratch        odachi                1 file         5 B  emptied by --apply · 1 file would go",
		"  old agent notes      an agent              2 files       14 B  removed only by --apply --legacy",
		"  the store's state    odachi                1 file        30 B  never touched: the client choice and the lock",
		"  the rest of .git/    git, other tools      2 files        9 B  never touched",
		"  proof/               the gate              1 file         9 B  never touched: the proof the gate reads",
		"  odd/tasks            the team              1 file        12 B  never touched: the plan",
		"",
		"  old agent notes are .git/PR*.md, .git/PLAN*.txt, .git/CLOSE*.txt, .git/REL*.txt, .git/*MSG.txt, .git/T[0-9]_*.txt, .git/ROWBASE.txt, .git/HOLE.txt and .git/fence.generated",
		"",
		"  odachi's records: 212 B in 7 files · --apply would free 15 B in 2 files",
	}
	got := strings.Join(result.Lines, "\n")
	if got != strings.Join(want, "\n") {
		t.Errorf("the report =\n%s\nwant\n%s", got, strings.Join(want, "\n"))
	}
}

// A directory that cannot be read is named and never counted as empty: the
// class keeps what it could read, and the report says so.
func TestAFailedReadIsNamedAndNotCountedAsEmpty(t *testing.T) {
	f := newFake(map[string]string{
		".git/odachi/reviews/aaaa.json": strings.Repeat("v", 100),
	}).brokenDir(".git/odachi/backups", errors.New("access is denied"))

	result := run(f)
	if result.Code != clean.CodeOK {
		t.Fatalf("code = %d, want %d: %s", result.Code, clean.CodeOK, result.Message)
	}

	printed := strings.Join(result.Lines, "\n")
	if !strings.Contains(printed, ".git/odachi/backups: access is denied") {
		t.Errorf("the report does not name the directory it could not read:\n%s", printed)
	}
	if !strings.Contains(result.Message, "1 path it could not read") {
		t.Errorf("message = %q, want it to name the unreadable path", result.Message)
	}

	for _, class := range result.Report.Classes {
		if class.Name != "undo backups" {
			continue
		}
		if class.Files != 0 || len(class.Problems) != 1 {
			t.Errorf("undo backups = %d files, problems %v; want 0 files and one problem", class.Files, class.Problems)
		}
	}
}

// The store is listed by its own classes, so the rest of .git never counts it
// twice: git's files and another tool's directory are one sum the run never
// touches.
func TestTheRestOfGitIsCountedWithoutTheStore(t *testing.T) {
	f := newFake(map[string]string{
		".git/odachi/reviews/aaaa.json":     strings.Repeat("v", 100),
		".git/COMMIT_EDITMSG":               strings.Repeat("m", 6),
		".git/objects/ab/cdef":              strings.Repeat("o", 3),
		".git/gentle-ai/state.json":         strings.Repeat("g", 4),
		".git/engram-project-identity.json": strings.Repeat("e", 2),
	})

	result := run(f)
	for _, class := range result.Report.Classes {
		if class.Name != "the rest of .git/" {
			continue
		}
		if class.Files != 4 {
			t.Errorf("the rest of .git = %d files, want 4 (git's three and the other tool's one)", class.Files)
		}
		if class.Bytes != 15 {
			t.Errorf("the rest of .git = %d bytes, want 15", class.Bytes)
		}
		return
	}
	t.Fatal("the report has no class for the rest of .git")
}

// An empty clone is a report of zeros, not an error: nothing to clean is an
// answer, and every class still teaches what it is.
func TestAnEmptyCloneReportsZeros(t *testing.T) {
	result := run(newFake(nil))
	if result.Code != clean.CodeOK {
		t.Fatalf("code = %d, want %d: %s", result.Code, clean.CodeOK, result.Message)
	}
	if len(result.Report.Classes) != 9 {
		t.Fatalf("classes = %d, want the nine the map names", len(result.Report.Classes))
	}
	if result.Report.TotalFiles != 0 || result.Report.TotalBytes != 0 {
		t.Errorf("totals = %d files, %d bytes; want zero", result.Report.TotalFiles, result.Report.TotalBytes)
	}
	if result.Message != "odachi's records: 0 B in 0 files" {
		t.Errorf("message = %q", result.Message)
	}
}

// A folder that is not a repository has no store, and the run says so instead of
// reporting a clone that does not exist.
func TestAFolderWithNoStoreIsRefused(t *testing.T) {
	result := clean.Run(clean.Options{Dir: dir}, newFake(nil), nil)
	if result.Code != clean.CodeBlocked {
		t.Fatalf("code = %d, want %d", result.Code, clean.CodeBlocked)
	}
	if !strings.Contains(result.Message, "not a git repository") {
		t.Errorf("message = %q, want it to name the missing repository", result.Message)
	}
}

// A directory inside a class is walked, not counted as a file: the backups keep
// the path of the artifact they save, and one nested version is one file.
func TestADirectoryIsWalkedAndNotCounted(t *testing.T) {
	f := newFake(map[string]string{
		".git/odachi/backups/odd/tasks/0132-a-feature.md.bak-aaaa": strings.Repeat("b", 40),
		".git/odachi/backups/odd/fence.toml.bak-bbbb":              strings.Repeat("f", 10),
	})
	result := run(f)
	for _, class := range result.Report.Classes {
		if class.Name != "undo backups" {
			continue
		}
		if class.Files != 2 || class.Bytes != 50 {
			t.Errorf("undo backups = %d files, %d bytes; want 2 and 50", class.Files, class.Bytes)
		}
		return
	}
	t.Fatal("the report has no class for the undo backups")
}

// A directory whose name fits a pattern belongs to somebody else: the class of
// old notes is files, and only files.
func TestADirectoryThatFitsAPatternIsNotALegacyNote(t *testing.T) {
	f := newFake(map[string]string{
		".git/PLAN/notes.txt": strings.Repeat("n", 5),
		".git/T1_100.txt":     strings.Repeat("t", 7),
	})

	result := run(f)
	for _, class := range result.Report.Classes {
		if class.Name != "old agent notes" {
			continue
		}
		if class.Files != 1 || class.Bytes != 7 {
			t.Errorf("old agent notes = %d files, %d bytes; want the T1 note only", class.Files, class.Bytes)
		}
	}
}

// The names an older agent used are the ones the class is made of: a file git
// owns is never one of them, and the notes of a step are.
func TestTheLegacyPatternsMatchTheNotesAnAgentLeft(t *testing.T) {
	f := newFake(map[string]string{
		".git/PR100.md":        strings.Repeat("p", 10),
		".git/CLOSEMSG.txt":    strings.Repeat("c", 4),
		".git/T1MSG.txt":       strings.Repeat("m", 6),
		".git/T1_100.txt":      strings.Repeat("t", 7),
		".git/ROWBASE.txt":     strings.Repeat("r", 8),
		".git/fence.generated": strings.Repeat("f", 9),
		".git/COMMIT_EDITMSG":  strings.Repeat("g", 3),
		".git/config":          strings.Repeat("o", 2),
	})

	result := run(f)
	for _, class := range result.Report.Classes {
		if class.Name != "old agent notes" {
			continue
		}
		if class.Files != 6 {
			t.Errorf("old agent notes = %d files, want the six the agent left", class.Files)
		}
	}
}

// The report says what --apply would free before it frees anything: a human
// decides with the number in front of them.
func TestTheReportSaysWhatApplyWouldFree(t *testing.T) {
	f := newFake(nil)
	for i := 0; i < 12; i++ {
		f.put(fmt.Sprintf(".git/odachi/reviews/tree%02d.json", i), strings.Repeat("v", 10), day(i))
	}

	result := run(f)
	if result.Report.DoomedFiles != 2 || result.Report.DoomedBytes != 20 {
		t.Errorf("doomed = %d files, %d bytes; want the two oldest records", result.Report.DoomedFiles, result.Report.DoomedBytes)
	}
	if !strings.Contains(result.Message, "would free 20 B in 2 files") {
		t.Errorf("message = %q, want it to say what --apply would free", result.Message)
	}
	if !strings.Contains(strings.Join(result.Lines, "\n"), "2 files would go") {
		t.Errorf("the row does not say what would go:\n%s", strings.Join(result.Lines, "\n"))
	}
}

// Applying a lifetime keeps the newest records whole: the tree that goes takes
// its verdict, its prompt and every record beside them.
func TestApplyKeepsTheNewestRecordsAndRemovesTheRest(t *testing.T) {
	f := newFake(nil)
	for i := 0; i < 12; i++ {
		tree := fmt.Sprintf("tree%02d", i)
		f.put(".git/odachi/reviews/"+tree+".json", strings.Repeat("v", 10), day(i))
		f.put(".git/odachi/reviews/"+tree+".prompt.md", strings.Repeat("p", 5), day(i))
	}

	result := apply(f, clean.Options{Apply: true})
	if result.Code != clean.CodeOK {
		t.Fatalf("code = %d, want %d: %s", result.Code, clean.CodeOK, result.Message)
	}
	if len(f.removed) != 4 {
		t.Errorf("removed %d files, want 4 (two records of two files)", len(f.removed))
	}
	for _, old := range []string{"tree00", "tree01"} {
		if f.has(".git/odachi/reviews/" + old + ".json") {
			t.Errorf("%s survived although it is the oldest", old)
		}
		if f.has(".git/odachi/reviews/" + old + ".prompt.md") {
			t.Errorf("the prompt of %s survived its verdict", old)
		}
	}
	if !f.has(".git/odachi/reviews/tree11.json") {
		t.Error("the newest record was removed")
	}
	if !strings.Contains(result.Message, "removed 4 files, 30 B freed") {
		t.Errorf("message = %q, want it to say what was freed", result.Message)
	}
}

// The versions of one artifact keep the newest five, and a name that is not a
// version odachi wrote is never one of them.
func TestApplyKeepsTheNewestFiveBackupsOfEachFile(t *testing.T) {
	f := newFake(map[string]string{
		".git/odachi/backups/notes.txt": strings.Repeat("n", 3),
	})
	for i := 0; i < 7; i++ {
		f.put(fmt.Sprintf(".git/odachi/backups/brief.md.bak-%08d", i), strings.Repeat("b", 10), day(i))
	}
	f.put(".git/odachi/backups/odd/fence.toml.bak-aaaaaaaa", strings.Repeat("f", 4), day(0))
	f.put(".git/odachi/backups/odd/fence.toml.bak-bbbbbbbb", strings.Repeat("f", 6), day(5))

	result := apply(f, clean.Options{Apply: true})
	if result.Code != clean.CodeOK {
		t.Fatalf("code = %d, want %d: %s", result.Code, clean.CodeOK, result.Message)
	}
	for i := 0; i < 2; i++ {
		if f.has(fmt.Sprintf(".git/odachi/backups/brief.md.bak-%08d", i)) {
			t.Errorf("version %d of the brief survived although five newer ones exist", i)
		}
	}
	for i := 2; i < 7; i++ {
		if !f.has(fmt.Sprintf(".git/odachi/backups/brief.md.bak-%08d", i)) {
			t.Errorf("version %d of the brief was removed although it is among the newest five", i)
		}
	}
	if !f.has(".git/odachi/backups/odd/fence.toml.bak-aaaaaaaa") {
		t.Error("a fence version was counted against the brief versions")
	}
	if !f.has(".git/odachi/backups/notes.txt") {
		t.Error("a file that is not a version was removed")
	}
}

// The two classes with no lifetime are emptied, and the report says so.
func TestApplyEmptiesScratchAndAnswers(t *testing.T) {
	f := newFake(map[string]string{
		".git/odachi/scratch/commit.md": strings.Repeat("s", 5),
		".git/odachi/scratch/pr.md":     strings.Repeat("p", 5),
		".git/odachi/answers/aaaa.json": strings.Repeat("a", 5),
	})

	result := apply(f, clean.Options{Apply: true})
	if result.Code != clean.CodeOK {
		t.Fatalf("code = %d, want %d: %s", result.Code, clean.CodeOK, result.Message)
	}
	if len(f.removed) != 3 {
		t.Errorf("removed %d files, want the three of the two classes", len(f.removed))
	}
	if !strings.Contains(result.Message, "removed 3 files") {
		t.Errorf("message = %q", result.Message)
	}
}

// What odachi does not own is never removed, whatever the lifetimes say.
func TestApplyRemovesNothingItDoesNotOwn(t *testing.T) {
	f := newFake(map[string]string{
		".git/odachi/agents.toml":     strings.Repeat("c", 30),
		".git/COMMIT_EDITMSG":         strings.Repeat("m", 6),
		".git/objects/ab/cdef":        strings.Repeat("o", 3),
		".git/gentle-ai/state.json":   strings.Repeat("g", 4),
		"proof/0132-a-feature.txt":    strings.Repeat("g", 9),
		"odd/tasks/0132-a-feature.md": strings.Repeat("t", 12),
		"odd/fence.toml":              strings.Repeat("f", 8),
	})

	result := apply(f, clean.Options{Apply: true})
	if result.Code != clean.CodeOK {
		t.Fatalf("code = %d, want %d: %s", result.Code, clean.CodeOK, result.Message)
	}
	if len(f.removed) != 0 {
		t.Errorf("removed %v, want nothing", f.removed)
	}
	if !strings.Contains(result.Message, "nothing to remove") {
		t.Errorf("message = %q, want it to say there was nothing to remove", result.Message)
	}
}

// The notes an older agent left are named by every run and removed only when the
// human asked for them.
func TestApplyKeepsTheOldAgentNotesWithoutLegacy(t *testing.T) {
	f := newFake(map[string]string{
		".git/PR100.md":       strings.Repeat("p", 10),
		".git/CLOSE99.txt":    strings.Repeat("c", 4),
		".git/COMMIT_EDITMSG": strings.Repeat("m", 6),
	})

	if result := apply(f, clean.Options{Apply: true}); result.Code != clean.CodeOK {
		t.Fatalf("code = %d, want %d: %s", result.Code, clean.CodeOK, result.Message)
	}
	if len(f.removed) != 0 {
		t.Errorf("removed %v without --legacy", f.removed)
	}

	result := apply(f, clean.Options{Apply: true, Legacy: true})
	if result.Code != clean.CodeOK {
		t.Fatalf("code = %d, want %d: %s", result.Code, clean.CodeOK, result.Message)
	}
	if len(f.removed) != 2 {
		t.Errorf("removed %d files, want the two notes", len(f.removed))
	}
	if !f.has(".git/COMMIT_EDITMSG") {
		t.Error("git's own file was removed")
	}
}

// --legacy alone is a run that names the class and removes nothing, so it is
// refused instead of being silently ignored.
func TestLegacyWithoutApplyIsRefused(t *testing.T) {
	result := clean.Run(clean.Options{Dir: dir, Store: store, Legacy: true}, newFake(nil), nil)
	if result.Code != clean.CodeBlocked {
		t.Fatalf("code = %d, want %d", result.Code, clean.CodeBlocked)
	}
	if !strings.Contains(result.Message, "--apply") {
		t.Errorf("message = %q, want it to name the flag that unblocks it", result.Message)
	}
}

// A run that cannot see a class it maintains removes nothing: a failed read is an
// answer, and deleting on a partial view is the defect the report exists to
// prevent.
func TestApplyRefusesWhileAClassItMaintainsCannotBeRead(t *testing.T) {
	f := newFake(map[string]string{
		".git/odachi/scratch/commit.md": strings.Repeat("s", 5),
	}).brokenDir(".git/odachi/reviews", errors.New("access is denied"))

	result := apply(f, clean.Options{Apply: true})
	if result.Code != clean.CodeBlocked {
		t.Fatalf("code = %d, want %d", result.Code, clean.CodeBlocked)
	}
	if len(f.removed) != 0 {
		t.Errorf("removed %v while a class could not be read", f.removed)
	}
	if !strings.Contains(result.Message, "review records") {
		t.Errorf("message = %q, want it to name the class it could not read", result.Message)
	}
}

// A file that cannot be removed stops the run and names itself; the files removed
// before it are reported as removed, because they are.
func TestAFailedRemovalNamesTheFileAndKeepsTheCount(t *testing.T) {
	f := newFake(map[string]string{
		".git/odachi/scratch/commit.md": strings.Repeat("s", 5),
		".git/odachi/scratch/pr.md":     strings.Repeat("p", 5),
	}).refuseRemove(".git/odachi/scratch/pr.md", errors.New("access is denied"))

	result := apply(f, clean.Options{Apply: true})
	if result.Code != clean.CodeBlocked {
		t.Fatalf("code = %d, want %d", result.Code, clean.CodeBlocked)
	}
	if !strings.Contains(result.Message, ".git/odachi/scratch/pr.md") {
		t.Errorf("message = %q, want it to name the file", result.Message)
	}
	if !strings.Contains(result.Message, "removed 1 of 2 files") {
		t.Errorf("message = %q, want it to count what was removed", result.Message)
	}
}
