package archtest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/reviewtoml"
	"github.com/biggs-100/odachi/internal/domain/review"
)

// TestThisRepositoryNamesAModelForEveryLens: the review this project pays for
// must say who judges each lens, and the file has to keep decoding as the
// vocabulary grows. A retired model is a measurement no offline test can see -the
// run names it and where it came from- but a lens left without a judge is a
// mistake a reader finds only when the reviewer starts.
func TestThisRepositoryNamesAModelForEveryLens(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "odd", "review.toml"))
	if err != nil {
		t.Fatalf("read odd/review.toml: %v", err)
	}

	config, err := reviewtoml.Decode(data)
	if err != nil {
		t.Fatalf("odd/review.toml does not decode: %v", err)
	}

	for _, lens := range review.AllLenses() {
		if config.Lenses[lens] == "" {
			t.Errorf("odd/review.toml names no model for %s: the reviewer would judge it with whatever the session runs", lens)
		}
	}
}
