package skills_test

import (
	"strings"
	"testing"
)

// TestTUIDesignKeepsItsAttribution: the method is adapted from another project, and
// the notice lives in the body because the installer copies SKILL.md, references/ and
// assets/ only -a NOTICE.md beside them would never reach a project-. The name, the
// license and the second credit are pinned so a rewrite cannot drop them: an audit
// that compares file names reads the missing NOTICE.md as a gap, and this is the test
// that answers it (0309).
func TestTUIDesignKeepsItsAttribution(t *testing.T) {
	skill := mustSkill(t, "tui-design")
	main := string(skill.Main())

	flat := strings.Join(strings.Fields(main), " ")
	for _, want := range []string{
		"## Attribution",
		"gentle-shell",
		"Apache-2.0",
		"impeccable",
		"Paul Bakaus",
		"no file from either project is redistributed",
		"the installer copies `SKILL.md`, `references/` and `assets/` only",
	} {
		if !strings.Contains(flat, want) {
			t.Errorf("tui-design does not carry %q in its attribution", want)
		}
	}
}
