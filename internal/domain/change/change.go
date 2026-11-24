// Package change owns the lines a commit writes about itself: the trailer a
// change without a document carries, and the risk its body declares, with the
// refusals a reader owes each one.
package change

import (
	"fmt"
	"slices"
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

// RiskKey is the line a commit closes with when its body names the risk the
// change carries, or says it carries none.
const RiskKey = "Risk:"

// noRisk is the one declaration that names nothing: the change claims it cannot
// hurt. It answers the question, so it is a declaration, and it is never a name.
const noRisk = "none"

// Risk is what the commits of a range declared about their blast radius. Declared
// says the range answered the question; Named is what it answered with, lowercase
// and without repeats, in the order the commits first said it. NoneReason is the
// first written explanation attached to none, and NoneWithoutReason records a
// none declaration that did not carry one.
type Risk struct {
	Declared          bool
	Named             []string
	NoneReason        string
	NoneWithoutReason bool
}

// ParseRisk reads the risk line from the commit messages of a range. It answers
// what the range declared, and refuses an empty line: `Risk:` with nothing after
// it answered nothing, and reading it as `none` would let a missing answer pass as
// a harmless change. A missing line is not a refusal -it is the silence of a
// project that never wrote one-, and the callers decide what silence means.
func ParseRisk(messages []string) (Risk, error) {
	risk := Risk{}

	for _, message := range messages {
		for _, raw := range strings.Split(message, "\n") {
			text := strings.TrimSpace(raw)
			if !strings.HasPrefix(text, RiskKey) {
				continue
			}
			value := strings.TrimSpace(strings.TrimPrefix(text, RiskKey))
			name := riskName(value)
			if name == "" {
				return Risk{}, fmt.Errorf("change: the line %q is empty: write the risk it carries, or none", RiskKey)
			}
			risk.Declared = true

			if name == noRisk {
				reason := noneReason(value)
				if risk.NoneReason == "" {
					risk.NoneReason = reason
				}
				if reason == "" {
					risk.NoneWithoutReason = true
				}
				continue
			}
			if slices.Contains(risk.Named, name) {
				continue
			}
			risk.Named = append(risk.Named, name)
		}
	}

	return risk, nil
}

// riskName is the first word of the declaration: the table names one risk, and
// the rest of the line is the human's prose about it. Case does not decide, and
// the colon a writer puts after the name is punctuation, not part of it.
func riskName(value string) string {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return ""
	}
	return strings.ToLower(strings.TrimRight(fields[0], ",;.:"))
}

// noneReason reads the prose after a none declaration. A colon attached to the
// name and a separated hyphen, colon or em dash are punctuation, not part of the
// reason. Anything else after none is not a written reason the gate can rely on.
func noneReason(value string) string {
	fields := strings.Fields(value)
	if len(fields) < 2 {
		return ""
	}

	rest := strings.TrimSpace(value[len(fields[0]):])
	if strings.HasSuffix(fields[0], ":") {
		return rest
	}
	for _, separator := range []string{"-", ":", "—"} {
		if strings.HasPrefix(rest, separator) {
			return strings.TrimSpace(strings.TrimPrefix(rest, separator))
		}
	}
	return ""
}
