package oneline

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestLineCollapsesWhitespace pins the contract the local oneLine helpers used:
// every run of whitespace becomes one space and the ends are trimmed.
func TestLineCollapsesWhitespace(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"only whitespace", " \t\n ", ""},
		{"single word", "hola", "hola"},
		{"repeated spaces", "  a   b  ", "a b"},
		{"tabs and newlines", "a\t\tb\nc", "a b c"},
		{"unicode", "café\tña\n日本語", "café ña 日本語"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Line(tc.in); got != tc.want {
				t.Fatalf("Line(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// TestOneLineHasOneHome walks the repository's Go files and fails on any
// function named oneLine whose body is not a delegation to oneline.Line.
// A new copy therefore fails the same test that guards the originals.
func TestOneLineHasOneHome(t *testing.T) {
	root := repoRoot(t)
	fset := token.NewFileSet()
	checked := 0
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != root {
				name := entry.Name()
				if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "oneLine" {
				continue
			}
			checked++
			if err := isDelegation(fn); err != nil {
				t.Errorf("%s: %v", fset.Position(fn.Pos()), err)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	if checked == 0 {
		t.Log("no oneLine wrappers found")
	}
}

// isDelegation accepts exactly one shape: return oneline.Line(param), where
// param is the function's single named parameter.
func isDelegation(fn *ast.FuncDecl) error {
	if fn.Body == nil || len(fn.Body.List) != 1 {
		return fmt.Errorf("body must be a single return statement")
	}
	ret, ok := fn.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(ret.Results) != 1 {
		return fmt.Errorf("body must return exactly one value")
	}
	call, ok := ret.Results[0].(*ast.CallExpr)
	if !ok {
		return fmt.Errorf("body must return a call")
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Line" {
		return fmt.Errorf("body must return oneline.Line(...)")
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok || pkg.Name != "oneline" {
		return fmt.Errorf("body must return oneline.Line(...)")
	}
	if len(call.Args) != 1 {
		return fmt.Errorf("oneline.Line must receive one argument")
	}
	arg, ok := call.Args[0].(*ast.Ident)
	if !ok {
		return fmt.Errorf("oneline.Line argument must be the parameter")
	}
	params := fn.Type.Params
	if params == nil || len(params.List) != 1 || len(params.List[0].Names) != 1 {
		return fmt.Errorf("oneLine must declare a single named parameter")
	}
	if want := params.List[0].Names[0].Name; arg.Name != want {
		return fmt.Errorf("oneline.Line argument %q must be the parameter %q", arg.Name, want)
	}
	return nil
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller could not locate this test file")
	}
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(file))))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("go.mod not found at %s: %v", root, err)
	}
	return root
}
