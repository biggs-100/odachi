package archtest_test

import (
	"os"
	"path/filepath"
	"testing"

	assets "github.com/biggs-100/odachi/internal/assets/review"
	"github.com/biggs-100/odachi/internal/domain/budget"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/rules"
)

// baselineFile is the frozen size of every text odachi hands to a model, in the
// order a session pays for it. It only moves on purpose: a text that grew fails
// here until somebody raises the pin in a commit a reviewer sees, and one that
// shrank fails until the pin follows it down (0295).
const baselineFile = "bytes.baseline"

// TestTheTextAnAgentReadsKeepsItsBudget: the block in AGENTS.md, the four lens
// prompts and the skills index are read on every turn and nothing measured them.
// Doctor reports their size and max_prompt_kb caps the review prompt, but a report
// nobody compares is a number that permits anything.
//
// What this does not prove: that the text is worth its bytes. A smaller prompt and
// a better one are different questions, and only a measurement against outcomes
// answers the second one.
func TestTheTextAnAgentReadsKeepsItsBudget(t *testing.T) {
	measured, err := measuredSurfaces()
	if err != nil {
		t.Fatalf("measuring the surfaces: %v", err)
	}
	pinned, err := budget.Read(baselineFile)
	if err != nil {
		t.Fatalf("reading %s: %v", baselineFile, err)
	}

	for _, issue := range budget.Issues(measured, pinned) {
		t.Error(issue)
	}
}

// namedSurface is one text a model is handed, with the name a person reads it by.
// One list feeds the size pin and the content pin, so the two cannot drift into
// describing different sets of artifacts.
type namedSurface struct {
	Artifact string
	Bytes    []byte
}

// surfaces renders every text a model is handed at the session's start and yields
// its exact bytes. Both pins measure this one list: the byte test counts it and the
// hash test digests it. The values are measured, never typed: a pin copied by hand
// from a document is the second list this project keeps removing.
func surfaces() ([]namedSurface, error) {
	var out []namedSurface
	for _, voice := range []string{persona.Novato, persona.Senior} {
		block, err := rules.BlockFor(voice, persona.SkillsHome)
		if err != nil {
			return nil, err
		}
		out = append(out, namedSurface{Artifact: "block:" + voice, Bytes: []byte(block)})
	}

	for _, name := range assets.Names() {
		lens, err := assets.Lens(name)
		if err != nil {
			return nil, err
		}
		out = append(out, namedSurface{Artifact: "lens:" + name, Bytes: []byte(lens)})
	}

	index, err := os.ReadFile(filepath.Join("..", "..", "SKILLS.md"))
	if err != nil {
		return nil, err
	}
	out = append(out, namedSurface{Artifact: "index:SKILLS.md", Bytes: index})
	return out, nil
}

// measuredSurfaces turns the one list of bytes into the size pins the byte test
// compares. Its order and values are unchanged from the original measurement.
func measuredSurfaces() ([]budget.Pin, error) {
	named, err := surfaces()
	if err != nil {
		return nil, err
	}
	pins := make([]budget.Pin, 0, len(named))
	for _, surface := range named {
		pins = append(pins, budget.Pin{Artifact: surface.Artifact, Bytes: len(surface.Bytes)})
	}
	return pins, nil
}
