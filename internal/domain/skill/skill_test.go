package skill_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/skill"
)

// acceptance is a complete skill in the canonical form: frontmatter with name
// and a Trigger: description, then the three sections.
const acceptance = `---
name: acceptance
description: Trigger: cuando hay que decidir si algo está terminado. Convierte un pedido en criterios observables.
---

## When

Cuando alguien dice "ya está" y no hay nada que lo pruebe.

## Rules

- Cada criterio se puede observar.
- Un pedido vago vuelve con una pregunta.

## Never

- Aceptar "funciona en mi máquina" como criterio.
`

// validSkill renders a minimal, valid skill under any name.
func validSkill(name, trigger string) string {
	return "---\nname: " + name + "\ndescription: Trigger: " + trigger +
		"\n---\n\n## When\n\ncuando\n\n## Rules\n\nregla\n\n## Never\n\nnunca\n"
}

func parse(dir, text string) skill.Skill {
	return skill.Parse(dir, "skills/"+dir+"/SKILL.md", text)
}

func joined(problems []skill.Problem) string {
	parts := make([]string, 0, len(problems))
	for _, problem := range problems {
		parts = append(parts, problem.File+": "+problem.Reason)
	}
	return strings.Join(parts, "\n")
}

// TestParseReadsTheFrontmatter: name, description, the trigger without its
// prefix, and the body with the three sections left intact.
func TestParseReadsTheFrontmatter(t *testing.T) {
	got := parse("acceptance", acceptance)

	if !got.HasFrontmatter {
		t.Fatal("HasFrontmatter = false, want true")
	}
	if got.Name != "acceptance" {
		t.Errorf("Name = %q, want acceptance", got.Name)
	}
	wantTrigger := "cuando hay que decidir si algo está terminado. Convierte un pedido en criterios observables."
	if got.Trigger != wantTrigger {
		t.Errorf("Trigger = %q, want %q", got.Trigger, wantTrigger)
	}
	if !strings.HasPrefix(got.Description, "Trigger: ") {
		t.Errorf("Description = %q, want the full description including its prefix", got.Description)
	}
	for _, section := range []string{"## When", "## Rules", "## Never"} {
		if !strings.Contains(got.Body, section) {
			t.Errorf("Body lost %q:\n%s", section, got.Body)
		}
	}
	if strings.Contains(got.Body, "---") {
		t.Errorf("Body still carries the frontmatter:\n%s", got.Body)
	}
	if problems := got.Problems(); len(problems) != 0 {
		t.Errorf("a complete skill has problems: %s", joined(problems))
	}
}

// TestParseWithoutFrontmatterIsOneReason: a file that never declares itself is
// not guessed at, and the reason names the missing header.
func TestParseWithoutFrontmatterIsOneReason(t *testing.T) {
	got := parse("suelta", "## When\n\ncuando\n\n## Rules\n\nregla\n\n## Never\n\nnunca\n")

	if got.HasFrontmatter {
		t.Fatal("HasFrontmatter = true for a file without ---")
	}
	problems := got.Problems()
	if len(problems) != 1 {
		t.Fatalf("problems = %d (%s), want exactly one about the frontmatter", len(problems), joined(problems))
	}
	if !strings.Contains(problems[0].Reason, "frontmatter") {
		t.Errorf("Reason = %q, want it to name the frontmatter", problems[0].Reason)
	}
	if problems[0].File != "skills/suelta/SKILL.md" {
		t.Errorf("File = %q, want the skill path", problems[0].File)
	}
}

// TestParseAcceptsCRLF: a Windows checkout validates exactly like a Unix one.
func TestParseAcceptsCRLF(t *testing.T) {
	crlf := strings.ReplaceAll(acceptance, "\n", "\r\n")

	got := parse("acceptance", crlf)

	if got.Name != "acceptance" {
		t.Errorf("Name = %q, want acceptance", got.Name)
	}
	if got.Trigger == "" {
		t.Error("Trigger is empty for a CRLF file")
	}
	if problems := got.Problems(); len(problems) != 0 {
		t.Errorf("CRLF file has problems: %s", joined(problems))
	}
}

