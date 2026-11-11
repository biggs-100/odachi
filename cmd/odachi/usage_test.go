package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// TestTheUsageNamesEveryVerbTheBinaryAnswers: the usage block is what a stranger
// reads first, and a verb the dispatcher answers but the block forgets is a surface
// nobody finds -`design` and `review probe` were missing for weeks. The list is read
// from the dispatcher's own switch, so there is no third list to keep in step (0273).
func TestTheUsageNamesEveryVerbTheBinaryAnswers(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}

	verbs := dispatchVerbs(t, string(source))
	if len(verbs) < 10 {
		t.Fatalf("read %d verbs from the dispatcher, want every one: %v", len(verbs), verbs)
	}

	var out bytes.Buffer
	usage(&out)

	block := usageBlock(t, out.String())
	for _, verb := range verbs {
		if !strings.Contains(block, "odachi "+verb) {
			t.Errorf("the usage block does not name %q, and the binary answers it", verb)
		}
	}
}

// dispatchVerbs reads the case labels of the `switch args[0]` block: the verbs the
// binary answers, taken from the one place that answers them.
func dispatchVerbs(t *testing.T, source string) []string {
	t.Helper()

	verbs := make([]string, 0, 32)
	inSwitch := false
	for _, line := range strings.Split(source, "\n") {
		trimmed := strings.TrimSpace(line)
		if !inSwitch {
			if trimmed == "switch args[0] {" {
				inSwitch = true
			}
			continue
		}
		if trimmed == "}" {
			break
		}
		if strings.HasPrefix(trimmed, `case "`) && strings.HasSuffix(trimmed, `":`) {
			verbs = append(verbs, strings.TrimSuffix(strings.TrimPrefix(trimmed, `case "`), `":`))
		}
	}
	return verbs
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
