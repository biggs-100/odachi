package units_test

import (
	"testing"

	"github.com/biggs-100/odachi/internal/domain/units"
)

// TestBytesIsTheFormatAPersonReads pins the four cases: bytes under a kilobyte,
// one decimal below ten of them, whole kilobytes below a megabyte, one decimal
// above it. A formatter that drifts here would change every line that prints a
// size, in two verbs at once.
func TestBytesIsTheFormatAPersonReads(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{5 * 1024, "5.0 KB"},
		{10 * 1024, "10 KB"},
		{1536 * 1024, "1.5 MB"},
		{2 * 1024 * 1024, "2.0 MB"},
	}
	for _, c := range cases {
		if got := units.Bytes(c.n); got != c.want {
			t.Errorf("Bytes(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}