// TestDescriptionWithoutTriggerIsRefused: without the prefix nobody knows when
// to load the skill.
func TestDescriptionWithoutTriggerIsRefused(t *testing.T) {
	text := strings.Replace(acceptance,
		"description: Trigger: cuando hay que decidir si algo está terminado. Convierte un pedido en criterios observables.",
		"description: decide si algo está terminado", 1)

	problems := parse("acceptance", text).Problems()

	if !strings.Contains(joined(problems), `does not start with "Trigger:"`) {
		t.Errorf("problems = %q, want the Trigger: reason", joined(problems))
	}
}

// TestNameMustMatchTheDirectory: the directory is how the human finds it, so the
// declared name cannot disagree with it.
func TestNameMustMatchTheDirectory(t *testing.T) {
	text := strings.Replace(acceptance, "name: acceptance", "name: otro", 1)

	problems := parse("acceptance", text).Problems()

	if !strings.Contains(joined(problems), `does not match the directory`) {
		t.Errorf("problems = %q, want the directory mismatch reason", joined(problems))
	}
}

// TestNameMustMatchThePattern pins the slug shape, lowercase and 3..41 chars.
func TestNameMustMatchThePattern(t *testing.T) {
	text := strings.Replace(acceptance, "name: acceptance", "name: Acceptance", 1)

	problems := parse("Acceptance", text).Problems()

	if !strings.Contains(joined(problems), "does not match the shape") {
		t.Errorf("problems = %q, want the pattern reason", joined(problems))
	}
}

// TestBodyOverSixtyLines: a skill that needs more than a screen is not a skill.
func TestBodyOverSixtyLines(t *testing.T) {
	lines := []string{"## When", "cuando", "## Rules"}
	for len(lines) < 59 {
		lines = append(lines, "- regla")
	}
	lines = append(lines, "## Never", "nunca")
	body := strings.Join(lines, "\n")
	if got := len(strings.Split(body, "\n")); got != 61 {
		t.Fatalf("the fixture has %d body lines, want 61", got)
	}

	text := "---\nname: largo\ndescription: Trigger: una skill larga\n---\n\n" + body + "\n"
	problems := parse("largo", text).Problems()

	if !strings.Contains(joined(problems), "61 lines") || !strings.Contains(joined(problems), "maximum is 60") {
		t.Errorf("problems = %q, want the 61-line reason", joined(problems))
	}
}

// TestTodoIsRefused: a shipped skill is complete; TODO is a note to the author.
func TestTodoIsRefused(t *testing.T) {
	text := strings.Replace(acceptance, "- Cada criterio se puede observar.", "- TODO: escribir los criterios.", 1)

	problems := parse("acceptance", text).Problems()

	if !strings.Contains(joined(problems), "TODO") {
		t.Errorf("problems = %q, want the TODO reason", joined(problems))
	}
}

// TestEmptySectionIsRefused: the three sections are the format, and each one
// carries content.
func TestEmptySectionIsRefused(t *testing.T) {
	text := "---\nname: vacia\ndescription: Trigger: una skill con un hueco\n---\n\n" +
		"## When\n\ncuando\n\n## Rules\n\n## Never\n\nnunca\n"

	problems := parse("vacia", text).Problems()

	if !strings.Contains(joined(problems), `"Rules"`) {
		t.Errorf("problems = %q, want the empty Rules reason", joined(problems))
	}
}

// TestMissingSectionIsRefused: absence is emptiness, and the reason names it.
func TestMissingSectionIsRefused(t *testing.T) {
	text := "---\nname: corta\ndescription: Trigger: una skill sin Nunca\n---\n\n" +
		"## When\n\ncuando\n\n## Rules\n\nregla\n"

	problems := parse("corta", text).Problems()

	if !strings.Contains(joined(problems), `"Never"`) {
		t.Errorf("problems = %q, want the missing Never reason", joined(problems))
	}
}

