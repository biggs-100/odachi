package review_test

import (
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
		"opencode/deepseek-v4.1-flash-free",
		"free by the host's name",
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
