package review_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
)

// lightBase is the commit a light feature started from: the document records it,
// and the review is asked about that same base.
const lightBase = "4c731875f8d24af429d3d0fc7991d42b234a995d"

// lightDocument is what `odachi start --light` writes: one requirement quoted and
// one step with the same words.
const lightDocument = `# three-tests-pin
Base: ` + lightBase + `
Light: yes

## Requirements (verbatim, from the user)
R1 · "three tests pin the light lane's refusals"

## Steps
- [ ] T1 · three tests pin the light lane's refusals   R1

## History
L1 · "three tests pin the light lane's refusals"
`

// TestALightFeatureIsJudgedAgainstItsOwnLine: a light feature writes no brief, so
// brief.md still carries the previous feature's vision. The review reads the
// document whose base is this candidate's and carries the requirement it quotes,
// naming the file it came from.
func TestALightFeatureIsJudgedAgainstItsOwnLine(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("odd/tasks/0167-three-tests-pin.md")] = []byte(lightDocument)

	result := p.run(review.Options{Base: lightBase})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Prompt, "three tests pin the light lane's refusals") {
		t.Errorf("the prompt does not carry the feature's own line:\n%s", result.Prompt)
	}
	if !strings.Contains(result.Prompt, "odd/tasks/0167-three-tests-pin.md") {
		t.Errorf("the prompt does not name the document the vision came from:\n%s", result.Prompt)
	}
	if strings.Contains(result.Prompt, "reservar turnos") {
		t.Errorf("the prompt carries the previous feature's brief:\n%s", result.Prompt)
	}
}

// TestAClosedLightFeatureIsStillJudgedAgainstItsLine: a feature whose steps are
// all checked still records the base it started from, and the review of that tree
// must read its line instead of the previous brief. Nothing in the lookup may ask
// how much of the work is done: closing a feature is what the gate judges, not
// what the reviewer is told.
func TestAClosedLightFeatureIsStillJudgedAgainstItsLine(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("odd/tasks/0167-three-tests-pin.md")] = []byte(
		strings.Replace(lightDocument, "- [ ] T1", "- [x] T1", 1))

	result := p.run(review.Options{Base: lightBase})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Prompt, "three tests pin the light lane's refusals") {
		t.Errorf("the prompt does not carry the closed feature's own line:\n%s", result.Prompt)
	}
	if strings.Contains(result.Prompt, "reservar turnos") {
		t.Errorf("the prompt carries the previous feature's brief:\n%s", result.Prompt)
	}
}

// TestADocumentLessChangeIsJudgedAgainstItsTrailer: the change wrote no document,
// and brief.md still carries the previous feature's vision. The review reads the
// trailer from the commits of the range it judges, and says the diff is the record.
func TestADocumentLessChangeIsJudgedAgainstItsTrailer(t *testing.T) {
	p := newProject()
	p.git.messages = []string{"the wait state\n\nOdachi-Change: bring back the wait state on the pay button\n"}

	result := p.run(review.Options{Base: lightBase})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Prompt, "bring back the wait state on the pay button") {
		t.Errorf("the prompt does not carry the change's own line:\n%s", result.Prompt)
	}
	if !strings.Contains(result.Prompt, "the diff is the record") {
		t.Errorf("the prompt does not say the diff is the record:\n%s", result.Prompt)
	}
	if strings.Contains(result.Prompt, "reservar turnos") {
		t.Errorf("the prompt carries the previous feature's brief:\n%s", result.Prompt)
	}
}

// A trailer that is empty or names two different changes is refused by name: a
// review judged against a line the reader had to choose is a guess.
func TestAnAmbiguousTrailerIsRefusedByTheReview(t *testing.T) {
	p := newProject()
	p.git.messages = []string{
		"first\n\nOdachi-Change: one line\n",
		"second\n\nOdachi-Change: another line\n",
	}

	result := p.run(review.Options{Base: lightBase})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want %d (%s)", result.Code, review.CodeBlocked, result.Message)
	}
	for _, want := range []string{"one line", "another line"} {
		if !strings.Contains(result.Message, want) {
			t.Errorf("the refusal does not name %q: %s", want, result.Message)
		}
	}
}

// TestANormalFeatureStillCarriesItsBrief: a feature with a brief is judged exactly
// as before, whatever its task document says.
func TestANormalFeatureStillCarriesItsBrief(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("odd/tasks/0166-a-line-that-can-close.md")] = []byte(`# a-line-that-can-close
Base: ` + lightBase + `

## Requirements (verbatim, from the user)
R1 · "a change that fits one line is written as one line"

## Steps
- [ ] T1 · the lane   R1
`)

	result := p.run(review.Options{Base: lightBase})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Prompt, "reservar turnos") {
		t.Errorf("the prompt does not carry brief.md:\n%s", result.Prompt)
	}
	if strings.Contains(result.Prompt, "a change that fits one line is written as one line") {
		t.Errorf("the prompt carries the task document instead of the brief:\n%s", result.Prompt)
	}
}

// TestATaskDocumentThatCannotBeReadIsNamed: a failed read is not an absence. The
// review says which document it could not read instead of falling back to a brief
// it did not judge this feature against.
func TestATaskDocumentThatCannotBeReadIsNamed(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("odd/tasks/0167-three-tests-pin.md")] = []byte(lightDocument)
	p.files.failRead = func(path string) error {
		if strings.HasSuffix(path, "0167-three-tests-pin.md") {
			return errors.New("access denied")
		}
		return nil
	}

	result := p.run(review.Options{Base: lightBase})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want %d (%s)", result.Code, review.CodeBlocked, result.Message)
	}
	if !strings.Contains(result.Message, "0167-three-tests-pin.md") {
		t.Errorf("message %q does not name the document that could not be read", result.Message)
	}
}
