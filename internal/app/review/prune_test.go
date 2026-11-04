package review_test

import (
	"errors"
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

// TestTheSweepKeepsEveryPromptWhenASizeCannotBeRead: the retention orders what
// it can read. When one ModTime fails the sweep prunes nothing and names the
// file, because a question removed on a guess costs a paid answer while a prompt
// left behind costs a file.
func TestTheSweepKeepsEveryPromptWhenASizeCannotBeRead(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = verdictAnswer

	oldTree := strings.Repeat("b", 40)
	p.files.seedStore(domainreview.Dir+"/"+oldTree+".prompt.md", "the old question", now.Add(-48*time.Hour))
	unreadable := storeFile(domainreview.Dir + "/" + oldTree + ".prompt.md")
	p.files.failModTime = func(path string) error {
		if path == unreadable {
			return errors.New("permission denied")
		}
		return nil
	}

	result := p.run(review.Options{Exec: "revisor --model {model}"})
	if result.Code != review.CodeOK {
		t.Fatalf("a size nobody could read blocked the verdict: code = %d (%s)", result.Code, result.Message)
	}

	joined := strings.Join(result.Lines, "\n")
	for _, want := range []string{"the prompts stayed", "could not read", oldTree + ".prompt.md", "permission denied"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the note does not name %q:\n%s", want, joined)
		}
	}
	if _, ok := p.files.files[unreadable]; !ok {
		t.Errorf("the prompt whose size nobody read was removed: a failed read is never a reason to delete a question")
	}
	if prompts := storePrompts(p); len(prompts) != 2 {
		t.Errorf("prompts = %v, want the old question and this run's", prompts)
	}
}

// TestAMidSweepRemovalNamesWhatWentAndWhatStayed: a removal that fails stops the
// sweep and keeps the file. The verdict already landed, so the note is the only
// place where a half-done sweep is visible, and it has to say the arithmetic
// instead of promising a store it did not leave.
func TestAMidSweepRemovalNamesWhatWentAndWhatStayed(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = verdictAnswer

	oldTree := strings.Repeat("b", 40)
	p.files.seedStore(domainreview.Dir+"/"+oldTree+".prompt.md", "the old question", now.Add(-48*time.Hour))
	doomed := storeFile(domainreview.Dir + "/" + oldTree + ".prompt.md")
	p.files.failRemove = func(path string) error {
		if path == doomed {
			return errors.New("permission denied")
		}
		return nil
	}

	result := p.run(review.Options{Exec: "revisor --model {model}"})
	if result.Code != review.CodeOK {
		t.Fatalf("a failed removal blocked the verdict: code = %d (%s)", result.Code, result.Message)
	}

	joined := strings.Join(result.Lines, "\n")
	for _, want := range []string{"disposed of 1 prompt and kept 1", oldTree + ".prompt.md", "permission denied"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the note does not name %q:\n%s", want, joined)
		}
	}
	if _, ok := p.files.files[doomed]; !ok {
		t.Errorf("the prompt the sweep could not remove is gone")
	}
	if _, ok := p.files.files[storeFile(domainreview.Dir+"/"+currentTree+".prompt.md")]; ok {
		t.Errorf("the first prompt the sweep removed is still there: the note says one went")
	}
}

// filepathSlash is the spelling the store's names use in a test, whatever the
// separator of the machine running it.
func filepathSlash(path string) string {
	return strings.ReplaceAll(path, `\`, "/")
}
