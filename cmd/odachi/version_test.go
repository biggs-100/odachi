package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/start"
)

// TestVersionSurfacesPrintTheBuildVersion pins the two shapes a human types and
// the default of a plain `go build`: without ldflags the binary says "dev", so
// whoever builds from source never gets a version that is not theirs.
func TestVersionSurfacesPrintTheBuildVersion(t *testing.T) {
	t.Parallel()
	for _, arg := range []string{"--version", "version"} {
		t.Run(arg, func(t *testing.T) {
			dir := t.TempDir()

			code, out, errOut := runDir(t, dir, []string{arg}, "")

			if code != start.CodeOK {
				t.Fatalf("code = %d, want %d (stderr: %s)", code, start.CodeOK, errOut)
			}
			if want := versionLine(version); out != want {
				t.Errorf("stdout = %q, want %q: a build without ldflags says dev", out, want)
			}
			if errOut != "" {
				t.Errorf("stderr = %q, want nothing", errOut)
			}
			if files := tree(t, dir); len(files) != 0 {
				t.Errorf("version wrote files: %v", files)
			}
		})
	}
}

// TestVersionLineRendersTheLDflagsValue guards the seam goreleaser uses:
// `-X main.version=<tag>` must land on the variable the two surfaces render.
// The value is passed in instead of written into the global on purpose: a
// parallel test that mutates package state makes every test that reads it
// order-dependent, and the suite must not depend on who wins the race.
func TestVersionLineRendersTheLDflagsValue(t *testing.T) {
	t.Parallel()
	if version != "dev" {
		t.Fatalf("version = %q, want %q: the test binary is built without ldflags", version, "dev")
	}
	if got, want := versionLine("1.2.3"), "odachi 1.2.3\n"; got != want {
		t.Errorf("versionLine(%q) = %q, want %q", "1.2.3", got, want)
	}
}

// TestVersionSurfacesRenderThePackageVariable closes the hole the pure test
// leaves: `-X main.version=<tag>` is only real if both surfaces hand THAT
// variable to versionLine. A build inside the suite is out (it is in the
// fence), so this reads the source: a call site that hardcodes a string passes
// the CLI test and fails here.
func TestVersionSurfacesRenderThePackageVariable(t *testing.T) {
	t.Parallel()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "main.go", nil, 0)
	if err != nil {
		t.Fatalf("parse main.go: %v", err)
	}

	sites := 0
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		fn, ok := call.Fun.(*ast.Ident)
		if !ok || fn.Name != "versionLine" {
			return true
		}
		sites++
		arg, ok := call.Args[0].(*ast.Ident)
		if !ok || arg.Name != "version" {
			t.Errorf("main.go:%d: versionLine is called with %s, want the package variable",
				fset.Position(call.Pos()).Line, types.ExprString(call.Args[0]))
		}
		return true
	})
	if sites != 2 {
		t.Errorf("versionLine call sites = %d, want 2: the ask before the cwd and the dispatch", sites)
	}
}

// TestUnknownCommandStillBlocks keeps the new verb from swallowing the error
// path: a word that is not a command is still blocked, not printed as version.
func TestUnknownCommandStillBlocks(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	code, out, errOut := runDir(t, dir, []string{"versions"}, "")

	if code != start.CodeBlocked {
		t.Fatalf("code = %d, want %d (stdout: %s)", code, start.CodeBlocked, out)
	}
	if !strings.Contains(errOut, "unknown command") {
		t.Errorf("stderr = %q, want the unknown-command message", errOut)
	}
	if strings.Contains(out, "odachi dev") {
		t.Errorf("stdout = %q: an unknown command must not print the version", out)
	}
}
