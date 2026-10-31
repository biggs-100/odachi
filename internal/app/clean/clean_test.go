package clean_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/clean"
)

const (
	dir   = "/proyecto"
	store = "/proyecto/.git/odachi"
)

// fakeFiles is the disk in memory: every path the report reads and the
// directories that refuse to be read.
type fakeFiles struct {
	files  map[string]string
	broken map[string]error
}

func newFake(seed map[string]string) *fakeFiles {
	f := &fakeFiles{files: map[string]string{}, broken: map[string]error{}}
	for at, content := range seed {
		f.files[filepath.FromSlash(filepath.Join(dir, at))] = content
	}
	return f
}

// brokenDir makes one directory unreadable, so a test can prove the report says
// so instead of counting it as empty.
func (f *fakeFiles) brokenDir(at string, err error) *fakeFiles {
	f.broken[filepath.FromSlash(filepath.Join(dir, at))] = err
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
	// first segment of every path below at, and a nested file contributes its
	// top directory, not its own name.
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

func run(f *fakeFiles) clean.Result {
	return clean.Run(clean.Options{Dir: dir, Store: store}, f)
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
		"  review scaffolding   odachi                1 file        10 B  emptied by --apply",
		"  undo backups         odachi                2 files       27 B  keeps the newest 5 of each file",
		"  agent scratch        odachi                1 file         5 B  emptied by --apply",
		"  old agent notes      an agent              2 files       14 B  removed only by --apply --legacy",
		"  the store's state    odachi                1 file        30 B  never touched: the client choice and the lock",
		"  the rest of .git/    git, other tools      2 files        9 B  never touched",
		"  proof/               the gate              1 file         9 B  never touched: the proof the gate reads",
		"  odd/tasks            the team              1 file        12 B  never touched: the plan",
		"",
		"  old agent notes are .git/PR*.md, .git/PLAN*.txt, .git/CLOSE*.txt, .git/REL*.txt, .git/*MSG.txt, .git/T[0-9]_*.txt, .git/ROWBASE.txt, .git/HOLE.txt and .git/fence.generated",
		"",
		"  odachi's records: 212 B in 7 files",
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
	result := clean.Run(clean.Options{Dir: dir}, newFake(nil))
	if result.Code != clean.CodeBlocked {
		t.Fatalf("code = %d, want %d", result.Code, clean.CodeBlocked)
	}
	if !strings.Contains(result.Message, "not a git repository") {
		t.Errorf("message = %q, want it to name the missing repository", result.Message)
	}
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
