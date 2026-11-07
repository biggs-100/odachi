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
