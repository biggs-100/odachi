// Package probe carries the fixture a lens probe is measured on: two arms of
// the same small project, a clean one and one with planted defects, and the
// manifest that names them.
//
// The files are compiled into the binary with //go:embed and live under
// testdata, so the go tool never builds them and the probe measures the same
// bytes on every machine. The measurement is about the lens prompts; the
// fixture is the exam those prompts answer.
package probe

import (
	"embed"
	"io/fs"
)

// Clean is the arm with no defect: a candidate the reviewer should approve
// without inventing a finding.
const Clean = "clean"

// Seeded is the arm with the planted defects: the same project with the
// manifest's defects in it.
const Seeded = "seeded"

//go:embed testdata/lenses
var bundled embed.FS

// Fixture returns the fixture tree, rooted at the fixture that ships with the
// binary. The two arms are the top-level directories Clean and Seeded.
func Fixture() fs.FS {
	return bundled
}

// Root is the path of the bundled fixture inside Fixture.
const Root = "testdata/lenses"

// ManifestPath is where the manifest of the bundled fixture lives, relative to
// Fixture.
const ManifestPath = Root + "/manifest.json"
