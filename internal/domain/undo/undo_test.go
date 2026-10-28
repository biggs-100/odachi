package undo_test

import (
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/domain/undo"
)

// at is a fixed date, so a test about ordering never depends on the clock.
func at(day, hour, minute int) time.Time {
	return time.Date(2026, 10, day, hour, minute, 0, 0, time.UTC)
}

// TestTheAllowListIsExactlyFour: Contract X names the only artifacts undo may
// read or touch. These four strings are that list; anything else is not odachi's.
func TestTheAllowListIsExactlyFour(t *testing.T) {
	kinds := undo.Kinds()
	if len(kinds) != 4 {
		t.Fatalf("Kinds() = %v, want four", kinds)
	}

	want := []string{"brief.md", "odd/profile.toml", "odd/fence.toml", "odd/tasks/*.md"}
	for i, kind := range kinds {
		if got := kind.Target(); got != want[i] {
			t.Errorf("Kinds()[%d].Target() = %q, want %q", i, got, want[i])
		}
	}
}

// TestBackupSubdirPreservesTheArtifactPath: the store's backups directory is one,
// but each kind keeps its place inside it so a backup can never collide with a
// name from another directory.
func TestBackupSubdirPreservesTheArtifactPath(t *testing.T) {
	cases := map[undo.Kind]string{
		undo.Brief:   ".",
		undo.Profile: "odd",
		undo.Fence:   "odd",
		undo.Task:    "odd/tasks",
	}
	for kind, want := range cases {
		if got := kind.BackupSubdir(); got != want {
			t.Errorf("%s.BackupSubdir() = %q, want %q", kind, got, want)
		}
	}
}

// TestParseKindReadsTheFlag: --file takes the short name, and a miss is a miss
// so the caller can list the four.
func TestParseKindReadsTheFlag(t *testing.T) {
	for _, value := range []string{"brief", "profile", "fence", "task"} {
		kind, ok := undo.ParseKind(value)
		if !ok {
			t.Fatalf("ParseKind(%q) = false, want true", value)
		}
		if string(kind) != value {
			t.Errorf("ParseKind(%q) = %q, want %q", value, kind, value)
		}
	}

	for _, value := range []string{"", "inventado", "Brief", "taskdoc", "odd/tasks"} {
		if kind, ok := undo.ParseKind(value); ok {
			t.Errorf("ParseKind(%q) = %q, true; want false", value, kind)
		}
	}
}

// TestArtifactOfKeepsOnlyTheFour: a name inside an artifact's directory is an
// artifact only when Contract X says so.
func TestArtifactOfKeepsOnlyTheFour(t *testing.T) {
	cases := []struct {
		kind    undo.Kind
		name    string
		wantRel string
		wantOK  bool
	}{
		{undo.Brief, "brief.md", "brief.md", true},
		{undo.Brief, "README.md", "", false},
		{undo.Brief, "brief.md.tmp", "", false},
		{undo.Profile, "profile.toml", "odd/profile.toml", true},
		{undo.Profile, "fence.toml", "", false},
		{undo.Profile, "otro.toml", "", false},
		{undo.Fence, "fence.toml", "odd/fence.toml", true},
		{undo.Fence, "fence.yaml", "", false},
		{undo.Task, "0001-reserva-de-turno.md", "odd/tasks/0001-reserva-de-turno.md", true},
		{undo.Task, "notas.txt", "", false},
		{undo.Task, ".0001-oculto.md", "", false},
	}

	for _, c := range cases {
		rel, ok := c.kind.Artifact(c.name)
		if ok != c.wantOK || rel != c.wantRel {
			t.Errorf("%s.Artifact(%q) = %q, %v; want %q, %v", c.kind, c.name, rel, ok, c.wantRel, c.wantOK)
		}
	}
}

// TestSplitBackupIsStrict: a backup is `<archivo>.bak-<8 hex>` and nothing else.
func TestSplitBackupIsStrict(t *testing.T) {
	valid := map[string]string{
		"brief.md.bak-466dffe6":  "brief.md",
		"brief.md.bak-a317e32c":  "brief.md",
		"0001-x.md.bak-0000000a": "0001-x.md",
		"brief.md.bak-466DFFE6":  "brief.md",
	}
	for name, want := range valid {
		got, ok := undo.SplitBackup(name)
		if !ok || got != want {
			t.Errorf("SplitBackup(%q) = %q, %v; want %q, true", name, got, ok, want)
		}
	}

	invalid := []string{
		"brief.md",
		"brief.md.bak-466dffe",   // 7 hex
		"brief.md.bak-466dffe67", // 9 hex
		"brief.md.bak-466dffeg",  // not hex
		"brief.md.bak-",
		".bak-466dffe6",
		"brief.md.bak-466dffe6.tmp",
	}
	for _, name := range invalid {
		if got, ok := undo.SplitBackup(name); ok {
			t.Errorf("SplitBackup(%q) = %q, true; want false", name, got)
		}
	}
}

