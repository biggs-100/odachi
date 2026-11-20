package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// tuiPalette is odachi's own colour set, not a copy of another tool's theme.
// The conventions are gentle-ai's -an accent on the cursor, a dim tone for the
// context and the labels, amber for a warning, red for a failure- but the values
// are ours: a steel blue for the blade odachi is named after, chosen for roughly
// 4.5:1 contrast on both a light and a dark background, with the warning and the
// failure far enough from it that the eye can tell a cursor from a problem.
var tuiPalette = struct {
	accent  lipgloss.AdaptiveColor
	value   lipgloss.AdaptiveColor
	muted   lipgloss.AdaptiveColor
	warning lipgloss.AdaptiveColor
	failure lipgloss.AdaptiveColor
}{
	accent:  lipgloss.AdaptiveColor{Light: "#1d5f8a", Dark: "#7fb8e0"},
	value:   lipgloss.AdaptiveColor{Light: "#2f3b45", Dark: "#c8d1da"},
	muted:   lipgloss.AdaptiveColor{Light: "#5f6b76", Dark: "#8b98a5"},
	warning: lipgloss.AdaptiveColor{Light: "#8a5a00", Dark: "#e5b567"},
	failure: lipgloss.AdaptiveColor{Light: "#b3261e", Dark: "#f28b82"},
}

// cursorGlyph is the only mark before a row. U+203A SINGLE RIGHT-POINTING ANGLE
// QUOTATION MARK is in Windows-1252, like the "·" the help lines already ship,
// so it renders on a Windows terminal that has not been switched to UTF-8. It is
// one cell wide and is followed by one space, which keeps the two-cell gutter the
// plain "> " marker used: a label does not move when the cursor arrives.
const cursorGlyph = "›"

// The markers a notice earns. They stay one cell wide so a notice line lines up
// with the body it explains.
const (
	warningMark = "!"
	failureMark = "×"
)

// rowWidth is the label column the plain layout already used. It is part of the
// look, not of any one screen: the value column of a row starts here.
const rowWidth = 12

// tuiStyle is the visual layer every screen draws from: named pieces instead of
// inline styles, so the header of one screen cannot drift from another's and a
// test can assert the pieces instead of the bytes (0338).
type tuiStyle struct {
	title       lipgloss.Style
	subtitle    lipgloss.Style
	cursor      lipgloss.Style
	rowPlain    lipgloss.Style
	rowSelected lipgloss.Style
	value       lipgloss.Style
	factLabel   lipgloss.Style
	warning     lipgloss.Style
	failure     lipgloss.Style
	help        lipgloss.Style
}

// tuiLook is the one instance the screens use. It holds styles, not state: the
// global lipgloss profile decides at render time how much of it reaches the
// terminal, which is what lets a colourless terminal fall back to plain text.
var tuiLook = tuiStyle{
	title:       lipgloss.NewStyle().Bold(true).Foreground(tuiPalette.accent),
	subtitle:    lipgloss.NewStyle().Foreground(tuiPalette.muted),
	cursor:      lipgloss.NewStyle().Bold(true).Foreground(tuiPalette.accent),
	rowPlain:    lipgloss.NewStyle(),
	rowSelected: lipgloss.NewStyle().Bold(true),
	value:       lipgloss.NewStyle().Foreground(tuiPalette.value),
	factLabel:   lipgloss.NewStyle().Foreground(tuiPalette.muted),
	warning:     lipgloss.NewStyle().Foreground(tuiPalette.warning),
	failure:     lipgloss.NewStyle().Foreground(tuiPalette.failure),
	help:        lipgloss.NewStyle().Foreground(tuiPalette.muted),
}

// titleLine is the first line of every screen: the name, the version and the
// section, in the accent colour.
func (s tuiStyle) titleLine(title string) string {
	return s.title.Render(title)
}

// subtitleLine names the context under the title, dim so it never competes with
// it: today the project the wizard is about to touch.
func (s tuiStyle) subtitleLine(text string) string {
	return s.subtitle.Render(text)
}

// cursorPrefix is the gutter every row shares. The selected row carries the
// accent glyph; an unselected one carries the same width in spaces, so the label
// column stays exactly where it was.
func (s tuiStyle) cursorPrefix(selected bool) string {
	if selected {
		return "  " + s.cursor.Render(cursorGlyph) + " "
	}
	return "    "
}

// cell renders a row's label: bold when the cursor is on it, plain otherwise.
func (s tuiStyle) cell(selected bool, text string) string {
	if selected {
		return s.rowSelected.Render(text)
	}
	return s.rowPlain.Render(text)
}

// row renders a single-column option: the glyph when selected, then the label.
func (s tuiStyle) row(selected bool, label string) string {
	return s.cursorPrefix(selected) + s.cell(selected, label) + "\n"
}

// rowValues renders a label followed by value columns. The label keeps the
// column width the plain layout used; the first value is the piece the value
// style names, the rest stay plain so a long allowed-list is not a wall of
// colour. A padded first value keeps a second column aligned.
func (s tuiStyle) rowValues(selected bool, label string, values ...string) string {
	var out strings.Builder
	out.WriteString(s.cursorPrefix(selected))
	out.WriteString(s.cell(selected, fmt.Sprintf("%-*s", rowWidth, label)))
	for i, value := range values {
		out.WriteString(" ")
		if i == 0 {
			// An empty value is a row with nothing after its label; styling it
			// would leave an escape pair around nothing.
			if value != "" {
				out.WriteString(s.value.Render(value))
			}
			continue
		}
		out.WriteString(value)
	}
	out.WriteString("\n")
	return out.String()
}

// rowNotice is a value that is not a value: a missing model or a file that could
// not be read. It keeps the row shape and the label column, and tones the value
// as a warning instead of as a fact.
func (s tuiStyle) rowNotice(selected bool, label, value string) string {
	var out strings.Builder
	out.WriteString(s.cursorPrefix(selected))
	out.WriteString(s.cell(selected, fmt.Sprintf("%-*s", rowWidth, label)))
	out.WriteString(" ")
	out.WriteString(s.warning.Render(value))
	out.WriteString("\n")
	return out.String()
}

// subRow is one row of a nested list -the profile's value picker-: four more
// cells of gutter, then the glyph or its blank, then the label.
func (s tuiStyle) subRow(selected bool, label string) string {
	if selected {
		return "    " + s.cursor.Render(cursorGlyph) + " " + s.rowSelected.Render(label) + "\n"
	}
	return "      " + s.rowPlain.Render(label) + "\n"
}

// warningLine is a condition worth a look: the file that could not be read, the
// value that would not keep. It sits on its own line so it cannot be mistaken
// for the row it explains.
func (s tuiStyle) warningLine(text string) string {
	return "  " + s.warning.Render(warningMark+" "+text) + "\n"
}

// failureLine is a refusal or a failed operation: the marker the apply screen
// already uses for a failed step, in the failure colour.
func (s tuiStyle) failureLine(text string) string {
	return "  " + s.failure.Render(failureMark+" "+text) + "\n"
}

// successLine is a result that landed. It has no marker because it needs none:
// the eye only has to find the exceptions.
func (s tuiStyle) successLine(text string) string {
	return "  " + s.value.Render(text) + "\n"
}

// helpLine is the key footer every screen ends with. The leading newline is the
// blank line the layout convention puts between the body and the keys, so a
// screen appends it right after its last row.
func (s tuiStyle) helpLine(text string) string {
	return "\n  " + s.help.Render(text) + "\n"
}
