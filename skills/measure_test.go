package skills_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/skills"
)

// TestMeasureTeachesTheShapeOfAMeasurement: a number this project takes had no
// shape. The criterion arrived with the result, the runs that were dropped went
// unrecorded, and nothing said what the number cannot answer - which is how a
// document ends up claiming counts nobody measures.
//
// The skill is the door an agent loads before it takes a number, so the test pins
// the five sections and the two sentences that keep a byte count from becoming a
// cost claim.
func TestMeasureTeachesTheShapeOfAMeasurement(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		byName[s.Name] = string(s.Main())
	}

	content, ok := byName["measure"]
	if !ok {
		t.Fatal("the payload has no measure skill: a measurement is still a story")
	}
	for _, want := range []string{
		"the criterion before the run",
		"bytes are not tokens",
		"instruction accounting is not a cost measurement",
		"the exact command",
		"excluded run",
		"what the data does not show",
		"quotes the tool",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("measure does not say %q: the shape of a measurement is missing", want)
		}
	}
}

// TestTheIndexNamesEverySkillOnDisk: SKILLS.md is what a session reads to know
// which craft exists, and it is generated. A skill added to the folder and missing
// from the index is a door nobody can find; a row with no folder is a promise
// nobody keeps.
func TestTheIndexNamesEverySkillOnDisk(t *testing.T) {
	index, err := os.ReadFile(filepath.Join("..", "SKILLS.md"))
	if err != nil {
		t.Fatalf("reading SKILLS.md: %v", err)
	}
	text := string(index)

	for _, name := range bundled {
		row := "| " + name + " |"
		if !strings.Contains(text, row) {
			t.Errorf("SKILLS.md has no row for %q: run odachi skills --write", name)
		}
	}
	if !strings.Contains(text, "skills/measure/SKILL.md") {
		t.Error("SKILLS.md does not name the measure skill's path")
	}
}
