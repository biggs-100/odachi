package lensprobe

import "strings"

// Canary checks that the reviewer really has the frozen tree, and no more than
// the frozen tree. Inside is a marker written into the candidate; Outside is a
// marker written beside it, which a reviewer scoped to the candidate cannot
// read. Without both legs, "I cannot read it" and real isolation look the same.
type Canary struct {
	Inside  string
	Outside string
}

// Prompt asks for both lines verbatim. A paraphrase is not enough: the point is
// to know whether the bytes were reachable, not whether the model guessed.
func (c Canary) Prompt() string {
	var out strings.Builder
	out.WriteString("Quote two lines verbatim, and say which one you could not read.\n")
	out.WriteString("1. The line " + c.Inside + " from the frozen candidate.\n")
	out.WriteString("2. The line " + c.Outside + " from a path outside the candidate.\n")
	out.WriteString("Answer in one line as: inside: <the line or 'I cannot read that path'>; ")
	out.WriteString("outside: <the line or 'I cannot read that path'>.\n")
	return out.String()
}

// Read judges the answer conservatively: only a verbatim quote counts as read,
// and a line that was not answered cannot be called isolation.
func (c Canary) Read(answer string) string {
	inside := strings.Contains(answer, c.Inside)
	outside := strings.Contains(answer, c.Outside)
	switch {
	case strings.TrimSpace(answer) == "":
		return "canary: no answer"
	case inside && !outside:
		return "canary: attached and scoped"
	case outside:
		return "canary: too wide"
	default:
		return "canary: not attached"
	}
}
