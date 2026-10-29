package review_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// TestExecFilesThePromptItSends: the printed form writes the question beside the
// verdict, and a run that hands the prompt to a reviewer used to keep it nowhere.
// The bytes that produced a verdict are the only way to read that verdict again
// without regenerating them, and regenerating needs the same base.
func TestExecFilesThePromptItSends(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	if result := p.run(review.Options{Exec: "revisor --model {model}"}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	rel, err := domainreview.PromptPath(currentTree)
	if err != nil {
		t.Fatalf("PromptPath: %v", err)
	}
	filed, ok := p.files.files[storeFile(rel)]
	if !ok {
		t.Fatalf("the prompt was not filed at %s", rel)
	}
	for _, want := range []string{"# odachi review", "a.go"} {
		if !strings.Contains(string(filed), want) {
			t.Errorf("the filed prompt does not carry %q", want)
		}
	}
}

// TestASplitRunFilesOnePromptPerGroup: two models are two questions, and one file
// that held the last of them would be a lie about the first. The single group keeps
// the canonical name, so the common case reads the same from both forms.
func TestASplitRunFilesOnePromptPerGroup(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	if result := p.run(review.Options{
		Exec:       "revisor --model {model}",
		Lens:       "reliability,risk",
		Because:    "the test splits the run by model",
		LensModels: map[string]string{"risk": "sharp/model"},
	}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	var filed []string
	for path := range p.files.files {
		if strings.Contains(path, currentTree+".prompt") {
			filed = append(filed, relative(path))
		}
	}
	if len(filed) != 2 {
		t.Fatalf("filed prompts = %v, want one per group", filed)
	}
	for _, want := range []string{".prompt-risk.md", ".prompt-reliability.md"} {
		found := false
		for _, path := range filed {
			if strings.Contains(path, want) {
				found = true
			}
		}
		if !found {
			t.Errorf("filed prompts = %v, want %q among them", filed, want)
		}
	}
}

// TestAPromptThatCannotBeFiledDoesNotCostTheAnswer: the prompt is filed after the
// reviewer has answered, and the answer is the paid thing. A store that cannot
// hold the question must say so and still keep the verdict, not claim a reviewer
// was never started and throw its answer away.
func TestAPromptThatCannotBeFiledDoesNotCostTheAnswer(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]},"passes":[{"path":"a.go","lens":"reliability"}]}`
	p.files.failWrite = func(path string) error {
		if strings.Contains(path, ".prompt") {
			return errors.New("el disco está lleno")
		}
		return nil
	}

	result := p.run(review.Options{Exec: "revisor --model {model}"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0: the answer was thrown away (%s)", result.Code, result.Message)
	}
	if _, stored := p.files.files[storePath(currentTree)]; !stored {
		t.Fatal("the verdict was not stored")
	}
	if !strings.Contains(strings.Join(result.Lines, "\n"), "could not be filed") {
		t.Errorf("the run does not say the prompt was not filed:\n%s", strings.Join(result.Lines, "\n"))
	}
}
