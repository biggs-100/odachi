package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/start"
	bundledskills "github.com/biggs-100/odachi/skills"
)

// gitTemplate is the `.git` directory TestMain prepares: a repository without a
// process per test.
var gitTemplate string

// The curriculum grows (drive) and the usage cannot pin a number by hand: the
// help names the folder, not how many go there. A number here goes stale in the
// next version and nobody notices until a human reads it.
func TestInitUsageDoesNotHardCodeTheSkillCount(t *testing.T) {
	dir := configuredProject(t)

	out, _ := output(t, dir, []string{"init", "-h"})
	if !strings.Contains(out, ".agents/skills") {
		t.Fatalf("the usage does not name .agents/skills:\n%s", out)
	}
	if hardCoded := regexp.MustCompile(`\b\d+ skills\b`); hardCoded.MatchString(out) {
		t.Fatalf("the usage hard-codes how many skills go: %q", hardCoded.FindString(out))
	}
}

// TestMain points the home adapter at a directory this process owns before any
// test runs. `odachi init` writes the bootstrap skill into the human's home now,
// and a suite that reaches the real home of whoever runs it has escaped the
// repository. A test that needs its own home still overrides it with withHome.
//
// It also prepares one empty repository template: `git init` costs ~80ms per
// process on Windows, and the e2e suite creates a repository per test. Copying
// the template is the same repository without the process.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "odachi-home-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not create a test home: "+err.Error())
		os.Exit(1)
	}
	previous := homeDir
	homeDir = func() (string, error) { return home, nil }

	template, err := os.MkdirTemp("", "odachi-git-template-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not create a git template: "+err.Error())
		os.Exit(1)
	}
	if out, err := exec.Command("git", "-C", template, "init", "-q").CombinedOutput(); err != nil {
		fmt.Fprintln(os.Stderr, "could not prepare the git template: "+err.Error()+": "+string(out))
		os.Exit(1)
	}
	gitTemplate = filepath.Join(template, ".git")

	code := m.Run()

	homeDir = previous
	os.RemoveAll(home)
	os.RemoveAll(template)
	os.Exit(code)
}

// homeTree lists every file under a home as a slash path, sorted: the fence is
// easier to read as the whole list than as one missing file.
func homeTree(t *testing.T, home string) []string {
	t.Helper()
	paths := []string{}
	err := filepath.WalkDir(home, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(home, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", home, err)
	}
	slices.Sort(paths)
	return paths
}

// skillPath is where one bundled skill lands under a root: the curriculum in the
// project's cross-client directory, the bootstrap one in the home.
func skillPath(root, name string) string {
	return filepath.Join(root, ".agents", "skills", name, "SKILL.md")
}

// TestInitInstallsTheSkillsInOneCommand: `odachi init` leaves the project ready.
// The thirteen curriculum skills go to the cross-client project location and the
// bootstrap skill goes to the home, so an agent learns how to start a project
// before a project exists. One command, not two.
func TestInitInstallsTheSkillsInOneCommand(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}

	curriculum, err := bundledskills.Curriculum()
	if err != nil {
		t.Fatalf("Curriculum(): %v", err)
	}
	if len(curriculum) == 0 {
		t.Fatalf("the curriculum came back empty: the project would receive no skill")
	}
	for _, skill := range curriculum {
		got, err := os.ReadFile(skillPath(dir, skill.Name))
		if err != nil {
			t.Errorf("%s was not installed in .agents/skills: %v", skill.Name, err)
			continue
		}
		if !bytes.Equal(got, skill.Content) {
			t.Errorf("%s difiere de skills/%s/SKILL.md", skill.Name, skill.Name)
		}
	}

	// The project has one skills location, and the bootstrap skill is not part
	// of the craft a project carries.
	if _, err := os.Stat(skillPath(dir, bundledskills.BootstrapName)); err == nil {
		t.Errorf("the project received the bootstrap skill: that one lives in the home")
	}
	// The client directories hold the native agent, never a second copy of the
	// skills: the project's one location is the cross-client folder.
	for _, rel := range []string{".opencode/skills", ".opencode/skill", ".claude/skills"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err == nil {
			t.Errorf("init installed skills in %s: the project uses a single folder", rel)
		}
	}

	bootstrap, ok := bundledskills.Bootstrap()
	if !ok {
		t.Fatalf("the binary does not carry the %q skill", bundledskills.BootstrapName)
	}
	got, err := os.ReadFile(skillPath(home, bootstrap.Name))
	if err != nil {
		t.Fatalf("the bootstrap skill did not reach the home: %v", err)
	}
	if !bytes.Equal(got, bootstrap.Content) {
		t.Errorf("the home skill is not the binary's one")
	}

	if !strings.Contains(out, "./.agents/skills") || !strings.Contains(out, "~/.agents/skills") {
		t.Errorf("the output does not name both destinations:\n%s", out)
	}
}

