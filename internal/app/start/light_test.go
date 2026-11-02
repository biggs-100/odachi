package start_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/start"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// lightOutcome is the one line a light feature carries.
const lightOutcome = "el saludo dice la versión"

// TestStartLightWritesOneLineAndTouchesNothingElse: a light feature is one
// requirement and one step with the same words, no brief, and the project's fence
// and voice stay exactly as they were.
func TestStartLightWritesOneLineAndTouchesNothingElse(t *testing.T) {
	const fence = "# the project's fence\n"
	const profile = "name = \"novato\"\n"
	files := newFakeFiles(map[string]string{
		"odd/fence.toml":   fence,
		"odd/profile.toml": profile,
	}, nil)

	got := start.Run(start.Options{
		Light:   true,
		Outcome: lightOutcome,
		Base:    "8ee3e55f31ee50a088bb1f8aa8bca4afdbc40319",
	}, nil, files, nil)

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}

	if exists, err := files.Exists("brief.md"); err != nil || exists {
		t.Errorf("a light run wrote a brief (exists=%v err=%v)", exists, err)
	}
	if got := files.content(t, "odd/fence.toml"); got != fence {
		t.Errorf("the fence was rewritten:\n%s", got)
	}
	if got := files.content(t, "odd/profile.toml"); got != profile {
		t.Errorf("the profile was rewritten:\n%s", got)
	}

	doc := taskdoc.Parse(files.content(t, got.Task))
	if !doc.Light {
		t.Error("the document does not carry the lane")
	}
	if len(doc.Requirements) != 1 || doc.Requirements[0].Text != lightOutcome {
		t.Errorf("Requirements = %+v, want the one line verbatim", doc.Requirements)
	}
	if len(doc.Steps) != 1 {
		t.Fatalf("Steps = %+v, want one: a feature with no step never closes", doc.Steps)
	}
	if step := doc.Steps[0]; step.Text != lightOutcome || step.Requirement != "R1" || step.Done {
		t.Errorf("T1 = %+v, want the same words, linked to R1 and open", step)
	}
	if doc.Base != "8ee3e55f31ee50a088bb1f8aa8bca4afdbc40319" {
		t.Errorf("Base = %q, want the commit the run read", doc.Base)
	}
}

// TestStartLightRefusesABrief: a light feature has no brief, because the line is
// the vision. A run that brings both is refused before it writes anything, and the
// answer names the flag to drop.
func TestStartLightRefusesABrief(t *testing.T) {
	files := newFakeFiles(nil, nil)

	got := start.Run(start.Options{
		Light:     true,
		Outcome:   lightOutcome,
		BriefPath: "brief.md",
	}, nil, files, nil)

	if got.Code != start.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeBlocked, got.Message)
	}
	assertOneLine(t, got.Message)
	if !strings.Contains(got.Message, "--brief") {
		t.Errorf("message %q does not name the flag to drop", got.Message)
	}
	if paths := files.paths(); len(paths) != 0 {
		t.Errorf("a refused run wrote %v, want nothing", paths)
	}
}

// TestStartLightRefusesANameThatIsAlreadyOpen: repeating the same line is the
// realistic way to arrive here, and it must not write a second document beside the
// one that still carries the change. The answer names the document on disk.
func TestStartLightRefusesANameThatIsAlreadyOpen(t *testing.T) {
	files := newFakeFiles(map[string]string{
		"odd/tasks/0001-el-saludo-dice.md": "# el-saludo-dice\n",
	}, nil)

	got := start.Run(start.Options{Light: true, Outcome: lightOutcome}, nil, files, nil)

	if got.Code != start.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeBlocked, got.Message)
	}
	assertOneLine(t, got.Message)
	if !strings.Contains(got.Message, "already open") {
		t.Errorf("message %q does not say the feature is still open", got.Message)
	}
	if !strings.Contains(got.Message, "0001-el-saludo-dice.md") {
		t.Errorf("message %q does not name the document that holds it", got.Message)
	}
	if paths := files.paths(); len(paths) != 1 {
		t.Errorf("a refused run wrote %v, want only the document that was already there", paths)
	}
}

// TestStartLightNeedsItsLine: the lane is the requirement; a run with no words to
// write says so instead of opening an empty feature.
func TestStartLightNeedsItsLine(t *testing.T) {
	files := newFakeFiles(nil, nil)

	got := start.Run(start.Options{Light: true}, nil, files, nil)

	if got.Code != start.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeBlocked, got.Message)
	}
	if !strings.Contains(got.Message, "--light") {
		t.Errorf("message %q does not name the call that works", got.Message)
	}
}
