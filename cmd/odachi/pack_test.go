package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/agent"
)

// withHome points the Home adapter at a directory the test owns, so a pack run
// never reads or writes the real home of whoever runs the suite.
func withHome(t *testing.T, dir string) {
	t.Helper()
	previous := homeDir
	homeDir = func() (string, error) { return dir, nil }
	t.Cleanup(func() { homeDir = previous })
}

// repoSkill is the exact bytes of a bundled skill in this repository: the embed
// carries these files, so this is what the binary must write.
func repoSkill(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), "skills", name, "SKILL.md"))
	if err != nil {
		t.Fatalf("read skills/%s/SKILL.md: %v", name, err)
	}
	return data
}

// agentFile is where one agent reads one skill, joined the way the app joins it.
func agentFile(t *testing.T, home, agentName, skill string) string {
	t.Helper()
	a, err := agent.Resolve(agentName)
	if err != nil {
		t.Fatalf("Resolve(%q): %v", agentName, err)
	}
	return filepath.Join(home, filepath.FromSlash(a.Dest), skill, "SKILL.md")
}

// walkNames lists every file under dir as a slash path, keyed to its content.
func walkNames(t *testing.T, dir string) map[string]string {
	t.Helper()
	found := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		found[filepath.ToSlash(rel)] = string(data)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	return found
}

// TestPackWritesTheBundledSkills: the binary carries this repository's
// skills, and one run puts their exact bytes where the agent looks.
func TestPackWritesTheBundledSkills(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)

	code, out, errOut := runDir(t, t.TempDir(), []string{"pack", "--agents", "claude"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}
	for _, name := range bundled {
		got, err := os.ReadFile(agentFile(t, home, "claude", name))
		if err != nil {
			t.Errorf("%s was not installed: %v", name, err)
			continue
		}
		if string(got) != string(repoSkill(t, name)) {
			t.Errorf("%s differs from skills/%s/SKILL.md", name, name)
		}
	}
	if !strings.Contains(out, fmt.Sprintf("claude · %d installed · ~/.claude/skills", len(bundled))) {
		t.Errorf("stdout lacks the summary line:\n%s", out)
	}
}

// TestPackSecondRunIsIdentical: a repeated run writes nothing and leaves no
// backup behind, so `pack` is safe to run after every pull.
func TestPackSecondRunIsIdentical(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()

	if code, _, errOut := runDir(t, dir, []string{"pack", "--agents", "claude"}, ""); code != 0 {
		t.Fatalf("first run code = %d (stderr %q)", code, errOut)
	}
	before := walkNames(t, home)

	code, out, errOut := runDir(t, dir, []string{"pack", "--agents", "claude"}, "")
	if code != 0 {
		t.Fatalf("second run code = %d (stderr %q)", code, errOut)
	}
	if !strings.Contains(out, fmt.Sprintf("claude · %d unchanged · ~/.claude/skills", len(bundled))) {
		t.Errorf("the second run does not report sin cambios:\n%s", out)
	}

	after := walkNames(t, home)
	if len(after) != len(before) {
		t.Errorf("the second run changed the tree: before %d files, after %d", len(before), len(after))
	}
	for path, content := range before {
		if after[path] != content {
			t.Errorf("%s changed on an identical run", path)
		}
		if strings.Contains(path, ".bak-") {
			t.Errorf("an identical run left a backup: %s", path)
		}
	}
}

// TestPackBacksUpAModifiedFile: the human's bytes are saved before they are
// replaced, and the file ends up with the bundled version.
func TestPackBacksUpAModifiedFile(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)

	path := agentFile(t, home, "claude", "acceptance")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("editado a mano\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	code, out, errOut := runDir(t, t.TempDir(), []string{"pack", "--agents", "claude"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stderr %q)", code, errOut)
	}
	if !strings.Contains(out, "1 backed up") {
		t.Errorf("stdout does not report the backup:\n%s", out)
	}

	backups, err := filepath.Glob(path + ".bak-*")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("backups = %v, want exactly one", backups)
	}
	saved, err := os.ReadFile(backups[0])
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(saved) != "editado a mano\n" {
		t.Errorf("the backup does not hold the human's bytes: %q", saved)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read installed file: %v", err)
	}
	if string(got) != string(repoSkill(t, "acceptance")) {
		t.Errorf("the file was not replaced with the bundled bytes")
	}
}

