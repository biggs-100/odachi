package doctor

import (
	"fmt"
	"strings"
	"testing"

	domaindoctor "github.com/biggs-100/odachi/internal/domain/doctor"
	"github.com/biggs-100/odachi/internal/domain/rules"
	"github.com/biggs-100/odachi/internal/domain/skill"
)

// TestCheckNineSaysWhatTheHarnessCosts: the block and the registry are what an
// agent reads on every turn. The line names their bytes beside the count it
// already prints, so the always-on cost is a number and not an estimate.
func TestCheckNineSaysWhatTheHarnessCosts(t *testing.T) {
	files := configured()
	result := run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi")

	body, ok := rules.BodyIn(files.files[at("AGENTS.md")])
	if !ok {
		t.Fatal("the fixture has no block")
	}
	registry := files.files[at(skill.RegistryFile)]

	detail := checkOf(t, result, domaindoctor.CheckSkills).Detail
	for _, want := range []string{
		"the block is " + humanSize(len(body)),
		"up to date (" + humanSize(len(registry)) + ")",
	} {
		if !strings.Contains(detail, want) {
			t.Errorf("check 9 does not say %q: %s", want, detail)
		}
	}
}

// TestCheckNineTakesTheProblemOfAnUnreadableBlock: a size nobody could read is not
// zero; the check takes the read's problem and names the file.
func TestCheckNineTakesTheProblemOfAnUnreadableBlock(t *testing.T) {
	files := configured()
	files.errOn = at("AGENTS.md")

	check := checkOf(t, run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi"), domaindoctor.CheckSkills)

	if check.Status != domaindoctor.StatusProblem {
		t.Fatalf("check 9 is %v with an unreadable block, want a problem: %s", check.Status, domaindoctor.Line(check))
	}
	if !strings.Contains(check.Detail, "could not read AGENTS.md") {
		t.Errorf("the problem does not name the file: %s", check.Detail)
	}
}

// TestCheckNineSaysNoSizeWithoutABlock: a file with no odachi block has no cost to
// name; check 4 is what says the block is missing, and this line never invents one.
func TestCheckNineSaysNoSizeWithoutABlock(t *testing.T) {
	files := configured()
	files.put("AGENTS.md", "# notes\n\nno block here\n")

	check := checkOf(t, run(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi"), domaindoctor.CheckSkills)

	if strings.Contains(check.Detail, "the block is") {
		t.Errorf("check 9 names a block the file does not carry: %s", check.Detail)
	}
}

// humanSize is the format a person reads, asserted here so a change to the
// numbers is a red test and not a silently different line.
func humanSize(n int) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 10*1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	case n < 1024*1024:
		return fmt.Sprintf("%.0f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	}
}
