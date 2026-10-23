package review

import (
	"fmt"
	"strings"
)

// The project's review switch. It is a human decision, not an inference: a
// missing file means off, and the only way to turn the review on is to write the
// file with `odachi review mode on`. The gate reads the same file, so the switch
// is not a flag somebody has to remember on every commit.
const (
	// ModePath is where the switch lives, relative to the project root.
	ModePath = ".odachi/review.toml"
	// ModeOn makes `odachi check` require the verdict.
	ModeOn = "on"
	// ModeOff leaves the review optional, the same behavior as no file.
	ModeOff = "off"
)

// NormalizeMode accepts the two values the switch knows. Matching is case
// insensitive and ignores surrounding space because a human writes this file;
// anything else is refused, naming the two values, instead of being guessed.
func NormalizeMode(value string) (string, error) {
	switch normalized := strings.ToLower(strings.TrimSpace(value)); normalized {
	case ModeOn:
		return ModeOn, nil
	case ModeOff:
		return ModeOff, nil
	default:
		return "", fmt.Errorf("invalid review mode %q: use on or off", strings.TrimSpace(value))
	}
}