// TestDuplicateNamesAreRefused: a name is a handle, and two skills with one
// handle make the registry useless.
func TestDuplicateNamesAreRefused(t *testing.T) {
	skills := []skill.Skill{
		parse("uno", validSkill("uno", "primero")),
		parse("dos", validSkill("uno", "segundo")),
	}

	problems := skill.Check(skills, skill.Registry(skills), true)

	if !strings.Contains(joined(problems), "repeated") {
		t.Errorf("problems = %q, want the duplicate-name reason", joined(problems))
	}
	if !strings.Contains(joined(problems), "skills/dos/SKILL.md") {
		t.Errorf("problems = %q, want it to name the second file", joined(problems))
	}
}

// TestRegistryIsSortedAndExact pins the generated shape: one marker line, then
// the table by name, and nothing else.
func TestRegistryIsSortedAndExact(t *testing.T) {
	skills := []skill.Skill{
		parse("ship", validSkill("ship", "cuándo se publica")),
		parse("acceptance", validSkill("acceptance", "cuándo se acepta")),
	}

	got := skill.Registry(skills)

	want := "<!-- generated by odachi skills --write; do not edit it by hand -->\n" +
		"| Skill | Trigger | Path |\n" +
		"| --- | --- | --- |\n" +
		"| acceptance | cuándo se acepta | skills/acceptance/SKILL.md |\n" +
		"| ship | cuándo se publica | skills/ship/SKILL.md |\n"
	if got != want {
		t.Errorf("Registry =\n%s\nwant\n%s", got, want)
	}
}

// TestRegistryMustBeExactlyWhatWriteGenerates: a hand-edited registry is an
// error that names --write, never a merge.
func TestRegistryMustBeExactlyWhatWriteGenerates(t *testing.T) {
	skills := []skill.Skill{parse("acceptance", acceptance)}

	handEdited := skill.Registry(skills) + "| inventada | a mano | skills/inventada/SKILL.md |\n"
	problems := skill.Check(skills, handEdited, true)
	if !strings.Contains(joined(problems), "SKILLS.md") || !strings.Contains(joined(problems), "--write") {
		t.Errorf("problems = %q, want the stale-registry reason naming --write", joined(problems))
	}

	missing := skill.Check(skills, "", false)
	if !strings.Contains(joined(missing), "SKILLS.md is missing") {
		t.Errorf("problems = %q, want the missing-registry reason", joined(missing))
	}
}

// TestCheckIsGreen: valid skills plus their own registry have nothing to report.
func TestCheckIsGreen(t *testing.T) {
	skills := []skill.Skill{
		parse("acceptance", validSkill("acceptance", "cuándo se acepta")),
		parse("ship", validSkill("ship", "cuándo se publica")),
	}

	if problems := skill.Check(skills, skill.Registry(skills), true); len(problems) != 0 {
		t.Errorf("Check reported %s, want nothing", joined(problems))
	}
}

// TestValidSlug pins the one shape both `skill new` and the name rule use.
func TestValidSlug(t *testing.T) {
	cases := map[string]bool{
		"ejemplo":                     true,
		"a-b-c":                       true,
		"abc":                         true,
		"ab":                          false,
		"Ejemplo":                     false,
		"1abc":                        false,
		"a_b":                         false,
		"con espacio":                 false,
		"":                            false,
		"a" + strings.Repeat("b", 40): true,
		"a" + strings.Repeat("b", 41): false,
	}
	for slug, want := range cases {
		if got := skill.ValidSlug(slug); got != want {
			t.Errorf("ValidSlug(%q) = %v, want %v", slug, got, want)
		}
	}
}

// TestTemplateIsAValidStartingPoint: `skill new` must leave something --check
// accepts, with the three sections ready to fill.
func TestTemplateIsAValidStartingPoint(t *testing.T) {
	text := skill.Template("ejemplo")

	got := parse("ejemplo", text)
	if problems := got.Problems(); len(problems) != 0 {
		t.Errorf("the template is not valid: %s\n%s", joined(problems), text)
	}
	if got.Name != "ejemplo" {
		t.Errorf("Name = %q, want ejemplo", got.Name)
	}
	for _, section := range []string{"## When", "## Rules", "## Never"} {
		if !strings.Contains(text, section) {
			t.Errorf("the template has no %q:\n%s", section, text)
		}
	}
	if strings.Contains(text, "TODO") {
		t.Errorf("the template contains TODO:\n%s", text)
	}
}
