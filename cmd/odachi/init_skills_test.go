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
	"github.com/biggs-100/odachi/internal/domain/persona"
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

// TestInitInstallsTheSkillsInOneCommand: `odachi init` leaves the machine ready.
// Since 0242 the craft is one copy per machine, so every bundled skill goes to
// ~/.agents/skills and the project carries no copy.
func TestInitInstallsTheSkillsInOneCommand(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"init"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}

	all, err := bundledskills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	if len(all) == 0 {
		t.Fatalf("the payload came back empty: the machine would receive no skill")
	}
	for _, skill := range all {
		got, err := os.ReadFile(skillPath(home, skill.Name))
		if err != nil {
			t.Errorf("%s was not installed in ~/.agents/skills: %v", skill.Name, err)
			continue
		}
		if !bytes.Equal(got, skill.Main()) {
			t.Errorf("%s difiere de skills/%s/SKILL.md", skill.Name, skill.Name)
		}
	}

	// The project carries no copy: the machine is the one place the craft lives.
	if _, err := os.Stat(filepath.Join(dir, ".agents")); err == nil {
		t.Errorf("a plain init wrote .agents/ in the project")
	}
	for _, rel := range []string{".opencode/skills", ".opencode/skill", ".claude/skills"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err == nil {
			t.Errorf("init installed skills in %s: the project uses no folder", rel)
		}
	}

	if !strings.Contains(out, "~/.agents/skills") {
		t.Errorf("the output does not name the destination:\n%s", out)
	}
	if !strings.Contains(out, "skills: home (the default)") {
		t.Errorf("the output does not say where the default lives:\n%s", out)
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
		if !bytes.Equal(got, skill.Main()) {
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
// form: init writes outside the project, so the home it touches has to be
// exactly the payload, the door of each chosen client, the subagents, the
// OpenCode config and the design plugin, and nothing else.
func TestInitWritesOnlyTheExpectedThingsIntoAnIsolatedHome(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()
	initRepo(t, dir)

	if code, out, errOut := runDir(t, dir, []string{"init"}, ""); code != 0 {
		t.Fatalf("code = %d (stdout %q, stderr %q)", code, out, errOut)
	}

	want := []string{
		".claude/output-styles/odachi.md",
		".config/opencode/agents/odachi.md",
		".config/opencode/opencode.json",
		".config/opencode/plugins/odachi-design/index.ts",
	}
	all, err := bundledskills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	for _, skill := range all {
		for _, file := range skill.Files {
			want = append(want, ".agents/skills/"+skill.Name+"/"+file.Path)
		}
	}
	for _, client := range persona.Clients() {
		for _, file := range persona.GlobalSubagentFilesFor(client.Agent) {
			want = append(want, file.Path)
		}
	}
	slices.Sort(want)
	if got := homeTree(t, home); !slices.Equal(got, want) {
		t.Fatalf("the home has %v, want exactly %v", got, want)
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
	if len(payload.Skills) != 1 {
		t.Fatalf("JSON skills = %+v, want one destination: the machine", payload.Skills)
	}

	all, err := bundledskills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	want := []string{"~/.agents/skills"}
	installed := []int{len(all)}
	for i, dest := range payload.Skills {
		if dest.Dest != want[i] {
			t.Errorf("skills[%d].dest = %q, want %q", i, dest.Dest, want[i])
		}
		if dest.Installed != installed[i] {
			t.Errorf("skills[%d].installed = %d, want %d", i, dest.Installed, installed[i])
		}
	}

	joined := strings.Join(payload.Files, " ")
	for _, path := range []string{"~/.agents/skills/acceptance/SKILL.md", "~/.agents/skills/bootstrap/SKILL.md"} {
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
	// installs the craft in the machine, and never in the project.
	if _, err := os.Stat(skillPath(home, "acceptance")); err != nil {
		t.Errorf("init did not install the curriculum in ~/.agents/skills: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents")); err == nil {
		t.Errorf("init installed the curriculum in the project")
	}
}

// TestInitKeepsTheProjectArtifactsWhenTheHomeIsNotWritable: the home is where the
// agent and the craft live, so a home that cannot be written is a failed run. The
// rules and the hook are already on disk, and the line names both facts.
func TestInitKeepsTheProjectArtifactsWhenTheHomeIsNotWritable(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	// A home that is a file: nothing can be created inside it, on any platform.
	parent := t.TempDir()
	home := filepath.Join(parent, "home")
	writeFile(t, parent, "home", "no soy una carpeta\n")
	withHome(t, home)

	code, out, errOut := runDir(t, dir, []string{"init"}, "")

	if code == 0 {
		t.Fatalf("code = 0, want a refusal: the only place the agent lives could not be written")
	}
	for _, rel := range []string{"AGENTS.md", ".gitignore", ".git/hooks/commit-msg"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("%s was not written: %v", rel, err)
		}
	}
	if !strings.Contains(out+errOut, "~/.agents/skills") {
		t.Errorf("the output does not say what is still pending:\n%s", out+errOut)
	}
	if !strings.Contains(out+errOut, "odachi init") {
		t.Errorf("the output does not name the command that unblocks it:\n%s", out+errOut)
	}
}
