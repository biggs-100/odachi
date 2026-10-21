package brief_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/biggz/odachi/internal/domain/brief"
)

// example reads the canonical brief artifact example. The path is relative to
// this package: internal/domain/brief -> module root -> testdata/examples.
func example(t *testing.T) string {
	t.Helper()
	path := filepath.Join("..", "..", "..", "testdata", "examples", "brief.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func TestParseCanonicalBrief(t *testing.T) {
	got := brief.Parse(example(t))

	if got.Name != "turnos-consultorio" {
		t.Errorf("Name = %q, want %q", got.Name, "turnos-consultorio")
	}

	counts := []struct {
		field string
		got   int
		want  int
	}{
		{"Outcome", len(got.Outcome), 4},
		{"DoneWhen", len(got.DoneWhen), 5},
		{"NonGoals", len(got.NonGoals), 5},
		{"Decisions", len(got.Decisions), 5},
	}
	for _, c := range counts {
		if c.got != c.want {
			t.Errorf("len(%s) = %d, want %d", c.field, c.got, c.want)
		}
	}

	if len(got.Outcome) > 0 && got.Outcome[0] != "Un paciente reserva un turno con un médico en menos de 1 minuto desde el teléfono." {
		t.Errorf("Outcome[0] = %q", got.Outcome[0])
	}
	if len(got.NonGoals) > 0 && got.NonGoals[2] != "Aplicación móvil nativa." {
		t.Errorf("NonGoals[2] = %q", got.NonGoals[2])
	}
	if len(got.Decisions) > 0 && got.Decisions[4] != "Los datos de los pacientes no salen del servidor del consultorio." {
		t.Errorf("Decisions[4] = %q", got.Decisions[4])
	}

	if missing := got.Missing(); len(missing) != 0 {
		t.Errorf("Missing() = %q, want none", missing)
	}
}

func TestMissingReportsEveryRequiredPartOfAHeaderOnlyBrief(t *testing.T) {
	got := brief.Parse("# turnos-consultorio\n")

	if got.Name != "turnos-consultorio" {
		t.Errorf("Name = %q, want %q", got.Name, "turnos-consultorio")
	}
	want := []string{"qué tiene que existir", "listo cuando"}
	if missing := got.Missing(); !reflect.DeepEqual(missing, want) {
		t.Errorf("Missing() = %q, want %q", missing, want)
	}
}

func TestMissingReportsOnlyTheEmptyPart(t *testing.T) {
	md := `# demo

## Qué tiene que existir
- algo

## NO incluye (primera versión)
- nada
`
	got := brief.Parse(md)
	want := []string{"listo cuando"}
	if missing := got.Missing(); !reflect.DeepEqual(missing, want) {
		t.Errorf("Missing() = %q, want %q", missing, want)
	}
}

func TestParseMalformedInputNeverPanicsAndYieldsZeroBrief(t *testing.T) {
	inputs := map[string]string{
		"empty":            "",
		"blank lines":      "\n\n\n",
		"header only":      "#\n",
		"deep header":      "### no title\n",
		"orphan bullet":    "- sin sección\n",
		"section no items": "## Listo cuando\n",
		"empty bullets":    "## Qué tiene que existir\n-\n- \n",
		"no newline":       "## NO incluye (primera versión)",
	}
	for name, md := range inputs {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Parse panicked: %v", r)
				}
			}()
			got := brief.Parse(md)
			if !reflect.DeepEqual(got, brief.Brief{}) {
				t.Errorf("Parse(%q) = %+v, want zero Brief", md, got)
			}
		})
	}
}

func TestParseIgnoresUnknownSections(t *testing.T) {
	md := `# demo

## Notas internas
- esto no es parte del brief

## Qué tiene que existir
- algo

## Listo cuando
- algo observable

## NO incluye (primera versión)
- otra cosa
`
	got := brief.Parse(md)

	if !reflect.DeepEqual(got.Outcome, []string{"algo"}) {
		t.Errorf("Outcome = %q, want [algo]", got.Outcome)
	}
	if !reflect.DeepEqual(got.DoneWhen, []string{"algo observable"}) {
		t.Errorf("DoneWhen = %q, want [algo observable]", got.DoneWhen)
	}
	if !reflect.DeepEqual(got.NonGoals, []string{"otra cosa"}) {
		t.Errorf("NonGoals = %q, want [otra cosa]", got.NonGoals)
	}
	if len(got.Decisions) != 0 {
		t.Errorf("Decisions = %q, want none", got.Decisions)
	}
}

func TestParseAcceptsAnySectionOrderAndCRLF(t *testing.T) {
	md := "# demo\r\n\r\n## Decisiones tomadas\r\n- decidido\r\n\r\n## Listo cuando\r\n- listo\r\n\r\n## Qué tiene que existir\r\n- existe\r\n\r\n## NO incluye (primera versión)\r\n- afuera\r\n"
	got := brief.Parse(md)

	if got.Name != "demo" {
		t.Errorf("Name = %q, want demo", got.Name)
	}
	if !reflect.DeepEqual(got.Outcome, []string{"existe"}) {
		t.Errorf("Outcome = %q, want [existe]", got.Outcome)
	}
	if !reflect.DeepEqual(got.DoneWhen, []string{"listo"}) {
		t.Errorf("DoneWhen = %q, want [listo]", got.DoneWhen)
	}
	if !reflect.DeepEqual(got.NonGoals, []string{"afuera"}) {
		t.Errorf("NonGoals = %q, want [afuera]", got.NonGoals)
	}
	if !reflect.DeepEqual(got.Decisions, []string{"decidido"}) {
		t.Errorf("Decisions = %q, want [decidido]", got.Decisions)
	}
}
