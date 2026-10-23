package skills_test

import (
	"testing"

	skillsapp "github.com/biggs-100/odachi/internal/app/skills"
)

// A file inside skills/ is not a skill. The scanner has to decide that by asking
// whether the entry is a directory, never by the error class of a stat: Windows
// reports "skills/<file>/SKILL.md" as missing while POSIX reports ENOTDIR, so the
// same code passes on one platform and fails on the other.
//
// This repository caught exactly that: skills/embed.go is a file, linux CI read
// ENOTDIR as a hard failure, and windows CI read it as "nothing there".
func TestAFileInsideTheSkillsDirectoryIsNotASkill(t *testing.T) {
	files := seededProject()
	files.put("skills/embed.go", "package skills\n")

	got := skillsapp.Run(skillsapp.Options{Dir: dir, Check: true}, files)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0: a file inside skills/ is not a skill (lines %q)", got.Code, got.Lines)
	}
}

// The listing must skip it too, instead of showing a skill that does not exist.
func TestAFileInsideTheSkillsDirectoryIsNotListed(t *testing.T) {
	files := seededProject()
	files.put("skills/embed.go", "package skills\n")

	got := skillsapp.Run(skillsapp.Options{Dir: dir}, files)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (lines %q)", got.Code, got.Lines)
	}
	for _, line := range got.Lines {
		if len(line) > 0 && line[0] == 'e' {
			t.Fatalf("the listing included a file as a skill: %q", line)
		}
	}
}
