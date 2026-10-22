package doctor

import (
	"errors"
	"strings"
	"testing"

	domaindoctor "github.com/biggs-100/odachi/internal/domain/doctor"
)

// TestEmptyFolderGetsOneSituationInsteadOfTenChecks is the beginner case: a
// folder where odachi was never started. Ten checks name three different
// commands; one situation names the two that matter, in the order a person runs
// them, and that is the whole point of the pre-flight.
func TestEmptyFolderGetsOneSituationInsteadOfTenChecks(t *testing.T) {
	result := run(newFakeFiles(), &fakeLookup{err: errors.New("not found")}, cleanProfiles(), "")

	if result.Code != CodeProblems {
		t.Fatalf("una carpeta sin proyecto todavía no está lista: Code = %d, want %d", result.Code, CodeProblems)
	}
	if len(result.Checks) != 0 {
		t.Fatalf("the checks are not what a beginner needs, got %d:\n%s", len(result.Checks), strings.Join(result.Lines, "\n"))
	}

	want := []string{
		"esta carpeta todavía no es un proyecto de odachi",
		"corré odachi init (deja las reglas que el agente lee) y después odachi start (la entrevista)",
	}
	if len(result.Lines) != len(want) {
		t.Fatalf("the beginner reads one situation and one step, got %d lines:\n%s", len(result.Lines), strings.Join(result.Lines, "\n"))
	}
	for i := range want {
		if result.Lines[i] != want[i] {
			t.Fatalf("line %d = %q, want %q", i+1, result.Lines[i], want[i])
		}
	}
	if result.Message != want[0] {
		t.Fatalf("Message = %q, want the situation", result.Message)
	}
	if result.Problems != 1 || result.Warnings != 0 {
		t.Fatalf("problems=%d warnings=%d, want 1 and 0", result.Problems, result.Warnings)
	}
}

// The pre-flight decides whether there is a project at all, and nothing else: a
// git repository without odachi artifacts, or a half-configured project, keeps
// every check with its per-check fix.
func TestThePreflightOnlyFiresWhenThereIsNothing(t *testing.T) {
	cases := []struct {
		name  string
		files *fakeFiles
	}{
		{name: "only .git as a directory", files: newFakeFiles().mkdir(".git")},
		{name: "only .git as a file (a linked worktree)", files: newFakeFiles().put(".git", "gitdir: /otro/repo/.git/worktrees/rama\n")},
		{name: "only brief.md", files: newFakeFiles().put("brief.md", briefComplete)},
		{name: "only AGENTS.md", files: newFakeFiles().put("AGENTS.md", "# mi proyecto\n")},
		{name: "only .odachi/", files: newFakeFiles().mkdir(".odachi")},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := run(c.files, &fakeLookup{err: errors.New("not found")}, cleanProfiles(), "")

			if len(result.Checks) != domaindoctor.CheckCount {
				t.Fatalf("a folder with something in it gets every check, got %d:\n%s", len(result.Checks), strings.Join(result.Lines, "\n"))
			}
			if result.Lines[0] == domaindoctor.NotAProjectMessage {
				t.Fatalf("the pre-flight fired on a project that exists:\n%s", strings.Join(result.Lines, "\n"))
			}
		})
	}
}

// A read error is not a verdict. Saying "esta carpeta todavía no es un proyecto"
// about a folder odachi could not read would be a lie, so it falls through to the
// checks, which report the unreadable path as its own line.
func TestAnUnreadableFolderFallsThroughToTheChecks(t *testing.T) {
	files := newFakeFiles()
	files.errOn = at(".git")

	result := run(files, &fakeLookup{err: errors.New("not found")}, cleanProfiles(), "")

	if len(result.Checks) != domaindoctor.CheckCount {
		t.Fatalf("an unreadable folder is not an empty one, got %d checks", len(result.Checks))
	}
}

// The pre-flight is the first thing a beginner runs, and it cannot be the thing
// that leaves files behind: it asks four questions and writes nothing.
func TestThePreflightNeverWrites(t *testing.T) {
	files := newFakeFiles()
	dirsBefore := len(files.dirs)

	run(files, &fakeLookup{err: errors.New("not found")}, cleanProfiles(), "")

	if len(files.files) != 0 {
		t.Fatalf("the pre-flight wrote %v", files.files)
	}
	if len(files.dirs) != dirsBefore {
		t.Fatalf("the pre-flight created %v", files.dirs)
	}
}
