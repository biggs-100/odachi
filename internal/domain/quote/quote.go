// Package quote compares the requirement a task document quotes with the words
// the brief carries for that feature.
//
// `odachi start` writes the same sentence twice: the bullet under the feature's
// "What must exist", and the R1 line of the task document. Only a human edits
// either, and `odachi ask --change` is the door for the second one, so the two
// drifting apart means an agent rewrote the requirement it was given. The
// comparison is bytes, not judgment: that is the whole point, because "the agent
// kept the human's words" is exactly the kind of rule that sounds right and
// cannot be checked by reading.
//
// It is pure: it reads the two texts it is given and nothing else.
package quote

import (
	"strings"

	"github.com/biggs-100/odachi/internal/domain/brief"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// Mismatch is one requirement whose words are not the brief's.
type Mismatch struct {
	// ID is the requirement, R1 or R2, as the task document names it.
	ID string
	// Brief is the bullet the brief carries for that position, empty when the
	// brief has none.
	Brief string
	// Doc is what the task document says the human said.
	Doc string
}

// Block returns the body of the brief block for one feature number, or "" when
// the brief carries no block for it. A block starts at "## <number>" and ends at
// the next header that starts with a number: the section headers inside it
// ("## What must exist") start with a word, so they do not end it.
func Block(md, feature string) string {
	feature = strings.TrimSpace(feature)
	if feature == "" {
		return ""
	}

	var out strings.Builder
	inside := false
	for _, raw := range strings.Split(md, "\n") {
		line := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if strings.HasPrefix(line, "## ") {
			header := strings.TrimSpace(strings.TrimPrefix(line, "## "))
			if startsWithNumber(header) {
				if inside {
					break
				}
				inside = strings.HasPrefix(header, feature)
				continue
			}
		}
		if inside {
			out.WriteString(line + "\n")
		}
	}
	return out.String()
}

// startsWithNumber reports whether a header begins with a feature number, which
// is what separates one brief block from the next.
func startsWithNumber(header string) bool {
	if len(header) < 4 {
		return false
	}
	for _, r := range header[:4] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// Compare returns the requirements of the task document that are not the brief's
// words, in the document's order. A feature the brief carries no block for is not
// judged, and neither is a light one: those two never wrote the sentence twice,
// and inventing a mismatch for them would make the gate refuse its own past.
func Compare(briefMD, docMD, feature string) []Mismatch {
	body := Block(briefMD, feature)
	if strings.TrimSpace(body) == "" {
		return nil
	}
	doc := taskdoc.Parse(docMD)
	if doc.Light {
		return nil
	}
	if len(doc.Requirements) == 0 {
		return nil
	}

	outcome := brief.Parse(body).Outcome
	var mismatches []Mismatch
	for i, requirement := range doc.Requirements {
		if i >= len(outcome) {
			mismatches = append(mismatches, Mismatch{ID: requirement.ID, Doc: requirement.Text})
			continue
		}
		if sameSentence(outcome[i], requirement.Text) {
			continue
		}
		mismatches = append(mismatches, Mismatch{ID: requirement.ID, Brief: outcome[i], Doc: requirement.Text})
	}
	return mismatches
}

// sameSentence compares two sentences the way the files write them: the brief
// wraps a long bullet and the tool joins it back, so the only difference that is
// not a rewrite is spacing.
func sameSentence(briefText, docText string) bool {
	return strings.Join(strings.Fields(briefText), " ") == strings.Join(strings.Fields(docText), " ")
}
