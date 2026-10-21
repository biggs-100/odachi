// Package review carries the four lens prompts: the content that turns one
// generic design review into a specialist's.
//
// Each file is a prompt, not code: frontmatter with its `name` and a
// `description` that starts with `Trigger:`, then `## Qué mirar` with concrete
// rules and `## Qué NO reportar` with what to stay quiet about. The files are
// compiled into the binary with //go:embed, so a review needs no network and no
// repository — the same reason the skills are embedded.
//
// The order below is the canonical order of the four lenses and it is the order
// the domain selects: internal/domain/review cannot import this package (the
// domain is stdlib-only), so a test pins the two lists together instead.
package review

import (
	"embed"
	"fmt"
	"strings"
)

// order is the fixed order of the four lenses, matching review.Lenses(Full).
var order = []string{"risk", "readability", "reliability", "resilience"}

// bundled is every lens prompt, one file per lens.
//
//go:embed *.md
var bundled embed.FS

// Names returns the lens names in the canonical order. The slice is a copy: a
// caller must not be able to sort the product.
func Names() []string {
	return append([]string(nil), order...)
}

// Lens returns one lens prompt verbatim, frontmatter included. The caller that
// builds a prompt strips the frontmatter; the file stays complete on disk, so it
// can be read, validated and edited like any other content file.
func Lens(name string) (string, error) {
	key := strings.ToLower(strings.TrimSpace(name))
	for _, lens := range order {
		if lens != key {
			continue
		}
		data, err := bundled.ReadFile(lens + ".md")
		if err != nil {
			return "", fmt.Errorf("no pude leer la lente %s: %w", lens, err)
		}
		return string(data), nil
	}
	return "", fmt.Errorf("la lente %q no existe: las lentes son %s", name, strings.Join(order, ", "))
}
