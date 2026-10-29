package schema_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/schema"
)

// TestAnAbsentValueIsTheFirstSchema: a file written before the field existed
// carries no number, and it must read as the first schema forever - not as the
// current one, or the day Current moves the oldest files gain fields they never
// had.
func TestAnAbsentValueIsTheFirstSchema(t *testing.T) {
	got, err := schema.Read(0)
	if err != nil {
		t.Fatalf("Read(0) = %v, want no error", err)
	}
	if got != schema.First {
		t.Errorf("Read(0) = %d, want %d", got, schema.First)
	}
}

// TestAValueThisBinaryUnderstandsIsItsOwn: the version a file declares is
// answered unchanged, so a reader can compare it with what it can handle.
func TestAValueThisBinaryUnderstandsIsItsOwn(t *testing.T) {
	got, err := schema.Read(schema.Current)
	if err != nil {
		t.Fatalf("Read(Current) = %v, want no error", err)
	}
	if got != schema.Current {
		t.Errorf("Read(Current) = %d, want %d", got, schema.Current)
	}
}

// TestAValueFromANewerOdachiAsksForAnUpgrade: this is the whole point of the
// field. A binary that meets a number above its own must say so in the words a
// human can act on, and must not name the newer file's keys as unknown.
func TestAValueFromANewerOdachiAsksForAnUpgrade(t *testing.T) {
	future := schema.Current + 1
	_, err := schema.Read(future)
	if err == nil {
		t.Fatalf("Read(%d) = no error, want a refusal", future)
	}
	message := err.Error()
	for _, want := range []string{"upgrade", "odachi", "schema " + strconv.Itoa(future), "schema " + strconv.Itoa(schema.Current)} {
		if !strings.Contains(message, want) {
			t.Errorf("the refusal %q does not name %q", message, want)
		}
	}
	if strings.Contains(message, "unknown") {
		t.Errorf("the refusal %q names an unknown key, which sends the human to fix a file that is not broken", message)
	}
}

// TestAVersionBelowTheFirstIsATypo: a number this tool never writes is not a
// file from the past, it is a broken file; reading it as absence would hide it.
func TestAVersionBelowTheFirstIsATypo(t *testing.T) {
	if _, err := schema.Read(-1); err == nil {
		t.Fatal("Read(-1) = no error, want a refusal")
	}
}
