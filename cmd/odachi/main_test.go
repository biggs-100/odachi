package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/profiletoml"
	"github.com/biggs-100/odachi/internal/app/start"
)

// runDir runs the CLI inside dir with a non-terminal stdin, so the machine path
// is exercised exactly like CI would.
func runDir(t *testing.T, dir string, args []string, stdin string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = runIn(dir, args, strings.NewReader(stdin), &out, &errOut)
	return code, out.String(), errOut.String()
}

// allFlags answers the five required questions and the two optional ones.
func allFlags() []string {
	return []string{
		"--outcome", "Quiero reservar turnos desde el teléfono",
		"--audience", "phone",
		"--done", "Reservo un turno y aparece en la agenda",
		"--exclude", "pagos",
		"--sensitive", "none",
		"--deadline", "today",
		"--learn", "tests",
		"--json",
	}
}

// tree reads every file under dir, keyed by a slash path.
func tree(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = string(data)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	return out
}

func findOne(t *testing.T, files map[string]string, prefix, suffix string) string {
	t.Helper()
	for path := range files {
		if strings.HasPrefix(path, prefix) && strings.HasSuffix(path, suffix) {
			return path
		}
	}
	return ""
}

// TestNoArgsEqualsStart: the first thing a human types is the whole command.
func TestNoArgsEqualsStart(t *testing.T) {
	implicit, explicit := t.TempDir(), t.TempDir()

	codeA, outA, errA := runDir(t, implicit, allFlags(), "")
	codeB, outB, errB := runDir(t, explicit, append([]string{"start"}, allFlags()...), "")

	if codeA != start.CodeOK || codeB != start.CodeOK {
		t.Fatalf("codes = %d and %d, want %d (stderr: %s / %s)", codeA, codeB, start.CodeOK, errA, errB)
	}
	if outA != outB {
		t.Errorf("output differs:\n%s\nvs\n%s", outA, outB)
	}

	treeA, treeB := tree(t, implicit), tree(t, explicit)
	if len(treeA) != 4 {
		t.Errorf("odachi wrote %d files, want 4: %v", len(treeA), treeA)
	}
	for path, content := range treeA {
		if treeB[path] != content {
			t.Errorf("%s differs between `odachi` and `odachi start`", path)
		}
	}
}

// TestUnknownCommandListsWhatExists: no silent fallback, and the message names
// the command that does exist.
func TestUnknownCommandListsWhatExists(t *testing.T) {
	dir := t.TempDir()

	code, out, errOut := runDir(t, dir, []string{"deploy"}, "")

	if code != start.CodeBlocked {
		t.Fatalf("code = %d, want %d", code, start.CodeBlocked)
	}
	if !strings.Contains(errOut, "start") || !strings.Contains(errOut, "comandos disponibles") {
		t.Errorf("stderr %q does not list the available commands", errOut)
	}
	if out != "" {
		t.Errorf("stdout = %q, want nothing", out)
	}
	if files := tree(t, dir); len(files) != 0 {
		t.Errorf("unknown command wrote files: %v", files)
	}
}

