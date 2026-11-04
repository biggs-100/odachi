package review_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
)

// verdict is one answer the fake reviewer gives: a complete inspection with no
// findings, which is what a run needs to store a verdict.
const verdict = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

// TestAModelTheHostDoesNotServeIsRefusedBeforeTheReviewerStarts is the defect the
// preflight exists for: a provider retired a model, the run waited for the whole
// timeout, and the failure named no way out. The host can now say which models it
// serves, so a model it no longer lists is refused before the call is paid for,
// and the refusal names the lens, the model, its source and the options.
func TestAModelTheHostDoesNotServeIsRefusedBeforeTheReviewerStarts(t *testing.T) {
	p := newProject()
	p.host.model = "opencode-go/deepseek-v4.1-flash#max"
	p.host.models = []string{
		"opencode-go/beam-free",
		"opencode-go/deepseek-v4.1-flash",
		"opencode/deepseek-v4.1-flash-free",
	}
	p.host.known = true

	result := p.run(review.Options{
		Exec:    "revisor --model {model}",
		Lens:    "reliability,risk",
		Because: "the test refuses a retired model",
		LensModels: map[string]string{
			"risk":        "opencode-go/retired-model",
			"reliability": "opencode-go/deepseek-v4.1-flash",
		},
	})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want %d (%s)", result.Code, review.CodeBlocked, result.Message)
	}
	for _, want := range []string{
		"risk",
		"opencode-go/retired-model",
		"odd/review.toml",
		"opencode-go/deepseek-v4.1-flash#max",
		"the host serves",
		"opencode-go/beam-free",
		"opencode-go/deepseek-v4.1-flash",
		"opencode/deepseek-v4.1-flash-free",
		"--model",
		"ODACHI_MODEL",
		"does not switch by itself",
		"the reviewer was not started",
	} {
		if !strings.Contains(result.Message, want) {
			t.Errorf("the refusal does not carry %q:\n%s", want, result.Message)
		}
	}
	if p.shell.asked {
		t.Errorf("the reviewer was started for a model the host does not serve: %v", p.shell.commands)
	}
}

// TestAModelFromAProviderTheHostDoesNotServeIsLeftToTheCommand: the catalog is
// OpenCode's, and a project may run another CLI. A provider the host does not
// know is not a model the host can judge, so the run starts.
func TestAModelFromAProviderTheHostDoesNotServeIsLeftToTheCommand(t *testing.T) {
	p := newProject()
	p.host.model = "opencode-go/deepseek-v4.1-flash#max"
	p.host.models = []string{"opencode-go/deepseek-v4.1-flash"}
	p.host.known = true
	p.shell.output = verdict

	result := p.run(review.Options{
		Exec:       "revisor --model {model}",
		Lens:       "risk",
		Because:    "the test leaves another provider to its CLI",
		LensModels: map[string]string{"risk": "other/retired-model"},
	})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !p.shell.asked {
		t.Errorf("the reviewer never ran for a provider the host cannot judge")
	}
}

// TestAHostThatCannotListItsModelsDoesNotRefuse: the catalog is an extra
// question, not a new requirement. When the host cannot answer, the run keeps
// the behaviour it had and the failure keeps what it knows.
func TestAHostThatCannotListItsModelsDoesNotRefuse(t *testing.T) {
	p := newProject()
	p.host.model = "opencode-go/deepseek-v4.1-flash#max"
	p.host.known = false
	p.shell.output = verdict

	result := p.run(review.Options{
		Exec:       "revisor --model {model}",
		Lens:       "risk",
		Because:    "the test runs without a catalog",
		LensModels: map[string]string{"risk": "opencode-go/retired-model"},
	})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !p.shell.asked {
		t.Errorf("the reviewer never ran when the host could not list its models")
	}
}

