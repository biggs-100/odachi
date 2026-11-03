package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// TestAnAttachedModelThatIsNotAnIdentifierIsRefused: a model replaces {model}
// inside a reviewer command, so a display name with spaces is not one. Attach
// refuses it before anything is written, naming the shape, because a verdict that
// stored it would make every later correction die on the lens table (0184).
func TestAnAttachedModelThatIsNotAnIdentifierIsRefused(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"agente","model":"DeepSeek V4.1 Flash","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}`)

	result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want %d (%s)", result.Code, review.CodeBlocked, result.Message)
	}
	for _, want := range []string{"DeepSeek V4.1 Flash", "identifier"} {
		if !strings.Contains(result.Message, want) {
			t.Errorf("the refusal does not carry %q: %s", want, result.Message)
		}
	}
	if _, stored := p.files.files[storePath(currentTree)]; stored {
		t.Errorf("a verdict with a model that is not an identifier was stored")
	}
}

// TestAnAttachedModelThatIsAnIdentifierIsStored: the refusal must not refuse a
// model the tool could substitute. The identifier is what the prompt asks the
// reviewer to write, and the stored verdict keeps it (0184).
func TestAnAttachedModelThatIsAnIdentifierIsStored(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"agente","model":"opencode-go/deepseek-v4.1-flash","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}`)

	result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	stored, ok := p.files.files[storePath(currentTree)]
	if !ok {
		t.Fatalf("an attach with an identifier stored no verdict")
	}
	parsed, err := domainreview.Parse(stored)
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if parsed.Model != "opencode-go/deepseek-v4.1-flash" {
		t.Errorf("model = %q, want the identifier the answer named", parsed.Model)
	}
}
