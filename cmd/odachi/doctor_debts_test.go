package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/layout"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/rules"
	bundledskills "github.com/biggs-100/odachi/skills"
)

// A linked worktree has .git as a file pointing at the main repository. Telling
// that project "there is no .git, run git init" is a lie and a dead end: git init is
// already done, in the main repository.
func TestDoctorUnderstandsAWorktree(t *testing.T) {
	dir := t.TempDir()
	pointer := "gitdir: /otro/repo/.git/worktrees/rama\n"
	if err := os.WriteFile(filepath.Join(dir, ".git"), []byte(pointer), 0o644); err != nil {
		t.Fatalf("write .git: %v", err)
	}

	_, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	out := stdout + stderr

	if !strings.Contains(out, "worktree") {
		t.Fatalf("the report does not recognise the worktree: %q", out)
	}
	if strings.Contains(out, "run git init") {
		t.Fatalf("the report sends a worktree to git init: %q", out)
	}
}

// A profile whose name nobody shipped and that inherits from nothing resolves to
// a profile with every field empty: it validates and configures nothing. Doctor
// has to call that a problem, because every verb would run with a blank ceremony.
func TestDoctorRefusesAProfileThatInheritsFromNobody(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash(layout.Odd)), 0o755); err != nil {
		t.Fatalf("mkdir odd: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(layout.Profile)), []byte("name = \"inventado\"\n"), 0o644); err != nil {
		t.Fatalf("write profile.toml: %v", err)
	}

	code, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	out := stdout + stderr

	if code != 1 {
		t.Fatalf("code = %d, want 1: a profile that inherits from nobody is a problem", code)
	}
	if !strings.Contains(out, "inherits from nobody") {
		t.Fatalf("the report does not say why: %q", out)
	}
	if !strings.Contains(out, "extends") {
		t.Fatalf("the report does not say how to fix it: %q", out)
	}
}

// A profile that inherits from a shipped one is a legitimate configuration, and
// must stay green so the rule only catches the empty ceremony.
func TestDoctorAcceptsAProfileThatInherits(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash(layout.Odd)), 0o755); err != nil {
		t.Fatalf("mkdir odd: %v", err)
	}
	content := "name = \"mio\"\nextends = \"senior\"\n"
	if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(layout.Profile)), []byte(content), 0o644); err != nil {
		t.Fatalf("write profile.toml: %v", err)
	}

	_, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	out := stdout + stderr

	if !strings.Contains(out, "profile mio") {
		t.Fatalf("the report does not resolve the inherited profile: %q", out)
	}
	if strings.Contains(out, "there is no file") {
		t.Fatalf("the report says there is no file when there is one: %q", out)
	}
}

// checkLine returns the one line of the report that carries a check number. The
// number is how a report and a test name the same check, so the helper matches on
// it and never on the position: a check inserted in the middle cannot make a test
// read its neighbour's line.
func checkLine(t *testing.T, report string, number int) string {
	t.Helper()

	prefix := " " + strconv.Itoa(number) + " · "
	for _, line := range strings.Split(report, "\n") {
		if strings.Contains(line, prefix) {
			return line
		}
	}
	t.Fatalf("the report does not carry check %d:\n%s", number, report)
	return ""
}

// A project that just ran `odachi init` has the skills in .agents/skills and no
// registry of its own, and check 9 used to say only "none: there is no skills/":
// the human read that there was nothing when init had installed its own. They are
// two different questions —the project registry and what is installed where the
// client reads it— and the line answers both without mixing them.
func TestDoctorCheckNineSaysBothSkillFacts(t *testing.T) {
	dir := configuredProject(t)

	out, _ := output(t, dir, []string{"doctor"})
	line := checkLine(t, out, 9)

	curriculum, err := bundledskills.Curriculum()
	if err != nil {
		t.Fatalf("the bundled skills: %v", err)
	}
	for _, want := range []string{
		"skills/: does not exist",
		fmt.Sprintf(".agents/skills: %d installed", len(curriculum)),
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("check 9 does not say %q: %s", want, line)
		}
	}
}

