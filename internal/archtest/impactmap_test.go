package archtest_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/skills"
)

// TestTheImpactMapNamesEverySkillTheBinaryCarries: the skills row is the list a
// reader trusts and the last line anyone updates -it had already lost `design`
// when this test was written-. Pinning it against the payload turns a new skill
// into one line of the map instead of a count nobody measures.
func TestTheImpactMapNamesEverySkillTheBinaryCarries(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "docs", "impact-map.md"))
	if err != nil {
		t.Fatalf("read docs/impact-map.md: %v", err)
	}

	want := "`skills/{" + strings.Join(skills.Names(), ",") + "}/SKILL.md`"
	if !strings.Contains(string(data), want) {
		t.Fatalf("docs/impact-map.md does not name every skill the binary carries;\nthe row must carry: %s", want)
	}
}
