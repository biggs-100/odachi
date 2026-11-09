package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

// initRepoWithCommit is a repository that already lives: the case the human
// reported, a project with history that is adopting odachi.
func initRepoWithCommit(t *testing.T, dir string) {
	t.Helper()
	initRepo(t, dir)
	writeFile(t, dir, "README.md", "un proyecto que ya vive\n")
	for _, args := range [][]string{
		{"-C", dir, "add", "-A"},
		{"-C", dir, "-c", "user.email=t@example.com", "-c", "user.name=T", "commit", "-q", "-m", "one"},
	} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
}

// TestAdoptingAnExistingRepoKeepsTheMinimalFootprint: a project that already
// lives keeps the rules, the ignore and the hook, and none of odachi's folders.
// The craft stays in the machine, and the run says so.
func TestAdoptingAnExistingRepoKeepsTheMinimalFootprint(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepoWithCommit(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}

	for _, rel := range []string{"AGENTS.md", ".gitignore", ".git/hooks/commit-msg"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("%s was not written: %v", rel, err)
		}
	}
	for _, rel := range []string{".agents", ".opencode", ".claude"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			t.Errorf("%s was written in a project that already lives", rel)
		}
	}
	report := out + errOut
	if !strings.Contains(report, "this repository already lives") {
		t.Errorf("the run does not say why the footprint is minimal:\n%s", report)
	}
	if _, err := os.Stat(filepath.Join(home, ".agents", "skills", "acceptance", "SKILL.md")); err != nil {
		t.Errorf("the craft did not go to the machine: %v", err)
	}
	// The block a teammate reads has to name the folder this project will use.
	agents, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("reading AGENTS.md: %v", err)
	}
	if !strings.Contains(string(agents), "`~/.agents/skills/`") {
		t.Errorf("the block does not name the machine's copy:\n%s", agents)
	}
	if strings.Contains(string(agents), "`.agents/skills/`") {
		t.Errorf("the block names a folder this project will not have:\n%s", agents)
	}
	// The inference is not a choice: a store that recorded it would turn the next
	// plain init into a decision the human never made.
	store, err := os.ReadFile(storePath(dir))
	if err == nil && strings.Contains(string(store), "skills =") {
		t.Errorf("the run stored an inference:\n%s", store)
	}
}

// TestAFreshRepoKeepsTheFullInstall: a folder with no history is a project
// starting now. It keeps the rules, the ignore and the hook here; the agent and
// the craft are installed once per machine.
func TestAFreshRepoKeepsTheFullInstall(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	for _, rel := range []string{"AGENTS.md", ".gitignore", ".git/hooks/commit-msg"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("%s was not written in a project starting now: %v", rel, err)
		}
	}
	for _, rel := range []string{".agents", ".opencode", ".claude"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			t.Errorf("%s was written in the project: the agent and the craft live in the home", rel)
		}
	}
	if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(".agents/skills/acceptance/SKILL.md"))); err != nil {
		t.Errorf("a fresh project did not get the craft in the machine: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(".config/opencode/agents/odachi.md"))); err != nil {
		t.Errorf("a fresh project did not get the door in the machine: %v", err)
	}
	if strings.Contains(out+errOut, "this repository already lives") {
		t.Errorf("a folder with no history was called a project that lives:\n%s%s", out, errOut)
	}
}

// TestAProjectThatAlreadyCarriesTheCraftIsNotMoved: an inference never moves a
// project that adopted odachi the old way. The folder stays the project's, its
// out-of-date skill is refreshed like any init does (with the old bytes set
// aside), and the run does not send the craft to the machine behind the human.
func TestAProjectThatAlreadyCarriesTheCraftIsNotMoved(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepoWithCommit(t, dir)
	writeFile(t, dir, ".agents/skills/acceptance/SKILL.md", "the old copy\n")

	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	// The craft did not move: the project carries it, so the block names this
	// project's copy. The agent half is separate, and with no agent file here it
	// stays in the machine -that is the split, not a contradiction.
	body, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("reading AGENTS.md: %v", err)
	}
	if !strings.Contains(string(body), "`.agents/skills/`") {
		t.Errorf("the block does not name the project's copy:\n%s", body)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "acceptance", "SKILL.md")); err != nil {
		t.Fatalf("the project copy is gone: %v", err)
	}
	skill := filepath.Join(dir, ".agents", "skills", "acceptance", "SKILL.md")
	if _, err := os.Stat(skill); err != nil {
		t.Fatalf("the project copy is gone: %v", err)
	}
	// pack backs a replaced skill up in the clone's store, not beside the file.
	backups, err := filepath.Glob(filepath.Join(dir, ".git", "odachi", "backups", "*"))
	if err != nil || len(backups) == 0 {
		t.Errorf("the old bytes were replaced without a backup in the store: %v (%v)", backups, err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "agents", "odachi.md")); err == nil {
		t.Errorf("the agent came into the project: the two halves are separate, and this one was never asked for")
	}
}

