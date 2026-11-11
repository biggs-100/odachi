package main

import (
	"bytes"
	"os"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"
)

// TestTheUsageNamesEveryVerbTheBinaryAnswers: the usage block is what a stranger
// reads first, and a verb the dispatcher answers but the block forgets is a surface
// nobody finds -`design` and `review probe` were missing. The list is read from the
// dispatcher's own switch, so there is no third list to keep in step (0273).
func TestTheUsageNamesEveryVerbTheBinaryAnswers(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}

	verbs := dispatchVerbs(t, string(source))
	// `version` closes the dispatcher today: a scan that stopped at a nested brace
	// would end before it, and fewer verbs would pass the loop below in silence.
	if !slices.Contains(verbs, "version") {
		t.Fatalf("the scan did not reach the end of the dispatcher: %v", verbs)
	}

	var out bytes.Buffer
	usage(&out)

	block := usageBlock(t, out.String())
	for _, verb := range verbs {
		if !namesVerb(block, verb) {
			t.Errorf("the usage block does not name %q, and the binary answers it", verb)
		}
	}
}

// namesVerb reports whether a line of the block offers exactly this verb. The
// boundary matters: `skill` is a prefix of `skills`, and a substring test would
// call a forgotten verb documented (0273).
func namesVerb(block, verb string) bool {
	for _, line := range strings.Split(block, "\n") {
		rest, ok := strings.CutPrefix(strings.TrimSpace(line), "odachi "+verb)
		if !ok {
			continue
		}
		if rest == "" {
			return true
		}
		next, _ := utf8.DecodeRuneInString(rest)
		if next != ' ' && next != '\t' {
			continue
		}
		return true
	}
	return false
}

// usageBlock is the list under "usage:" and nothing else: the body further down
// documents the same verbs in prose, and a verb that only lives there is a verb a
// stranger reading the block never sees (0273).
func usageBlock(t *testing.T, help string) string {
	t.Helper()

	start := strings.Index(help, "usage:")
	if start < 0 {
		t.Fatalf("the help has no usage block:\n%s", help)
	}
	rest := help[start:]
	if end := strings.Index(rest, "\n\n"); end >= 0 {
		return rest[:end]
	}
	return rest
}

// dispatchVerbs reads the case labels of the `switch args[0]` block: the verbs the
// binary answers, taken from the one place that answers them. The scan ends at the
// brace that closes the switch, which sits at one tab; a nested block inside a case
// closes deeper, and reading it as the end would drop every verb after it (0273).
func dispatchVerbs(t *testing.T, source string) []string {
	t.Helper()

	verbs := make([]string, 0, 32)
	inSwitch := false
	for _, line := range strings.Split(source, "\n") {
		if !inSwitch {
			if strings.TrimSpace(line) == "switch args[0] {" {
				inSwitch = true
			}
			continue
		}
		if line == "\t}" {
			break
		}
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, `case "`) && strings.HasSuffix(trimmed, `":`) {
			verbs = append(verbs, strings.TrimSuffix(strings.TrimPrefix(trimmed, `case "`), `":`))
		}
	}
	return verbs
}
