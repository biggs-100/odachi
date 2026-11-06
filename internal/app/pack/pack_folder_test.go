package pack_test

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/biggs-100/odachi/internal/app/pack"
)

// routerMain is a skill with one reference: the smallest folder that is more
// than a lone SKILL.md.
const routerMain = "---\nname: router\ndescription: Trigger: cuando hay que enrutar\n---\n\n## When\n\ncuando\n\n## Rules\n\nregla\n\n## Never\n\nnunca\n\n## Output\n\nthe file.\n\n## References\n\n- `references/numbers.md` — the scale.\n"

func routerSkill() pack.Skill {
	return pack.Skill{
		Name: "router",
		Files: []pack.File{
			{Path: pack.SkillFile, Content: []byte(routerMain)},
			{Path: "references/numbers.md", Content: []byte("the scale\n")},
		},
	}
}

// installedPath is where one file of one skill lands for the claude agent.
func installedPath(root, skill, file string) string {
	return filepath.Join(dest(root, "claude"), skill, filepath.FromSlash(file))
}

// TestASkillFolderTravelsWhole: the reference is installed beside the SKILL.md
// that links it, in the same run, and the skill is reported once.
func TestASkillFolderTravelsWhole(t *testing.T) {
	root := home()
	files := newFakeFiles(nil)

	result := run(files, fakeHome{dir: root}, pack.Options{Agents: "claude", Skills: []pack.Skill{routerSkill()}})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q, lines %v)", result.Code, result.Message, result.Lines)
	}
	for file, want := range map[string]string{
		"SKILL.md":              routerMain,
		"references/numbers.md": "the scale\n",
	} {
		got, ok := files.content(installedPath(root, "router", file))
		if !ok {
			t.Errorf("%s was not installed", file)
			continue
		}
		if got != want {
			t.Errorf("%s = %q, want %q", file, got, want)
		}
	}
	if got := result.Dests[0].Installed; got != 1 {
		t.Errorf("Installed = %d, want 1: the summary counts skills, not files", got)
	}
	if !slices.Contains(result.Lines, "claude · installed · ~/.claude/skills/router/") {
		t.Errorf("the skill is not reported as one folder, lines: %v", result.Lines)
	}
}

// TestAChangedReferenceIsBackedUpAndWritten: a skill that already lived keeps
// its bytes aside when a reference changes, and the next run is quiet again.
func TestAChangedReferenceIsBackedUpAndWritten(t *testing.T) {
	root := home()
	files := newFakeFiles(nil)
	opts := pack.Options{Agents: "claude", Skills: []pack.Skill{routerSkill()}}
	run(files, fakeHome{dir: root}, opts)

	reference := installedPath(root, "router", "references/numbers.md")
	files.files[reference] = []byte("an older scale\n")
	files.writes = nil
	files.backups = nil

	result := run(files, fakeHome{dir: root}, opts)

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q)", result.Code, result.Message)
	}
	if !slices.Contains(files.backups, reference) {
		t.Errorf("the changed reference was not backed up: %v", files.backups)
	}
	if got := string(files.files[reference]); got != "the scale\n" {
		t.Errorf("reference = %q, want the bundled bytes", got)
	}
	if !slices.Contains(result.Lines, "claude · backed up and written · ~/.claude/skills/router/") {
		t.Errorf("the skill is not reported as one folder, lines: %v", result.Lines)
	}
}

// TestANewReferenceIsWrittenAndNotBackedUp: a skill that already lived and gains
// a file has nothing of the human's to set aside. The report says what happened,
// not what the skill became (the row 0205 left at pack.go:461).
func TestANewReferenceIsWrittenAndNotBackedUp(t *testing.T) {
	root := home()
	files := newFakeFiles(nil)
	run(files, fakeHome{dir: root}, pack.Options{Agents: "claude", Skills: []pack.Skill{routerSkill()}})
	files.writes = nil
	files.backups = nil

	grown := routerSkill()
	grown.Files = append(grown.Files, pack.File{Path: "references/extra.md", Content: []byte("more\n")})

	result := run(files, fakeHome{dir: root}, pack.Options{Agents: "claude", Skills: []pack.Skill{grown}})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q)", result.Code, result.Message)
	}
	if len(files.backups) != 0 {
		t.Errorf("backups = %v, want none: nothing of the human's was replaced", files.backups)
	}
	if got := result.Dests[0].BackedUp; got != 0 {
		t.Errorf("BackedUp = %d, want 0: no backup ran", got)
	}
	if got := result.Dests[0].Written; got != 1 {
		t.Errorf("Written = %d, want 1", got)
	}
	if !slices.Contains(result.Lines, "claude · 1 written · ~/.claude/skills") {
		t.Errorf("the summary does not say written, lines: %v", result.Lines)
	}
}

// TestANewReferenceInADryRunSaysWriteAndNeverBackUp: the rehearsal names what the
// real run would do, and a backup that cannot happen is not printed.
func TestANewReferenceInADryRunSaysWriteAndNeverBackUp(t *testing.T) {
	root := home()
	files := newFakeFiles(nil)
	run(files, fakeHome{dir: root}, pack.Options{Agents: "claude", Skills: []pack.Skill{routerSkill()}})
	files.writes = nil
	files.backups = nil

	grown := routerSkill()
	grown.Files = append(grown.Files, pack.File{Path: "references/extra.md", Content: []byte("more\n")})

	result := run(files, fakeHome{dir: root}, pack.Options{Agents: "claude", Skills: []pack.Skill{grown}, DryRun: true})

	if result.Code != pack.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q)", result.Code, result.Message)
	}
	if len(files.writes) != 0 || len(files.backups) != 0 {
		t.Errorf("a dry run wrote %v and backed up %v, want nothing", files.writes, files.backups)
	}
	if got := result.Dests[0].Written; got != 1 {
		t.Errorf("Written = %d, want 1: the rehearsal names what it would write", got)
	}
	if !slices.Contains(result.Lines, "claude · 1 to write · ~/.claude/skills") {
		t.Errorf("the dry run does not say to write, lines: %v", result.Lines)
	}
}

// TestASkillFolderWithAPathOutsideIsRefused: the payload is fixed at build time,
// and a path that would leave the agent's directory is a broken build, not a
// write.
func TestASkillFolderWithAPathOutsideIsRefused(t *testing.T) {
	for _, path := range []string{"../evil.md", "references/../../evil.md", "/etc/passwd", `references\evil.md`} {
		files := newFakeFiles(nil)
		result := run(files, fakeHome{dir: home()}, pack.Options{Agents: "claude", Skills: []pack.Skill{{
			Name: "router",
			Files: []pack.File{
				{Path: pack.SkillFile, Content: []byte("x")},
				{Path: path, Content: []byte("x")},
			},
		}}})

		if result.Code != pack.CodeBlocked {
			t.Errorf("path %q: Code = %d, want blocked", path, result.Code)
		}
		if len(files.writes) != 0 {
			t.Errorf("path %q: wrote %v, want nothing", path, files.writes)
		}
	}
}
