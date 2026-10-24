package review_test

import "testing"

// TestRedProof is deliberately red. The pull request that carries it exists to
// show that a tree with a failing suite cannot merge; it is closed, never
// merged.
func TestRedProof(t *testing.T) {
	t.Fatal("this tree is red on purpose")
}