// TestStartJSONWritesTheFourArtifacts is the machine path end to end.
func TestStartJSONWritesTheFourArtifacts(t *testing.T) {
	dir := t.TempDir()

	code, out, errOut := runDir(t, dir, allFlags(), "")

	if code != start.CodeOK {
		t.Fatalf("code = %d, want %d (stderr: %s)", code, start.CodeOK, errOut)
	}

	var payload struct {
		Code    int      `json:"code"`
		Message string   `json:"message"`
		Profile string   `json:"profile"`
		Task    string   `json:"task"`
		Files   []string `json:"files"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out)
	}
	if payload.Code != start.CodeOK {
		t.Errorf("JSON code = %d, want %d", payload.Code, start.CodeOK)
	}
	if payload.Profile != "ci" {
		t.Errorf("JSON profile = %q, want %q (no terminal, so the machine profile)", payload.Profile, "ci")
	}
	if !strings.Contains(payload.Task, "odd/tasks/0001-quiero-reservar-turnos.md") {
		t.Errorf("JSON task = %q, want the numbered task doc", payload.Task)
	}
	if len(payload.Files) != 4 {
		t.Errorf("JSON files = %v, want the four artifacts", payload.Files)
	}

	files := tree(t, dir)
	for _, want := range []string{"brief.md", ".odachi/profile.toml", ".odachi/fence.toml"} {
		if _, ok := files[want]; !ok {
			t.Errorf("missing %s; files: %v", want, files)
		}
	}
	if findOne(t, files, "odd/tasks/0001-", ".md") == "" {
		t.Errorf("missing the task doc; files: %v", files)
	}

	loaded, err := profiletoml.Load(filepath.Join(dir, ".odachi", "profile.toml"))
	if err != nil {
		t.Fatalf("the written profile does not decode: %v", err)
	}
	if loaded.Name != "ci" {
		t.Errorf("profile.toml name = %q, want ci", loaded.Name)
	}
}

// TestForceBacksUpTheBriefAndKeepsItsBytes: the human's file is copied aside,
// and the same answers produce the same bytes.
func TestForceBacksUpTheBriefAndKeepsItsBytes(t *testing.T) {
	dir := t.TempDir()
	args := allFlags()[:len(allFlags())-1] // human output, no --json

	if code, _, errOut := runDir(t, dir, args, ""); code != start.CodeOK {
		t.Fatalf("first run code = %d (stderr: %s)", code, errOut)
	}
	first := tree(t, dir)
	firstBrief, ok := first["brief.md"]
	if !ok {
		t.Fatalf("brief.md was not written: %v", first)
	}

	if code, _, errOut := runDir(t, dir, append(args, "--force"), ""); code != start.CodeOK {
		t.Fatalf("second run code = %d (stderr: %s)", code, errOut)
	}
	second := tree(t, dir)

	backup := ""
	for path := range second {
		if strings.HasPrefix(filepath.Base(path), "brief.md.bak-") {
			backup = path
			break
		}
	}
	if backup == "" {
		t.Fatalf("brief.md was not backed up: %v", second)
	}
	if len(filepath.Base(backup)) != len("brief.md.bak-")+8 {
		t.Errorf("backup name %q is not brief.md.bak-<8 hex>", backup)
	}
	if second["brief.md"] != firstBrief {
		t.Errorf("brief.md changed with the same answers:\n%s\nvs\n%s", second["brief.md"], firstBrief)
	}
	if second[backup] != firstBrief {
		t.Errorf("backup content = %q, want the previous brief %q", second[backup], firstBrief)
	}
	taskDoc := findOne(t, first, "odd/tasks/0001-", ".md")
	if taskDoc == "" {
		t.Fatalf("first run wrote no task doc: %v", first)
	}
	if second[taskDoc] != first[taskDoc] {
		t.Errorf("the task doc was rewritten on a rerun; its steps and history are the memory of the work")
	}
	if findOne(t, second, "odd/tasks/0002-", ".md") != "" {
		t.Errorf("a rerun with the same outcome created a second task doc: %v", second)
	}
}

// TestExistingProfileWinsOverInference: the human owns .odachi/profile.toml.
func TestExistingProfileWinsOverInference(t *testing.T) {
	dir := t.TempDir()
	profileDir := filepath.Join(dir, ".odachi")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	seed := `name    = "novato"
extends = ""
`
	if err := os.WriteFile(filepath.Join(profileDir, "profile.toml"), []byte(seed), 0o644); err != nil {
		t.Fatalf("seed profile: %v", err)
	}

	if code, _, errOut := runDir(t, dir, allFlags()[:len(allFlags())-1], ""); code != start.CodeOK {
		t.Fatalf("code = %d, want %d (stderr: %s)", code, start.CodeOK, errOut)
	}

	loaded, err := profiletoml.Load(filepath.Join(profileDir, "profile.toml"))
	if err != nil {
		t.Fatalf("profile after start: %v", err)
	}
	if loaded.Name != "novato" {
		t.Errorf("profile name = %q, want novato: start must not re-infer over the human's file", loaded.Name)
	}
	if loaded.Questions != "one-by-one" {
		t.Errorf("profile questions = %q, want one-by-one", loaded.Questions)
	}
}

// TestProfileWithoutNameBlocksWithAHumanSentence covers the deferred T4 defect:
// Decode accepts a nameless file, and start is where the human is told.
func TestProfileWithoutNameBlocksWithAHumanSentence(t *testing.T) {
	dir := t.TempDir()
	profileDir := filepath.Join(dir, ".odachi")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(profileDir, "profile.toml")
	if err := os.WriteFile(path, []byte("extends = \"\"\n"), 0o644); err != nil {
		t.Fatalf("seed profile: %v", err)
	}

	code, _, errOut := runDir(t, dir, allFlags()[:len(allFlags())-1], "")

	if code != start.CodeBlocked {
		t.Fatalf("code = %d, want %d (stderr: %s)", code, start.CodeBlocked, errOut)
	}
	if !strings.Contains(errOut, "name") {
		t.Errorf("stderr %q does not tell the human the profile has no name", errOut)
	}
	unchanged, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}
	if string(unchanged) != "extends = \"\"\n" {
		t.Errorf("the profile was rewritten by a blocked run: %q", unchanged)
	}
}

// TestNoHumanAndNoFlagsExitsTwo: the message names the flags, not a stack trace.
func TestNoHumanAndNoFlagsExitsTwo(t *testing.T) {
	dir := t.TempDir()

	code, _, errOut := runDir(t, dir, []string{"start"}, "")

	if code != start.CodeNeedsAnswer {
		t.Fatalf("code = %d, want %d (stderr: %s)", code, start.CodeNeedsAnswer, errOut)
	}
	if !strings.Contains(errOut, "--outcome") {
		t.Errorf("stderr %q does not name --outcome", errOut)
	}
	if files := tree(t, dir); len(files) != 0 {
		t.Errorf("wrote files without a human: %v", files)
	}
}

// TestInvalidFlagIsBlockedInJSON keeps the machine output shape for failures.
func TestInvalidFlagIsBlockedInJSON(t *testing.T) {
	dir := t.TempDir()

	code, out, _ := runDir(t, dir, []string{"--audience", "tablet", "--json"}, "")

	if code != start.CodeBlocked {
		t.Fatalf("code = %d, want %d", code, start.CodeBlocked)
	}

	var payload struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out)
	}
	if payload.Code != start.CodeBlocked || !strings.Contains(payload.Message, "--audience") {
		t.Errorf("payload = %+v, want code 1 naming --audience", payload)
	}
	if files := tree(t, dir); len(files) != 0 {
		t.Errorf("invalid flag wrote files: %v", files)
	}
}

// TestInterviewPrintsTheWhyInTheSameBlock pins the tutor rule at the adapter:
// question, why and options are one block, and the why is one line.
func TestInterviewPrintsTheWhyInTheSameBlock(t *testing.T) {
	var out bytes.Buffer
	interview := &terminalInterview{
		in:  bufio.NewReader(strings.NewReader("1\n")),
		out: &out,
	}

	answer, err := interview.Ask(start.Question{
		Text:    "¿Quién lo usa y desde dónde? (1 celular · 2 compu · 3 los dos)",
		Why:     "cambia todo lo que viene después, y es la decisión más cara de cambiar tarde",
		Options: []string{"phone", "cpu", "both"},
	})
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if answer != "1" {
		t.Errorf("answer = %q, want %q", answer, "1")
	}

	lines := strings.Split(strings.Trim(out.String(), "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("the question block has %d lines, want 4:\n%s", len(lines), out.String())
	}
	if !strings.HasPrefix(lines[0], "? ") || !strings.Contains(lines[0], "¿Quién lo usa") {
		t.Errorf("line 1 = %q, want the question", lines[0])
	}
	if !strings.HasPrefix(lines[1], "  por qué: ") || !strings.Contains(lines[1], "decisión más cara") {
		t.Errorf("line 2 = %q, want the why on one line", lines[1])
	}
	if !strings.HasPrefix(lines[2], "  opciones: ") {
		t.Errorf("line 3 = %q, want the options", lines[2])
	}
	if lines[3] != "> " {
		t.Errorf("line 4 = %q, want the prompt", lines[3])
	}

	if err := interview.Say("una línea"); err != nil {
		t.Fatalf("Say: %v", err)
	}
	if !strings.HasSuffix(out.String(), "una línea\n") {
		t.Errorf("Say did not print one line: %q", out.String())
	}
}
