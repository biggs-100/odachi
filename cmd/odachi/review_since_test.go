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

// TestSinceRefusesABaseItWouldSilentlyReplace: a correction keeps the base of the
// verdict it completes. Accepting --base beside it would let the flag do nothing and
// say nothing, which is a lie about what the run obeyed.
func TestSinceRefusesABaseItWouldSilentlyReplace(t *testing.T) {
	code, _, errOut := runDir(t, t.TempDir(),
		[]string{"review", "--since", "1111111111111111111111111111111111111111", "--base", "main"}, "")

	if code != 1 {
		t.Errorf("code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "keeps the base of the verdict it corrects") {
		t.Errorf("the refusal does not name the substitution: %s", errOut)
	}
}
