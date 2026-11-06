package doctor

import (
	"strconv"
	"strings"
	"testing"
)

// TestTheEnvironmentNamesACutPath: a PATH longer than cmd can pass is the reason
// a proof fails with a binary that disappeared, and the line has to name the
// length, the limit and the repair (0210).
func TestTheEnvironmentNamesACutPath(t *testing.T) {
	long := strings.Repeat("a", MaxChildPath+1)

	got := Environment(long, "windows")

	if got.Status != StatusWarning {
		t.Fatalf("a cut PATH must be a warning, got %s: %s", got.Status, got.Detail)
	}
	for _, want := range []string{strconv.Itoa(len(long)), strconv.Itoa(MaxChildPath), "shorter PATH"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("the line must say %q, got %q", want, got.Detail)
		}
	}
}

// TestTheEnvironmentAtTheLimitIsOk: the boundary belongs to the platform, and one
// character less is still a PATH a child receives whole.
func TestTheEnvironmentAtTheLimitIsOk(t *testing.T) {
	got := Environment(strings.Repeat("a", MaxChildPath), "windows")

	if got.Status != StatusOK {
		t.Fatalf("a PATH at the limit is ok, got %s: %s", got.Status, got.Detail)
	}
	if !strings.Contains(got.Detail, strconv.Itoa(MaxChildPath)) {
		t.Errorf("the line must name the limit, got %q", got.Detail)
	}
}

// TestTheEnvironmentCountsCharactersNotBytes: cmd's ceiling is a character
// count, and a PATH with an accented folder name is longer in bytes than the
// platform sees. A path of MaxChildPath characters is at the limit whatever it
// weighs, and the line prints characters (0210's review).
func TestTheEnvironmentCountsCharactersNotBytes(t *testing.T) {
	// Each ñ is two bytes in UTF-8.
	atLimit := strings.Repeat("ñ", MaxChildPath)

	got := Environment(atLimit, "windows")

	if got.Status != StatusOK {
		t.Fatalf("a PATH of %d characters is at the limit, got %s: %s", MaxChildPath, got.Status, got.Detail)
	}
	if !strings.Contains(got.Detail, strconv.Itoa(MaxChildPath)) {
		t.Errorf("the line must print characters, not bytes, got %q", got.Detail)
	}

	if over := Environment(atLimit+"ñ", "windows"); over.Status != StatusWarning {
		t.Fatalf("one character over the limit is a warning, got %s: %s", over.Status, over.Detail)
	}
}

// TestTheEnvironmentDoesNotCutElsewhere: the limit is Windows', and the platform
// that does not cut the environment says which one it measured.
func TestTheEnvironmentDoesNotCutElsewhere(t *testing.T) {
	got := Environment(strings.Repeat("a", MaxChildPath*2), "linux")

	if got.Status != StatusOK {
		t.Fatalf("a platform that does not cut the PATH is ok, got %s: %s", got.Status, got.Detail)
	}
	for _, want := range []string{"does not cut", "linux"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("the line must say %q, got %q", want, got.Detail)
		}
	}
}
