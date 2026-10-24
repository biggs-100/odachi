package opencode_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/opencode"
)

// parse reads the document the way OpenCode would: the file may carry comments,
// so it strips them through the same helper the editor validates with.
func parse(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(opencode.StripComments(data), &doc); err != nil {
		t.Fatalf("the result is not valid JSON: %v\n%s", err, data)
	}
	return doc
}

func TestAnEmptyConfigGetsTheSchemaAndTheAgent(t *testing.T) {
	updated, out := opencode.EnsureDefaultAgent(nil, "odachi")
	if out.Status != opencode.Added {
		t.Fatalf("status = %v, want Added (%s)", out.Status, out.Reason)
	}
	doc := parse(t, updated)
	if doc["default_agent"] != "odachi" {
		t.Errorf("default_agent = %v, want odachi", doc["default_agent"])
	}
	if doc["$schema"] != opencode.SchemaURL {
		t.Errorf("$schema = %v, want %q", doc["$schema"], opencode.SchemaURL)
	}
}

func TestAnEmptyObjectGetsTheKeyWithoutATrailingComma(t *testing.T) {
	updated, out := opencode.EnsureDefaultAgent([]byte("{}"), "odachi")
	if out.Status != opencode.Added {
		t.Fatalf("status = %v, want Added", out.Status)
	}
	if doc := parse(t, updated); doc["default_agent"] != "odachi" {
		t.Errorf("default_agent = %v, want odachi", doc["default_agent"])
	}
}

func TestTheKeyGoesFirstAndTheRestIsUntouched(t *testing.T) {
	original := "{\n  \"$schema\": \"https://opencode.ai/config.json\",\n  \"mcp\": {\"servers\": {}}\n}\n"
	updated, out := opencode.EnsureDefaultAgent([]byte(original), "odachi")
	if out.Status != opencode.Added {
		t.Fatalf("status = %v, want Added", out.Status)
	}
	for _, want := range []string{`"$schema": "https://opencode.ai/config.json"`, `"mcp": {"servers": {}}`} {
		if !strings.Contains(string(updated), want) {
			t.Errorf("the result lost %q:\n%s", want, updated)
		}
	}
	if strings.Index(string(updated), "default_agent") > strings.Index(string(updated), "$schema") {
		t.Errorf("default_agent did not land first:\n%s", updated)
	}
	parse(t, updated)
}

func TestCRLFStaysCRLF(t *testing.T) {
	original := "{\r\n  \"$schema\": \"https://opencode.ai/config.json\"\r\n}\r\n"
	updated, out := opencode.EnsureDefaultAgent([]byte(original), "odachi")
	if out.Status != opencode.Added {
		t.Fatalf("status = %v, want Added", out.Status)
	}
	if bytes.Contains(bytes.ReplaceAll(updated, []byte("\r\n"), nil), []byte("\n")) {
		t.Errorf("the edit introduced a bare LF into a CRLF file: %q", updated)
	}
	if !bytes.Contains(updated, []byte("\"default_agent\": \"odachi\"")) {
		t.Errorf("the key is missing:\n%s", updated)
	}
	parse(t, updated)
}

func TestCommentsSurviveAndDoNotFoolTheKey(t *testing.T) {
	original := `{
  // "default_agent": "commented-out"
  "$schema": "https://opencode.ai/config.json", /* the schema */
  "mcp": {}
}`
	updated, out := opencode.EnsureDefaultAgent([]byte(original), "odachi")
	if out.Status != opencode.Added {
		t.Fatalf("status = %v, want Added (%s)", out.Status, out.Reason)
	}
	for _, want := range []string{`// "default_agent": "commented-out"`, "/* the schema */", `"https://opencode.ai/config.json"`} {
		if !strings.Contains(string(updated), want) {
			t.Errorf("the result lost %q:\n%s", want, updated)
		}
	}
	if doc := parse(t, updated); doc["default_agent"] != "odachi" {
		t.Errorf("default_agent = %v, want odachi", doc["default_agent"])
	}
}

func TestTheAgentAlreadyChosenIsKeptByteForByte(t *testing.T) {
	original := []byte("{\n  \"default_agent\": \"odachi\"\n}\n")
	updated, out := opencode.EnsureDefaultAgent(original, "odachi")
	if out.Status != opencode.Already {
		t.Fatalf("status = %v, want Already", out.Status)
	}
	if !bytes.Equal(updated, original) {
		t.Errorf("the file changed:\n%s", updated)
	}
}

func TestAnotherAgentIsNeverOverwritten(t *testing.T) {
	original := []byte("{\"default_agent\": \"build\"}\n")
	updated, out := opencode.EnsureDefaultAgent(original, "odachi")
	if out.Status != opencode.Chosen {
		t.Fatalf("status = %v, want Chosen", out.Status)
	}
	if out.Agent != "build" {
		t.Errorf("agent = %q, want build", out.Agent)
	}
	if !bytes.Equal(updated, original) {
		t.Errorf("another default was overwritten:\n%s", updated)
	}
}

func TestDuplicateKeysAreAmbiguous(t *testing.T) {
	original := []byte(`{"default_agent":"build","default_agent":"odachi"}`)
	updated, out := opencode.EnsureDefaultAgent(original, "odachi")
	if out.Status != opencode.Unreadable {
		t.Fatalf("status = %v, want Unreadable", out.Status)
	}
	if !bytes.Equal(updated, original) {
		t.Errorf("an ambiguous file was changed:\n%s", updated)
	}
}

func TestANonStringDefaultIsRefused(t *testing.T) {
	original := []byte(`{"default_agent": 3}`)
	updated, out := opencode.EnsureDefaultAgent(original, "odachi")
	if out.Status != opencode.Unreadable {
		t.Fatalf("status = %v, want Unreadable", out.Status)
	}
	if !bytes.Equal(updated, original) {
		t.Errorf("a refused file was changed:\n%s", updated)
	}
}

func TestAMalformedDocumentIsRefused(t *testing.T) {
	original := []byte(`{"mcp": {`)
	updated, out := opencode.EnsureDefaultAgent(original, "odachi")
	if out.Status != opencode.Unreadable {
		t.Fatalf("status = %v, want Unreadable", out.Status)
	}
	if !bytes.Equal(updated, original) {
		t.Errorf("a malformed file was changed:\n%s", updated)
	}
}

func TestCurrentAgentReadsWithoutWriting(t *testing.T) {
	agent, found, err := opencode.CurrentAgent([]byte("{\n  // who starts\n  \"default_agent\": \"build\"\n}\n"))
	if err != nil {
		t.Fatalf("CurrentAgent: %v", err)
	}
	if !found || agent != "build" {
		t.Errorf("agent = %q, found = %v, want build/true", agent, found)
	}

	if _, found, err := opencode.CurrentAgent([]byte(`{"mcp": {}}`)); err != nil || found {
		t.Errorf("an absent key = found %v, err %v; want false/nil", found, err)
	}
}

func TestTheRootMustBeAnObject(t *testing.T) {
	original := []byte(`["odachi"]`)
	updated, out := opencode.EnsureDefaultAgent(original, "odachi")
	if out.Status != opencode.Unreadable {
		t.Fatalf("status = %v, want Unreadable", out.Status)
	}
	if !bytes.Equal(updated, original) {
		t.Errorf("a non-object config was changed:\n%s", updated)
	}
}
