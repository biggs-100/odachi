package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// planDoc reads the current task document from disk.
func planDoc(t *testing.T, dir, taskPath string) taskdoc.TaskDoc {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, taskPath))
	if err != nil {
		t.Fatalf("read %s: %v", taskPath, err)
	}
	return taskdoc.Parse(string(data))
}

// TestPlanAppendsAndClosesStepsEndToEnd is the CLI contract for T15: the agent
// adds steps and closes them without editing Markdown by hand, and the document
// stays canonical by construction.
func TestPlanAppendsAndClosesStepsEndToEnd(t *testing.T) {
	t.Parallel()
	dir, taskPath := startedProject(t)

	code, stdout, stderr := runDir(t, dir, []string{"plan", "primer paso", "--req", "R1"}, "")
	if code != 0 {
		t.Fatalf("plan add: exit %d, stderr=%s", code, stderr)
	}
	if want := "step T1 added to " + taskPath; !strings.Contains(stdout, want) {
		t.Errorf("stdout %q does not say %q", stdout, want)
	}

	code, stdout, stderr = runDir(t, dir, []string{"plan", "segundo paso"}, "")
	if code != 0 {
		t.Fatalf("plan add 2: exit %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "step T2 added") {
		t.Errorf("stdout %q does not report T2", stdout)
	}

	doc := planDoc(t, dir, taskPath)
	if len(doc.Steps) != 2 {
		t.Fatalf("steps = %+v, want two", doc.Steps)
	}
	if doc.Steps[0].ID != "T1" || doc.Steps[0].Requirement != "R1" || doc.Steps[0].Done {
		t.Errorf("T1 = %+v, want open and linked to R1", doc.Steps[0])
	}
	if doc.Steps[1].ID != "T2" || doc.Steps[1].Requirement != "" || doc.Steps[1].Done {
		t.Errorf("T2 = %+v, want open and unlinked", doc.Steps[1])
	}
	if data, err := os.ReadFile(filepath.Join(dir, taskPath)); err != nil {
		t.Fatal(err)
	} else if got := string(data); got != taskdoc.Parse(got).Render() {
		t.Errorf("the document on disk is not canonical:\n%s", got)
	}

	code, stdout, stderr = runDir(t, dir, []string{"plan", "--done", "T1", "--commit", "deadbee"}, "")
	if code != 0 {
		t.Fatalf("plan done: exit %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "step T1 done · deadbee") {
		t.Errorf("stdout %q does not report the closed step and its commit", stdout)
	}

	step := planDoc(t, dir, taskPath).Steps[0]
	if !step.Done || step.Commit != "deadbee" {
		t.Errorf("T1 = %+v, want done with commit deadbee", step)
	}

	// An unknown id lists the open steps instead of guessing.
	code, _, stderr = runDir(t, dir, []string{"plan", "--done", "T9"}, "")
	if code != 1 {
		t.Fatalf("plan --done T9: exit %d, want 1 (stderr=%s)", code, stderr)
	}
	if !strings.Contains(stderr, "T9") || !strings.Contains(stderr, "T2") {
		t.Errorf("stderr %q does not name T9 and the open step T2", stderr)
	}

	// Closing a closed step is a no-op that says so and writes nothing.
	before, err := os.ReadFile(filepath.Join(dir, taskPath))
	if err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr = runDir(t, dir, []string{"plan", "--done", "T1"}, "")
	if code != 0 {
		t.Fatalf("plan --done T1 again: exit %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "was already done") {
		t.Errorf("stdout %q does not say it was already done", stdout)
	}
	after, err := os.ReadFile(filepath.Join(dir, taskPath))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Errorf("a no-op rewrote the document:\n%s", after)
	}
}

// TestPlanRouteFlagsEndToEnd: the two flags reach the document through the
// command line, and a contradiction is refused with the flag to change.
func TestPlanRouteFlagsEndToEnd(t *testing.T) {
	t.Parallel()
	dir, taskPath := startedProject(t)

	code, _, stderr := runDir(t, dir, []string{"plan", "primer paso", "--req", "R1", "--route", "delegated", "--agent", "odachi-worker"}, "")
	if code != 0 {
		t.Fatalf("plan --route delegated: exit %d, stderr=%s", code, stderr)
	}

	code, _, stderr = runDir(t, dir, []string{"plan", "segundo paso", "--route", "inline"}, "")
	if code != 0 {
		t.Fatalf("plan --route inline: exit %d, stderr=%s", code, stderr)
	}

	doc := planDoc(t, dir, taskPath)
	if len(doc.Steps) != 2 {
		t.Fatalf("steps = %+v, want two", doc.Steps)
	}
	if doc.Steps[0].Route != "delegated" || doc.Steps[0].Agent != "odachi-worker" {
		t.Errorf("T1 = %+v, want route delegated/odachi-worker", doc.Steps[0])
	}
	if doc.Steps[1].Route != "inline" || doc.Steps[1].Agent != "" {
		t.Errorf("T2 = %+v, want route inline with no agent", doc.Steps[1])
	}

	// A delegated step with no agent is refused, names the flag, and writes nothing.
	code, _, stderr = runDir(t, dir, []string{"plan", "tercer paso", "--route", "delegated"}, "")
	if code != 1 {
		t.Fatalf("plan --route delegated without --agent: exit %d, want 1", code)
	}
	if !strings.Contains(stderr, "--agent") {
		t.Errorf("stderr %q does not name --agent", stderr)
	}
	if got := len(planDoc(t, dir, taskPath).Steps); got != 2 {
		t.Errorf("the refused step was written: %d steps, want 2", got)
	}
}

// TestPlanRejectsBadTextBeforeWriting: the text is checked before anything is
// written, asserted on the filesystem.
func TestPlanRejectsBadTextBeforeWriting(t *testing.T) {
	t.Parallel()
	dir, taskPath := startedProject(t)
	path := filepath.Join(dir, taskPath)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	cases := [][]string{
		{"plan", ""},
		{"plan", "one line\ntwo lines"},
		{"plan", "-x"},
		{"plan", strings.Repeat("a", 121)},
	}
	for _, args := range cases {
		code, _, stderr := runDir(t, dir, args, "")
		if code != 1 {
			t.Errorf("%v: exit %d, want 1 (stderr=%s)", args, code, stderr)
		}
		if !strings.Contains(stderr, "odachi") {
			t.Errorf("%v: stderr %q does not read like an odachi error", args, stderr)
		}
		after, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(after) != string(before) {
			t.Fatalf("%v rewrote the document:\n%s", args, after)
		}
	}
}

// TestPlanRefusesAStraySectionAndWritesNothing: the shared guard at the CLI,
// asserted on the file, not only on the output.
func TestPlanRefusesAStraySectionAndWritesNothing(t *testing.T) {
	t.Parallel()
	dir, taskPath := startedProject(t)
	path := filepath.Join(dir, taskPath)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	handwritten := string(data) + "\n## Notas\nuna idea que el humano agregó a mano\n"
	if err := os.WriteFile(path, []byte(handwritten), 0o644); err != nil {
		t.Fatal(err)
	}

	code, _, stderr := runDir(t, dir, []string{"plan", "un paso más"}, "")

	if code != 1 {
		t.Fatalf("exit = %d, want 1 (stderr=%s)", code, stderr)
	}
	if !strings.Contains(stderr, "## History") {
		t.Errorf("stderr %q does not say where the notes belong", stderr)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != handwritten {
		t.Errorf("a blocked plan rewrote the document:\n%s", after)
	}
	if backups := backupFiles(t, dir); len(backups) != 0 {
		t.Errorf("a blocked plan left backups: %v", backups)
	}
}

// TestPlanWithoutATaskDocumentNamesStart: exit 1 and the command that creates
// the first document.
func TestPlanWithoutATaskDocumentNamesStart(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	code, _, stderr := runDir(t, dir, []string{"plan", "primer paso"}, "")

	if code != 1 {
		t.Fatalf("exit = %d, want 1 (stderr=%s)", code, stderr)
	}
	if !strings.Contains(stderr, "odachi start") {
		t.Errorf("stderr %q does not name odachi start", stderr)
	}
	if files := tree(t, dir); len(files) != 0 {
		t.Errorf("a blocked plan wrote files: %v", files)
	}
}

// TestPlanJSONReportsTheSameContent: machine output carries the same message
// the human reads, plus the task document and the file written.
func TestPlanJSONReportsTheSameContent(t *testing.T) {
	t.Parallel()
	dir, taskPath := startedProject(t)

	code, stdout, stderr := runDir(t, dir, []string{"plan", "un paso", "--json"}, "")
	if code != 0 {
		t.Fatalf("exit = %d, stderr=%s", code, stderr)
	}

	var payload struct {
		Code    int      `json:"code"`
		Message string   `json:"message"`
		Task    string   `json:"task"`
		Files   []string `json:"files"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout)
	}
	if payload.Code != 0 {
		t.Errorf("code = %d, want 0", payload.Code)
	}
	if want := "step T1 added to " + taskPath; payload.Message != want {
		t.Errorf("message = %q, want %q", payload.Message, want)
	}
	if payload.Task != taskPath {
		t.Errorf("task = %q, want %q", payload.Task, taskPath)
	}
	if len(payload.Files) != 1 || payload.Files[0] != taskPath {
		t.Errorf("files = %v, want [%s]", payload.Files, taskPath)
	}
}

// TestPlanLeavesTheOtherArtifactsAlone: brief, profile and fence are untouched,
// and only the task document is backed up.
func TestPlanLeavesTheOtherArtifactsAlone(t *testing.T) {
	t.Parallel()
	dir, taskPath := startedProject(t)

	before := tree(t, dir)
	code, _, stderr := runDir(t, dir, []string{"plan", "un paso"}, "")
	if code != 0 {
		t.Fatalf("exit = %d, stderr=%s", code, stderr)
	}
	after := tree(t, dir)

	for _, name := range []string{"brief.md", "odd/profile.toml", "odd/fence.toml"} {
		if after[name] != before[name] {
			t.Errorf("%s changed:\nbefore %q\nafter  %q", name, before[name], after[name])
		}
	}
	if backups := backupFiles(t, dir); len(backups) != 1 || !strings.Contains(backups[0], taskPath) {
		t.Errorf("backups = %v, want only one for %s", backups, taskPath)
	}
}

// backupFiles lists the backup files a verb left in the tree.
func backupFiles(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	for path := range tree(t, dir) {
		if strings.Contains(path, ".bak-") {
			out = append(out, path)
		}
	}
	return out
}

// TestPlanFlagRulesAtTheCLI: the three refusals that do not depend on the
// document's content, each naming the command or flag that unblocks it.
func TestPlanFlagRulesAtTheCLI(t *testing.T) {
	t.Parallel()
	dir, _ := startedProject(t)
	before := tree(t, dir)

	cases := []struct {
		args []string
		want string
	}{
		{[]string{"plan"}, `odachi plan "<text>"`},
		{[]string{"plan", "--commit", "deadbee"}, "--done"},
		{[]string{"plan", "un paso", "--req", "R9"}, "R9"},
	}

	for _, c := range cases {
		code, _, stderr := runDir(t, dir, c.args, "")
		if code != 1 {
			t.Errorf("%v: exit %d, want 1 (stderr=%s)", c.args, code, stderr)
		}
		if !strings.Contains(stderr, c.want) {
			t.Errorf("%v: stderr %q does not name %q", c.args, stderr, c.want)
		}
	}

	after := tree(t, dir)
	for path, content := range before {
		if after[path] != content {
			t.Errorf("%v rewrote %s", cases, path)
		}
	}
	if len(after) != len(before) {
		t.Errorf("a blocked plan left files: before %v, after %v", before, after)
	}
}
