package undone_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/undone"
)

// TestClassifyReadsTheThreeOutcomes is the row's whole judgment: a green suite is
// the only block, a build failure is unknown, and any other red is the suite
// noticing.
func TestClassifyReadsTheThreeOutcomes(t *testing.T) {
	cases := []struct {
		name     string
		exitCode int
		output   string
		want     undone.Verdict
	}{
		{
			name:     "a suite that stayed green",
			exitCode: 0,
			output:   "ok  github.com/x/y\t0.02s\n",
			want:     undone.Missed,
		},
		{
			name:     "a failing test",
			exitCode: 1,
			output:   "--- FAIL: TestPay (0.00s)\n    pay_test.go:12: got 1, want 2\nFAIL\n",
			want:     undone.Noticed,
		},
		{
			name:     "a package that refused to build",
			exitCode: 1,
			output:   "# github.com/x/y [build failed]\n./pay.go:9:2: undefined: Charge\nFAIL\tgithub.com/x/y [build failed]\n",
			want:     undone.Unknown,
		},
		{
			name:     "a command that never ran",
			exitCode: -1,
			output:   "",
			want:     undone.Unknown,
		},
		{
			name:     "a red without a quoted test is still a notice",
			exitCode: 1,
			output:   "FAIL\tgithub.com/x/y\t0.02s\n",
			want:     undone.Noticed,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := undone.Classify(tc.exitCode, tc.output); got != tc.want {
				t.Fatalf("Classify(%d, %q) = %q, want %q", tc.exitCode, tc.output, got, tc.want)
			}
		})
	}
}

// TestRenderWritesTheThreeHeadersThenTheOutputVerbatim: the output is a
// quotation, so a first line that looks like a header stays in it.
func TestRenderWritesTheThreeHeadersThenTheOutputVerbatim(t *testing.T) {
	e := undone.Evidence{
		Tree:    "4a1c9f7d2e3b5a6c8d9e0f1a2b3c4d5e6f7a8b9c",
		Command: "go test ./...",
		Verdict: undone.Missed,
		Output:  "# command a line the suite printed\nok  github.com/x/y\t0.02s\n",
	}

	got := undone.Render(e)

	want := "# undone 4a1c9f7d2e3b5a6c8d9e0f1a2b3c4d5e6f7a8b9c\n" +
		"# command go test ./...\n" +
		"# verdict missed\n" +
		"# output\n" +
		"# command a line the suite printed\nok  github.com/x/y\t0.02s\n"
	if got != want {
		t.Fatalf("Render =\n%q\nwant\n%q", got, want)
	}
}

// TestParseReadsBackWhatRenderWrote: a round trip keeps every field, including a
// verdict nobody has to guess.
func TestParseReadsBackWhatRenderWrote(t *testing.T) {
	cases := []undone.Evidence{
		{Tree: "4a1c9f", Command: "go test ./...", Verdict: undone.Noticed, Output: "--- FAIL: TestX\n"},
		{Tree: "4a1c9f", Command: "echo ok", Verdict: undone.Missed, Output: ""},
		{Tree: "4a1c9f", Command: `go test ./... -run "TestGate"`, Verdict: undone.Unknown, Output: "# output\n"},
		{Tree: "4a1c9f", Command: "go test ./...", Verdict: undone.NotOwed, Output: ""},
	}

	for _, want := range cases {
		got, err := undone.Parse(undone.Render(want))
		if err != nil {
			t.Fatalf("Parse(Render(%+v)): %v", want, err)
		}
		if got != want {
			t.Errorf("Parse(Render(x)) = %+v, want %+v", got, want)
		}
	}
}

// TestParseRefusesAProofByName: a stored proof is not an undone run, and reading
// one as the other would put a green in the row's file.
func TestParseRefusesAProofByName(t *testing.T) {
	_, err := undone.Parse("# tree 4a1c9f\n# command go test ./...\n# output\nok\n")
	if err == nil {
		t.Fatal("a proof parsed as an undone run")
	}
	if !strings.Contains(err.Error(), "proof") {
		t.Fatalf("the refusal does not name a proof: %v", err)
	}
}

// TestParseRefusesAFileWithNoVerdict: the verdict is what makes the file a
// record instead of an output. An old file does not exist, so nothing has to be
// guessed.
func TestParseRefusesAFileWithNoVerdict(t *testing.T) {
	_, err := undone.Parse("# undone 4a1c9f\n# command go test ./...\n# output\nok\n")
	if err == nil {
		t.Fatal("an undone run with no verdict parsed")
	}
	if !strings.Contains(err.Error(), "verdict") {
		t.Fatalf("the refusal does not name the verdict: %v", err)
	}
}
