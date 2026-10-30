package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

// TestSetupDryRunCountsASkillItWouldReplace: after an upgrade the skills in the
// home differ from the binary's, and a rehearsal that counts only missing and
// identical ones says "nothing to write" about the exact machine it exists for.
func TestSetupDryRunCountsASkillItWouldReplace(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	writeFile(t, home, ".config/opencode/opencode.json", "{}")
	writeFile(t, home, ".agents/skills/acceptance/SKILL.md", "an older binary wrote this\n")

	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"setup", "--dry-run"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	report := out + errOut
	if !strings.Contains(report, "would replace") {
		t.Errorf("the rehearsal does not count the skill it would replace:\n%s", report)
	}
	if strings.Contains(report, "nothing to write") {
		t.Errorf("the rehearsal calls an out-of-date machine current:\n%s", report)
	}

	got, err := os.ReadFile(filepath.Join(home, ".agents", "skills", "acceptance", "SKILL.md"))
	if err != nil {
		t.Fatalf("the old skill is gone: %v", err)
	}
	if string(got) != "an older binary wrote this\n" {
		t.Errorf("--dry-run rewrote the skill: %q", got)
	}
}

// TestSetupDryRunNamesTheDefaultAgentItWouldWrite: the config is a file this run
// changes when odachi is not the default yet, and the report has to name it. It
// must not lean on the door being on disk: in a rehearsal the door was written by
// the step before, and only in the report.
func TestSetupDryRunNamesTheDefaultAgentItWouldWrite(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	writeFile(t, home, ".config/opencode/opencode.json", "{}")

	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"setup", "--dry-run"}, "")
	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	report := out + errOut
	if !strings.Contains(report, "opencode.json") || !strings.Contains(report, "would write") {
		t.Errorf("the rehearsal does not name the config it would write:\n%s", report)
	}
	if strings.Contains(report, "is the default agent") {
		t.Errorf("the rehearsal claims the default is set:\n%s", report)
	}

	data, err := os.ReadFile(filepath.Join(home, ".config", "opencode", "opencode.json"))
	if err != nil {
		t.Fatalf("the config is gone: %v", err)
	}
	if strings.Contains(string(data), persona.AgentName) {
		t.Errorf("--dry-run wrote the default agent: %s", data)
	}
}

// TestSetupDryRunJSONSaysItIsARehearsal: the payload is what a machine reads, and
// a plan and a fact cannot look the same there: the files it lists are the ones it
// would write, and the flag says so.
func TestSetupDryRunJSONSaysItIsARehearsal(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	writeFile(t, home, ".config/opencode/opencode.json", "{}")

	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"setup", "--dry-run", "--json"}, "")
	if code != 0 {
		t.Fatalf("code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	var payload struct {
		DryRun  bool     `json:"dry_run"`
		Message string   `json:"message"`
		Files   []string `json:"files"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("the payload is not JSON: %v\n%s", err, out)
	}
	if !payload.DryRun {
		t.Errorf("dry_run is false in a rehearsal: %+v", payload)
	}
	if !strings.Contains(payload.Message, "dry run") {
		t.Errorf("message = %q, want the word dry run", payload.Message)
	}
	if len(payload.Files) == 0 {
		t.Errorf("the rehearsal names no file it would write")
	}
}

// TestTheRealRunSaysItReplacedASkill: the count is not only for the rehearsal. A
// real setup over a skill an older binary wrote reports it, and the old bytes are
// beside it.
func TestTheRealRunSaysItReplacedASkill(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	writeFile(t, home, ".config/opencode/opencode.json", "{}")
	writeFile(t, home, ".agents/skills/acceptance/SKILL.md", "an older binary wrote this\n")

	dir := t.TempDir()
	initRepo(t, dir)

	code, out, errOut := runDir(t, dir, []string{"setup"}, "")
	if code != 0 {
		t.Fatalf("code = %d (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(out+errOut, "replaced") {
		t.Errorf("a real run does not say what it replaced:\n%s%s", out, errOut)
	}
	backups, err := filepath.Glob(filepath.Join(home, ".agents", "skills", "acceptance", "SKILL.md.bak-*"))
	if err != nil || len(backups) == 0 {
		t.Errorf("the replaced skill was not backed up: %v (%v)", backups, err)
	}
}
