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

	got := Environment(long, true)

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
	got := Environment(strings.Repeat("a", MaxChildPath), true)

	if got.Status != StatusOK {
		t.Fatalf("a PATH at the limit is ok, got %s: %s", got.Status, got.Detail)
	}
	if !strings.Contains(got.Detail, strconv.Itoa(MaxChildPath)) {
		t.Errorf("the line must name the limit, got %q", got.Detail)
	}
}

// TestTheEnvironmentDoesNotCutElsewhere: the limit is Windows', and a platform
// that does not cut the environment says so instead of warning for nothing.
func TestTheEnvironmentDoesNotCutElsewhere(t *testing.T) {
	got := Environment(strings.Repeat("a", MaxChildPath*2), false)

	if got.Status != StatusOK {
		t.Fatalf("a platform that does not cut the PATH is ok, got %s: %s", got.Status, got.Detail)
	}
	if !strings.Contains(got.Detail, "does not cut") {
		t.Errorf("the line must say the platform does not cut it, got %q", got.Detail)
	}
}
