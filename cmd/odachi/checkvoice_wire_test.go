package main

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/app/start"
)

// TestReportCheckUsesTheVoiceOnlyForOfficePeru: the default profiles keep every
// byte of their human output; the office profile gets the Spanish line. The JSON
// payload is a contract and never changes with the profile (0355).
func TestReportCheckUsesTheVoiceOnlyForOfficePeru(t *testing.T) {
	result := check.Result{Code: start.CodeOK, Lines: []string{"ready: there is proof for this tree"}}

	var out, errOut bytes.Buffer
	reportCheck(result, "office-peru", false, false, &out, &errOut)
	if got := out.String(); !strings.Contains(got, "listo: hay prueba para este árbol") {
		t.Errorf("office-peru output = %q, want the Spanish line", got)
	}

	out.Reset()
	reportCheck(result, "novato", false, false, &out, &errOut)
	if got, want := out.String(), "ready: there is proof for this tree\n"; got != want {
		t.Errorf("novato output = %q, want %q", got, want)
	}
}

// TestReportCheckUsesTheVoiceBoundaryForQuietNotes pins the one line a quiet
// ready run writes to stderr for both translated and unchanged profiles.
func TestReportCheckUsesTheVoiceBoundaryForQuietNotes(t *testing.T) {
	result := check.Result{Code: start.CodeOK, Note: "ready: there is proof for this tree"}

	for _, tc := range []struct {
		profile string
		want    string
	}{
		{profile: "office-peru", want: "listo: hay prueba para este árbol"},
		{profile: "novato", want: "ready: there is proof for this tree"},
	} {
		t.Run(tc.profile, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if got := reportCheck(result, tc.profile, true, false, &out, &errOut); got != start.CodeOK {
				t.Errorf("exit code = %d, want %d", got, start.CodeOK)
			}
			if got := out.String(); got != "" {
				t.Errorf("stdout = %q, want no output", got)
			}
			if got, want := errOut.String(), "odachi: "+tc.want+"\n"; got != want {
				t.Errorf("stderr = %q, want %q", got, want)
			}
		})
	}
}

// TestReportCheckDoesNotSpecialCaseAProfile keeps profile selection outside the
// reporter so future voices do not require another reportCheck branch.
func TestReportCheckDoesNotSpecialCaseAProfile(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "main.go", source, 0)
	if err != nil {
		t.Fatalf("parse main.go: %v", err)
	}

	var report *ast.FuncDecl
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "reportCheck" {
			report = fn
			break
		}
	}
	if report == nil || report.Body == nil {
		t.Fatal("main.go does not define reportCheck with a body")
	}

	isProfileName := func(expr ast.Expr) bool {
		ident, ok := expr.(*ast.Ident)
		return ok && ident.Name == "profileName"
	}
	isStringLiteral := func(expr ast.Expr) bool {
		literal, ok := expr.(*ast.BasicLit)
		return ok && literal.Kind == token.STRING
	}

	ast.Inspect(report.Body, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.BinaryExpr:
			if node.Op != token.EQL && node.Op != token.NEQ {
				break
			}
			if (isProfileName(node.X) && isStringLiteral(node.Y)) ||
				(isStringLiteral(node.X) && isProfileName(node.Y)) {
				t.Errorf("reportCheck compares profileName with a string literal at %s", fileSet.Position(node.Pos()))
			}
		case *ast.BasicLit:
			if node.Kind != token.STRING {
				break
			}
			value, err := strconv.Unquote(node.Value)
			if err == nil && value == "office-peru" {
				t.Errorf("reportCheck mentions the office-peru literal at %s", fileSet.Position(node.Pos()))
			}
		}
		return true
	})
}

func TestReportCheckJSONKeepsItsKeysForEveryProfile(t *testing.T) {
	result := check.Result{
		Code:    start.CodeOK,
		Reason:  "ready",
		Message: "ready: there is proof for this tree",
		Lines:   []string{"ready: there is proof for this tree"},
	}
	for _, profile := range []string{"novato", "office-peru"} {
		var out, errOut bytes.Buffer
		reportCheck(result, profile, false, true, &out, &errOut)
		var payload struct {
			Message string   `json:"message"`
			Lines   []string `json:"lines"`
		}
		if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
			t.Fatalf("profile %q: JSON = %q: %v", profile, out.String(), err)
		}
		if payload.Message != result.Message || !reflect.DeepEqual(payload.Lines, result.Lines) {
			t.Errorf("profile %q: JSON was translated: %+v", profile, payload)
		}
	}
}
