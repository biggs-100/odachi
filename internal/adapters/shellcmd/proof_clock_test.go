package shellcmd_test

import (
	"io"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/shellcmd"
)

// TestAProofKeepsItsTotalDeadline: a proof command such as go test may stay
// silent for its whole run. A silence clock on the proof path would cut it in
// the middle and file a proof that says the command failed when it was still
// working. This is a guard: it pins the boundary that separates the two paths.
func TestAProofKeepsItsTotalDeadline(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		run  func() (bool, error)
	}{
		{"Run", func() (bool, error) {
			_, _, _, timedOut, err := shellcmd.Run(".", quietCommand(), 5)
			return timedOut, err
		}},
		{"RunStream", func() (bool, error) {
			_, _, _, timedOut, err := shellcmd.RunStream(".", quietCommand(), 5, io.Discard)
			return timedOut, err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			timedOut, err := tc.run()
			if err != nil {
				t.Fatalf("proof command: %v", err)
			}
			if timedOut {
				t.Error("a silent proof command was stopped before its total deadline")
			}
		})
	}
}
