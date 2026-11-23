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

// limitsFrameSentinel is the text the wait uses to know the limits section drew.
// It must be text the section writes and the hub menu does not (0353).
const limitsFrameSentinel = "prompt KB"

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
		if bytes.Contains(b, []byte(limitsFrameSentinel)) {
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

// TestLimitsFrameSentinelIsSectionOnly: the hub menu already names the limits
// section, so waiting for that word can capture the menu frame before the section
// draws. The sentinel has to be text only the section writes (0353).
func TestLimitsFrameSentinelIsSectionOnly(t *testing.T) {
	hub := "instalar y configurar\n\n   › límites\n"
	if !strings.Contains(hub, "límites") {
		t.Fatal("the fixture must show the trap: the hub menu names the section")
	}
	if strings.Contains(hub, limitsFrameSentinel) {
		t.Fatalf("the sentinel %q is already in the hub frame: the wait can capture the menu", limitsFrameSentinel)
	}
	section := "límites\nprompt KB\nveredictos\nprompts\ntimeout s\n"
	if !strings.Contains(section, limitsFrameSentinel) {
		t.Fatalf("the sentinel %q is not in the limits section", limitsFrameSentinel)
	}
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

// lastFrame is the final frame in a teatest transcript. The renderer writes a
// full fresh frame each time, so the last block is what the eye read last (0343).
func lastFrame(b []byte) string {
	text := stripANSI(string(b))
	if at := strings.LastIndex(text, "odachi "); at >= 0 {
		text = text[at:]
	}
	return strings.TrimSpace(text)
}

// waitFooter reads until the last frame ends with the footer named and returns
// that frame, so a state change is observed only after it was drawn (0343).
func waitFooter(t *testing.T, tm *teatest.TestModel, want string) string {
	t.Helper()
	var frame string
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		frame = lastFrame(b)
		return strings.HasSuffix(frame, want)
	}, teatest.WithDuration(3*time.Second))
	return frame
}

// openSection walks from the hub into the section `downs` rows below it (0343).
func openSection(t *testing.T, tm *teatest.TestModel, downs int) {
	t.Helper()
	waitFor(t, tm, "Proyecto")
	for i := 0; i < downs; i++ {
		tm.Send(downKey)
	}
	tm.Send(enterKey)
}

// TestTUIFooterFramesAcrossNestedStates captures each section from the terminal
// with its own footer, with the nested control open, and after it closes. The
// frame is the evidence: the three states must read as three different footers,
// and no line may run past the column budget (0343).
func TestTUIFooterFramesAcrossNestedStates(t *testing.T) {
	const (
		profileKeys = "↑↓ elegir · enter cambiar · esc volver · q salir"
		sectionKeys = "enter cambiar · esc volver · se aplican en el resumen"
		pickerKeys  = "↑↓ elegir · enter cambiar · esc volver"
		editorKeys  = "escribir · enter cambiar · esc volver"
		checksKeys  = "enter editar · esc volver · se aplican en el resumen"
		fieldKeys   = "↑↓ campo · enter editar · esc cerrar"
	)
	cases := []struct {
		name        string
		downs       int
		defaultKeys string
		nestedKeys  string
	}{
		{"profile picker", 1, profileKeys, pickerKeys},
		{"models editor", 2, sectionKeys, editorKeys},
		{"limits editor", 3, sectionKeys, editorKeys},
		{"checks fields", 4, checksKeys, fieldKeys},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := limitedState()
			if c.name == "checks fields" {
				state = checkedState()
			}
			model := newTUIModel(state)
			model.run = &fakeRun{}
			tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))

			openSection(t, tm, c.downs)
			normal := waitFooter(t, tm, c.defaultKeys)
			t.Logf("normal frame:\n%s", normal)
			if !strings.HasSuffix(normal, c.defaultKeys) {
				t.Fatalf("normal frame does not end with the section footer:\n%s", normal)
			}

			tm.Send(enterKey) // open the picker or editor on the first row
			nested := waitFooter(t, tm, c.nestedKeys)
			t.Logf("nested frame:\n%s", nested)
			if !strings.HasSuffix(nested, c.nestedKeys) {
				t.Fatalf("nested frame does not name the open control's keys:\n%s", nested)
			}
			if strings.HasSuffix(nested, c.defaultKeys) {
				t.Errorf("nested frame still shows the section footer:\n%s", nested)
			}

			tm.Send(escKey) // close the control without leaving the section
			restored := waitFooter(t, tm, c.defaultKeys)
			t.Logf("restored frame:\n%s", restored)
			if !strings.HasSuffix(restored, c.defaultKeys) {
				t.Fatalf("closing the control did not restore the section footer:\n%s", restored)
			}
			for _, line := range strings.Split(restored, "\n") {
				if n := len([]rune(line)); n > 80 {
					t.Errorf("a line is %d cells, wider than 80: %q", n, line)
				}
			}

			tm.Send(qKey)
			tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))
		})
	}
}

// checksTerminalFrame walks from the hub into the checks section and returns the
// last frame the terminal drew, at the width asked for. It uses the same teatest
// harness as the limits frame: keys in, output out, no second rendering path (0346).
func checksTerminalFrame(t *testing.T, cols, rows int) string {
	t.Helper()
	model := newTUIModel(checkedState())
	model.run = &fakeRun{}
	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(cols, rows))
	openSection(t, tm, 4) // checks is the fifth hub row

	var captured []byte
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		if bytes.Contains(b, []byte("checks")) {
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

// TestTUChecksFrameAtTwoWidths captures the new section from the terminal at the
// normal width and a narrow one: the declared check and the severity it keeps
// stay on the row, and no line runs past the column budget (0346).
func TestTUChecksFrameAtTwoWidths(t *testing.T) {
	for _, width := range []struct {
		name string
		cols int
		rows int
	}{
		{"normal", 80, 24},
		{"narrow", 60, 20},
	} {
		t.Run(width.name, func(t *testing.T) {
			frame := checksTerminalFrame(t, width.cols, width.rows)
			for _, want := range []string{"checks", "lint", "CRITICAL"} {
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
