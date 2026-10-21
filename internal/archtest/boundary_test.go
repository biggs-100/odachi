package archtest

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// modulePath is the module declared in go.mod. The boundary guard does not
// trust import strings it cannot attribute to this module.
const modulePath = "github.com/biggz/odachi"

// domainRoot is the only area allowed to be dependency-free.
const domainRoot = modulePath + "/internal/domain"

// TestDomainCoreImportsOnlyStdlibAndSiblings is the hard boundary guard: the
// domain core may import only the standard library and sibling domain
// packages. A profile, an adapter, or a vendor SDK reaching into the domain is
// a build failure, not a review comment. See
// docs/decisions/0001-stack-and-boundaries.md.
func TestDomainCoreImportsOnlyStdlibAndSiblings(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot read working directory: %v", err)
	}
	// The test runs inside <module root>/internal/archtest.
	root := filepath.Dir(filepath.Dir(wd))

	// The guard must exist before the code it guards, so an absent domain
	// directory is not a boundary violation.
	domainDir := filepath.Join(root, "internal", "domain")
	if info, err := os.Stat(domainDir); err != nil || !info.IsDir() {
		t.Logf("no domain packages yet: %s does not exist", domainDir)
		return
	}

	cmd := exec.Command("go", "list", "-deps", "-f", "{{if not .Standard}}{{.ImportPath}}{{end}}", "./internal/domain/...")
	cmd.Dir = root

	// Output keeps stdout only. stderr carries the harmless "matched no
	// packages" warning and must never be parsed as an import path.
	out, err := cmd.Output()
	if err != nil {
		stderr := ""
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		// Before the first domain package exists, `go list` fails on the
		// pattern itself. That is not a boundary violation.
		if strings.Contains(stderr, "matched no packages") {
			t.Logf("no domain packages yet: %s", strings.TrimSpace(stderr))
			return
		}
		t.Fatalf("go list failed: %v: %s", err, strings.TrimSpace(stderr))
	}

	for _, line := range strings.Split(string(out), "\n") {
		importPath := strings.TrimSpace(line)
		if importPath == "" {
			continue
		}
		if importPath == modulePath {
			continue
		}
		if strings.HasPrefix(importPath, domainRoot) {
			continue
		}
		t.Errorf("domain core imports %q; the domain may import only the standard library and sibling %s/... packages (see docs/decisions/0001-stack-and-boundaries.md)", importPath, domainRoot)
	}
}
