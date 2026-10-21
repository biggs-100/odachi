package taskdoc_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// TestUnaccountedNamesWhatTheCanonicalRenderCannotCarry: the write guard needs
// to know which lines Parse ignores, because those are the ones a rewrite would
// destroy. A malformed line inside a canonical section is unaccounted; the
// history never is, because it is kept verbatim.
func TestUnaccountedNamesWhatTheCanonicalRenderCannotCarry(t *testing.T) {
	md := "# reserva-de-turno\n\n" +
		"## Requisitos (textual, del usuario)\n" +
		"R1 · \"reservar turnos\"\n" +
		"esto no es un requisito\n\n" +
		"## Notas\n" +
		"una idea que el humano agregó a mano\n\n" +
		"## Listo cuando\n" +
		"- Reservo un turno (R1)\n" +
		"una línea suelta\n\n" +
		"## Pasos\n" +
		"- [ ] T1 · horario libre   R1\n" +
		"basura entre pasos\n\n" +
		"## Historial\n" +
		"L1 · \"reservar turnos\"\n" +
		"cualquier cosa escrita a mano\n"

	got := taskdoc.Unaccounted(md)
	want := []string{
		"esto no es un requisito",
		"## Notas",
		"una idea que el humano agregó a mano",
		"una línea suelta",
		"basura entre pasos",
	}

	if !slices.Equal(got, want) {
		t.Fatalf("Unaccounted =\n%v\nwant\n%v", got, want)
	}
}

// TestUnaccountedOnACanonicalDocumentIsEmpty: the documents the tool writes
// itself have nothing to lose, so the guard never blocks an ordinary rewrite.
func TestUnaccountedOnACanonicalDocumentIsEmpty(t *testing.T) {
	doc := taskdoc.TaskDoc{
		Title:        "reserva-de-turno",
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "reservar turnos"}},
		DoneWhen:     []string{"Reservo un turno (R1)"},
		Steps: []taskdoc.Step{
			{ID: "T1", Text: "horario libre", Requirement: "R1", Commit: "4a1c9f", Done: true},
			{ID: "T2", Text: "agenda del día", Requirement: "R2"},
		},
		History: []string{`L1 · "reservar turnos"`, "L2 · una nota suelta"},
	}

	if got := taskdoc.Unaccounted(doc.Render()); len(got) != 0 {
		t.Fatalf("Unaccounted(canonical) = %v, want nothing", got)
	}
}

// TestUnaccountedFlagsTheTitleThatWouldBeLost: Parse keeps the last title, so
// an earlier one is content the canonical render would drop.
func TestUnaccountedFlagsTheTitleThatWouldBeLost(t *testing.T) {
	md := "# primero\n\n## Historial\nL1 · \"x\"\n\n# segundo\n"

	got := taskdoc.Unaccounted(md)
	if !slices.Equal(got, []string{"# primero"}) {
		t.Fatalf("Unaccounted = %v, want the first title (the one Parse drops)", got)
	}
	if strings.Contains(strings.Join(got, "\n"), "# segundo") {
		t.Error("the last title is the one that survives and must not be reported")
	}
}
