package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPowerShellInstallerIsServedWithoutABOM pins the command the README
// documents:
//
//	irm https://raw.githubusercontent.com/biggs-100/odachi/main/scripts/install.ps1 | iex
//
// Invoke-RestMethod decodes the served bytes and Invoke-Expression parses the
// string, leading BOM included. PowerShell 5.1 reads that U+FEFF as part of the
// first token, so `#Requires -Version 5.1` becomes `?#Requires`: the documented
// install prints an error and then runs anyway. A command that teaches the
// reader to ignore an error is a defect of teaching, so the file opens with the
// directive and nothing before it.
func TestPowerShellInstallerIsServedWithoutABOM(t *testing.T) {
	path := filepath.Join(repoRoot(t), "scripts", "install.ps1")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		t.Fatalf("scripts/install.ps1 starts with a UTF-8 BOM: `irm | iex` reads it as `?#Requires`, and the documented install prints an error")
	}

	first := string(data)
	if end := strings.IndexAny(first, "\r\n"); end >= 0 {
		first = first[:end]
	}
	const want = "#Requires -Version 5.1"
	if first != want {
		t.Fatalf("scripts/install.ps1 opens with %q, want %q", first, want)
	}
}
