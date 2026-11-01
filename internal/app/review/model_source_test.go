package review_test

import (
	"strings"
	"testing"

	reviewapp "github.com/biggs-100/odachi/internal/app/review"
	"github.com/biggs-100/odachi/internal/domain/review"
)

// TestAFailedReviewerNamesTheModelAndItsSource is the defect of the day: a
// provider retired a model, and the run said only that the command exited. An
// offline tool cannot know what a provider serves; what it can do is name the
// model it asked for, the lens that asked, and the line that chose it, so the fix
// is one edit instead of an investigation.
func TestAFailedReviewerNamesTheModelAndItsSource(t *testing.T) {
	cases := []struct {
		name  string
		want  string
		setup func(p *project)
		opts  reviewapp.Options
	}{
		{
			name: "the project's file",
			want: "from " + review.ModePath,
			// Every lens the fixture's diff asks for needs an entry: a lens with
			// no model is a different refusal, and this test is about a model
			// that was chosen and did not answer.
			opts: reviewapp.Options{LensModels: map[string]string{"risk": "opencode-go/retired-model", "reliability": "opencode-go/retired-model"}},
		},
		{
			name: "the flag",
			want: "from --model",
			opts: reviewapp.Options{Model: "opencode-go/retired-model"},
		},
		{
			name: "the session",
			want: "from the session",
			setup: func(p *project) {
				p.host.model = "opencode-go/retired-model"
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := newProject()
			if c.setup != nil {
				c.setup(p)
			}
			p.shell.exitCode = 1
			p.shell.stderr = "error: model not found"

			opts := c.opts
			opts.Exec = "revisor --model {model}"
			opts.Lens = "risk"
			opts.Because = "the test pins a failure"

			result := p.run(opts)
			if result.Code == reviewapp.CodeOK {
				t.Fatalf("a failed reviewer answered ready: %s", result.Message)
			}

			joined := result.Message + "\n" + strings.Join(result.Lines, "\n")
			for _, want := range []string{"risk", "opencode-go/retired-model", c.want} {
				if !strings.Contains(joined, want) {
					t.Errorf("the failure does not name %q:\n%s", want, joined)
				}
			}
		})
	}
}
