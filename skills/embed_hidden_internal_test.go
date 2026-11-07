package skills

import (
	"io/fs"
	"slices"
	"testing"
)

// TestTheEmbedCarriesTheHiddenFiles: `*` skips names that begin with '.' or '_',
// and `all:*` keeps them. The loader lists what the payload carries, so a
// regression to `*` would ship a skill folder with files missing and the suite
// would say nothing. The fixture beside the skills is the measurement: it is not
// a skill -it carries no SKILL.md- and nothing installs it.
func TestTheEmbedCarriesTheHiddenFiles(t *testing.T) {
	for _, path := range []string{".embed-fixture/.hidden.txt", ".embed-fixture/_under.txt"} {
		if _, err := fs.ReadFile(bundled, path); err != nil {
			t.Errorf("the bundled payload does not carry %s: the directive is not all:* (%v)", path, err)
		}
	}

	// The fixture must not become a skill: a payload with an extra name installs
	// it, and a skill nobody wrote is worse than a missing file.
	if slices.Contains(Names(), ".embed-fixture") {
		t.Errorf("the fixture is listed as a skill: %q", Names())
	}
}