// The voice is the last section of the block and the binary writes it. If the
// block says one thing and the profile the binary reads says another, the agent
// speaks in a register the tool never promised, and the check has to say it before
// the human notices by reading the agent: it names both and the command that makes
// them agree.
func TestDoctorReportsAVoiceThatDoesNotMatchTheProfile(t *testing.T) {
	dir := configuredProject(t)

	// The project was initialized with the novato voice and the profile `start`
	// wrote. Putting the block back to novato is the state of a project whose
	// profile was changed without refreshing the voice.
	block, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Block(novato): %v", err)
	}
	writeFile(t, dir, "AGENTS.md", block)
	profile := readProfileName(t, dir)

	code, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	if code != 1 {
		t.Fatalf("a voice that does not match the profile: exit %d, want 1: %s %s", code, stdout, stderr)
	}
	line := checkLine(t, stdout+stderr, 11)
	for _, want := range []string{"bad 11", "novato", profile, "odachi level --to " + profile} {
		if !strings.Contains(line, want) {
			t.Fatalf("check 11 does not say %q: %s", want, line)
		}
	}

	// The command the line names fixes the problem, and the check goes back to ok.
	if code, _, stderr := runDir(t, dir, []string{"level", "--to", profile}, ""); code != 0 {
		t.Fatalf("level --to %s: exit %d: %s", profile, code, stderr)
	}
	out, _ := output(t, dir, []string{"doctor"})
	if line := checkLine(t, out, 11); !strings.HasPrefix(line, "ok 11") {
		t.Fatalf("after level --to check 11 is still bad: %s", line)
	}
}

// A block without the voice section and a profile that exists: the voice is not in
// the project, and the command that writes it is init.
func TestDoctorReportsABlockWithoutTheVoiceSection(t *testing.T) {
	dir := configuredProject(t)
	writeFile(t, dir, "AGENTS.md", "<!-- odachi:begin -->\n# Reglas del proyecto (odachi)\n\nreglas\n<!-- odachi:end -->\n")

	code, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	if code != 1 {
		t.Fatalf("a block without the voice and with a profile: exit %d, want 1: %s %s", code, stdout, stderr)
	}
	line := checkLine(t, stdout+stderr, 11)
	for _, want := range []string{"bad 11", rules.Section, "odachi init"} {
		if !strings.Contains(line, want) {
			t.Fatalf("check 11 does not say %q: %s", want, line)
		}
	}
}

// The per-profile personas from v0.10.0 are no longer this version's agent. They are the human's files, not
// a defect: the check says so on the same green line, names the folders, and
// does not fail because of it.
func TestDoctorSaysTheV0100PersonaFilesAreLeftovers(t *testing.T) {
	dir := configuredProject(t)
	writeFile(t, dir, ".opencode/agents/odachi-senior.md", "viejo")

	code, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	if code != 0 {
		t.Fatalf("the leftovers are not a problem: exit %d: %s %s", code, stdout, stderr)
	}
	line := checkLine(t, stdout, 11)
	if !strings.HasPrefix(line, "ok 11") {
		t.Fatalf("check 11 has to be ok with leftovers: %s", line)
	}
	for _, want := range []string{"v0.10.0", ".opencode/agents", ".claude/output-styles", "you can delete them"} {
		if !strings.Contains(line, want) {
			t.Fatalf("check 11 does not say %q: %s", want, line)
		}
	}
}

// A project without odd/profile.toml has nothing to compare the voice against:
// the block carries the one init wrote and nothing is broken. Failing here would
// tell someone who just ran init that their project is wrong.
func TestDoctorSaysNoProfileIsNothingToCompare(t *testing.T) {
	dir := configuredProject(t)
	if err := os.Remove(filepath.Join(dir, filepath.FromSlash(layout.Profile))); err != nil {
		t.Fatalf("delete odd/profile.toml: %v", err)
	}

	code, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	if code != 0 {
		t.Fatalf("without a profile there is no voice to compare: exit %d: %s%s", code, stdout, stderr)
	}
	line := checkLine(t, stdout, 11)
	if !strings.HasPrefix(line, "ok 11") {
		t.Fatalf("check 11 has to be ok without a profile: %s", line)
	}
	if !strings.Contains(line, "there is no profile to compare") {
		t.Fatalf("check 11 does not say there is nothing to compare against: %s", line)
	}
}

// readProfileName is the profile the binary reads, for the lines that name it.
func readProfileName(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(layout.Profile)))
	if err != nil {
		t.Fatalf("read odd/profile.toml: %v", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if ok && strings.TrimSpace(key) == "name" {
			return strings.Trim(strings.TrimSpace(value), `"`)
		}
	}
	t.Fatalf("the profile does not say its name: %s", data)
	return ""
}

// The closing line says how many checks ran: with a new one there are twelve, and
// the verb's usage cannot keep saying ten.
func TestDoctorUsageAndSummaryCountTwelveChecks(t *testing.T) {
	dir := configuredProject(t)

	out, _ := output(t, dir, []string{"doctor"})
	if !strings.Contains(out, "all clear: 12 checks") {
		t.Fatalf("the summary counts twelve checks:\n%s", out)
	}

	usageOut, _ := output(t, dir, []string{"doctor", "-h"})
	if !strings.Contains(usageOut, "twelve health checks") {
		t.Fatalf("the usage cannot keep saying ten:\n%s", usageOut)
	}
}
