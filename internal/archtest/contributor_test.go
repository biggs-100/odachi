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
	for _, want := range []string{"odachi check --fence-from", "--base origin/main", "--run", "--timeout 900"} {
		if !strings.Contains(contributing, want) {
			t.Errorf("CONTRIBUTING does not say how to reproduce the gate locally (%q)", want)
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