// TestInitInstallsAllSkillsInTheHomeWhenAsked: --skills home is the one-copy
// option. The project gets no .agents/skills, the home gets the whole payload,
// and a second run writes nothing.
func TestInitInstallsAllSkillsInTheHomeWhenAsked(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init", "--skills", "home"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}

	all, err := bundledskills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	for _, skill := range all {
		got, err := os.ReadFile(skillPath(home, skill.Name))
		if err != nil {
			t.Errorf("%s did not reach the home: %v", skill.Name, err)
			continue
		}
		if !bytes.Equal(got, skill.Content) {
			t.Errorf("%s differs from the payload", skill.Name)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents")); err == nil {
		t.Errorf("--skills home wrote .agents/ in the project")
	}
	if !strings.Contains(out, "~/.agents/skills") {
		t.Errorf("the output does not name the home destination:\n%s", out)
	}

	before := homeTree(t, home)
	code, out, errOut = runDir(t, dir, []string{"init", "--skills", "home"}, "")
	if code != 0 {
		t.Fatalf("second run code = %d (stdout %q, stderr %q)", code, out, errOut)
	}
	if !slices.Equal(homeTree(t, home), before) {
		t.Errorf("the second run touched the home: %v vs %v", homeTree(t, home), before)
	}
	if !strings.Contains(out, "unchanged") {
		t.Errorf("the second run does not say nothing changed:\n%s", out)
	}
}

// TestInitHomeScopeKeepsTheProjectCopy: the option changes where init writes,
// not what it deletes. A .agents/skills from an earlier run is the human's now.
func TestInitHomeScopeKeepsTheProjectCopy(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, dir, ".agents/skills/acceptance/SKILL.md", "viejo\n")

	code, out, errOut := runDir(t, dir, []string{"init", "--skills", "home"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}
	for _, want := range []string{"it already had a copy", "~/.agents/skills"} {
		if !strings.Contains(out, want) {
			t.Errorf("the output does not say %q:\n%s", want, out)
		}
	}
	got, err := os.ReadFile(filepath.Join(dir, ".agents", "skills", "acceptance", "SKILL.md"))
	if err != nil {
		t.Fatalf("the project copy was deleted: %v", err)
	}
	if string(got) != "viejo\n" {
		t.Errorf("the project copy changed: %q", got)
	}
}

// TestInitRefusesAnUnknownSkillsScope: a typo is stopped before any write, and
// the message names both values.
func TestInitRefusesAnUnknownSkillsScope(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	code, out, errOut := runDir(t, dir, []string{"init", "--skills", "fuera"}, "")

	if code != start.CodeBlocked {
		t.Fatalf("code = %d, want %d (stdout %q, stderr %q)", code, start.CodeBlocked, out, errOut)
	}
	for _, want := range []string{"project", "home", "fuera"} {
		if !strings.Contains(out+errOut, want) {
			t.Errorf("the output does not say %q: %s%s", want, out, errOut)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err == nil {
		t.Errorf("an invalid scope wrote AGENTS.md")
	}
}

// TestInitWritesOnlyTheExpectedThingsIntoAnIsolatedHome is the fence in test
// form: init writes outside the project now, so the home it touches has to be
// exactly the bootstrap skill, and nothing else. The door of the home is opt-in,
// so a plain init never lands there.
func TestInitWritesOnlyTheExpectedThingsIntoAnIsolatedHome(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	if code, out, errOut := runDir(t, dir, []string{"init"}, ""); code != 0 {
		t.Fatalf("code = %d (stdout %q, stderr %q)", code, out, errOut)
	}

	want := []string{
		".agents/skills/bootstrap/SKILL.md",
	}
	if got := homeTree(t, home); !slices.Equal(got, want) {
		t.Fatalf("the home has %v, want exactly %v", got, want)
	}

	// The same fence with the flag: --global is what adds the two doors, one per
	// chosen client, and the OpenCode config that makes one of them the default.
	withGlobal := t.TempDir()
	withHome(t, withGlobal)
	if code, out, errOut := runDir(t, dir, []string{"init", "--global"}, ""); code != 0 {
		t.Fatalf("--global: code = %d (stdout %q, stderr %q)", code, out, errOut)
	}
	wantGlobal := []string{
		".agents/skills/bootstrap/SKILL.md",
		".claude/output-styles/odachi.md",
		".config/opencode/agents/odachi.md",
		".config/opencode/opencode.json",
	}
	if got := homeTree(t, withGlobal); !slices.Equal(got, wantGlobal) {
		t.Fatalf("the home with --global has %v, want exactly %v", got, wantGlobal)
	}
}

// TestInitSecondRunWritesNothing: a rerun changes no bytes anywhere, skills
// included, and says so.
func TestInitSecondRunWritesNothing(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	if code, _, errOut := runDir(t, dir, []string{"init"}, ""); code != 0 {
		t.Fatalf("first run code = %d (stderr %q)", code, errOut)
	}
	agentsBefore, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	homeBefore := homeTree(t, home)

	code, out, errOut := runDir(t, dir, []string{"init"}, "")

	if code != 0 {
		t.Fatalf("second run code = %d (stdout %q, stderr %q)", code, out, errOut)
	}
	if !strings.Contains(out, "unchanged") {
		t.Errorf("the second run does not say nothing changed:\n%s", out)
	}
	agentsAfter, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if !bytes.Equal(agentsAfter, agentsBefore) {
		t.Errorf("AGENTS.md changed on a second run")
	}
	if got := homeTree(t, home); !slices.Equal(got, homeBefore) {
		t.Errorf("the home changed on a second run: %v, before %v", got, homeBefore)
	}
	// No backup of a skill either: identical bytes are never rewritten.
	backups, err := filepath.Glob(filepath.Join(home, ".agents", "skills", bundledskills.BootstrapName, "SKILL.md.bak-*"))
	if err != nil || len(backups) != 0 {
		t.Errorf("home backups = %v (err %v), want none", backups, err)
	}
}

// TestInitJSONCarriesTheSkillDestinations: machine output has to carry the same
// destinations a human reads, or a caller cannot tell what the run installed.
func TestInitJSONCarriesTheSkillDestinations(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init", "--json"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stderr %q)", code, errOut)
	}

	var payload struct {
		Code   int      `json:"code"`
		Files  []string `json:"files"`
		Skills []struct {
			Dest      string   `json:"dest"`
			Installed int      `json:"installed"`
			Unchanged int      `json:"unchanged"`
			Files     []string `json:"files"`
		} `json:"skills"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out)
	}
	if payload.Code != 0 {
		t.Errorf("JSON code = %d, want 0", payload.Code)
	}
	if len(payload.Skills) != 2 {
		t.Fatalf("JSON skills = %+v, want two destinations", payload.Skills)
	}

	curriculum, err := bundledskills.Curriculum()
	if err != nil {
		t.Fatalf("Curriculum(): %v", err)
	}
	want := []string{"./.agents/skills", "~/.agents/skills"}
	installed := []int{len(curriculum), 1}
	for i, dest := range payload.Skills {
		if dest.Dest != want[i] {
			t.Errorf("skills[%d].dest = %q, want %q", i, dest.Dest, want[i])
		}
		if dest.Installed != installed[i] {
			t.Errorf("skills[%d].installed = %d, want %d", i, dest.Installed, installed[i])
		}
	}

	joined := strings.Join(payload.Files, " ")
	for _, path := range []string{".agents/skills/acceptance/SKILL.md", "~/.agents/skills/bootstrap/SKILL.md"} {
		if !strings.Contains(joined, path) {
			t.Errorf("JSON files do not list %s: %v", path, payload.Files)
		}
	}
}

// TestInitAcknowledgesTheOpenCodeCopy: a project where `pack --agents opencode`
// already ran has a second copy of the curriculum. init keeps one project
// location and says the other one is there instead of pretending it is not.
func TestInitAcknowledgesTheOpenCodeCopy(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	if code, out, errOut := runDir(t, dir, []string{"pack", "--agents", "opencode"}, ""); code != 0 {
		t.Fatalf("pack code = %d (stdout %q, stderr %q)", code, out, errOut)
	}

	code, out, errOut := runDir(t, dir, []string{"init"}, "")

	if code != 0 {
		t.Fatalf("init code = %d (stdout %q, stderr %q)", code, out, errOut)
	}
	if !strings.Contains(out, ".opencode/skills") {
		t.Errorf("the output does not mention the .opencode/skills copy:\n%s", out)
	}
	if !strings.Contains(out, "odachi pack --agents opencode") {
		t.Errorf("the output does not say where that copy came from:\n%s", out)
	}
	// The copy that was already there is not touched and not duplicated: init
	// still installs the project under .agents/skills only.
	if _, err := os.Stat(skillPath(dir, "acceptance")); err != nil {
		t.Errorf("init did not install the curriculum in .agents/skills: %v", err)
	}
}

// TestInitKeepsTheProjectWhenTheHomeIsNotWritable: the home is the only part of
// the job outside the project, and a home that cannot be written must not cost
// the human the project. The line names the block and the command that fixes it.
func TestInitKeepsTheProjectWhenTheHomeIsNotWritable(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	// A home that is a file: nothing can be created inside it, on any platform.
	parent := t.TempDir()
	home := filepath.Join(parent, "home")
	writeFile(t, parent, "home", "no soy una carpeta\n")
	withHome(t, home)

	code, out, errOut := runDir(t, dir, []string{"init"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0: the project half was done (stdout %q, stderr %q)", code, out, errOut)
	}
	for _, rel := range []string{"AGENTS.md", ".gitignore", ".agents/skills/acceptance/SKILL.md"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("%s was not written: %v", rel, err)
		}
	}
	if !strings.Contains(out, "~/.agents/skills") {
		t.Errorf("the output does not say what is still pending:\n%s", out)
	}
	if !strings.Contains(out, "odachi init") {
		t.Errorf("the output does not name the command that unblocks it:\n%s", out)
	}
}
