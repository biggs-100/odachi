package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// bundled is every skill this repository ships, read from disk so adding one does
// not require editing this list. The registry is dogfooded:
// the tests run the real command over the real skills/ directory.
var bundled = func() []string {
	entries, err := os.ReadDir(filepath.Join("..", "..", "skills"))
	if err != nil {
		panic("read the skills directory: " + err.Error())
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	slices.Sort(names)
	return names
}()

// repoRoot is the module root, one level above cmd/odachi.
func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	return root
}

func skillFile(name, trigger string) string {
	return "---\nname: " + name + "\ndescription: Trigger: " + trigger +
		"\n---\n\n## When\n\ncuando\n\n## Rules\n\nregla\n\n## Never\n\nnunca\n"
}

// TestSkillsCheckIsGreenInThisRepo: the bundled skills and the committed
// registry agree, so the command CI runs is green on the real tree.
func TestSkillsCheckIsGreenInThisRepo(t *testing.T) {
	t.Parallel()
	code, out, errOut := runDir(t, repoRoot(t), []string{"skills", "--check"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}
}

// TestSkillsListsTheBundledFive: the human surface is the five names, each with
// its trigger and its path.
func TestSkillsListsTheBundledFive(t *testing.T) {
	t.Parallel()
	code, out, errOut := runDir(t, repoRoot(t), []string{"skills"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stderr %q)", code, errOut)
	}
	for _, name := range bundled {
		if !strings.Contains(out, name+" · ") {
			t.Errorf("stdout does not list %s:\n%s", name, out)
		}
		if !strings.Contains(out, "skills/"+name+"/SKILL.md") {
			t.Errorf("stdout does not show the path of %s:\n%s", name, out)
		}
	}
}

// TestSkillsJSON is the machine shape: one entry per skill with name, trigger
// and path.
func TestSkillsJSON(t *testing.T) {
	t.Parallel()
	code, out, errOut := runDir(t, repoRoot(t), []string{"skills", "--json"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stderr %q)", code, errOut)
	}

	var payload struct {
		Code   int `json:"code"`
		Skills []struct {
			Name    string `json:"name"`
			Trigger string `json:"trigger"`
			Path    string `json:"path"`
		} `json:"skills"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out)
	}
	if payload.Code != 0 {
		t.Errorf("JSON code = %d, want 0", payload.Code)
	}
	seen := map[string]bool{}
	for _, entry := range payload.Skills {
		seen[entry.Name] = true
		if entry.Trigger == "" {
			t.Errorf("%s has no trigger in JSON", entry.Name)
		}
		if entry.Path != "skills/"+entry.Name+"/SKILL.md" {
			t.Errorf("%s path = %q, want skills/%s/SKILL.md", entry.Name, entry.Path, entry.Name)
		}
	}
	for _, name := range bundled {
		if !seen[name] {
			t.Errorf("JSON does not list %s: %+v", name, payload.Skills)
		}
	}
}

// TestSkillsCheckReportsOneLinePerViolation: each broken file gets its own
// English line naming the file and what is wrong.
func TestSkillsCheckReportsOneLinePerViolation(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		seed   map[string]string
		reason string
	}{
		{
			name:   "without frontmatter",
			seed:   map[string]string{"skills/suelta/SKILL.md": "## When\n\ncuando\n\n## Rules\n\nregla\n\n## Never\n\nnunca\n"},
			reason: "frontmatter",
		},
		{
			name:   "description without Trigger",
			seed:   map[string]string{"skills/aceptar/SKILL.md": strings.Replace(skillFile("aceptar", "x"), "description: Trigger: x", "description: aceptar cosas", 1)},
			reason: `does not start with "Trigger:"`,
		},
		{
			name:   "name that does not match the directory",
			seed:   map[string]string{"skills/aceptar/SKILL.md": skillFile("otro", "x")},
			reason: "does not match the directory",
		},
		{
			name:   "name outside the shape",
			seed:   map[string]string{"skills/Aceptar/SKILL.md": strings.Replace(skillFile("Aceptar", "x"), "description: Trigger: x", "description: Trigger: x", 1)},
			reason: "does not match the shape",
		},
		{
			name:   "body of 61 lines",
			seed:   map[string]string{"skills/larga/SKILL.md": longSkill("larga", 61)},
			reason: "the maximum is 60",
		},
		{
			name:   "TODO",
			seed:   map[string]string{"skills/aceptar/SKILL.md": strings.Replace(skillFile("aceptar", "x"), "regla", "TODO regla", 1)},
			reason: "TODO",
		},
		{
			name:   "empty section",
			seed:   map[string]string{"skills/vacia/SKILL.md": "---\nname: vacia\ndescription: Trigger: un hueco\n---\n\n## When\n\ncuando\n\n## Rules\n\n## Never\n\nnunca\n"},
			reason: `"Rules" section is empty`,
		},
		{
			name:   "repeated name",
			seed:   map[string]string{"skills/uno/SKILL.md": skillFile("uno", "primero"), "skills/dos/SKILL.md": skillFile("uno", "segundo")},
			reason: "is repeated",
		},
		{
			name:   "hand-edited registry",
			seed:   map[string]string{"skills/aceptar/SKILL.md": skillFile("aceptar", "x"), "SKILLS.md": "| a mano |\n"},
			reason: "--write",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range tc.seed {
				writeFile(t, dir, name, content)
			}

			code, out, errOut := runDir(t, dir, []string{"skills", "--check"}, "")

			if code != 1 {
				t.Fatalf("code = %d, want 1 (stdout %q, stderr %q)", code, out, errOut)
			}
			if !strings.Contains(errOut, tc.reason) {
				t.Errorf("stderr %q does not contain %q", errOut, tc.reason)
			}
			if !strings.Contains(errOut, "SKILLS.md") && !strings.Contains(errOut, "skills/") {
				t.Errorf("stderr %q does not name a file", errOut)
			}
			if out != "" {
				t.Errorf("stdout = %q, want nothing on a failed check", out)
			}
		})
	}
}

// longSkill builds a skill whose body has exactly bodyLines lines.
func longSkill(name string, bodyLines int) string {
	lines := []string{"## When", "cuando", "## Rules"}
	for len(lines) < bodyLines-2 {
		lines = append(lines, "- regla")
	}
	lines = append(lines, "## Never", "nunca")
	return "---\nname: " + name + "\ndescription: Trigger: una skill larga\n---\n\n" +
		strings.Join(lines, "\n") + "\n"
}

// TestSkillsWriteIsIdempotent: two runs, same bytes, and the second says so.
func TestSkillsWriteIsIdempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, dir, "skills/ship/SKILL.md", skillFile("ship", "cuándo se publica"))
	writeFile(t, dir, "skills/acceptance/SKILL.md", skillFile("acceptance", "cuándo se acepta"))

	if code, _, errOut := runDir(t, dir, []string{"skills", "--write"}, ""); code != 0 {
		t.Fatalf("first write code = %d (stderr %q)", code, errOut)
	}
	first, err := os.ReadFile(filepath.Join(dir, "SKILLS.md"))
	if err != nil {
		t.Fatalf("SKILLS.md was not written: %v", err)
	}
	if !strings.HasPrefix(string(first), "<!-- generated by odachi skills --write; do not edit it by hand -->\n") {
		t.Errorf("SKILLS.md does not start with the marker:\n%s", first)
	}

	code, out, errOut := runDir(t, dir, []string{"skills", "--write"}, "")
	if code != 0 {
		t.Fatalf("second write code = %d (stderr %q)", code, errOut)
	}
	second, err := os.ReadFile(filepath.Join(dir, "SKILLS.md"))
	if err != nil {
		t.Fatalf("read SKILLS.md: %v", err)
	}
	if string(second) != string(first) {
		t.Errorf("the second --write changed the bytes:\n%s\nvs\n%s", second, first)
	}
	if !strings.Contains(out, "unchanged") {
		t.Errorf("stdout = %q, want it to say nothing changed", out)
	}
}

// TestSkillsHandEditedRegistryFailsAndWriteRestores: the registry is generated,
// so a hand edit is an error that names --write, never a merge.
func TestSkillsHandEditedRegistryFailsAndWriteRestores(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, dir, "skills/acceptance/SKILL.md", skillFile("acceptance", "cuándo se acepta"))
	if code, _, errOut := runDir(t, dir, []string{"skills", "--write"}, ""); code != 0 {
		t.Fatalf("write code = %d (stderr %q)", code, errOut)
	}
	good, err := os.ReadFile(filepath.Join(dir, "SKILLS.md"))
	if err != nil {
		t.Fatalf("read SKILLS.md: %v", err)
	}
	edited := string(good) + "| inventada | a mano | skills/inventada/SKILL.md |\n"
	writeFile(t, dir, "SKILLS.md", edited)

	code, _, errOut := runDir(t, dir, []string{"skills", "--check"}, "")
	if code != 1 {
		t.Fatalf("check on a hand-edited registry = %d, want 1", code)
	}
	if !strings.Contains(errOut, "--write") {
		t.Errorf("stderr %q does not tell the human to run --write", errOut)
	}

	if code, _, errOut := runDir(t, dir, []string{"skills", "--write"}, ""); code != 0 {
		t.Fatalf("restoring write code = %d (stderr %q)", code, errOut)
	}
	restored, err := os.ReadFile(filepath.Join(dir, "SKILLS.md"))
	if err != nil {
		t.Fatalf("read restored SKILLS.md: %v", err)
	}
	if string(restored) != string(good) {
		t.Errorf("--write did not restore the generated bytes:\n%s\nvs\n%s", restored, good)
	}
	if code, _, errOut := runDir(t, dir, []string{"skills", "--check"}, ""); code != 0 {
		t.Errorf("check after restoring = %d, want 0 (stderr %q)", code, errOut)
	}
}

// TestSkillNewCreatesThenRefuses: the creator writes once, updates the registry
// and never clobbers what exists.
func TestSkillNewCreatesThenRefuses(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	code, out, errOut := runDir(t, dir, []string{"skill", "new", "ejemplo"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stderr %q)", code, errOut)
	}
	path := filepath.Join(dir, "skills", "ejemplo", "SKILL.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("the skill was not created: %v", err)
	}
	if !strings.Contains(string(content), "name: ejemplo") {
		t.Errorf("SKILL.md does not declare its name:\n%s", content)
	}
	if !strings.Contains(out, "skills/ejemplo/SKILL.md") || !strings.Contains(out, "SKILLS.md") {
		t.Errorf("stdout %q does not report the skill and the registry", out)
	}
	registry, err := os.ReadFile(filepath.Join(dir, "SKILLS.md"))
	if err != nil {
		t.Fatalf("the registry was not generated: %v", err)
	}
	if !strings.Contains(string(registry), "| ejemplo |") {
		t.Errorf("SKILLS.md does not list ejemplo:\n%s", registry)
	}

	code, _, errOut = runDir(t, dir, []string{"skill", "new", "ejemplo"}, "")
	if code != 1 {
		t.Fatalf("second create code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "already exists") {
		t.Errorf("stderr %q does not say the skill exists", errOut)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after refusal: %v", err)
	}
	if string(after) != string(content) {
		t.Errorf("the refused run rewrote the skill:\n%s\nvs\n%s", after, content)
	}
}

// TestSkillNewRejectsABadSlugBeforeWriting: the regex is checked first, so a bad
// slug leaves no directory behind.
func TestSkillNewRejectsABadSlugBeforeWriting(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	code, _, errOut := runDir(t, dir, []string{"skill", "new", "Slug Invalido"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "^[a-z][a-z0-9-]{2,40}$") {
		t.Errorf("stderr %q does not state the slug shape", errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, "skills")); !os.IsNotExist(err) {
		t.Errorf("a rejected slug created skills/: %v", err)
	}
}

// TestSkillsWithoutDirectorySaysHowToStart: no skills is a starting point, not
// a failure, and the line names the command that creates the first one.
func TestSkillsWithoutDirectorySaysHowToStart(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{{"skills"}, {"skills", "--check"}} {
		dir := t.TempDir()

		code, out, errOut := runDir(t, dir, args, "")

		if code != 0 {
			t.Errorf("%v: code = %d, want 0 (stderr %q)", args, code, errOut)
		}
		if !strings.Contains(out, "odachi skill new") {
			t.Errorf("%v: stdout %q does not say how to start", args, out)
		}
	}
}

// TestSkillsCheckAndWriteCannotBeCombined: two intents in one run is a typo, and
// it is reported as one line.
func TestSkillsCheckAndWriteCannotBeCombined(t *testing.T) {
	t.Parallel()
	code, _, errOut := runDir(t, t.TempDir(), []string{"skills", "--check", "--write"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "--check") || !strings.Contains(errOut, "--write") {
		t.Errorf("stderr %q does not name both flags", errOut)
	}
}
