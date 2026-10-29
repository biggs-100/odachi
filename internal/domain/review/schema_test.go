package review_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
	"github.com/biggs-100/odachi/internal/domain/schema"
)

// TestAVerdictWrittenTodayCarriesItsSchema: the tool stamps the version, never
// the reviewer, the same way it stamps the tree. Without it the next field this
// file gains is an unknown key for every binary older than the change.
func TestAVerdictWrittenTodayCarriesItsSchema(t *testing.T) {
	sealed, err := review.Seal(review.Verdict{Reviewer: "humano"}, tree, "HEAD", now)
	if err != nil {
		t.Fatalf("Seal = %v", err)
	}

	rendered := string(review.Render(sealed))
	if !strings.Contains(rendered, `"schema": `+strconv.Itoa(schema.Current)) {
		t.Fatalf("the stored verdict does not carry schema %d:\n%s", schema.Current, rendered)
	}

	back, err := review.Parse(review.Render(sealed))
	if err != nil {
		t.Fatalf("Parse = %v", err)
	}
	if back.Schema != schema.Current {
		t.Errorf("Schema = %d, want %d", back.Schema, schema.Current)
	}
}

// TestAVerdictFromANewerOdachiAsksForAnUpgrade: the strict decoder meets two
// pieces of a newer verdict -the version and a field it does not know- and must
// name the version first. Naming the field sends the human to fix a file that is
// not broken; naming the upgrade sends them to the right place.
func TestAVerdictFromANewerOdachiAsksForAnUpgrade(t *testing.T) {
	future := schema.Current + 1
	answer := `{"schema":` + strconv.Itoa(future) + `,"tree":"` + tree + `","base":"HEAD","reviewer":"humano","at":"` +
		now.Format("2006-01-02T15:04:05Z07:00") + `","findings":[],"verdict":"approved","future_key":true}`

	_, err := review.Parse([]byte(answer))
	if err == nil {
		t.Fatal("Parse = no error, want a refusal")
	}
	message := err.Error()
	if !strings.Contains(message, "upgrade") {
		t.Errorf("the refusal %q does not name the upgrade", message)
	}
	if strings.Contains(message, "unknown field") {
		t.Errorf("the refusal %q names an unknown field: a newer verdict is refused by its version", message)
	}
}

// TestAVerdictWithoutASchemaIsTheFirst: every verdict stored before the field
// existed carries no version and must keep parsing, and the answer a reviewer
// writes today carries none either: the tool stamps it.
func TestAVerdictWithoutASchemaIsTheFirst(t *testing.T) {
	old := `{"tree":"` + tree + `","base":"HEAD","reviewer":"humano","at":"` +
		now.Format("2006-01-02T15:04:05Z07:00") + `","findings":[],"verdict":"approved"}`

	back, err := review.Parse([]byte(old))
	if err != nil {
		t.Fatalf("Parse of a verdict without a schema: %v", err)
	}
	if back.Schema != schema.First {
		t.Errorf("Schema = %d, want %d for a verdict written before the field", back.Schema, schema.First)
	}
}
