package taskdoc_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// TestParseRecordsTheLane: the document says which lane it was opened in, so the
// gate reads the shape where it reads every other decision.
func TestParseRecordsTheLane(t *testing.T) {
	md := `# un cambio chico
Base: 8ee3e55f31ee50a088bb1f8aa8bca4afdbc40319
Light: yes

## Requirements (verbatim, from the user)
R1 · "el saludo dice la versión"

## Steps
- [ ] T1 · el saludo dice la versión   R1
`
	got := taskdoc.Parse(md)
	if !got.Light {
		t.Fatalf("Light = false, want true: the preamble line was not read")
	}
	if lost := taskdoc.Unaccounted(md); len(lost) != 0 {
		t.Errorf("the write guard flags the lane line: %v", lost)
	}
	if again := taskdoc.Parse(got.Render()); !reflect.DeepEqual(again, got) {
		t.Errorf("render then parse changed the document:\ngot  %+v\nwant %+v\nrendered:\n%s",
			again, got, got.Render())
	}
	if back := taskdoc.Parse(got.Render()); !back.Light {
		t.Error("the canonical form lost the lane")
	}
}

// TestADocumentWithoutTheLaneIsUnchanged: every document written before the lane
// existed says nothing about it, and that is not the same as saying "light".
func TestADocumentWithoutTheLaneIsUnchanged(t *testing.T) {
	md := `# turnos
Base: 8ee3e55f31ee50a088bb1f8aa8bca4afdbc40319

## Requirements (verbatim, from the user)
R1 · "reservar turnos"

## Steps
- [ ] T1 · reservar   R1
`
	if got := taskdoc.Parse(md); got.Light {
		t.Error("Light = true, want false: no line means the normal lane")
	}
}

// TestTheWriteGuardKeepsThePreambleDecisions: the guard exists to refuse a
// rewrite that would drop a line. The decisions Parse reads in the preamble are
// carried by Render, so they are not lost -and a guard that flags them refuses
// every later write to the document-.
func TestTheWriteGuardKeepsThePreambleDecisions(t *testing.T) {
	md := `# un cambio chico
Base: 8ee3e55f31ee50a088bb1f8aa8bca4afdbc40319
Review: waived
Light: yes

## Requirements (verbatim, from the user)
R1 · "el saludo dice la versión"

## Steps
- [ ] T1 · el saludo dice la versión   R1
`
	if lost := taskdoc.Unaccounted(md); len(lost) != 0 {
		t.Errorf("the write guard would refuse a rewrite of a canonical document: %v", lost)
	}
}

// TestTheLaneOutsideThePreambleIsProse: the field lives between the title and
// the first section, like the base and the waiver. A note later in the document
// that says "Light: yes" is something a human wrote, not a decision the gate must
// obey.
func TestTheLaneOutsideThePreambleIsProse(t *testing.T) {
	md := `# turnos
Base: 8ee3e55f31ee50a088bb1f8aa8bca4afdbc40319

## Notas
Light: yes

## Steps
- [ ] T1 · reservar   R1
`
	if got := taskdoc.Parse(md); got.Light {
		t.Error("Light = true for a line in a note: the gate would obey prose")
	}
	var flagged bool
	for _, line := range taskdoc.Unaccounted(md) {
		if strings.TrimSpace(line) == "Light: yes" {
			flagged = true
		}
	}
	if !flagged {
		t.Error("the write guard does not flag the lane line a rewrite would move")
	}
}
