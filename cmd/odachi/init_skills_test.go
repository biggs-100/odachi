package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/start"
	bundledskills "github.com/biggs-100/odachi/skills"
)

// El currículum crece (drive) y el uso no puede quedarse con un número
// escrito a mano: la ayuda nombra la carpeta, no cuántas van. Un número acá se
// queda viejo en la versión siguiente y nadie lo nota hasta que un humano lo lee.
func TestInitUsageDoesNotHardCodeTheSkillCount(t *testing.T) {
	dir := configuredProject(t)

	out, _ := output(t, dir, []string{"init", "-h"})
	if !strings.Contains(out, ".agents/skills") {
		t.Fatalf("el uso no nombra .agents/skills:\n%s", out)
	}
	if hardCoded := regexp.MustCompile(`\b\d+ skills\b`); hardCoded.MatchString(out) {
		t.Fatalf("el uso escribe a mano cuántas skills van: %q", hardCoded.FindString(out))
	}
}

// TestMain points the home adapter at a directory this process owns before any
// test runs. `odachi init` writes the bootstrap skill into the human's home now,
// and a suite that reaches the real home of whoever runs it has escaped the
// repository. A test that needs its own home still overrides it with withHome.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "odachi-home-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "no pude crear un home de prueba: "+err.Error())
		os.Exit(1)
	}
	previous := homeDir
	homeDir = func() (string, error) { return home, nil }

	code := m.Run()

	homeDir = previous
	os.RemoveAll(home)
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
		t.Fatalf("el currículum vino vacío: el proyecto no recibiría ninguna skill")
	}
	for _, skill := range curriculum {
		got, err := os.ReadFile(skillPath(dir, skill.Name))
		if err != nil {
			t.Errorf("%s no se instaló en .agents/skills: %v", skill.Name, err)
			continue
		}
		if !bytes.Equal(got, skill.Content) {
			t.Errorf("%s difiere de skills/%s/SKILL.md", skill.Name, skill.Name)
		}
	}

	// The project has one skills location, and the bootstrap skill is not part
	// of the craft a project carries.
	if _, err := os.Stat(skillPath(dir, bundledskills.BootstrapName)); err == nil {
		t.Errorf("el proyecto recibió la skill de arranque: esa vive en el home")
	}
	// The client directories hold the native agent, never a second copy of the
	// skills: the project's one location is the cross-client folder.
	for _, rel := range []string{".opencode/skills", ".opencode/skill", ".claude/skills"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err == nil {
			t.Errorf("init instaló skills en %s: el proyecto usa una sola carpeta", rel)
		}
	}

	bootstrap, ok := bundledskills.Bootstrap()
	if !ok {
		t.Fatalf("el binario no trae la skill %q", bundledskills.BootstrapName)
	}
	got, err := os.ReadFile(skillPath(home, bootstrap.Name))
	if err != nil {
		t.Fatalf("la skill de arranque no llegó al home: %v", err)
	}
	if !bytes.Equal(got, bootstrap.Content) {
		t.Errorf("la skill del home no es la del binario")
	}

	if !strings.Contains(out, "./.agents/skills") || !strings.Contains(out, "~/.agents/skills") {
		t.Errorf("la salida no nombra los dos destinos:\n%s", out)
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
			t.Errorf("%s no llegó al home: %v", skill.Name, err)
			continue
		}
		if !bytes.Equal(got, skill.Content) {
			t.Errorf("%s difiere del payload", skill.Name)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents")); err == nil {
		t.Errorf("--skills home escribió .agents/ en el proyecto")
	}
	if !strings.Contains(out, "~/.agents/skills") {
		t.Errorf("la salida no nombra el destino del home:\n%s", out)
	}

	before := homeTree(t, home)
	code, out, errOut = runDir(t, dir, []string{"init", "--skills", "home"}, "")
	if code != 0 {
		t.Fatalf("second run code = %d (stdout %q, stderr %q)", code, out, errOut)
	}
	if !slices.Equal(homeTree(t, home), before) {
		t.Errorf("la segunda corrida tocó el home: %v vs %v", homeTree(t, home), before)
	}
	if !strings.Contains(out, "sin cambios") {
		t.Errorf("la segunda corrida no dice que no cambió nada:\n%s", out)
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
	for _, want := range []string{"ya tenía una copia", "~/.agents/skills"} {
		if !strings.Contains(out, want) {
			t.Errorf("la salida no dice %q:\n%s", want, out)
		}
	}
	got, err := os.ReadFile(filepath.Join(dir, ".agents", "skills", "acceptance", "SKILL.md"))
	if err != nil {
		t.Fatalf("la copia del proyecto se borró: %v", err)
	}
	if string(got) != "viejo\n" {
		t.Errorf("la copia del proyecto cambió: %q", got)
	}
}

