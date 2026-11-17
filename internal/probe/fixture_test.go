package probe_test

import (
	"io/fs"
	"strings"
	"testing"

	domainprobe "github.com/biggs-100/odachi/internal/domain/probe"
	"github.com/biggs-100/odachi/internal/probe"
)

// TestTheBundledFixtureIsScorable: the fixture is the exam, so it is checked
// like one before any measurement uses it. Every defect names a file that exists
// in both arms and a line that exists in the seeded one; a key pointing at a
// line that is not there would score a reviewer against a place nobody can find.
func TestTheBundledFixtureIsScorable(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(probe.Fixture(), probe.ManifestPath)
	if err != nil {
		t.Fatalf("read the manifest: %v", err)
	}
	manifest, err := domainprobe.Parse(data)
	if err != nil {
		t.Fatalf("parse the manifest: %v", err)
	}

	for _, defect := range manifest.Defects {
		t.Run(defect.ID, func(t *testing.T) {
			t.Parallel()
			for _, arm := range []string{probe.Clean, probe.Seeded} {
				path := probe.Root + "/" + arm + "/" + defect.Path
				if _, err := fs.ReadFile(probe.Fixture(), path); err != nil {
					t.Fatalf("the %s arm has no %s: %v", arm, defect.Path, err)
				}
			}

			body, err := fs.ReadFile(probe.Fixture(), probe.Root+"/"+probe.Seeded+"/"+defect.Path)
			if err != nil {
				t.Fatalf("read the seeded %s: %v", defect.Path, err)
			}
			if lines := strings.Count(string(body), "\n"); defect.Line > lines {
				t.Errorf("defect %q points at line %d and the seeded file has %d", defect.ID, defect.Line, lines)
			}
		})
	}
}

// TestADefectThatNeedsContextLivesInTheBase: the base is what a defect outside
// the diff lives in. A fixture that named a context file in an arm would put it
// back in the diff, and the two readers would measure the same thing.
func TestADefectThatNeedsContextLivesInTheBase(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(probe.Fixture(), probe.ManifestPath)
	if err != nil {
		t.Fatalf("read the manifest: %v", err)
	}
	manifest, err := domainprobe.Parse(data)
	if err != nil {
		t.Fatalf("parse the manifest: %v", err)
	}

	contextual := 0
	for _, defect := range manifest.Defects {
		for _, path := range defect.Context {
			contextual++
			if _, err := fs.ReadFile(probe.Fixture(), probe.Root+"/"+probe.Base+"/"+path); err != nil {
				t.Errorf("defect %q needs %s, and the base has no such file: %v", defect.ID, path, err)
				continue
			}
			for _, arm := range []string{probe.Clean, probe.Seeded} {
				if _, err := fs.ReadFile(probe.Fixture(), probe.Root+"/"+arm+"/"+path); err == nil {
					t.Errorf("defect %q needs %s, and the %s arm carries it: the file the key calls context would be in the diff", defect.ID, path, arm)
				}
			}
		}
	}
	if contextual == 0 {
		t.Error("no defect needs a file outside the diff, so the two arms of the reader cannot differ")
	}
}
