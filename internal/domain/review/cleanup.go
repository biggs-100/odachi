package review

import (
	"sort"
	"strings"
	"time"
)

// StoreFile is one file of the reviews directory as a caller measured it: its
// name relative to that directory, and when it was written. The retention counts
// trees, and a tree is every file whose name starts with its hash.
type StoreFile struct {
	Name string
	At   time.Time
}

// TreeOf is the tree a store file belongs to: the hash every name starts with.
func TreeOf(name string) string {
	if at := strings.Index(name, "."); at > 0 {
		return name[:at]
	}
	return name
}

// IsPrompt reports whether a store file is a question: its name says so right
// after the tree. A verdict, a refutation or a name somebody else wrote is not.
func IsPrompt(name string) bool {
	rest := strings.TrimPrefix(name, TreeOf(name))
	return strings.HasPrefix(rest, ".prompt")
}

// NewestTrees orders the store's trees by the newest file each holds, and breaks
// a tie on the name, so the answer never depends on the order the directory was
// listed in. The order is the one every lifetime counts: a prompt and a verdict
// about the same tree are the same day's evidence.
func NewestTrees(files []StoreFile) []string {
	newest := map[string]time.Time{}
	for _, file := range files {
		tree := TreeOf(file.Name)
		if at, seen := newest[tree]; !seen || file.At.After(at) {
			newest[tree] = file.At
		}
	}

	trees := make([]string, 0, len(newest))
	for tree := range newest {
		trees = append(trees, tree)
	}
	sort.Slice(trees, func(i, j int) bool {
		left, right := newest[trees[i]], newest[trees[j]]
		if !left.Equal(right) {
			return left.After(right)
		}
		return trees[i] > trees[j]
	})
	return trees
}

// PromptsToForget answers the prompt files a retention of keep stopped keeping:
// every question but the newest keep trees' own. A tree counts with every file it
// holds -a verdict, a refutation, a prompt- because the lifetime counts the
// store's trees and not the files a run happened to write. It is the rule
// `odachi clean` reports and the store obeys when a verdict lands, so the two
// cannot disagree about keep_prompts.
func PromptsToForget(files []StoreFile, keep int) map[string]bool {
	prompts := map[string][]string{}
	for _, file := range files {
		if !IsPrompt(file.Name) {
			continue
		}
		tree := TreeOf(file.Name)
		prompts[tree] = append(prompts[tree], file.Name)
	}

	forget := map[string]bool{}
	for index, tree := range NewestTrees(files) {
		if index < keep {
			continue
		}
		for _, name := range prompts[tree] {
			forget[name] = true
		}
	}
	return forget
}
