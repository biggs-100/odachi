//go:build windows

package main

import (
	"strings"
	"testing"
)

// A reviewer reads its prompt from stdin. findstr prints what it received, and the
// failure path shows it, so the test can look at exactly what the reviewer saw.
func TestReviewExecFeedsTheReviewerItsPrompt(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)

	code, out, errOut := runDir(t, dir, []string{"review", "--exec", "findstr /r .* & exit 1"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	for _, want := range []string{"# odachi review", "internal/app/turnos.go"} {
		if !strings.Contains(out+errOut, want) {
			t.Errorf("the reviewer never saw %q in its prompt (stdout: %q, stderr: %q)", want, out, errOut)
		}
	}
	if verdicts := reviewsIn(t, dir); len(verdicts) != 0 {
		t.Errorf("a failed reviewer stored %v, want nothing", verdicts)
	}
}

// The whole form: a reviewer that reads stdin and answers with the verdict. The
// findstr is the reviewer's own check that the prompt was there; without it the
// command would answer about a candidate it never read.
func TestReviewExecStoresTheVerdictOfAReviewerThatReadsStdin(t *testing.T) {
	t.Parallel()
	dir := changedProject(t)

	command := `findstr /c:"odachi review" >nul && echo {"findings":[],"passes":[{"path":"internal/app/turnos.go","lens":"reliability"}]}`

	code, out, errOut := runDir(t, dir, []string{"review", "--exec", command}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if verdicts := reviewsIn(t, dir); len(verdicts) != 1 {
		t.Errorf("verdicts = %v, want exactly one", verdicts)
	}
}
