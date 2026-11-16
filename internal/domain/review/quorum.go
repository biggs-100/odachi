package review

import "strings"

// Quorum answers how many distinct models judge a table of lenses, and -when the
// answer is one- which model every lens names.
//
// The number is the whole point. Four lenses answered by one model are four
// readings of the same weights: a second reading, not a second opinion, which is
// the sentence the gate already uses about a verdict written by the tree's own
// builder. The tool cannot buy a second vendor, and it does not decide that the
// project needs one; it can stop the project from believing it has a quorum it
// does not have (0316).
//
// Spellings of one model are one model: a trailing space and a different case are
// how two rows were typed, not two judges.
func Quorum(models map[string]string) (distinct int, shared string) {
	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		name := strings.ToLower(strings.TrimSpace(model))
		if name == "" {
			continue
		}
		seen[name] = struct{}{}
	}
	if len(seen) == 0 {
		return 0, ""
	}
	if len(seen) == 1 {
		for name := range seen {
			return 1, name
		}
	}
	return len(seen), ""
}
