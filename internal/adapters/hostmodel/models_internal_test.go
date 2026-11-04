package hostmodel

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// TestTheCatalogIsKnownOnlyWhenTheHostAnswers pins the two answers the review
// reads: a list is a catalog, and an error or an empty answer is not. The
// difference decides whether a model is refused or the run keeps going, so the
// test answers the probe instead of starting a host.
func TestTheCatalogIsKnownOnlyWhenTheHostAnswers(t *testing.T) {
	original := list
	t.Cleanup(func() { list = original })

	cases := []struct {
		name  string
		out   []byte
		err   error
		want  []string
		known bool
	}{
		{name: "a list", out: []byte("opencode-go/beam-free\nopencode-go/deepseek-v4.1-flash\n"), want: []string{"opencode-go/beam-free", "opencode-go/deepseek-v4.1-flash"}, known: true},
		{name: "an error", err: errors.New("no such executable")},
		{name: "an empty answer", out: []byte("\n\n")},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var probed string
			list = func(dir string) ([]byte, error) {
				probed = dir
				return c.out, c.err
			}

			got, known := Models("C:\\project")
			if known != c.known {
				t.Fatalf("Models known = %v, want %v (models: %v)", known, c.known, got)
			}
			if len(got) != len(c.want) {
				t.Fatalf("Models = %v, want %v", got, c.want)
			}
			for i := range c.want {
				if got[i] != c.want[i] {
					t.Fatalf("Models = %v, want %v", got, c.want)
				}
			}

			// The catalog is global: the probe leaves the project's folder so a
			// host with a background service cannot hold it open on Windows.
			home, err := os.UserHomeDir()
			if err == nil && strings.TrimSpace(home) != "" && probed != home {
				t.Errorf("the probe ran in %q, want the user's home %q", probed, home)
			}
		})
	}
}