// backup is a one-line constructor for the selection tests.
func backup(artifact, backupPath string, when time.Time) undo.Backup {
	// Every caller names a real artifact; the kind is taken from the directory
	// conventions the domain already owns.
	kind := undo.Brief
	switch {
	case len(artifact) > 0 && artifact[0] == '.':
		kind = undo.Profile
	case len(artifact) > 8 && artifact[:8] == "odd/task":
		kind = undo.Task
	}
	return undo.Backup{Kind: kind, Artifact: artifact, Path: backupPath, ModTime: when}
}

// TestLatestIsTheMostRecentByModTime: choosing by the date the filesystem
// recorded is the whole selection rule.
func TestLatestIsTheMostRecentByModTime(t *testing.T) {
	backups := []undo.Backup{
		backup("brief.md", "brief.md.bak-aaaaaaaa", at(21, 20, 31)),
		backup("brief.md", "brief.md.bak-bbbbbbbb", at(21, 22, 5)),
		backup("brief.md", "brief.md.bak-cccccccc", at(20, 9, 0)),
	}

	got, ok := undo.Latest(backups)
	if !ok {
		t.Fatal("Latest() = false, want a backup")
	}
	if got.Path != "brief.md.bak-bbbbbbbb" {
		t.Errorf("Latest() = %q, want the newest by mtime", got.Path)
	}

	if _, ok := undo.Latest(nil); ok {
		t.Error("Latest(nil) = true, want false")
	}
}

// TestLatestDoesNotDependOnDirectoryOrder: two backups taken in the same second
// must still answer the same way on every machine.
func TestLatestDoesNotDependOnDirectoryOrder(t *testing.T) {
	same := at(21, 20, 31)
	a := backup("brief.md", "brief.md.bak-aaaaaaaa", same)
	b := backup("brief.md", "brief.md.bak-bbbbbbbb", same)

	first, _ := undo.Latest([]undo.Backup{a, b})
	second, _ := undo.Latest([]undo.Backup{b, a})
	if first.Path != second.Path {
		t.Errorf("the answer depends on input order: %q vs %q", first.Path, second.Path)
	}
}

// TestOfKindFiltersToOneArtifact: `--file brief` must ignore every other
// artifact, however recent.
func TestOfKindFiltersToOneArtifact(t *testing.T) {
	backups := []undo.Backup{
		backup("brief.md", "brief.md.bak-aaaaaaaa", at(20, 10, 0)),
		backup("odd/tasks/0001-x.md", "odd/tasks/0001-x.md.bak-bbbbbbbb", at(21, 23, 0)),
	}

	only := undo.OfKind(backups, undo.Brief)
	if len(only) != 1 || only[0].Artifact != "brief.md" {
		t.Fatalf("OfKind(brief) = %+v, want only brief.md", only)
	}

	if got := undo.OfKind(backups, undo.Fence); len(got) != 0 {
		t.Errorf("OfKind(fence) = %+v, want nothing", got)
	}
}

// TestGroupsShowsOneEntryPerArtifactNewestFirst: what `--list` prints a line of.
func TestGroupsShowsOneEntryPerArtifactNewestFirst(t *testing.T) {
	backups := []undo.Backup{
		backup("brief.md", "brief.md.bak-aaaaaaaa", at(20, 10, 0)),
		backup("brief.md", "brief.md.bak-bbbbbbbb", at(21, 20, 31)),
		backup("odd/tasks/0001-x.md", "odd/tasks/0001-x.md.bak-cccccccc", at(22, 8, 15)),
	}

	groups := undo.Groups(backups)
	if len(groups) != 2 {
		t.Fatalf("Groups() = %+v, want two artifacts", groups)
	}

	if groups[0].Artifact != "odd/tasks/0001-x.md" {
		t.Errorf("first group = %q, want the newest artifact", groups[0].Artifact)
	}
	if groups[1].Artifact != "brief.md" {
		t.Errorf("second group = %q, want brief.md", groups[1].Artifact)
	}
	if len(groups[1].Backups) != 2 {
		t.Fatalf("brief.md backups = %d, want 2", len(groups[1].Backups))
	}
	if groups[1].Backups[0].Path != "brief.md.bak-bbbbbbbb" {
		t.Errorf("brief.md backups[0] = %q, want the newest first", groups[1].Backups[0].Path)
	}
	if !groups[1].Last().Equal(at(21, 20, 31)) {
		t.Errorf("Last() = %v, want the newest backup's date", groups[1].Last())
	}
}
