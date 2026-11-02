package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// briefRel is the artifact a human edits most, so it carries the end-to-end.
const briefRel = "brief.md"

// readRel reads a file of the throwaway project.
func readRel(t *testing.T, dir, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}

// writeRel writes a file of the throwaway project, creating parents: this is
// the human editing by hand, outside odachi.
func writeRel(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

// backupsOf lists the backups of one artifact, sorted.
func backupsOf(t *testing.T, dir, rel string) []string {
	t.Helper()
	prefix := rel + ".bak-"
	var found []string
	for path := range tree(t, dir) {
		// New backups keep the artifact name inside .odachi/backups; old ones
		// are still beside the artifact. Both are this artifact's backups.
		if strings.HasPrefix(filepath.Base(path), prefix) {
			found = append(found, path)
		}
	}
	slices.Sort(found)
	return found
}

// backupProject runs `start` once, then once more with --force so odachi has
// overwritten its own artifacts and left backups behind. It returns the project
// directory and the bytes odachi wrote.
func backupProject(t *testing.T) (dir, written string) {
	t.Helper()
	dir, _ = startedProject(t)
	initRepo(t, dir)
	written = readRel(t, dir, briefRel)

	if code, _, stderr := runDir(t, dir, append([]string{"start", "--force"}, allFlags()...), ""); code != 0 {
		t.Fatalf("start --force: exit %d, stderr=%s", code, stderr)
	}
	return dir, written
}

// TestUndoRestoresTheLastArtifactOdachiWrote is the CLI contract for T16:
// restore the newest backup, keep the state that was replaced as a new backup,
// and let a second undo undo the undo.
func TestUndoRestoresTheLastArtifactOdachiWrote(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)

	// A project with the four artifacts but no backups has nothing to undo, and
	// says what odachi writes and which command writes it.
	if code, _, stderr := runDir(t, dir, allFlags(), ""); code != 0 {
		t.Fatalf("start: exit %d, stderr=%s", code, stderr)
	}
	written := readRel(t, dir, briefRel)

	code, _, stderr := runDir(t, dir, []string{"undo"}, "")
	if code != 1 {
		t.Fatalf("undo without backups: exit %d, want 1 (stderr=%s)", code, stderr)
	}
	for _, want := range []string{"odachi start", "brief.md", "odd/profile.toml", "odd/fence.toml", "odd/tasks/*.md"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr %q does not say %q", stderr, want)
		}
	}
	if strings.Count(strings.TrimSpace(stderr), "\n") != 0 {
		t.Errorf("the blocked message must be one line: %q", stderr)
	}

	// A second start with --force overwrites its own artifacts, so there is now
	// something to undo.
	if code, _, stderr := runDir(t, dir, append([]string{"start", "--force"}, allFlags()...), ""); code != 0 {
		t.Fatalf("start --force: exit %d, stderr=%s", code, stderr)
	}
	before := backupsOf(t, dir, briefRel)
	if len(before) != 1 {
		t.Fatalf("backups of %s = %v, want one", briefRel, before)
	}

	// The human edits the brief by hand, after the agent's last write. This is
	// the content that must never be lost.
	const hand = "# Mi brief\n\nEsto lo escribí yo, a mano.\n"
	writeRel(t, dir, briefRel, hand)

	code, stdout, stderr := runDir(t, dir, []string{"undo", "--list"}, "")
	if code != 0 {
		t.Fatalf("undo --list: exit %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, briefRel+" · 1 backup · last ") {
		t.Errorf("--list = %q, want a line for %s with its count and date", stdout, briefRel)
	}
	for _, path := range backupsOf(t, dir, briefRel) {
		if strings.Contains(stdout, path) {
			t.Errorf("--list = %q, want the artifact's name, not the backup's (%s)", stdout, path)
		}
	}

	code, stdout, stderr = runDir(t, dir, []string{"undo", "--file", "brief"}, "")
	if code != 0 {
		t.Fatalf("undo --file brief: exit %d, stderr=%s", code, stderr)
	}
	if want := "restored " + briefRel + " from " + before[0] + "; what was there was left in "; !strings.Contains(stdout, want) {
		t.Errorf("stdout = %q, want it to say %q", stdout, want)
	}
	if !strings.Contains(stdout, "git revert <sha>") {
		t.Errorf("stdout = %q, want it to name git revert <sha> as what undo does not do", stdout)
	}
	if got := readRel(t, dir, briefRel); got != written {
		t.Errorf("%s = %q, want the bytes odachi wrote back", briefRel, got)
	}

	after := backupsOf(t, dir, briefRel)
	if len(after) != 2 {
		t.Fatalf("backups of %s = %v, want the hand edit saved as a second one", briefRel, after)
	}
	var handBackup string
	for _, path := range after {
		if path != before[0] {
			handBackup = path
		}
	}
	if handBackup == "" {
		t.Fatalf("backups of %s = %v, want one of them new", briefRel, after)
	}
	if got := readRel(t, dir, handBackup); got != hand {
		t.Errorf("%s = %q, want the hand edit", handBackup, got)
	}
	if !strings.Contains(stdout, handBackup) {
		t.Errorf("stdout = %q, want it to name %s", stdout, handBackup)
	}

	// Undo is itself undoable: the second run brings the hand edit back. It names
	// the artifact because with more than one candidate undo refuses to guess.
	code, stdout, stderr = runDir(t, dir, []string{"undo", "--file", "brief"}, "")
	if code != 0 {
		t.Fatalf("second undo: exit %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "restored "+briefRel+" from "+handBackup) {
		t.Errorf("stdout = %q, want the undo of the undo to start from %s", stdout, handBackup)
	}
	if got := readRel(t, dir, briefRel); got != hand {
		t.Errorf("%s = %q, want the hand edit back", briefRel, got)
	}

	// A name outside the allow-list is refused, and the message says the four.
	code, _, stderr = runDir(t, dir, []string{"undo", "--file", "invented"}, "")
	if code != 1 {
		t.Fatalf("undo --file invented: exit %d, want 1", code)
	}
	for _, want := range []string{"invented", "brief", "profile", "fence", "task"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr %q does not name %q", stderr, want)
		}
	}
}