// TestPackDryRunWritesNothing: the whole home stays empty, and the report still
// says what would happen.
func TestPackDryRunWritesNothing(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)

	code, out, errOut := runDir(t, t.TempDir(), []string{"pack", "--agents", "claude", "--dry-run"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stderr %q)", code, errOut)
	}
	if files := walkNames(t, home); len(files) != 0 {
		t.Errorf("dry-run wrote %d files: %v", len(files), files)
	}
	if !strings.Contains(out, fmt.Sprintf("%d to install", len(bundled))) {
		t.Errorf("stdout does not list the planned installs:\n%s", out)
	}
	if !strings.Contains(out, "dry-run") {
		t.Errorf("stdout does not say it was a dry run:\n%s", out)
	}
}

// TestPackRefusesPiAndUnknownAgents: both are exit 1, and both name the way out
// instead of silently skipping.
func TestPackRefusesPiAndUnknownAgents(t *testing.T) {
	cases := []struct {
		agents string
		want   string
	}{
		{"pi", "Pi does not use a skills directory"},
		{"inventado", "unknown agent: inventado"},
	}
	for _, tc := range cases {
		t.Run(tc.agents, func(t *testing.T) {
			home := t.TempDir()
			withHome(t, home)

			code, out, errOut := runDir(t, t.TempDir(), []string{"pack", "--agents", tc.agents}, "")

			if code != 1 {
				t.Fatalf("code = %d, want 1 (stdout %q)", code, out)
			}
			if !strings.Contains(errOut, tc.want) {
				t.Errorf("stderr does not say %q:\n%s", tc.want, errOut)
			}
			for _, name := range agent.Names() {
				if !strings.Contains(errOut, name) {
					t.Errorf("stderr does not name %q as supported:\n%s", name, errOut)
				}
			}
			if files := walkNames(t, home); len(files) != 0 {
				t.Errorf("a refused agent wrote %d files: %v", len(files), files)
			}
		})
	}
}

// TestPackTwoAgents: one run serves both, in one summary line each.
func TestPackTwoAgents(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)

	code, out, errOut := runDir(t, t.TempDir(), []string{"pack", "--agents", "claude,codex"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stderr %q)", code, errOut)
	}
	for _, agentName := range []string{"claude", "codex"} {
		for _, name := range bundled {
			if _, err := os.Stat(agentFile(t, home, agentName, name)); err != nil {
				t.Errorf("%s: %s was not installed: %v", agentName, name, err)
			}
		}
		if !strings.Contains(out, fmt.Sprintf("%s · %d installed · ~/.%s/skills", agentName, len(bundled), agentName)) {
			t.Errorf("stdout lacks the summary of %s:\n%s", agentName, out)
		}
	}
}

// agentProjectFile is where a project-scoped agent reads one skill: the project
// root joined with the table's destination.
func agentProjectFile(t *testing.T, project, agentName, skill string) string {
	t.Helper()
	a, err := agent.Resolve(agentName)
	if err != nil {
		t.Fatalf("Resolve(%q): %v", agentName, err)
	}
	if a.Scope != agent.ScopeProject {
		t.Fatalf("%s is not project-scoped", agentName)
	}
	return filepath.Join(project, filepath.FromSlash(a.Dest), skill, "SKILL.md")
}

