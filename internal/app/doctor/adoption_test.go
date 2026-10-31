package doctor

import (
	"errors"
	"strings"
	"testing"

	domaindoctor "github.com/biggs-100/odachi/internal/domain/doctor"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/rules"
)

// fakeGit answers the one repository question: does this folder have a commit?
type fakeGit struct {
	headErr error
}

func (g fakeGit) Head(string) (string, error) {
	if g.headErr != nil {
		return "", g.headErr
	}
	return "0123456789abcdef0123456789abcdef01234567", nil
}

// runInARepository is the doctor with the repository port wired: the fact the
// adoption is drawn from.
func runInARepository(files *fakeFiles, lookup *fakeLookup, profiles *fakeProfiles, exe string, git fakeGit) Result {
	return Run(Options{Dir: root, Home: homeRoot, Version: "0.5.0", Exe: exe, Agents: fakeAgentSelection{}, Git: git}, files, lookup, profiles)
}

// adopted is a project that adopted odachi: a repository that already lives, the
// block and the hook in the tree, and the craft and the agent in the machine. No
// agent file, no .agents/skills, and nothing recorded -because the inference is
// never stored.
func adopted(t *testing.T) *fakeFiles {
	t.Helper()
	files := configured()
	body, err := rules.Body(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Body: %v", err)
	}
	for _, file := range persona.Files(body) {
		delete(files.files, at(file.Path))
	}
	for _, file := range persona.SubagentFiles() {
		delete(files.files, at(file.Path))
	}
	return files
}

// withTheDoor puts what `odachi setup` installs into the fake home, so the machine
// really is where the agent comes from.
func withTheDoor(files *fakeFiles) *fakeFiles {
	for _, client := range persona.Clients() {
		for _, file := range persona.GlobalSubagentFilesFor(client.Agent) {
			files.putHome(file.Path, file.Content)
		}
	}
	return files
}

// adoptionLine is not needed: the notes a check carries are joined with "; " and
// this feature's sentence has its own "; " inside it, so a test asks for the words
// this note owns and for the old sentence's marker, never for a prefix.

// TestACorrectAdoptionIsNotCalledUnfinished: 0124 leaves a project with the block,
// the hook and no agent file, by design. Check 11 used to call that a gap and advise
// `init --global`, which records a choice nobody made.
func TestACorrectAdoptionIsNotCalledUnfinished(t *testing.T) {
	files := withTheDoor(adopted(t))

	result := runInARepository(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi", fakeGit{})

	if result.Code != CodeOK {
		t.Fatalf("an adoption does not break the project: %s", strings.Join(result.Lines, " / "))
	}
	got := checkOf(t, result, domaindoctor.CheckVoice)
	for _, want := range []string{"adopted odachi", "the craft and the agent live in the machine", "nothing to run"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("the note must say %q: %q", want, got.Detail)
		}
	}
	if strings.Contains(got.Detail, "does not say so") {
		t.Errorf("the note still calls the silent store a defect: %q", got.Detail)
	}
}

// TestAnAdoptionWithoutTheDoorIsToldSetup: the machine is the only place the agent
// comes from, and `setup` installs the door without recording a choice in the
// project's store.
func TestAnAdoptionWithoutTheDoorIsToldSetup(t *testing.T) {
	files := adopted(t)
	for _, file := range persona.GlobalFiles() {
		delete(files.files, atHome(file.Path))
	}

	result := runInARepository(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi", fakeGit{})

	if result.Code != CodeOK {
		t.Fatalf("a missing door does not break the project: %s", strings.Join(result.Lines, " / "))
	}
	got := checkOf(t, result, domaindoctor.CheckVoice)
	for _, want := range []string{"adopted odachi", "the door is not installed", "run odachi setup once"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("the note must say %q: %q", want, got.Detail)
		}
	}
	if strings.Contains(got.Detail, "does not say so") {
		t.Errorf("a correct adoption is still called a store that says nothing: %q", got.Detail)
	}
}

// TestARepositoryDoctorCannotReadIsNotCalledAdopted: a note that cannot tell does
// not claim an adoption. A repository git refuses is check 2's business, and this
// one says only what it sees.
func TestARepositoryDoctorCannotReadIsNotCalledAdopted(t *testing.T) {
	files := adopted(t)

	result := runInARepository(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi", fakeGit{headErr: errors.New("not a git repository")})

	got := checkOf(t, result, domaindoctor.CheckVoice)
	if strings.Contains(got.Detail, "adopted odachi") {
		t.Errorf("the note claims an adoption it could not read: %q", got.Detail)
	}
}

// TestTheBlockMustNameTheCopyTheProjectChose: the block tells every agent where the
// craft lives, and a block naming the other folder points at a copy nothing refreshes.
// The doctor asks the same scope init asks -the store, then the adoption- so a project
// that keeps the craft in the machine is told when its block still names the project's
// copy, and the warning names the command that rewrites it.
func TestTheBlockMustNameTheCopyTheProjectChose(t *testing.T) {
	files := withTheDoor(adopted(t))
	files.put("AGENTS.md", blockOf(persona.Novato)) // the block for the project's own copy

	result := runInARepository(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi", fakeGit{})

	got := checkOf(t, result, domaindoctor.CheckAgents)
	if got.Status != domaindoctor.StatusWarning {
		t.Fatalf("a block naming the project's copy passed in a project that keeps the craft in the machine: %s %s", got.Status, got.Detail)
	}
	for _, want := range []string{"~/.agents/skills", "odachi init"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("the warning must say %q: %q", want, got.Detail)
		}
	}
}
