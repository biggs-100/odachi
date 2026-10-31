package review_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// TestAReviewOverTheLimitRunsOneReviewerPerPart: a candidate whose prompt does not
// fit the project's limit is reviewed in parts -one reviewer per part, with the
// same lens and model- and the store holds one verdict for the tree with every
// changed path answered.
func TestAReviewOverTheLimitRunsOneReviewerPerPart(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.git.changed = []string{"a.go", "b.go"}
	p.git.blobs = map[string]string{"a.go": gitBlob, "b.go": gitBlob}
	p.git.diff = bigDiff("a.go", "b.go")
	p.shell.script = []fakeAnswer{
		{output: `{"findings":[],"passes":[{"path":"a.go","lens":"reliability"}],"inspection":{"status":"completed","paths":["a.go"]}}`},
		{output: `{"findings":[],"passes":[{"path":"b.go","lens":"reliability"}],"inspection":{"status":"completed","paths":["b.go"]}}`},
	}

	result := p.run(review.Options{Exec: "revisor --model {model}", MaxPromptKB: 25})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(p.shell.commands) != 2 {
		t.Fatalf("commands = %d, want one reviewer per part", len(p.shell.commands))
	}

	stored, ok := p.files.files[storePath(currentTree)]
	if !ok {
		t.Fatalf("no verdict stored at .git/odachi/reviews/%s.json", currentTree)
	}
	parsed, err := domainreview.Parse(stored)
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	for _, path := range []string{"a.go", "b.go"} {
		answered := slices.ContainsFunc(parsed.Passes, func(pass domainreview.Pass) bool { return pass.Path == path })
		if !answered {
			t.Errorf("the merged verdict does not answer %s: %+v", path, parsed.Passes)
		}
	}
}

// bigDiff is a candidate of two files whose hunks are big enough that a limit of
// 20 KB holds one of them and not both.
func bigDiff(paths ...string) string {
	var out strings.Builder
	for _, path := range paths {
		out.WriteString("diff --git a/" + path + " b/" + path + "\n")
		out.WriteString("index 1111111..2222222 100644\n")
		out.WriteString("--- a/" + path + "\n")
		out.WriteString("+++ b/" + path + "\n")
		out.WriteString("@@ -1,2 +1,2 @@\n")
		for i := 0; i < 120; i++ {
			out.WriteString("+var somethingThatTakesSpace = \"" + strings.Repeat("x", 100) + "\"\n")
		}
	}
	return out.String()
}

// TestAReviewOverTheSizeLimitStartsNoReviewer: a model that stalls at a prompt
// size does not answer late, it does not answer at all, and odachi knows the size
// before it spends anything. A single file whose diff cannot fit is the one case a
// split cannot answer, so the refusal names the file, both numbers and the two
// ways out, and no reviewer runs.
func TestAReviewOverTheSizeLimitStartsNoReviewer(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = `{"findings":[]}`

	result := p.run(review.Options{Exec: "revisor --model {model}", MaxPromptKB: 1})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if len(p.shell.commands) != 0 {
		t.Fatalf("commands = %v, want none: the reviewer must not be started", p.shell.commands)
	}
	for _, want := range []string{"KB", "max_prompt_kb", "a.go"} {
		if !strings.Contains(strings.Join(result.Lines, "\n"), want) {
			t.Errorf("the refusal does not carry %q:\n%s", want, strings.Join(result.Lines, "\n"))
		}
	}
}

// TestAReviewUnderTheSizeLimitRuns: the limit is the project's, and a prompt that
// fits it is the normal case.
func TestAReviewUnderTheSizeLimitRuns(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	if result := p.run(review.Options{Exec: "revisor --model {model}", MaxPromptKB: 1024}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
}

// TestThePrintedFormStillPrintsAndNamesTheSize: printing a prompt spends nothing,
// so a size over the limit cannot refuse the human their own debugging tool. It
// says the size where the run's notes go, so the human learns before pasting it
// into a reviewer that will stall.
func TestThePrintedFormStillPrintsAndNamesTheSize(t *testing.T) {
	p := newProject()

	result := p.run(review.Options{MaxPromptKB: 1})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0: the printed form spends nothing (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Prompt, "odachi review") {
		t.Error("the prompt was not printed")
	}
	if !strings.Contains(strings.Join(result.Lines, "\n"), "KB") {
		t.Errorf("the run does not name the size:\n%s", strings.Join(result.Lines, "\n"))
	}
}

// TestNoLimitMeansNoRefusal: a project that never wrote the key reviews as it
// always did. A default limit invented here would refuse prompts the human never
// agreed to refuse.
func TestNoLimitMeansNoRefusal(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	if result := p.run(review.Options{Exec: "revisor --model {model}"}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
}
