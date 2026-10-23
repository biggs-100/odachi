package brief_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/brief"
)

// TestRenderUsesTheCanonicalHeaderOrder pins the exact bytes of a rendered
// brief: the four canonical English headers, in the order a human reads them,
// one bullet per line, and nothing else.
func TestRenderUsesTheCanonicalHeaderOrder(t *testing.T) {
	b := brief.Brief{
		Name:      "demo",
		Outcome:   []string{"algo que existe", "otra cosa que existe"},
		DoneWhen:  []string{"lo puedo ver funcionando"},
		NonGoals:  []string{"pagos", "login"},
		Decisions: []string{"el celular primero"},
	}

	want := `# demo

## What must exist
- algo que existe
- otra cosa que existe

## Done when
- lo puedo ver funcionando

## Not included (first version)
- pagos
- login

## Decisions taken
- el celular primero
`

	if got := brief.Render(b); got != want {
		t.Errorf("Render() =\n%s\nwant\n%s", got, want)
	}
}

// TestRenderRoundTrips is the contract that makes Render safe to write to disk:
// what Parse reads back is the same Brief the caller handed in.
func TestRenderRoundTrips(t *testing.T) {
	b := brief.Brief{
		Name:      "demo",
		Outcome:   []string{"algo"},
		DoneWhen:  []string{"listo"},
		NonGoals:  []string{"nada"},
		Decisions: []string{"decidido"},
	}

	got := brief.Parse(brief.Render(b))
	if !reflect.DeepEqual(got, b) {
		t.Errorf("Parse(Render(b)) = %+v, want %+v", got, b)
	}
}

// TestRenderRoundTripsTheCanonicalExample proves Render against the real
// artifact a human reads, not only against a synthetic one.
func TestRenderRoundTripsTheCanonicalExample(t *testing.T) {
	parsed := brief.Parse(example(t))

	got := brief.Parse(brief.Render(parsed))
	if !reflect.DeepEqual(got, parsed) {
		t.Errorf("Parse(Render(Parse(example))) = %+v, want %+v", got, parsed)
	}
}

// TestRenderOmitsAnEmptyTitle keeps a brief without a name parseable: a
// dangling "# " line is not a title, so writing it would lose nothing while
// looking broken to the human.
func TestRenderOmitsAnEmptyTitle(t *testing.T) {
	got := brief.Render(brief.Brief{Outcome: []string{"algo"}})

	if strings.HasPrefix(got, "# ") {
		t.Errorf("Render with an empty Name started with a title: %q", got)
	}
	if !strings.HasPrefix(got, "## What must exist") {
		t.Errorf("Render with an empty Name should start with the first section, got %q", got)
	}
	if parsed := brief.Parse(got); parsed.Name != "" {
		t.Errorf("Parse(Render(...)).Name = %q, want empty", parsed.Name)
	}
}

// TestRenderOfAnEmptyBriefRoundTrips keeps the zero value safe: a human who
// deletes every bullet still gets a file the tool can read.
func TestRenderOfAnEmptyBriefRoundTrips(t *testing.T) {
	got := brief.Render(brief.Brief{})

	for _, header := range []string{
		"## What must exist",
		"## Done when",
		"## Not included (first version)",
		"## Decisions taken",
	} {
		if !strings.Contains(got, header) {
			t.Errorf("Render(Brief{}) is missing %q:\n%s", header, got)
		}
	}

	if parsed := brief.Parse(got); !reflect.DeepEqual(parsed, brief.Brief{}) {
		t.Errorf("Parse(Render(Brief{})) = %+v, want the zero Brief", parsed)
	}
}

// TestSpanishHeadersStillParse: the language migration changed what Render
// writes, not what Parse accepts. A brief written before the migration stays
// readable.
func TestSpanishHeadersStillParse(t *testing.T) {
	got := brief.Parse(`# viejo

## Qué tiene que existir
- algo

## Listo cuando
- listo

## NO incluye (primera versión)
- nada

## Decisiones tomadas
- decidido
`)

	if len(got.Outcome) != 1 || len(got.DoneWhen) != 1 || len(got.NonGoals) != 1 || len(got.Decisions) != 1 {
		t.Errorf("Parse(old brief) = %+v, want all four sections", got)
	}
}
