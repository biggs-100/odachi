package probe

import "testing"

func TestData(t *testing.T) {
	n := 5000
	want := n
	if want > maxSize {
		want = maxSize
	}
	if got := data("manual.txt"); got != want {
		t.Fatalf("data(manual.txt) = %d, want %d", got, want)
	}
}