// TestPackOpenCodeWritesIntoTheProject: the project path is the one OpenCode
// reads, so it is the one `pack` writes — and the home stays byte-for-byte
// empty, which is the half of the promise a home-based test could not show.
func TestPackOpenCodeWritesIntoTheProject(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	project := t.TempDir()

	code, out, errOut := runDir(t, project, []string{"pack", "--agents", "opencode"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}
	for _, name := range bundled {
		got, err := os.ReadFile(agentProjectFile(t, project, "opencode", name))
		if err != nil {
			t.Errorf("%s was not installed in the project: %v", name, err)
			continue
		}
		if string(got) != string(repoSkill(t, name)) {
			t.Errorf("%s differs from skills/%s/SKILL.md", name, name)
		}
	}
	if files := walkNames(t, home); len(files) != 0 {
		t.Errorf("the home was touched: %v", files)
	}
	if !strings.Contains(out, fmt.Sprintf("opencode · %d installed · ./.opencode/skills", len(bundled))) {
		t.Errorf("stdout lacks the project-relative summary line:\n%s", out)
	}
	if strings.Contains(out, project) || strings.Contains(out, home) {
		t.Errorf("stdout prints an absolute path:\n%s", out)
	}
}

// TestPackOpenCodeSecondRunIsIdentical: the write rule does not change with the
// scope — a repeated run says `sin cambios` and leaves no backup behind.
func TestPackOpenCodeSecondRunIsIdentical(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	project := t.TempDir()

	if code, _, errOut := runDir(t, project, []string{"pack", "--agents", "opencode"}, ""); code != 0 {
		t.Fatalf("first run code = %d (stderr %q)", code, errOut)
	}
	before := walkNames(t, project)

	code, out, errOut := runDir(t, project, []string{"pack", "--agents", "opencode"}, "")
	if code != 0 {
		t.Fatalf("second run code = %d (stderr %q)", code, errOut)
	}
	if !strings.Contains(out, fmt.Sprintf("opencode · %d unchanged · ./.opencode/skills", len(bundled))) {
		t.Errorf("the second run does not report sin cambios:\n%s", out)
	}
	after := walkNames(t, project)
	if len(after) != len(before) {
		t.Errorf("the second run changed the tree: before %d files, after %d", len(before), len(after))
	}
	for path, content := range before {
		if after[path] != content {
			t.Errorf("%s changed on an identical run", path)
		}
		if strings.Contains(path, ".bak-") {
			t.Errorf("an identical run left a backup: %s", path)
		}
	}
}

// TestPackOpenCodeWorksWithoutTheDirectory: naming the agent is enough even when
// `.opencode/` does not exist yet — that directory is what this command is
// creating.
func TestPackOpenCodeWorksWithoutTheDirectory(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	project := t.TempDir()

	if _, err := os.Stat(filepath.Join(project, ".opencode")); !os.IsNotExist(err) {
		t.Fatalf(".opencode already exists in the project: %v", err)
	}

	code, out, errOut := runDir(t, project, []string{"pack", "--agents", "opencode"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}
	for _, name := range bundled {
		if _, err := os.Stat(agentProjectFile(t, project, "opencode", name)); err != nil {
			t.Errorf("%s was not installed: %v", name, err)
		}
	}
}

// TestPackDetectsOpenCodeInTheProject: with no --agents, `.opencode/` in the
// project is the signal, and the home is still untouched.
func TestPackDetectsOpenCodeInTheProject(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	project := t.TempDir()
	if err := os.MkdirAll(filepath.Join(project, ".opencode"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	code, out, errOut := runDir(t, project, []string{"pack"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}
	if !strings.Contains(out, fmt.Sprintf("opencode · %d installed · ./.opencode/skills", len(bundled))) {
		t.Errorf("the detected agent was not served:\n%s", out)
	}
	if _, err := os.Stat(agentProjectFile(t, project, "opencode", "acceptance")); err != nil {
		t.Errorf("skills were not installed: %v", err)
	}
	if files := walkNames(t, home); len(files) != 0 {
		t.Errorf("detection wrote into the home: %v", files)
	}
}

// TestPackJSON is the machine shape: the same numbers the summary line shows.
func TestPackJSON(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)

	code, out, errOut := runDir(t, t.TempDir(), []string{"pack", "--agents", "claude", "--json"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stderr %q)", code, errOut)
	}

	var payload struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		DryRun  bool   `json:"dry_run"`
		Files   []string
		Dests   []struct {
			Agents    []string `json:"agents"`
			Dest      string   `json:"dest"`
			Scope     string   `json:"scope"`
			Installed int      `json:"installed"`
			Failed    int      `json:"failed"`
		} `json:"dests"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out)
	}
	if payload.Code != 0 {
		t.Errorf("JSON code = %d, want 0", payload.Code)
	}
	if payload.DryRun {
		t.Errorf("JSON says dry_run for a real run")
	}
	if len(payload.Dests) != 1 || len(payload.Dests[0].Agents) != 1 || payload.Dests[0].Agents[0] != "claude" {
		t.Fatalf("JSON dests = %+v, want one claude", payload.Dests)
	}
	if payload.Dests[0].Scope != string(agent.ScopeHome) {
		t.Errorf("JSON scope = %q, want %q", payload.Dests[0].Scope, agent.ScopeHome)
	}
	if payload.Dests[0].Installed != len(bundled) {
		t.Errorf("JSON installed = %d, want %d", payload.Dests[0].Installed, len(bundled))
	}
	if len(payload.Files) != len(bundled) {
		t.Errorf("JSON files = %d, want %d: %v", len(payload.Files), len(bundled), payload.Files)
	}
}

// TestPackDoesNotTouchTheProject: `pack` installs into the agent's directory and
// nowhere else. The project's skills/ and SKILLS.md belong to `odachi skills`.
func TestPackDoesNotTouchTheProject(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	project := t.TempDir()
	seed := map[string]string{
		"skills/acceptance/SKILL.md": "mi skill local\n",
		"SKILLS.md":                  "registro\n",
	}
	for path, content := range seed {
		full := filepath.Join(project, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	before := walkNames(t, project)

	if code, _, errOut := runDir(t, project, []string{"pack", "--agents", "claude"}, ""); code != 0 {
		t.Fatalf("code = %d (stderr %q)", code, errOut)
	}

	after := walkNames(t, project)
	if len(after) != len(before) {
		t.Fatalf("pack changed the project: before %v, after %v", before, after)
	}
	for path, content := range before {
		if after[path] != content {
			t.Errorf("pack changed %s", path)
		}
	}
}

// TestPackDetectsAnAgentOrSaysHowToName: with no flag, an empty home is a
// starting point that names --agents; an installed agent is served.
func TestPackDetectsAnAgentOrSaysHowToName(t *testing.T) {
	empty := t.TempDir()
	withHome(t, empty)
	code, out, errOut := runDir(t, t.TempDir(), []string{"pack"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stderr %q)", code, errOut)
	}
	if !strings.Contains(out, "--agents") {
		t.Errorf("stdout does not name --agents:\n%s", out)
	}

	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".codex"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	withHome(t, home)
	code, out, errOut = runDir(t, t.TempDir(), []string{"pack"}, "")
	if code != 0 {
		t.Fatalf("detected run code = %d (stderr %q)", code, errOut)
	}
	if !strings.Contains(out, fmt.Sprintf("codex · %d installed · ~/.codex/skills", len(bundled))) {
		t.Errorf("the detected agent was not served:\n%s", out)
	}
	if _, err := os.Stat(agentFile(t, home, "claude", "acceptance")); err == nil {
		t.Errorf("pack served an agent that is not installed here")
	}
}

// TestUnknownCommandListsPack: the new verb is part of the list a human reads
// when they type something that does not exist.
func TestUnknownCommandListsPack(t *testing.T) {
	code, _, errOut := runDir(t, t.TempDir(), []string{"inventado"}, "")
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "pack") {
		t.Errorf("the available-command list does not mention pack:\n%s", errOut)
	}
}

// TestPackUsage: `pack -h` exists and names its flags, including the two escape
// hatches and the portable pseudo-agent.
func TestPackUsage(t *testing.T) {
	code, out, errOut := runDir(t, t.TempDir(), []string{"pack", "-h"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stderr %q)", code, errOut)
	}
	for _, want := range []string{"odachi pack", "--agents", "--dry-run", "--json", "--scope", "--dest", "standard"} {
		if !strings.Contains(out, want) {
			t.Errorf("usage does not mention %q:\n%s", want, out)
		}
	}
}

// scopedAgentFile is where one agent reads one skill once --scope has been
// applied: the root the override names, joined with its destination under it.
func scopedAgentFile(t *testing.T, root, agentName string, scope agent.Scope, skill string) string {
	t.Helper()
	a, err := agent.Resolve(agentName)
	if err != nil {
		t.Fatalf("Resolve(%q): %v", agentName, err)
	}
	return filepath.Join(root, filepath.FromSlash(a.In(scope).Dest), skill, "SKILL.md")
}

// TestPackScopeHomeServesOpenCodeFromTheHome: the case the human asked for —
// OpenCode reading skills from every project without running pack per project.
// The destination is OpenCode's documented global config dir.
func TestPackScopeHomeServesOpenCodeFromTheHome(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	project := t.TempDir()

	code, out, errOut := runDir(t, project, []string{"pack", "--agents", "opencode", "--scope", "home"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}
	for _, name := range bundled {
		got, err := os.ReadFile(scopedAgentFile(t, home, "opencode", agent.ScopeHome, name))
		if err != nil {
			t.Errorf("%s was not installed in ~/.config/opencode/skills: %v", name, err)
			continue
		}
		if string(got) != string(repoSkill(t, name)) {
			t.Errorf("%s differs from skills/%s/SKILL.md", name, name)
		}
	}
	if files := walkNames(t, project); len(files) != 0 {
		t.Errorf("--scope home wrote into the project: %v", files)
	}
	if !strings.Contains(out, fmt.Sprintf("opencode · %d installed · ~/.config/opencode/skills", len(bundled))) {
		t.Errorf("stdout lacks the home summary line:\n%s", out)
	}
}

// TestPackStandardInstallsTheCrossClientLocation: the portable default, in both
// roots, and nothing written into the other one.
func TestPackStandardInstallsTheCrossClientLocation(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	project := t.TempDir()

	code, out, errOut := runDir(t, project, []string{"pack", "--agents", "standard"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}
	for _, name := range bundled {
		if _, err := os.Stat(filepath.Join(project, ".agents", "skills", name, "SKILL.md")); err != nil {
			t.Errorf("%s was not installed in ./.agents/skills: %v", name, err)
		}
	}
	if files := walkNames(t, home); len(files) != 0 {
		t.Errorf("the project copy touched the home: %v", files)
	}
	if !strings.Contains(out, fmt.Sprintf("standard · %d installed · ./.agents/skills", len(bundled))) {
		t.Errorf("stdout lacks the cross-client summary:\n%s", out)
	}

	home2, project2 := t.TempDir(), t.TempDir()
	withHome(t, home2)
	code, out, errOut = runDir(t, project2, []string{"pack", "--agents", "standard", "--scope", "home"}, "")
	if code != 0 {
		t.Fatalf("home run code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}
	for _, name := range bundled {
		if _, err := os.Stat(filepath.Join(home2, ".agents", "skills", name, "SKILL.md")); err != nil {
			t.Errorf("%s was not installed in ~/.agents/skills: %v", name, err)
		}
	}
	if files := walkNames(t, project2); len(files) != 0 {
		t.Errorf("the home copy touched the project: %v", files)
	}
	if !strings.Contains(out, fmt.Sprintf("standard · %d installed · ~/.agents/skills", len(bundled))) {
		t.Errorf("stdout lacks the home cross-client summary:\n%s", out)
	}
}

// TestPackDestServesAClientTheTableDoesNotKnow: the refusal of an unknown name
// names the way out, and the way out works — the files land in the folder the
// human typed, and a second identical run writes nothing.
func TestPackDestServesAClientTheTableDoesNotKnow(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	project := t.TempDir()

	code, out, errOut := runDir(t, project, []string{"pack", "--agents", "inventado"}, "")
	if code != 1 {
		t.Fatalf("code = %d, want 1 (stdout %q)", code, out)
	}
	for _, want := range []string{"unknown agent: inventado", "--agents standard", "--dest"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("the refusal does not say %q:\n%s", want, errOut)
		}
	}
	if files := walkNames(t, home); len(files) != 0 {
		t.Errorf("the refusal wrote into the home: %v", files)
	}

	code, out, errOut = runDir(t, project, []string{"pack", "--agents", "inventado", "--dest", "./skills-x"}, "")
	if code != 0 {
		t.Fatalf("escape-hatch run code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}
	for _, name := range bundled {
		got, err := os.ReadFile(filepath.Join(project, "skills-x", name, "SKILL.md"))
		if err != nil {
			t.Errorf("%s did not land in --dest: %v", name, err)
			continue
		}
		if string(got) != string(repoSkill(t, name)) {
			t.Errorf("%s differs from skills/%s/SKILL.md", name, name)
		}
	}
	if !strings.Contains(out, fmt.Sprintf("inventado · %d installed · ./skills-x", len(bundled))) {
		t.Errorf("stdout lacks the --dest summary:\n%s", out)
	}

	before := walkNames(t, project)
	code, out, errOut = runDir(t, project, []string{"pack", "--agents", "inventado", "--dest", "./skills-x"}, "")
	if code != 0 {
		t.Fatalf("second run code = %d, want 0 (stderr %q)", code, errOut)
	}
	if !strings.Contains(out, fmt.Sprintf("inventado · %d unchanged · ./skills-x", len(bundled))) {
		t.Errorf("the second run does not report sin cambios:\n%s", out)
	}
	after := walkNames(t, project)
	if len(after) != len(before) {
		t.Errorf("the second run changed the tree: before %d files, after %d", len(before), len(after))
	}
	for path := range after {
		if strings.Contains(path, ".bak-") {
			t.Errorf("an identical run left a backup: %s", path)
		}
	}
}

// TestPackRefusesDestAndScopeTogether: two flags that both say where are a
// contradiction, and nothing is written while it is unresolved.
func TestPackRefusesDestAndScopeTogether(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	project := t.TempDir()

	code, out, errOut := runDir(t, project, []string{"pack", "--agents", "claude", "--dest", "./skills-x", "--scope", "home"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1 (stdout %q)", code, out)
	}
	for _, want := range []string{"--dest", "--scope", "choose one"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr does not say %q:\n%s", want, errOut)
		}
	}
	if files := walkNames(t, home); len(files) != 0 {
		t.Errorf("a refused combination wrote into the home: %v", files)
	}
	if files := walkNames(t, project); len(files) != 0 {
		t.Errorf("a refused combination wrote into the project: %v", files)
	}
}

// TestPackScopeNeedsNamedAgents: --scope and --dest speak about the agents the
// human named, so on their own they refuse and name --agents.
func TestPackScopeNeedsNamedAgents(t *testing.T) {
	for _, flag := range [][]string{{"--scope", "home"}, {"--dest", "./skills-x"}} {
		t.Run(flag[0], func(t *testing.T) {
			home := t.TempDir()
			withHome(t, home)
			project := t.TempDir()

			code, out, errOut := runDir(t, project, append([]string{"pack"}, flag...), "")

			if code != 1 {
				t.Fatalf("code = %d, want 1 (stdout %q)", code, out)
			}
			for _, want := range []string{flag[0], "--agents"} {
				if !strings.Contains(errOut, want) {
					t.Errorf("stderr does not say %q:\n%s", want, errOut)
				}
			}
			if files := walkNames(t, project); len(files) != 0 {
				t.Errorf("wrote without a named agent: %v", files)
			}
		})
	}
}
