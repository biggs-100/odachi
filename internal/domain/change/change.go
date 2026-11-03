// Package change owns the one line a change carries when it has no document:
// the trailer a commit writes, and the three refusals a reader owes it.
package change

import (
	"fmt"
	"strings"
)

// Trailer is the key a commit writes when the diff is the whole record.
const Trailer = "Odachi-Change:"

// Document is the change a person declared in a commit message.
type Document struct {
	Line string
}

// Parse reads the trailer from the commit messages of a range. It answers the
// line, whether one was there, and the refusal when the answer is ambiguous: an
// empty trailer names nothing, and two different lines are two changes in one
// range, which a reader cannot choose between.
func Parse(messages []string) (Document, bool, error) {
	line := ""
	seen := false

	for _, message := range messages {
		for _, raw := range strings.Split(message, "\n") {
			text := strings.TrimSpace(raw)
			if !strings.HasPrefix(text, Trailer) {
				continue
			}
			value := strings.TrimSpace(strings.TrimPrefix(text, Trailer))
			if value == "" {
				return Document{}, false, fmt.Errorf("change: the trailer %q is empty: write the line it names", Trailer)
			}
			if seen && value != line {
				return Document{}, false, fmt.Errorf("change: two trailers name different changes: %q and %q", line, value)
			}
			line = value
			seen = true
		}
	}

	if !seen {
		return Document{}, false, nil
	}
	return Document{Line: line}, true, nil
}
