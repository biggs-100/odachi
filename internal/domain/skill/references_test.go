package skill_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/skill"
)

// routerBase is a skill that satisfies the three sections every skill carries, so
// a test about references fails only for the reference rule it names.
const routerBase = `## When

cuando hay que elegir.

## Rules

- una regla.

## Never

- nada.
`

func code(path string) string { return "`" + path + "`" }

func bodyWith(parts ...string) string {
	return routerBase + "\n" + strings.Join(parts, "\n\n")
}

// referencesSection renders the inventory the grammar asks for.
func referencesSection(paths ...string) string {
	lines := []string{"## References", ""}
	for _, path := range paths {
		lines = append(lines, "- "+code(path)+" — what it holds.")
	}
	return strings.Join(lines, "\n")
}

// gatesSection renders the router table: one row per situation, the path in
// backticks.
func gatesSection(paths ...string) string {
	lines := []string{"## Decision gates", "", "| Situation | Load |", "| --- | --- |"}
	for _, path := range paths {
		lines = append(lines, "| a "+path+" question | "+code(path)+" |")
	}
	return strings.Join(lines, "\n")
}

const outputSection = "## Output\n\nthe file and the tradeoff it makes."

// router parses a skill under the name router and returns it with the files the
// folder holds, the way the loader fills them.
func router(references []string, sections ...string) skill.Skill {
	s := parse("router", "---\nname: router\ndescription: Trigger: cuando hay que enrutar\n---\n\n"+bodyWith(sections...))
	s.References = references
	return s
}

func wants(t *testing.T, problems []skill.Problem, text string) {
	t.Helper()
	for _, problem := range problems {
		if strings.Contains(problem.Reason, text) {
			return
		}
	}
	t.Errorf("no problem names %q; problems: %s", text, joined(problems))
}

// TestACoherentRouterHasNoProblems: the grammar accepts a skill that lists every
// file, gates every file, and says what it returns.
func TestACoherentRouterHasNoProblems(t *testing.T) {
	s := router(
		[]string{"references/motion.md", "references/numbers.md"},
		gatesSection("references/numbers.md", "references/motion.md"),
		outputSection,
		referencesSection("references/numbers.md", "references/motion.md"),
	)

	if problems := s.Problems(); len(problems) != 0 {
		t.Errorf("a coherent router has problems: %s", joined(problems))
	}
}

// TestAReferenceTheSkillNamesMustExist: a path in the body is a claim about the
// folder, and a claim without a file is refused.
func TestAReferenceTheSkillNamesMustExist(t *testing.T) {
	s := router(nil, outputSection, referencesSection("references/numbers.md"))

	wants(t, s.Problems(), "does not exist")
}

// TestAFileUnderReferencesMustBeListed: a file nobody declares is depth the
// agent will never load, so the folder and the skill must agree.
func TestAFileUnderReferencesMustBeListed(t *testing.T) {
	s := router(
		[]string{"references/numbers.md"},
		outputSection,
		"## References\n\n- something else.",
	)

	wants(t, s.Problems(), "is not a bullet of ## References")
}

// TestAReferenceMustBeAMarkdown: references are the text an agent reads, not a
// place for binaries.
func TestAReferenceMustBeAMarkdown(t *testing.T) {
	s := router(
		[]string{"references/logo.png"},
		outputSection,
		referencesSection("references/logo.png"),
	)

	wants(t, s.Problems(), "is not a .md")
}

// TestARouterNeedsDecisionGates: more than one reference is a choice, and a
// choice without a table is the agent guessing.
func TestARouterNeedsDecisionGates(t *testing.T) {
	s := router(
		[]string{"references/motion.md", "references/numbers.md"},
		outputSection,
		referencesSection("references/motion.md", "references/numbers.md"),
	)

	wants(t, s.Problems(), "Decision gates")
}

// TestEveryReferenceIsAGateRow: a table that forgets one reference leaves it
// unreachable from the router.
func TestEveryReferenceIsAGateRow(t *testing.T) {
	s := router(
		[]string{"references/motion.md", "references/numbers.md"},
		gatesSection("references/numbers.md"),
		outputSection,
		referencesSection("references/motion.md", "references/numbers.md"),
	)

	wants(t, s.Problems(), "is not a row of ## Decision gates")
}

// TestASkillWithReferencesDeclaresItsOutput: a skill that carries depth has to
// say what the agent hands back after reading it.
func TestASkillWithReferencesDeclaresItsOutput(t *testing.T) {
	s := router(
		[]string{"references/numbers.md"},
		referencesSection("references/numbers.md"),
	)

	wants(t, s.Problems(), `"Output"`)
}

// TestAProseSentenceIsNotAReferenceBullet: the inventory is a list, one line per
// file; a sentence that names the path declares nothing (0207).
func TestAProseSentenceIsNotAReferenceBullet(t *testing.T) {
	s := router(
		[]string{"references/numbers.md"},
		"## References\n\nThe scale lives in `references/numbers.md`, read it first.",
		outputSection,
	)

	wants(t, s.Problems(), "is not a bullet")
}

// TestAProseSentenceIsNotAGateRow: the router table is a table; a path named in a
// sentence routes nothing (0207).
func TestAProseSentenceIsNotAGateRow(t *testing.T) {
	s := router(
		[]string{"references/motion.md", "references/numbers.md"},
		"## Decision gates\n\nLoad `references/motion.md` for motion and `references/numbers.md` for numbers.",
		outputSection,
		referencesSection("references/motion.md", "references/numbers.md"),
	)

	wants(t, s.Problems(), "is not a row of ## Decision gates")
}

// TestAReferenceInTheHeaderIsNotARow: the header names the columns, not a route.
func TestAReferenceInTheHeaderIsNotARow(t *testing.T) {
	s := router(
		[]string{"references/motion.md", "references/numbers.md"},
		"## Decision gates\n\n| Situation | Load `references/motion.md` |\n| --- | --- |\n| the numbers | `references/numbers.md` |",
		outputSection,
		referencesSection("references/motion.md", "references/numbers.md"),
	)

	wants(t, s.Problems(), "is not a row of ## Decision gates")
}

// TestASkillWithoutReferencesStillOwesItsOutput: the contract belongs to every
// skill, not only to one that carries depth. A reference is an addition; it is
// never the reason the agent has to be told what it hands back (0206).
func TestASkillWithoutReferencesStillOwesItsOutput(t *testing.T) {
	s := router(nil)

	wants(t, s.Problems(), `"Output"`)
}
