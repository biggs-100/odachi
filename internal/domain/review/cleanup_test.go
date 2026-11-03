package review_test

import (
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// day is a fixed clock: the retention is counted in the files a store holds, and
// a test that reads the wall clock would fail on a slow machine.
func atDay(n int) time.Time {
	return time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, n)
}

func hashOf(letter string) string {
	return strings.Repeat(letter, 40)
}

// TestNewestTreesOrdersByTheNewestFileAndBreaksTiesOnTheName pins the order both
// lifetimes count: the store's trees, newest first, with a tie that cannot depend
// on the order the directory was listed in.
func TestNewestTreesOrdersByTheNewestFileAndBreaksTiesOnTheName(t *testing.T) {
	files := []review.StoreFile{
		{Name: hashOf("a") + ".json", At: atDay(1)},
		{Name: hashOf("b") + ".json", At: atDay(2)},
		{Name: hashOf("c") + ".json", At: atDay(2)},
		{Name: hashOf("e") + ".prompt.md", At: atDay(3)},
		{Name: hashOf("e") + ".json", At: atDay(1)},
	}

	got := review.NewestTrees(files)
	want := []string{hashOf("e"), hashOf("c"), hashOf("b"), hashOf("a")}
	if len(got) != len(want) {
		t.Fatalf("NewestTrees = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("NewestTrees = %v, want %v", got, want)
		}
	}
}

// TestPromptsToForgetCountsTreesNotPrompts is the rule the verb and the store
// share: the newest keep trees keep their questions, a tree counts with every
// file it holds even when it has no prompt, and a keep of zero forgets them all.
func TestPromptsToForgetCountsTreesNotPrompts(t *testing.T) {
	newest, middle, oldest := hashOf("c"), hashOf("b"), hashOf("a")
	files := []review.StoreFile{
		// The newest tree was judged from an attached verdict: it holds no prompt
		// and still spends the first place of the retention.
		{Name: newest + ".json", At: atDay(4)},
		{Name: middle + ".prompt.md", At: atDay(3)},
		{Name: middle + ".prompt-risk+readability.md", At: atDay(3)},
		{Name: middle + ".json", At: atDay(3)},
		{Name: oldest + ".prompt.md", At: atDay(2)},
		{Name: oldest + ".json", At: atDay(2)},
	}

	cases := []struct {
		name string
		keep int
		want []string
	}{
		{
			name: "a keep of zero forgets every prompt",
			keep: 0,
			want: []string{middle + ".prompt.md", middle + ".prompt-risk+readability.md", oldest + ".prompt.md"},
		},
		{
			name: "the newest tree has no prompt, and still spends the first place",
			keep: 1,
			want: []string{middle + ".prompt.md", middle + ".prompt-risk+readability.md", oldest + ".prompt.md"},
		},
		{
			name: "two trees survive, and both prompts of the newest of them",
			keep: 2,
			want: []string{oldest + ".prompt.md"},
		},
		{
			name: "a keep larger than the store forgets nothing",
			keep: 9,
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := review.PromptsToForget(files, tc.keep)
			if len(got) != len(tc.want) {
				t.Fatalf("PromptsToForget(keep %d) = %v, want %v", tc.keep, got, tc.want)
			}
			for _, name := range tc.want {
				if !got[name] {
					t.Fatalf("PromptsToForget(keep %d) forgot %v of %v", tc.keep, name, tc.want)
				}
			}
		})
	}
}

// TestIsPromptReadsTheNamesTheStoreWrites keeps the classification where the
// naming lives: a question is a file whose name says prompt, and a verdict, a
// refutation or a file somebody else named is not.
func TestIsPromptReadsTheNamesTheStoreWrites(t *testing.T) {
	hash := hashOf("a")
	cases := []struct {
		name string
		want bool
	}{
		{hash + ".prompt.md", true},
		{hash + ".prompt-risk+readability.md", true},
		{hash + ".prompt-risk-part-1of2.md", true},
		{hash + ".json", false},
		{hash + ".refuted.json", false},
		{"prompt.md", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := review.IsPrompt(tc.name); got != tc.want {
			t.Errorf("IsPrompt(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}
