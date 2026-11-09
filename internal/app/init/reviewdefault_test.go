package init_test

import (
	"strings"
	"testing"

	initapp "github.com/biggs-100/odachi/internal/app/init"
)

// fakeReview renders the bytes a project is born with, the way the CLI wires the
// review adapter: the content is the test's, the port is init's.
type fakeReview struct{ data string }

func (f fakeReview) Default() ([]byte, error) { return []byte(f.data), nil }

// TestInitLeavesTheReviewOnForANewProject: a project with no review policy is
// born with the switch on, so its first close asks for a verdict instead of
// discovering the review late.
func TestInitLeavesTheReviewOnForANewProject(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := initapp.Run(initapp.Options{Dir: dir, Review: fakeReview{data: "mode = \"on\"\n"}}, files, &fakeSkills{})

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (lines %q)", got.Code, got.Lines)
	}
	if content, ok := files.get("odd/review.toml"); !ok || content != "mode = \"on\"\n" {
		t.Errorf("odd/review.toml = %q, present %v; want the policy the port renders", content, ok)
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), "on; a close asks for a verdict") {
		t.Errorf("lines %q do not say the review is on", got.Lines)
	}
}

// TestInitKeepsAReviewPolicyTheProjectAlreadyHas: the switch is the human's, so
// a file that exists is kept byte for byte, and the run says it kept it instead
// of staying silent about a decision it did not take.
func TestInitKeepsAReviewPolicyTheProjectAlreadyHas(t *testing.T) {
	own := "mode = \"off\"\n"
	files := newFakeFiles(map[string]string{
		".git/HEAD":       "ref: refs/heads/main\n",
		"odd/review.toml": own,
	})

	got := initapp.Run(initapp.Options{Dir: dir, Review: fakeReview{data: "mode = \"on\"\n"}}, files, &fakeSkills{})

	if content, _ := files.get("odd/review.toml"); content != own {
		t.Errorf("odd/review.toml = %q, want the project's own %q", content, own)
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), "kept; the review switch is the human's") {
		t.Errorf("lines %q do not say the policy was kept", got.Lines)
	}
}
