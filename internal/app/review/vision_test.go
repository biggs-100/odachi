package review_test

import (
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