// TestUndoArtifactRestoresTheDocumentItNames: a project with several task
// documents is answerable when the run says which one, and the document comes
// back exactly as odachi last wrote it.
func TestUndoArtifactRestoresTheDocumentItNames(t *testing.T) {
	t.Parallel()
	dir, taskPath := startedProject(t)
	initRepo(t, dir)
	before := readRel(t, dir, taskPath)

	if code, _, stderr := runDir(t, dir, []string{"plan", "un paso", "--req", "R1"}, ""); code != 0 {
		t.Fatalf("plan: exit %d, stderr=%s", code, stderr)
	}
	if after := readRel(t, dir, taskPath); after == before {
		t.Fatal("plan did not change the document, so there is nothing to undo")
	}

	code, stdout, stderr := runDir(t, dir, []string{"undo", "--artifact", taskPath}, "")
	if code != 0 {
		t.Fatalf("undo --artifact: exit %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, taskPath) {
		t.Errorf("stdout %q does not name the document it restored", stdout)
	}
	if got := readRel(t, dir, taskPath); got != before {
		t.Errorf("the document was not restored:\n%s", got)
	}
}

// TestUndoCollectMovesLegacyBackups: the migration command is the bridge for a
// project that already has backups beside its artifacts.
func TestUndoCollectMovesLegacyBackups(t *testing.T) {
	t.Parallel()
	dir, _ := backupProject(t)
	const legacy = "brief legado\n"
	legacyRel := briefRel + ".bak-00000000"
	storeRel := ".git/odachi/backups/" + legacyRel
	writeRel(t, dir, legacyRel, legacy)

	code, stdout, stderr := runDir(t, dir, []string{"undo", "--collect"}, "")
	if code != 0 {
		t.Fatalf("undo --collect: exit %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "moved 1 backups") {
		t.Errorf("stdout = %q, want the count", stdout)
	}
	if got := readRel(t, dir, storeRel); got != legacy {
		t.Errorf("%s = %q, want the legacy bytes", storeRel, got)
	}
	if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(legacyRel))); !os.IsNotExist(err) {
		t.Errorf("the legacy backup still exists after collect: %v", err)
	}

	code, stdout, stderr = runDir(t, dir, []string{"undo", "--collect"}, "")
	if code != 0 {
		t.Fatalf("second undo --collect: exit %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "there are no old backups to move") {
		t.Errorf("second collect = %q, want a no-op", stdout)
	}
}

// TestUndoJSONHasTheSameContent: a machine gets what the human sees.
func TestUndoJSONHasTheSameContent(t *testing.T) {
	t.Parallel()
	dir, _ := backupProject(t)
	writeRel(t, dir, briefRel, "# a mano\n")

	code, human, stderr := runDir(t, dir, []string{"undo", "--list"}, "")
	if code != 0 {
		t.Fatalf("undo --list: exit %d, stderr=%s", code, stderr)
	}

	code, machine, stderr := runDir(t, dir, []string{"undo", "--list", "--json"}, "")
	if code != 0 {
		t.Fatalf("undo --list --json: exit %d, stderr=%s", code, stderr)
	}

	var payload struct {
		Code    int      `json:"code"`
		Message string   `json:"message"`
		Lines   []string `json:"lines"`
		Files   []string `json:"files"`
	}
	if err := json.Unmarshal([]byte(machine), &payload); err != nil {
		t.Fatalf("json: %v (%s)", err, machine)
	}
	if payload.Code != 0 {
		t.Errorf("code = %d, want 0", payload.Code)
	}
	humanLines := strings.Split(strings.TrimSpace(human), "\n")
	if !slices.Equal(payload.Lines, humanLines) {
		t.Errorf("Lines = %v, want the human lines %v", payload.Lines, humanLines)
	}
	if strings.Join(payload.Lines, "; ") != payload.Message {
		t.Errorf("Message = %q, want the lines joined", payload.Message)
	}
}
