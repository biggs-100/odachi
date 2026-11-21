package main

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/review"
)

// ansiPattern removes the renderer's cursor moves and attributes so a captured
// transcript reads as the text a person saw. The frame is the evidence, and
// escape codes are not part of what the eye reads (0342).
var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;?]*[ -\/]*[@-~]|\x1b\][^\x07]*(?:\x07|\x1b\\)`)

func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}

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

// TestTUIWizardAsksTheScopeInATerminal: a folder without the block gets the
// setup-or-init screen before any write, and the terminal drives it too.
func TestTUIWizardAsksTheScopeInATerminal(t *testing.T) {
	in := wizard()
	in.Project = false
	model := newTUIModel(in)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))
	waitFor(t, tm, "Proyecto")

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // state -> scope
	waitFor(t, tm, "Qué configurar")

	tm.Send(tea.KeyMsg{Type: tea.KeyDown})  // machine -> project
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // scope -> clients
	waitFor(t, tm, "Clientes")

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

// limitsTerminalFrame walks from the hub into the limits section and returns the
// last frame the terminal drew, at the width asked for. The capture goes through
// the same teatest harness as the other terminal tests: keys in, output out, no
// second rendering path (0342).
func limitsTerminalFrame(t *testing.T, cols, rows int) string {
	t.Helper()
	model := newTUIModel(limitedState())
	model.run = &fakeRun{}
	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(cols, rows))
	waitFor(t, tm, "Proyecto")
	for i := 0; i < 3; i++ {
		tm.Send(downKey)
	}
	tm.Send(enterKey)

	var captured []byte
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		if bytes.Contains(b, []byte("límites")) {
			captured = append([]byte(nil), b...)
			return true
		}
		return false
	}, teatest.WithDuration(5*time.Second))

	tm.Send(qKey)
	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))

	text := stripANSI(string(captured))
	if at := strings.LastIndex(text, "odachi "); at >= 0 {
		text = text[at:]
	}
	return strings.TrimSpace(text)
}

// TestTUILimitsFrameAtTwoWidths captures the new section from the terminal at the
// normal width and a narrow one: the four rows keep their label column, and no
// line runs past the column budget (0342).
func TestTUILimitsFrameAtTwoWidths(t *testing.T) {
	for _, width := range []struct {
		name string
		cols int
		rows int
	}{
		{"normal", 80, 24},
		{"narrow", 60, 20},
	} {
		t.Run(width.name, func(t *testing.T) {
			frame := limitsTerminalFrame(t, width.cols, width.rows)
			for _, want := range []string{"límites", "prompt KB", "veredictos", "prompts", "timeout s"} {
				if !strings.Contains(frame, want) {
					t.Errorf("the frame does not show %q:\n%s", want, frame)
				}
			}
			for _, line := range strings.Split(frame, "\n") {
				if n := len([]rune(line)); n > width.cols {
					t.Errorf("a line is %d cells, wider than %d: %q", n, width.cols, line)
				}
			}
			t.Logf("frame %dx%d:\n%s", width.cols, width.rows, frame)
		})
	}
}
