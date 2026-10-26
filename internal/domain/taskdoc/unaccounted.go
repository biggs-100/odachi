package taskdoc

import "strings"

// Unaccounted returns, in file order, the non-empty lines of a document that
// Parse ignores. Render cannot carry them over, so they are exactly the content
// a canonical rewrite would destroy: an extra section, a note dropped between
// steps, a line that does not match the canonical shape of its section.
//
// It is the question the write guard asks before replacing a document. The
// `## Historial` never appears here: it is kept verbatim, which is why the guard
// tells the human to move their notes there.
func Unaccounted(md string) []string {
	lines := splitLines(md)

	// Parse keeps the LAST title, so every earlier one is lost. Knowing which
	// line survives needs the whole document, not a running flag.
	lastTitle := -1
	for i, line := range lines {
		if strings.HasPrefix(line, headerTitle) {
			lastTitle = i
		}
	}

	var out []string
	current := ""
	// preamble mirrors Parse: only the lines between the title and the first
	// section can be the base field, so every other position is unaccounted.
	preamble := true

	for i, line := range lines {
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "## ") {
			current = section(line)
			preamble = false
			if current == "" {
				out = append(out, line)
			}
			continue
		}
		if strings.HasPrefix(line, headerTitle) {
			if i != lastTitle {
				out = append(out, line)
			}
			// Parse closes whatever section was open at any title line.
			current = ""
			continue
		}
		if strings.HasPrefix(line, "#") {
			out = append(out, line)
			continue
		}

		switch current {
		case "requirements":
			if _, ok := parseRequirement(line); !ok {
				out = append(out, line)
			}
		case "donewhen":
			if _, ok := bullet(line); !ok {
				out = append(out, line)
			}
		case "steps":
			if _, ok := parseStep(line); !ok {
				out = append(out, line)
			}
		case "history":
			// The history is kept verbatim; nothing here is ever lost.
		default:
			// The base line is the one preamble line Parse reads; anything else
			// outside the canonical sections is lost by a rewrite.
			if preamble && strings.HasPrefix(line, basePrefix) {
				continue
			}
			out = append(out, line)
		}
	}

	return out
}

// splitLines is the line walk Parse and Unaccounted share: trailing CRs gone,
// surrounding spaces trimmed, blank lines dropped.
func splitLines(md string) []string {
	lines := make([]string, 0, strings.Count(md, "\n")+1)
	for _, raw := range strings.Split(md, "\n") {
		lines = append(lines, strings.TrimSpace(strings.TrimSuffix(raw, "\r")))
	}
	return lines
}
