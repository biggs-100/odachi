package probe

import "testing"

func TestLimitCapsAtTheMaximum(t *testing.T) {
	if got := limit(5000); got != maxSize {
		t.Fatalf("limit(5000) = %d, want %d", got, maxSize)
	}
}

func TestLimitKeepsASmallerValue(t *testing.T) {
	if got := limit(10); got != 10 {
		t.Fatalf("limit(10) = %d, want 10", got)
	}
}
