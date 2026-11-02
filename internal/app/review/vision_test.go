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
