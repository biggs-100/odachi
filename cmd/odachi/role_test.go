package main

import (
	"strings"
	"testing"
)

// TestRoleNewWritesTheSourceEndToEnd is the CLI contract for the first half of
// the loop: the skeleton lands under .odachi/roles, and the message names the
// command that continues.
func TestRoleNewWritesTheSourceEndToEnd(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	code, stdout, stderr := runDir(t, dir, []string{"role", "new", "marketing"}, "")
	if code != 0 {
		t.Fatalf("role new: exit %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, ".odachi/roles/marketing.md") {
		t.Errorf("stdout %q does not name the source", stdout)
	}
	if !strings.Contains(stdout, "odachi role marketing") {
		t.Errorf("stdout %q does not name the next command", stdout)
	}

	if _, ok := tree(t, dir)[".odachi/roles/marketing.md"]; !ok {
		t.Fatalf("the source was not written under .odachi/roles")
	}
}
