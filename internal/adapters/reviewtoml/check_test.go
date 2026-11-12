package reviewtoml_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/reviewtoml"
	"github.com/biggs-100/odachi/internal/domain/review"
)

// TestDecodeReadsTheChecksTheProjectDeclared: the tools run in the order the
// file wrote them, each with its command and the severity it is paid at.
func TestDecodeReadsTheChecksTheProjectDeclared(t *testing.T) {
	data := `mode = "on"

[[check]]
name = "lint"
command = "golangci-lint run ./..."
severity = "WARNING"

[[check]]
name = "vuln"
command = "govulncheck ./..."
severity = "CRITICAL"
`

	got, err := reviewtoml.Decode([]byte(data))
	if err != nil {
		t.Fatalf("Decode = %v, want no error", err)
	}
	if len(got.Checks) != 2 {
		t.Fatalf("Decode kept %d checks, want 2: %+v", len(got.Checks), got.Checks)
	}
	if got.Checks[0].Name != "lint" || got.Checks[0].Severity != review.Warning {
		t.Errorf("first check = %+v, want lint at WARNING", got.Checks[0])
	}
	if got.Checks[1].Name != "vuln" || got.Checks[1].Command != "govulncheck ./..." {
		t.Errorf("second check = %+v, want vuln running govulncheck", got.Checks[1])
	}
	if got.CheckTimeout != review.DefaultCheckTimeout {
		t.Errorf("CheckTimeout = %d, want the default %d", got.CheckTimeout, review.DefaultCheckTimeout)
	}
}

// TestDecodeReadsTheCeilingTheProjectChose: a tool that hangs must not hang a
// review, and the number is the project's.
func TestDecodeReadsTheCeilingTheProjectChose(t *testing.T) {
	data := `mode = "on"
check_timeout_s = 45

[[check]]
name = "lint"
command = "golangci-lint run ./..."
severity = "WARNING"
`

	got, err := reviewtoml.Decode([]byte(data))
	if err != nil {
		t.Fatalf("Decode = %v, want no error", err)
	}
	if got.CheckTimeout != 45 {
		t.Errorf("CheckTimeout = %d, want 45", got.CheckTimeout)
	}
}

// TestDecodeRefusesACheckItCannotRun: the refusal names which check is wrong and
// what it is missing, because a check nobody runs is a review that lies about
// what it measured.
func TestDecodeRefusesACheckItCannotRun(t *testing.T) {
	cases := map[string]string{
		"no severity": `mode = "on"

[[check]]
name = "lint"
command = "golangci-lint run ./..."
`,
		"no command": `mode = "on"

[[check]]
name = "lint"
severity = "WARNING"
`,
		"unknown key": `mode = "on"

[[check]]
name = "lint"
command = "golangci-lint run ./..."
severity = "WARNING"
timeout = 3
`,
	}

	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if got, err := reviewtoml.Decode([]byte(data)); err == nil {
				t.Fatalf("Decode = %+v, want a refusal", got)
			}
		})
	}
}

// TestAFileWithoutChecksReadsAsNone: the file every project has today decodes as
// it always did, and a project that declares nothing runs nothing.
func TestAFileWithoutChecksReadsAsNone(t *testing.T) {
	got, err := reviewtoml.Decode([]byte("mode = \"on\"\nkeep_verdicts = 3\n"))
	if err != nil {
		t.Fatalf("Decode = %v, want no error", err)
	}
	if len(got.Checks) != 0 {
		t.Errorf("Decode kept %+v, want no checks", got.Checks)
	}
}

// TestRenderRoundTripsTheChecks: what the writer wrote is what the reader reads,
// so a project that edits the file through odachi keeps its tools.
func TestRenderRoundTripsTheChecks(t *testing.T) {
	config := review.ProjectConfig{
		Mode:         review.ModeOn,
		KeepVerdicts: review.DefaultKeepVerdicts,
		Checks: []review.Check{
			{Name: "lint", Command: "golangci-lint run ./...", Severity: review.Warning},
			{Name: "vuln", Command: "govulncheck ./...", Severity: review.Critical},
		},
	}

	data, err := reviewtoml.Render(config)
	if err != nil {
		t.Fatalf("Render = %v, want no error", err)
	}
	if !strings.Contains(string(data), "[[check]]") {
		t.Fatalf("Render did not write a check table:\n%s", data)
	}

	back, err := reviewtoml.Decode(data)
	if err != nil {
		t.Fatalf("Decode(Render(config)) = %v, want no error:\n%s", err, data)
	}
	if len(back.Checks) != len(config.Checks) {
		t.Fatalf("round trip kept %d checks, want %d:\n%s", len(back.Checks), len(config.Checks), data)
	}
	for i := range config.Checks {
		if back.Checks[i] != config.Checks[i] {
			t.Errorf("check %d came back as %+v, want %+v", i+1, back.Checks[i], config.Checks[i])
		}
	}
}

// TestRenderSaysHowLongAToolMayRun: a project that declared tools reads the
// ceiling next to them instead of guessing what the run used.
func TestRenderSaysHowLongAToolMayRun(t *testing.T) {
	config := review.ProjectConfig{
		Mode:         review.ModeOn,
		KeepVerdicts: review.DefaultKeepVerdicts,
		Checks: []review.Check{
			{Name: "lint", Command: "golangci-lint run ./...", Severity: review.Warning},
		},
	}

	data, err := reviewtoml.Render(config)
	if err != nil {
		t.Fatalf("Render = %v, want no error", err)
	}
	if !strings.Contains(string(data), "check_timeout_s = ") {
		t.Errorf("Render does not write the ceiling:\n%s", data)
	}
}
