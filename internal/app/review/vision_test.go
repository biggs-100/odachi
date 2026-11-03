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
// document the gate would take and carries the requirement it quotes, naming the
// file it came from.
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

// A trailer never stands in for a document that is still open: `check` judges the
// document in that case, so the review must read the same vision or the two verbs
// answer different questions about one tree.
func TestATrailerDoesNotOverrideAnOpenDocument(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("odd/tasks/0167-three-tests-pin.md")] = []byte(lightDocument)
	p.git.messages = []string{"the wait state\n\nOdachi-Change: bring back the wait state on the pay button\n"}

	result := p.run(review.Options{Base: lightBase})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Prompt, "three tests pin the light lane's refusals") {
		t.Errorf("the prompt does not carry the open document's line:\n%s", result.Prompt)
	}
	if strings.Contains(result.Prompt, "bring back the wait state") {
		t.Errorf("the trailer overrode an open document:\n%s", result.Prompt)
	}
}

// A normal feature that is still open keeps its brief, whatever a commit in its
// range says: a trailer is the vision only when there is no open document.
func TestATrailerDoesNotOverrideAnOpenNormalFeature(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("odd/tasks/0166-a-line-that-can-close.md")] = []byte(`# a-line-that-can-close
Base: ` + lightBase + `

## Requirements (verbatim, from the user)
R1 · "a change that fits one line is written as one line"

## Steps
- [ ] T1 · the lane   R1
`)
	p.git.messages = []string{"the wait state\n\nOdachi-Change: bring back the wait state on the pay button\n"}

	result := p.run(review.Options{Base: lightBase})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Prompt, "reservar turnos") {
		t.Errorf("the prompt does not carry brief.md:\n%s", result.Prompt)
	}
	if strings.Contains(result.Prompt, "bring back the wait state") {
		t.Errorf("the trailer overrode an open feature's brief:\n%s", result.Prompt)
	}
}

// otherBase is the commit a document records when the review is asked about a
// different one: the gate never compares the two, and the review must not either.
const otherBase = "9b1e6c2a7f4d8e0b3c5a7d9e1f2b4c6d8e0f1a3b"

// TestATrailerDoesNotOverrideAnOpenDocumentAtAnotherBase: the candidate's base is
// not the base the document records, and `check` still judges the open document.
// The review asks the same question -the newest document, whatever its base- or a
// trailer stands in for the vision the gate uses (0182).
func TestATrailerDoesNotOverrideAnOpenDocumentAtAnotherBase(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("odd/tasks/0166-a-line-that-can-close.md")] = []byte(`# a-line-that-can-close
Base: ` + otherBase + `

## Requirements (verbatim, from the user)
R1 · "a change that fits one line is written as one line"

## Steps
- [ ] T1 · the lane   R1
`)
	p.git.messages = []string{"the wait state\n\nOdachi-Change: bring back the wait state on the pay button\n"}

	result := p.run(review.Options{Base: lightBase})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Prompt, "reservar turnos") {
		t.Errorf("the prompt does not carry brief.md:\n%s", result.Prompt)
	}
	if strings.Contains(result.Prompt, "bring back the wait state") {
		t.Errorf("the trailer overrode an open document at another base:\n%s", result.Prompt)
	}
}

// TestAClosedDocumentAtAnotherBaseStillLetsTheTrailerBeTheVision: the gate opens
// the trailer lane when the newest document is closed, and the document's base is
// not the question. A review that refused the trailer for every document present
// would be the same defect pointing the other way.
func TestAClosedDocumentAtAnotherBaseStillLetsTheTrailerBeTheVision(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("odd/tasks/0166-a-line-that-can-close.md")] = []byte(`# a-line-that-can-close
Base: ` + otherBase + `

## Requirements (verbatim, from the user)
R1 · "a change that fits one line is written as one line"

## Steps
- [x] T1 · the lane   R1
`)
	p.git.messages = []string{"the wait state\n\nOdachi-Change: bring back the wait state on the pay button\n"}

	result := p.run(review.Options{Base: lightBase})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Prompt, "bring back the wait state") {
		t.Errorf("the closed document at another base blocked the trailer:\n%s", result.Prompt)
	}
}

// TestACorrectionAsksTheDocumentItsPredecessorJudged: a correction belongs to the
// feature its verdict judged, and the newest document may be a later feature. The
// vision comes from the predecessor's base, not from the newest document the gate
// would take for a new run.
func TestACorrectionAsksTheDocumentItsPredecessorJudged(t *testing.T) {
	p := newProject()
	p.files.files[storePath(correctedTree)] = []byte(verdictFor(t, correctedTree, otherBase, []string{"risk"}))
	p.files.files[fullPath("odd/tasks/0166-a-line-that-can-close.md")] = []byte(`# a-line-that-can-close
Base: ` + lightBase + `

## Requirements (verbatim, from the user)
R1 · "a change that fits one line is written as one line"

## Steps
- [ ] T1 · the lane   R1
`)
	p.git.messages = []string{"the wait state\n\nOdachi-Change: bring back the wait state on the pay button\n"}

	result := p.run(review.Options{Since: correctedTree})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Prompt, "bring back the wait state") {
		t.Errorf("a newer open document took the vision from the corrected change:\n%s", result.Prompt)
	}
}

// TestALightFeatureIsJudgedAgainstItsOwnLineOnTheDefaultBase: `odachi review`
// with no --base asks about HEAD, and the document records the commit its feature
// started from. The two strings differ, so asking "is there a document at this
// base" finds nothing; the newest document is the question the gate asks, and a
// light feature's line is the vision (0182).
func TestALightFeatureIsJudgedAgainstItsOwnLineOnTheDefaultBase(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("odd/tasks/0167-three-tests-pin.md")] = []byte(lightDocument)

	result := p.run(review.Options{Base: review.DefaultBase})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Prompt, "three tests pin the light lane's refusals") {
		t.Errorf("the prompt does not carry the light feature's line on the default base:\n%s", result.Prompt)
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
