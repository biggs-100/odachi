package main

import "strings"

// The marks a clipped list writes. They sit on their own line, in the layer's
// muted voice and outside the four-cell row gutter, because they describe the
// window and not a row: a mark inside the gutter would look like a selectable
// entry (0338).
const (
	clipAbove = "↑ más"
	clipBelow = "↓ más"
)

// The most rows each screen draws before the window takes over. A terminal has
// to fit the title, the subtitle, the body and the key footer; eight rows leaves
// the header and the footer room to breathe on an 80x24 screen, which is the
// smallest one worth drawing. The hub, the choice screens, the review rows and
// the profile axes are shorter than that today, so the cap is invisible there;
// the models rows and the profile's value picker are the lists that grow, and
// they get a number instead of an unbounded loop (0338).
const (
	hubRows     = 8 // three sections today; the cap guards a fourth
	choiceRows  = 8 // the longest single-choice list is the voice table
	reviewRows  = 8 // the switch and the four lenses
	modelsRows  = 8 // the verifier plus one row per detected client
	profileRows = 8 // the six axes plus the apply row
	// valueRows is smaller on purpose: the picker is nested under its axis row,
	// so the axis window and the picker window share one screen.
	valueRows = 5
)

// tuiWindow is the slice of a list a screen draws, and whether anything is
// hidden above or below it. start is inclusive, end exclusive.
type tuiWindow struct {
	start int
	end   int
	above bool
	below bool
}

// windowFor answers which slice of a list is visible. It is pure and total: the
// caller passes the cursor, the length and the maximum of rows, and every corner
// -an empty list, a cursor outside the range, a maximum of zero, a list exactly
// as long as the window- has an answer instead of a panic.
//
// The window keeps the cursor on its last visible row once the list is longer
// than the window. That is the one offset a pure function can choose without
// keeping scroll state: moving down past the last visible row scrolls by one,
// moving up past the first scrolls back by one, and jumping to either end shows
// that end. A window that centred the cursor would scroll on every step, and a
// stateful one would need a second field the screens do not have (0338).
func windowFor(cursor, length, maxVisible int) tuiWindow {
	if length <= 0 {
		return tuiWindow{}
	}
	// A window with no rows is a screen with no rows. One row is the smallest
	// answer that keeps the promise that the cursor is always drawn.
	if maxVisible < 1 {
		maxVisible = 1
	}
	cursor = clamp(cursor, length)
	if length <= maxVisible {
		return tuiWindow{start: 0, end: length}
	}

	start := cursor - maxVisible + 1
	if start < 0 {
		start = 0
	}
	end := start + maxVisible
	if end > length {
		end = length
	}
	return tuiWindow{start: start, end: end, above: start > 0, below: end < length}
}

// windowLines draws one window through the layer: the mark for rows hidden
// above, the rows the window keeps, and the mark for the rows hidden below. The
// draw callback renders index i in the screen's own row shape, so every list
// keeps its cursor and its value columns while the window decides the extent. A
// list that fits writes neither mark, which is what keeps its bytes exactly the
// ones it drew before it had a window (0338).
func (s tuiStyle) windowLines(w tuiWindow, draw func(index int) string) string {
	var out strings.Builder
	if w.above {
		out.WriteString(s.mutedLine(clipAbove))
	}
	for i := w.start; i < w.end; i++ {
		out.WriteString(draw(i))
	}
	if w.below {
		out.WriteString(s.mutedLine(clipBelow))
	}
	return out.String()
}
