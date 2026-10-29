package reviewtoml_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/reviewtoml"
	"github.com/biggs-100/odachi/internal/domain/review"
	"github.com/biggs-100/odachi/internal/domain/schema"
)

// TestRenderWritesTheSchemaItWasWrittenWith: the file says which odachi wrote it,
// so the next key this file gains arrives as an upgrade and not as a typo.
func TestRenderWritesTheSchemaItWasWrittenWith(t *testing.T) {
	data, err := reviewtoml.Render(review.ProjectConfig{Mode: review.ModeOn})
	if err != nil {
		t.Fatalf("Render = %v", err)
	}
	if !strings.Contains(string(data), "schema = "+strconv.Itoa(schema.Current)) {
		t.Fatalf("Render = %q, want the schema written", data)
	}

	if _, err := reviewtoml.Decode(data); err != nil {
		t.Fatalf("Decode(Render) = %v", err)
	}
}

// TestDecodeOfANewerFileAsksForAnUpgrade: a file that carries a version this
// binary does not know is newer, not broken. The strict decoder meets both the
// version and the keys that came with it, and must name the version: naming a key
// sends the human to fix a file that is well formed.
func TestDecodeOfANewerFileAsksForAnUpgrade(t *testing.T) {
	future := schema.Current + 1
	_, err := reviewtoml.Decode([]byte("schema = " + strconv.Itoa(future) + "\nmode = \"on\"\nfuture_key = 1\n"))
	if err == nil {
		t.Fatal("Decode accepted a file from a newer odachi")
	}
	message := err.Error()
	if !strings.Contains(message, "upgrade") {
		t.Errorf("the refusal %q does not name the upgrade", message)
	}
	if strings.Contains(message, "unknown keys") {
		t.Errorf("the refusal %q names unknown keys: a newer file is refused by its version", message)
	}
}

// TestDecodeOfAFileWithoutASchemaStillWorks: every odd/review.toml written before
// the field existed carries no version and must keep working untouched.
func TestDecodeOfAFileWithoutASchemaStillWorks(t *testing.T) {
	got, err := reviewtoml.Decode([]byte("mode = \"on\"\nmax_prompt_kb = 96\n"))
	if err != nil {
		t.Fatalf("Decode = %v", err)
	}
	if got.Mode != review.ModeOn || got.MaxPromptKB != 96 {
		t.Errorf("Decode = %+v, want the file read as it always was", got)
	}
}
