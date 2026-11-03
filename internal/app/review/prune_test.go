package review_test

import (
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// verdictAnswer is the reviewer's answer every test here pays for: no finding,
// and the declaration the store requires.
const verdictAnswer = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

// storePrompts is every question the fake store holds, project-relative.
func storePrompts(p *project) []string {
	var prompts []string
	for path := range p.files.files {
		if strings.Contains(filepathSlash(path), ".prompt") {
			prompts = append(prompts, relative(path))
		}
	}
	return prompts
}

// TestKeepPromptsZeroLeavesNoQuestionBehind: the default lifetime keeps no
// prompt, and the moment the verdict lands the question it answered is disposed
// of. The verdict is the evidence, and a file that nobody reads is the biggest
// thing in the store.
func TestKeepPromptsZeroLeavesNoQuestionBehind(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = verdictAnswer

	result := p.run(review.Options{Exec: "revisor --model {model}"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if _, ok := p.files.files[storePath(currentTree)]; !ok {
		t.Fatalf("the verdict was not stored: %v", p.files.written())
	}
	if prompts := storePrompts(p); len(prompts) != 0 {
		t.Fatalf("the store kept a question keep_prompts = 0 disposes of: %v", prompts)
	}
}

// TestKeepPromptsOneKeepsTheNewestQuestionAndForgetsTheOlder: the retention is
// counted in trees, and the tree that just landed is the newest. The older
// tree's prompt goes while its verdict stays, because a record is evidence and a
// question is not.
func TestKeepPromptsOneKeepsTheNewestQuestionAndForgetsTheOlder(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = verdictAnswer

	oldTree := strings.Repeat("b", 40)
	seeded, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "primero",
		Lenses:   []string{domainreview.Reliability},
	}, oldTree, "HEAD", now.Add(-48*time.Hour))
	if err != nil {
		t.Fatalf("seeding the verdict: %v", err)
	}
	p.files.seedStore(domainreview.Dir+"/"+oldTree+".json", string(domainreview.Render(seeded)), now.Add(-48*time.Hour))
	p.files.seedStore(domainreview.Dir+"/"+oldTree+".prompt.md", "the old question", now.Add(-48*time.Hour))

	result := p.run(review.Options{Exec: "revisor --model {model}", KeepPrompts: 1})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	if _, ok := p.files.files[storePath(oldTree)]; !ok {
		t.Errorf("the older verdict was removed: a record is evidence")
	}
	if _, ok := p.files.files[storeFile(domainreview.Dir+"/"+oldTree+".prompt.md")]; ok {
		t.Errorf("the older question stayed: keep_prompts = 1 keeps the newest tree only")
	}
	if _, ok := p.files.files[storeFile(domainreview.Dir+"/"+currentTree+".prompt.md")]; !ok {
		t.Errorf("the newest question was removed: keep_prompts = 1 keeps it")
	}
}

// TestAReviewThatGetsNoVerdictKeepsItsQuestion: nothing was stored, so nothing
// is disposed of. The prompt is the only record of what the reviewer was asked
// when it failed, and the next run reads it instead of paying twice.
func TestAReviewThatGetsNoVerdictKeepsItsQuestion(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = "not a verdict"
	p.shell.exitCode = 1

	result := p.run(review.Options{Exec: "revisor --model {model}"})
	if result.Code == review.CodeOK {
		t.Fatalf("the run was green with no verdict: %s", result.Message)
	}
	if _, ok := p.files.files[storePath(currentTree)]; ok {
		t.Fatalf("a verdict was stored from a failure")
	}
	if prompts := storePrompts(p); len(prompts) != 1 {
		t.Fatalf("the failed run kept %v, want its one question", prompts)
	}
}

// filepathSlash is the spelling the store's names use in a test, whatever the
// separator of the machine running it.
func filepathSlash(path string) string {
	return strings.ReplaceAll(path, `\`, "/")
}