// TestInitRefusesAnUnknownSkillsScope: a typo is stopped before any write, and
// the message names both values.
func TestInitRefusesAnUnknownSkillsScope(t *testing.T) {
	dir := t.TempDir()

	code, out, errOut := runDir(t, dir, []string{"init", "--skills", "fuera"}, "")

	if code != start.CodeBlocked {
		t.Fatalf("code = %d, want %d (stdout %q, stderr %q)", code, start.CodeBlocked, out, errOut)
	}
	for _, want := range []string{"project", "home", "fuera"} {
		if !strings.Contains(out+errOut, want) {
			t.Errorf("la salida no dice %q: %s%s", want, out, errOut)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err == nil {
		t.Errorf("un scope inválido escribió AGENTS.md")
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
		t.Fatalf("el home tiene %v, want exactamente %v", got, want)
	}

	// The same fence with the flag: --global is what adds the two doors, one per
	// chosen client.
	withGlobal := t.TempDir()
	withHome(t, withGlobal)
	if code, out, errOut := runDir(t, dir, []string{"init", "--global"}, ""); code != 0 {
		t.Fatalf("--global: code = %d (stdout %q, stderr %q)", code, out, errOut)
	}
	wantGlobal := []string{
		".agents/skills/bootstrap/SKILL.md",
		".claude/output-styles/odachi.md",
		".config/opencode/agents/odachi.md",
	}
	if got := homeTree(t, withGlobal); !slices.Equal(got, wantGlobal) {
		t.Fatalf("el home con --global tiene %v, want exactamente %v", got, wantGlobal)
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
	if !strings.Contains(out, "sin cambios") {
		t.Errorf("la segunda corrida no dice que no cambió nada:\n%s", out)
	}
	agentsAfter, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if !bytes.Equal(agentsAfter, agentsBefore) {
		t.Errorf("AGENTS.md cambió en una segunda corrida")
	}
	if got := homeTree(t, home); !slices.Equal(got, homeBefore) {
		t.Errorf("el home cambió en una segunda corrida: %v, antes %v", got, homeBefore)
	}
	// No backup of a skill either: identical bytes are never rewritten.
	backups, err := filepath.Glob(filepath.Join(home, ".agents", "skills", bundledskills.BootstrapName, "SKILL.md.bak-*"))
	if err != nil || len(backups) != 0 {
		t.Errorf("backups del home = %v (err %v), want ninguno", backups, err)
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
			t.Errorf("JSON files no listan %s: %v", path, payload.Files)
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
		t.Errorf("la salida no menciona la copia de .opencode/skills:\n%s", out)
	}
	if !strings.Contains(out, "odachi pack --agents opencode") {
		t.Errorf("la salida no dice de dónde salió esa copia:\n%s", out)
	}
	// The copy that was already there is not touched and not duplicated: init
	// still installs the project under .agents/skills only.
	if _, err := os.Stat(skillPath(dir, "acceptance")); err != nil {
		t.Errorf("init no instaló el currículum en .agents/skills: %v", err)
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
		t.Fatalf("code = %d, want 0: la parte del proyecto quedó hecha (stdout %q, stderr %q)", code, out, errOut)
	}
	for _, rel := range []string{"AGENTS.md", ".gitignore", ".agents/skills/acceptance/SKILL.md"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("%s no se escribió: %v", rel, err)
		}
	}
	if !strings.Contains(out, "~/.agents/skills") {
		t.Errorf("la salida no dice qué quedó pendiente:\n%s", out)
	}
	if !strings.Contains(out, "odachi init") {
		t.Errorf("la salida no nombra el comando que lo destraba:\n%s", out)
	}
}
