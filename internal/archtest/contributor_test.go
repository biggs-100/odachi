package archtest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// moduleRoot is the directory that holds go.mod: the test runs inside
// <module root>/internal/archtest.
func moduleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot read working directory: %v", err)
	}
	return filepath.Dir(filepath.Dir(wd))
}

// TestTheContributorSurfaceExists: a stranger's path is three files and two
// templates. A missing file, a template that stops asking for the proof, or a
// README that stops linking them is a contributor walking into a closed door.
// The assertions are deliberately few: one per promise the project makes.
func TestTheContributorSurfaceExists(t *testing.T) {
	root := moduleRoot(t)

	files := []string{
		"CONTRIBUTING.md",
		"AI_POLICY.md",
		"SECURITY.md",
		".github/pull_request_template.md",
		".github/ISSUE_TEMPLATE/bug_report.yml",
		".github/ISSUE_TEMPLATE/feature_request.yml",
	}
	for _, name := range files {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(name))); err != nil {
			t.Errorf("%s is missing: %v", name, err)
		}
	}

	reading := func(name string) string {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		return string(data)
	}

	contributing := reading("CONTRIBUTING.md")
	for _, want := range []string{"odachi check --fence-from", "--base origin/main", "--no-review", "--run", "--timeout 900"} {
		if !strings.Contains(contributing, want) {
			t.Errorf("CONTRIBUTING does not say how to reproduce the gate locally (%q)", want)
		}
	}
	contributingProse := strings.Join(strings.Fields(contributing), " ")
	for _, want := range []string{
		"A real local close with review mode on must run review and any required verification",
		"`Risk: none` still needs a written reason on non-passive work",
	} {
		if !strings.Contains(contributingProse, want) {
			t.Errorf("CONTRIBUTING does not preserve the local close contract (%q)", want)
		}
	}

	if security := reading("SECURITY.md"); !strings.Contains(security, "security/advisories") {
		t.Error("SECURITY does not give the private advisories channel")
	}

	if policy := reading("AI_POLICY.md"); !strings.Contains(policy, "Disclose it") {
		t.Error("AI_POLICY does not ask for disclosure")
	}

	template := reading(".github/pull_request_template.md")
	for _, want := range []string{"## The proof", "Risk:", "## Rollback"} {
		if !strings.Contains(template, want) {
			t.Errorf("the pull request template does not ask for %q", want)
		}
	}

	readme := reading("README.md")
	for _, name := range []string{"CONTRIBUTING.md", "AI_POLICY.md", "SECURITY.md"} {
		if !strings.Contains(readme, "]("+name+")") {
			t.Errorf("README does not link %s", name)
		}
	}
}

// TestCILocalOnlyReviewContract keeps CI deterministic: it may bypass the
// clone-local review evidence gate, but it must never execute a reviewer or
// verifier. Local closes still owe that evidence when review mode is on.
func TestCILocalOnlyReviewContract(t *testing.T) {
	root := moduleRoot(t)
	data, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "gate.yml"))
	if err != nil {
		t.Fatalf("read CI gate workflow: %v", err)
	}
	gate := string(data)
	var gateCommand, executableLines string
	for _, line := range strings.Split(gate, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.Contains(trimmed, `"$ODACHI_GATE" check`) {
			gateCommand = trimmed
		}
		executableLines += trimmed + "\n"
	}

	if !strings.Contains(gateCommand, "--no-review") {
		t.Error("CI gate command does not explicitly bypass the clone-local review evidence gate")
	}
	for _, forbidden := range []string{
		"odachi review",
		"odachi verify",
		`"$ODACHI_GATE" review`,
		`"$ODACHI_GATE" verify`,
		"opencode run",
		"odachi-verify",
	} {
		if strings.Contains(executableLines, forbidden) {
			t.Errorf("CI gate executes a reviewer or verifier (%q)", forbidden)
		}
	}
}
