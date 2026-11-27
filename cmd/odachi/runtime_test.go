package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestOpenCodeRuntimeControls is the opt-in, deterministic proof of the runtime
// controls R1 promises. It builds a control executable in a temp directory, never
// touches the installed OpenCode binary, and never calls a live model.
func TestOpenCodeRuntimeControls(t *testing.T) {
	if os.Getenv("ODACHI_OPENCODE_RUNTIME_TEST") != "1" {
		t.Skip("set ODACHI_OPENCODE_RUNTIME_TEST=1 to run the deterministic OpenCode runtime-control harness")
	}

	root := t.TempDir()
	controlDir := filepath.Join(root, "control")
	executable, err := BuildRuntimeControl(t.Context(), controlDir)
	if err != nil {
		t.Fatalf("building the deterministic control executable: %v", err)
	}

	proof, err := RunOpenCodeRuntimeProofs(t.Context(), RuntimeControl{Executable: executable}, filepath.Join(root, "work"))
	if err != nil {
		t.Fatalf("running the runtime-control proofs: %v", err)
	}

	if len(proof.SpawnsPerGrant) != 3 {
		t.Fatalf("observed %d grants, want 3 (two actor runs and one mutation run)", len(proof.SpawnsPerGrant))
	}
	for index, spawns := range proof.SpawnsPerGrant {
		if spawns != 1 {
			t.Errorf("grant %d spawned the control executable %d times, want exactly one", index+1, spawns)
		}
	}

	if len(proof.Sessions) != 2 {
		t.Fatalf("observed %d actor sessions, want 2", len(proof.Sessions))
	}
	if proof.Sessions[0] == proof.Sessions[1] {
		t.Errorf("two distinct grants received the same session %q", proof.Sessions[0])
	}
	for _, session := range proof.Sessions {
		if !strings.HasPrefix(session, "ses_") {
			t.Errorf("the runtime reported %q, which is not a fresh OpenCode session identity", session)
		}
	}

	if !proof.ReplayRefused {
		t.Error("replaying a completed grant was not refused")
	}
	if proof.ReplaySpawns != 0 {
		t.Errorf("a replay spawned the control executable %d times, want 0", proof.ReplaySpawns)
	}

	if proof.MutationBefore == "" || proof.MutationAfter == "" {
		t.Fatalf("the mutation proof did not read both snapshots: before %q after %q", proof.MutationBefore, proof.MutationAfter)
	}
	if proof.MutationBefore == proof.MutationAfter {
		t.Errorf("mutating the frozen candidate kept the snapshot %q", proof.MutationBefore)
	}
	if !proof.MutationRefused {
		t.Error("a moved frozen candidate was not refused")
	}
	if proof.MutationSealed {
		t.Error("a moved frozen candidate still sealed evidence")
	}

	if len(proof.Argv) != 3 {
		t.Fatalf("the control executable recorded %d argv entries, want 3: %v", len(proof.Argv), proof.Argv)
	}
	want := []string{"api", "post", "/api/session", "--data", "{}"}
	for index, argv := range proof.Argv {
		if !slices.Equal(argv, want) {
			t.Errorf("spawn %d carried %q, want the fresh-session request only %q", index+1, argv, want)
		}
		for _, forbidden := range []string{"--session", "--continue", "--fork", "--resume", "sessionID", "parentID"} {
			if strings.Contains(strings.Join(argv, " "), forbidden) {
				t.Errorf("spawn %d carried the forbidden input %q: %q", index+1, forbidden, argv)
			}
		}
	}

	t.Logf("spawns per grant: %v", proof.SpawnsPerGrant)
	t.Logf("actor sessions: %v", proof.Sessions)
	t.Logf("replay: refused=%v spawns=%d", proof.ReplayRefused, proof.ReplaySpawns)
	t.Logf("mutation: before=%s after=%s refused=%v sealed=%v", proof.MutationBefore, proof.MutationAfter, proof.MutationRefused, proof.MutationSealed)
	t.Logf("argv: %v", proof.Argv)
}
