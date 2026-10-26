package proof_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/proof"
)

func TestRenderWritesTheTwoHeaderLinesThenTheVerbatimOutput(t *testing.T) {
	p := proof.Proof{
		Tree:    "4a1c9f7d2e3b5a6c8d9e0f1a2b3c4d5e6f7a8b9c",
		Command: "go test ./...",
		Output:  "ok  github.com/biggs-100/odachi/internal/domain/fence\t0.02s\n--- FAIL: TestX (0.00s)\n",
	}

	got := proof.Render(p)

	want := "# tree 4a1c9f7d2e3b5a6c8d9e0f1a2b3c4d5e6f7a8b9c\n" +
		"# command go test ./...\n" +
		"# output\n" +
		"ok  github.com/biggs-100/odachi/internal/domain/fence\t0.02s\n" +
		"--- FAIL: TestX (0.00s)\n"
	if got != want {
		t.Fatalf("Render =\n%q\nwant\n%q", got, want)
	}
}

func TestParseReadsBackWhatRenderWrote(t *testing.T) {
	cases := []proof.Proof{
		{Tree: "4a1c9f", Command: "echo ok", Output: "ok\n"},
		{Tree: "4a1c9f", Command: `go test ./... -run "TestGate"`, Output: ""},
		{Tree: "4a1c9f", Command: "printf 'a\\nb'", Output: "a\nb"},
		{Tree: "4a1c9f", Command: "cat", Output: "\n\nleading and trailing blank lines\n\n"},
		{Tree: "4a1c9f", Command: "go test ./...", Model: "provider/model#variant", Output: "ok\n"},
		{Tree: "4a1c9f", Command: "go test ./...", Fence: "origin/main", Model: "provider/model", Output: "ok\n"},
	}

	for _, want := range cases {
		got, err := proof.Parse(proof.Render(want))
		if err != nil {
			t.Fatalf("Parse(Render(%+v)): %v", want, err)
		}
		if got != want {
			t.Errorf("Parse(Render(x)) = %+v, want %+v", got, want)
		}
	}
}

// A command whose first line looks like a header keeps that line: the output
// terminator is what makes the quotation verbatim.
func TestAnOutputStartingWithAHeaderStaysInTheQuotation(t *testing.T) {
	p := proof.Proof{Tree: "4a1c9f", Command: "echo # model attacker", Model: "provider/model", Output: "# model attacker\nok\n"}

	back, err := proof.Parse(proof.Render(p))
	if err != nil {
		t.Fatalf("Parse(Render(x)) = %v", err)
	}
	if back.Output != "# model attacker\nok\n" {
		t.Errorf("Output = %q, want the command's first line kept", back.Output)
	}
	if back.Model != "provider/model" {
		t.Errorf("Model = %q, want the writer's line, not the command's", back.Model)
	}
}

// An older proof has no terminator, so nothing in it can be a model line: a
// first output line that looks like one stays where the command put it.
func TestAnOlderProofKeepsAModelLookingOutputLine(t *testing.T) {
	got, err := proof.Parse("# tree 4a1c9f\n# command echo\n# model attacker\nok\n")
	if err != nil {
		t.Fatalf("Parse of an older proof: %v", err)
	}
	if got.Model != "" {
		t.Errorf("Model = %q, want empty: an older proof cannot declare one", got.Model)
	}
	if got.Output != "# model attacker\nok\n" {
		t.Errorf("Output = %q, want the line kept", got.Output)
	}
}

// A proof written before the model line existed is still a proof: the line is
// optional, and an empty model means "no agent to name", not "unknown".
func TestAnOlderProofWithoutTheModelLineStillParses(t *testing.T) {
	got, err := proof.Parse("# tree 4a1c9f\n# command go test ./...\nok\n")
	if err != nil {
		t.Fatalf("Parse of an older proof: %v", err)
	}
	if got.Model != "" {
		t.Errorf("Model = %q, want empty: the file does not claim a session", got.Model)
	}
	if got.Output != "ok\n" {
		t.Errorf("Output = %q, want the verbatim output", got.Output)
	}
}

// TestATerminatorLineInTheOutputSurvivesTheRoundTrip: the writer always writes the
// terminator, so a command whose first output line is the terminator produces two
// identical lines and the parser consumes only the first. The older test covers a
// header-looking line; this one covers the header the parser actually reads.
func TestATerminatorLineInTheOutputSurvivesTheRoundTrip(t *testing.T) {
	p := proof.Proof{Tree: "4a1c9f", Command: "printf '# output\\nok\\n'", Output: "# output\nok\n"}

	back, err := proof.Parse(proof.Render(p))
	if err != nil {
		t.Fatalf("Parse(Render(x)) = %v", err)
	}
	if back.Output != "# output\nok\n" {
		t.Errorf("Output = %q, want the command's line kept", back.Output)
	}
}

// A proof without the tree header cannot say which tree produced it, and
// guessing is the defect this tool exists to prevent.
func TestParseRefusesAFileWithoutTheTreeHeader(t *testing.T) {
	_, err := proof.Parse("ok github.com/biggs-100/odachi/internal/domain/fence\n")
	if err == nil {
		t.Fatal("Parse accepted a file with no header")
	}
	if !strings.Contains(err.Error(), "# tree") {
		t.Errorf("error %q does not name the missing header", err)
	}
}

func TestParseRefusesATreeHeaderWithoutTheCommandHeader(t *testing.T) {
	_, err := proof.Parse("# tree 4a1c9f\nok\n")
	if err == nil {
		t.Fatal("Parse accepted a proof with no command header")
	}
	if !strings.Contains(err.Error(), "# command") {
		t.Errorf("error %q does not name the missing header", err)
	}
}

