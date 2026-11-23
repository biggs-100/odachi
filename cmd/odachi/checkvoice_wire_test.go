package main

import (
	"bytes"
	"encoding/json"
	"reflect"
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