// TestAFailedReviewerNamesTheOptions: the preflight cannot see a model that is
// listed and still does not answer. The three ways a reviewer fails after it
// started -could not run, timeout, non-zero exit- name the same options the
// refusal names, so the human does not have to read the store to find a way out.
func TestAFailedReviewerNamesTheOptions(t *testing.T) {
	cases := []struct {
		name  string
		setup func(p *project)
	}{
		{name: "could not run", setup: func(p *project) { p.shell.err = errors.New("no such executable") }},
		{name: "timeout", setup: func(p *project) { p.shell.timedOut = true }},
		{name: "non-zero exit", setup: func(p *project) { p.shell.exitCode = 1; p.shell.stderr = "error: model not found" }},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := newProject()
			p.host.model = "opencode-go/deepseek-v4.1-flash#max"
			p.host.models = []string{
				"opencode-go/beam-free",
				"opencode-go/deepseek-v4.1-flash",
				"opencode/deepseek-v4.1-flash-free",
			}
			p.host.known = true
			c.setup(p)

			result := p.run(review.Options{
				Exec:       "revisor --model {model}",
				Lens:       "risk",
				Because:    "the test pins a failure after the preflight",
				LensModels: map[string]string{"risk": "opencode-go/beam-free"},
			})

			if result.Code == review.CodeOK {
				t.Fatalf("a failed reviewer answered ready: %s", result.Message)
			}
			if !p.shell.asked {
				t.Fatalf("the reviewer never ran: the preflight refused a listed model")
			}

			joined := result.Message + "\n" + strings.Join(result.Lines, "\n")
			for _, want := range []string{
				"risk",
				"opencode-go/beam-free",
				"from odd/review.toml",
				"the session's model is opencode-go/deepseek-v4.1-flash#max",
				"opencode/deepseek-v4.1-flash-free",
				"does not switch by itself",
			} {
				if !strings.Contains(joined, want) {
					t.Errorf("the failure does not carry %q:\n%s", want, joined)
				}
			}
		})
	}
}

// TestAFailureWithNoHostStillNamesTheWaysToSwitch: the catalog is an extra
// question, and a host that cannot answer must not take the way out with it. The
// three ways to switch are a flag, the project's file and a variable, so they
// are named even when the tool knows nothing else.
func TestAFailureWithNoHostStillNamesTheWaysToSwitch(t *testing.T) {
	p := newProject()
	p.host.model = "opencode-go/deepseek-v4.1-flash#max"
	p.host.known = false
	p.shell.exitCode = 1

	result := p.run(review.Options{
		Exec:       "revisor --model {model}",
		Lens:       "risk",
		Because:    "the test pins the no-host failure",
		LensModels: map[string]string{"risk": "opencode-go/retired-model"},
	})

	if result.Code == review.CodeOK {
		t.Fatalf("a failed reviewer answered ready: %s", result.Message)
	}
	for _, want := range []string{"--model", "odd/review.toml", "ODACHI_MODEL", "does not switch by itself"} {
		if !strings.Contains(result.Message, want) {
			t.Errorf("the failure does not carry %q:\n%s", want, result.Message)
		}
	}
}

// TestTheCatalogIsAskedOnce: a message that names the options must not pay a
// probe per group, and two groups must see the same list.
func TestTheCatalogIsAskedOnce(t *testing.T) {
	p := newProject()
	p.host.model = "opencode-go/deepseek-v4.1-flash#max"
	p.host.models = []string{"opencode-go/beam-free", "opencode-go/deepseek-v4.1-flash"}
	p.host.known = true
	p.shell.output = verdict

	result := p.run(review.Options{
		Exec:    "revisor --model {model}",
		Lens:    "reliability,risk",
		Because: "the test splits the run by model",
		LensModels: map[string]string{
			"risk":        "opencode-go/beam-free",
			"reliability": "opencode-go/deepseek-v4.1-flash",
		},
	})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if p.host.calls != 1 {
		t.Errorf("the catalog was asked %d times, want once", p.host.calls)
	}
}