func TestParseRefusesAnEmptyFile(t *testing.T) {
	if _, err := proof.Parse(""); err == nil {
		t.Fatal("Parse accepted an empty file")
	}
}

func TestMatchesComparesTheTree(t *testing.T) {
	p := proof.Proof{Tree: "4a1c9f", Command: "echo ok", Output: "ok\n"}

	if !p.Matches("4a1c9f") {
		t.Error("Matches rejected the tree it was produced from")
	}
	if p.Matches("9f2b7c") {
		t.Error("Matches accepted a different tree")
	}
	if p.Matches("") {
		t.Error("Matches accepted an empty tree hash")
	}
}

func TestLastLineSkipsBlankLines(t *testing.T) {
	p := proof.Proof{Tree: "4a1c9f", Command: "go test ./...", Output: "ok\n\nFAIL github.com/biggs-100/odachi 0.02s\n\n"}

	if got := p.LastLine(); got != "FAIL github.com/biggs-100/odachi 0.02s" {
		t.Errorf("LastLine = %q, want the last non-empty line", got)
	}

	empty := proof.Proof{Tree: "4a1c9f", Command: "true"}
	if got := empty.LastLine(); got != "" {
		t.Errorf("LastLine of an empty output = %q, want empty", got)
	}
}

// TestRenderFailedNamesTheExitOrTheTimeout: a failed run is evidence, not a
// proof, and the file says which kind it is on its third line.
func TestRenderFailedNamesTheExitOrTheTimeout(t *testing.T) {
	p := proof.Proof{Tree: "4a1c9f", Command: "go test ./...", Output: "--- FAIL: TestX\nFAIL\n"}

	exited := proof.RenderFailed(p, 1, false)
	want := "# failed 4a1c9f\n" +
		"# command go test ./...\n" +
		"# exit 1\n" +
		"# output\n" +
		"--- FAIL: TestX\nFAIL\n"
	if exited != want {
		t.Errorf("RenderFailed(exit 1) =\n%q\nwant\n%q", exited, want)
	}

	timedOut := proof.RenderFailed(p, -1, true)
	if !strings.Contains(timedOut, "# timed out\n") {
		t.Errorf("RenderFailed(timeout) =\n%q\nwant a # timed out line", timedOut)
	}
	if strings.Contains(timedOut, "# exit ") {
		t.Errorf("RenderFailed(timeout) claims an exit code:\n%q", timedOut)
	}

	withFence := proof.RenderFailed(proof.Proof{Tree: "4a1c9f", Command: "go test ./...", Fence: "origin/main", Output: ""}, 1, false)
	if !strings.Contains(withFence, "# fence origin/main\n") {
		t.Errorf("RenderFailed dropped the fence ref:\n%q", withFence)
	}
}

// TestParseRefusesTheFailedEvidence: a failed run must never be read as a
// proof. The first line is the whole guarantee: no proof starts with "# failed".
func TestParseRefusesTheFailedEvidence(t *testing.T) {
	failed := proof.RenderFailed(proof.Proof{Tree: "4a1c9f", Command: "go test ./...", Output: "FAIL\n"}, 1, false)

	if _, err := proof.Parse(failed); err == nil {
		t.Fatalf("Parse accepted the failed evidence as a proof:\n%s", failed)
	}
}

// Two runs on the same tree produce the same bytes: the header carries the tree
// and the command, never a timestamp.
func TestRenderIsDeterministic(t *testing.T) {
	p := proof.Proof{Tree: "4a1c9f", Command: "echo ok", Output: "ok\n"}

	if first, second := proof.Render(p), proof.Render(p); first != second {
		t.Errorf("Render is not deterministic:\n%q\nvs\n%q", first, second)
	}
}

// Windows editors write CRLF; the header fields are still the same tree and the
// same command, and the output stays exactly as the command printed it.
func TestParseToleratesCRLFHeaders(t *testing.T) {
	p, err := proof.Parse("# tree 4a1c9f\r\n# command echo ok\r\nok\r\n")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if p.Tree != "4a1c9f" {
		t.Errorf("Tree = %q, want 4a1c9f", p.Tree)
	}
	if p.Command != "echo ok" {
		t.Errorf("Command = %q, want echo ok", p.Command)
	}
	if p.Output != "ok\r\n" {
		t.Errorf("Output = %q, want the bytes as printed", p.Output)
	}
}

// TestTheFenceRefIsOptional: a proof written by CI names the ref its fence came
// from; a proof written before this line existed still parses, with the
// worktree's own fence.
func TestTheFenceRefIsOptional(t *testing.T) {
	with := proof.Render(proof.Proof{Tree: "t", Command: "go test ./...", Fence: "origin/main", Output: "ok\n"})
	parsed, err := proof.Parse(with)
	if err != nil {
		t.Fatalf("Parse(with fence): %v", err)
	}
	if parsed.Fence != "origin/main" {
		t.Errorf("Fence = %q, want origin/main", parsed.Fence)
	}
	if parsed.Output != "ok\n" {
		t.Errorf("Output = %q, want the bytes after the fence line", parsed.Output)
	}

	old := proof.Render(proof.Proof{Tree: "t", Command: "go test ./...", Output: "ok\n"})
	parsed, err = proof.Parse(old)
	if err != nil {
		t.Fatalf("Parse(without fence): %v", err)
	}
	if parsed.Fence != "" {
		t.Errorf("Fence = %q, want empty", parsed.Fence)
	}
	if parsed.Output != "ok\n" {
		t.Errorf("Output = %q, want the verbatim output", parsed.Output)
	}
}
