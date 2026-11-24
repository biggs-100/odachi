package main

import (
	"bytes"
	"os"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/biggs-100/odachi/internal/app/start"
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

// TestTheVerbCatalogNamesExactlyTheDispatchedVerbs: the catalog and the switch
// are two views of one list. A verb the switch answers but the catalog omits is
// invisible to the help and the error; a catalog entry the switch does not
// answer promises a command that does not exist. Checking both directions keeps
// the two from drifting apart (0360).
func TestTheVerbCatalogNamesExactlyTheDispatchedVerbs(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	dispatched := dispatchVerbs(t, string(source))

	catalog := make([]string, 0, len(verbCatalog))
	for _, entry := range verbCatalog {
		catalog = append(catalog, entry.name)
	}

	for _, verb := range dispatched {
		if !slices.Contains(catalog, verb) {
			t.Errorf("the dispatcher answers %q but the verb catalog omits it", verb)
		}
	}
	for _, verb := range catalog {
		if !slices.Contains(dispatched, verb) {
			t.Errorf("the verb catalog names %q but the dispatcher does not answer it", verb)
		}
	}
}

// pinnedUsageBlock is the exact block the binary prints today. The refactor is
// only allowed to change where the bytes come from, never the bytes themselves,
// so a stray space in a catalog entry fails here instead of reaching a human
// reading `odachi start -h` (0360).
const pinnedUsageBlock = `usage:
  odachi                      the same as odachi start
  odachi start [flags]
  odachi start --light "<requirement>"
  odachi ask "<question>"
  odachi ask --answer "<answer>" [--for P1]
  odachi ask --craft "<decision>" --because "<reason>"
  odachi ask --change "<words>" --req R1
  odachi plan "<text>" [--req R1] [--route inline | delegated --agent <name> --skill <name>]
  odachi plan --done T2 [--commit abc1234]
  odachi delegate <taskdoc> --step T1 [--surface <path>] [--command <cmd>]
  odachi level [--to <profile>]
  odachi summary
  odachi design <list|comments|open|stop|handoff|sync>
  odachi progress
  odachi undo [--list | --collect] [--file brief] [--artifact <path>]
  odachi clean [--apply] [--legacy] [--json]
  odachi store move [--dry-run]
  odachi evidence export <tree> --out <path>
  odachi check [flags]
  odachi review [--exec "<command>" | --attach <file> | --show | --refute <verdict>]
  odachi review [--lens <lens>]
  odachi review mode [status|on|off]
  odachi review checks [status|approve]
  odachi review waive --because "<why>" [--row <fingerprint>]
  odachi review probe [--runs N] [--blind] [--exec "<command>"]
  odachi verify [--exec "<command>"] [--model <id>]
  odachi doctor [--json]
  odachi bench [--journeys <names>] [--binary <path>] [--json]
  odachi init [--skills home|project] [--agents opencode|claude|both] [--model <client>=<model>] [--json]
  odachi setup [--agents opencode|claude|both] [--model <client>=<model>] [--dry-run] [--json]
  odachi tui
  odachi skills [--check | --write | --resolve <step>]
  odachi skills eval <skill> [--exec "<command>"] [--runs N] [--timeout N] [--move-home] [--recover]
  odachi role new <slug>
  odachi role [<slug>]
  odachi skill new <slug> --scope project|home
  odachi pack [--agents claude,codex|standard] [--scope project|home] [--dest <folder>] [--dry-run]
  odachi --version
  odachi version`

// TestTheVerbCatalogRendersTheSameUsageBlock pins the bytes. Deriving the block
// from a table is a refactor only while the rendered help stays identical (0360).
func TestTheVerbCatalogRendersTheSameUsageBlock(t *testing.T) {
	var out bytes.Buffer
	usage(&out)

	if got := usageBlock(t, out.String()); got != pinnedUsageBlock {
		t.Errorf("the usage block moved:\n--- got ---\n%s\n--- want ---\n%s", got, pinnedUsageBlock)
	}
}

// TestTheUnknownCommandNamesEveryDispatchedVerb: the error a stranger gets for
// a typo is the other list of verbs. Today it omits delegate, store, setup, tui
// and verify, so a person who mistypes one of those is told only about the
// others. Both lists come from the catalog, and this guard keeps them complete
// (0360).
func TestTheUnknownCommandNamesEveryDispatchedVerb(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	verbs := dispatchVerbs(t, string(source))

	code, out, stderr := runDir(t, t.TempDir(), []string{"deploy"}, "")
	if code != start.CodeBlocked {
		t.Fatalf("code = %d, want %d", code, start.CodeBlocked)
	}
	if out != "" {
		t.Errorf("stdout = %q, want nothing", out)
	}
	for _, verb := range verbs {
		if !namesAvailableCommand(stderr, verb) {
			t.Errorf("the unknown-command message does not name %q, and the binary answers it:\n%s", verb, stderr)
		}
	}
}

// namesAvailableCommand reports whether the error's command list offers this
// verb as a named entry. The token includes the opening parenthesis so `skill`
// cannot be satisfied by `skills` (0360).
func namesAvailableCommand(message, verb string) bool {
	return strings.Contains(message, verb+" (")
}
