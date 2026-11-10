package main

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/review"
)

// TestTUIWizardRunsInATerminal walks the program the way a person does: the
// first screen, every choice, the summary and the apply, read from the terminal's
// own output.
func TestTUIWizardRunsInATerminal(t *testing.T) {
	in := wizard()
	in.Models = map[string]string{"reliability": "opencode-go/x"}
	model := newTUIModel(in)
	run := &fakeRun{}
	model.run = run

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))
	waitFor(t, tm, "Proyecto")

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // state -> clients
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})  // both -> opencode
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // clients -> skills
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})  // home -> project
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // skills -> voice
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})  // novato -> senior
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // voice -> review
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})   // review -> summary
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // summary -> apply

	waitFor(t, tm, "la configuración quedó aplicada")

	if got := len(run.calls); got != 4 {
		t.Fatalf("operations = %q, want four", run.calls)
	}
	if run.clients != persona.SelectionOpenCode || run.skills != persona.SkillsProject {
		t.Errorf("install = %q/%q, want opencode/project", run.clients, run.skills)
	}
	if run.voice != "senior" || run.review != review.ModeOn {
		t.Errorf("voice = %q review = %q", run.voice, run.review)
	}

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))
}

// TestTUIWizardQuitsFromTheFirstScreen: q closes the program before any choice,
// and the terminal returns to the shell.
func TestTUIWizardQuitsFromTheFirstScreen(t *testing.T) {
	model := newTUIModel(wizard())

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))
	waitFor(t, tm, "Proyecto")

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))
}

// waitFor reads the program's output until the line appears, or fails.
func waitFor(t *testing.T, tm *teatest.TestModel, want string) {
	t.Helper()
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte(want))
	}, teatest.WithDuration(5*time.Second))
}