// TestTheFlagsAskForTheCopiesAndAreStored: `--skills project` is the explicit pin
// for the craft, and the store keeps it; `--agents` names the clients of the door
// in the home. A plain init afterwards honours both answers.
func TestTheFlagsAskForTheCopiesAndAreStored(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepoWithCommit(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init", "--skills", "project", "--agents", "opencode"}, "")
	if code != 0 {
		t.Fatalf("code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "acceptance", "SKILL.md")); err != nil {
		t.Errorf("--skills project did not put the craft here: %v", err)
	}
	door, _ := persona.GlobalFilesFor(persona.SelectionOpenCode)
	if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(door.Path))); err != nil {
		t.Errorf("--agents did not put the OpenCode door in the home: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "agents", "odachi.md")); err == nil {
		t.Error("--agents put the native agent in the project: the agent lives in the home")
	}
	store, err := os.ReadFile(storePath(dir))
	if err != nil {
		t.Fatalf("the choice was not stored: %v", err)
	}
	for _, want := range []string{`skills = "project"`, `agents = "opencode"`} {
		if !strings.Contains(string(store), want) {
			t.Errorf("the store does not keep %q:\n%s", want, store)
		}
	}

	// A plain init afterwards honours them instead of going back to the defaults.
	// The project's copy is removed first, so the run can only bring it back from
	// the store: with it in place the inference would explain the work and the
	// store branch would never be exercised.
	if err := os.RemoveAll(filepath.Join(dir, ".agents")); err != nil {
		t.Fatalf("removing the project copy: %v", err)
	}

	code, out, errOut = runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("the plain init failed: %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "acceptance", "SKILL.md")); err != nil {
		t.Errorf("the store did not bring the craft back: %v", err)
	}
	if !strings.Contains(out+errOut, "from the store") {
		t.Errorf("the run does not say the store decided:\n%s%s", out, errOut)
	}
	if strings.Contains(out+errOut, "this repository already lives") {
		t.Errorf("the plain init ignored the stored choice:\n%s%s", out, errOut)
	}
}

// TestNamingAClientDoesNotDecideWhereTheCraftGoes: the two choices are separate.
// --agents says which clients receive the door in the home; it says nothing about
// the craft, so the store must not record a scope for it and the next plain init
// still leaves the craft in the machine.
func TestNamingAClientDoesNotDecideWhereTheCraftGoes(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepoWithCommit(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init", "--agents", "opencode"}, "")
	if code != 0 {
		t.Fatalf("code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	door, _ := persona.GlobalFilesFor(persona.SelectionOpenCode)
	if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(door.Path))); err != nil {
		t.Errorf("--agents did not put the OpenCode door in the home: %v", err)
	}
	store, err := os.ReadFile(storePath(dir))
	if err != nil {
		t.Fatalf("the client choice was not stored: %v", err)
	}
	if !strings.Contains(string(store), `agents = "opencode"`) {
		t.Errorf("the store does not keep the client choice:\n%s", store)
	}
	if strings.Contains(string(store), "skills =") {
		t.Errorf("naming a client recorded a scope nobody chose:\n%s", store)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents")); err == nil {
		t.Errorf("--agents put the craft in the project")
	}

	code, out, errOut = runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("the plain init failed: %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(out+errOut, "this repository already lives") {
		t.Errorf("the craft moved into the project because a client was named:\n%s%s", out, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents")); err == nil {
		t.Errorf("the plain init put the craft in the project")
	}
}

// TestAProjectWithOnlyItsAgentKeepsItAndTheRunNamesTheLeftover: an old install
// can leave the native agent in the project and the craft in the machine. This
// version writes no project agent and deletes nothing: the file keeps its bytes,
// and the run names the folder so the human decides.
func TestAProjectWithOnlyItsAgentKeepsItAndTheRunNamesTheLeftover(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepoWithCommit(t, dir)
	writeFile(t, dir, ".opencode/agents/odachi.md", "a door from an earlier release\n")

	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	report := out + errOut
	if !strings.Contains(report, "the agent lives only in the home") {
		t.Errorf("the run does not say where the agent lives:\n%s", report)
	}
	if !strings.Contains(report, "left from an older release") || !strings.Contains(report, ".opencode/agents/odachi.md") {
		t.Errorf("the run does not name the leftover file:\n%s", report)
	}
	// The folder holds a role's agent too; the line may not offer the folder for
	// deletion, only the file odachi wrote.
	if strings.Contains(report, "left from an older release: .opencode/agents;") {
		t.Errorf("the run names the folder instead of the file:\n%s", report)
	}
	got, err := os.ReadFile(filepath.Join(dir, ".opencode", "agents", "odachi.md"))
	if err != nil {
		t.Fatalf("the project agent is gone: %v", err)
	}
	if string(got) != "a door from an earlier release\n" {
		t.Errorf("the project agent was rewritten: %q", got)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents")); err == nil {
		t.Errorf("the craft came into the project because the agent was there")
	}
}

// TestAnAgentClientIsNotOdachisAgent: a repository that lives and already uses
// OpenCode or Claude has their folders for its own settings. The signal for
// "odachi's agent is here" is the file the client reads, never the folder: an
// ordinary adopting repo must get no mirror written into it.
func TestAnAgentClientIsNotOdachisAgent(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepoWithCommit(t, dir)
	writeFile(t, dir, ".opencode/settings.json", "{}")
	writeFile(t, dir, ".claude/settings.json", "{}")

	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	for _, rel := range []string{".opencode/agents/odachi.md", ".claude/output-styles/odachi.md"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err == nil {
			t.Errorf("%s was written: the client folder is not odachi's agent", rel)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents")); err == nil {
		t.Errorf("the craft came into the project")
	}
}

// TestAnotherToolsFolderIsNotTheCraft: `.agents/` belongs to other tools too. A
// repository that keeps its own notes there has no copy of the craft, so the
// minimal footprint still applies and nothing is written into `.agents/skills`.
func TestAnotherToolsFolderIsNotTheCraft(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepoWithCommit(t, dir)
	writeFile(t, dir, ".agents/notes.md", "our own notes\n")

	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills")); err == nil {
		t.Errorf("the craft came into a project that only had a folder named .agents:\n%s%s", out, errOut)
	}
	if _, err := os.Stat(filepath.Join(home, ".agents", "skills", "acceptance", "SKILL.md")); err != nil {
		t.Errorf("the craft did not go to the machine: %v", err)
	}
}

// TestTheHomeOnlyAgentIsRecordedEvenWhenOnlyTheCraftWasNamed: a project that chose
// where the craft goes has not named a client, but it still gets the agent from the
// home. The CLI records that answer every run, so the next command cannot rebuild a
// mirror nobody asked for.
func TestTheHomeOnlyAgentIsRecordedEvenWhenOnlyTheCraftWasNamed(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepoWithCommit(t, dir)

	if code, _, errOut := runDir(t, dir, []string{"init", "--skills", "home"}, ""); code != 0 {
		t.Fatalf("the first init failed: %s", errOut)
	}
	code, out, errOut := runDir(t, dir, []string{"init"}, "")
	if code != 0 {
		t.Fatalf("the plain init failed: %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	store, err := os.ReadFile(storePath(dir))
	if err != nil {
		t.Fatalf("reading the store: %v", err)
	}
	if !strings.Contains(string(store), `project_agents = "none"`) {
		t.Errorf("the store does not keep the home-only agent:\n%s", store)
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "agents", "odachi.md")); err == nil {
		t.Errorf("the mirrors came into the project without a flag asking for them")
	}
}
