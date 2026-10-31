package main

import (
	"strings"
	"testing"
)

// TestSinceDoesNotGoWithTheReadOnlyForms: --since corrects a verdict, and a
// correction of a read-only form would be a flag nobody reads. The refusal names
// the three forms it does not go with, before anything is read or written.
func TestSinceDoesNotGoWithTheReadOnlyForms(t *testing.T) {
	for _, other := range [][]string{
		{"--show"},
		{"--refute", "abc"},
		{"--dismiss", "a.go:1", "--because", "why"},
	} {
		args := append([]string{"review", "--since", "1111111111111111111111111111111111111111"}, other...)

		code, _, errOut := runDir(t, t.TempDir(), args, "")

		if code != 1 {
			t.Errorf("%v: code = %d, want 1", args, code)
		}
		if !strings.Contains(errOut, "--since corrects a verdict") {
			t.Errorf("%v: the refusal does not name the flag: %s", args, errOut)
		}
	}
}
