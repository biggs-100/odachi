package jsonmember_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/jsonmember"
)

// parse reads the document the way a client would: the file may carry comments,
// so it strips them through the same helper the editor validates with.
func parse(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(jsonmember.StripComments(data), &doc); err != nil {
		t.Fatalf("the result is not valid JSON: %v\n%s", err, data)
	}
	return doc
}

// TestAnEmptyDocumentGetsTheMember: a file that does not exist yet is the same
// question as a document with no bytes, and the answer is the smallest object.
func TestAnEmptyDocumentGetsTheMember(t *testing.T) {
	updated, out := jsonmember.Ensure(nil, "model", "anthropic/claude#fast")
	if out.Status != jsonmember.Added {
		t.Fatalf("status = %v, want Added (%s)", out.Status, out.Reason)
	}
	if doc := parse(t, updated); doc["model"] != "anthropic/claude#fast" {
		t.Errorf("model = %v, want anthropic/claude#fast", doc["model"])
	}
}

// TestTheMemberGoesFirstAndTheRestIsUntouched: the edit is an insertion right
// after the root brace. Comments and CRLF are the human's bytes, not formatting.
func TestTheMemberGoesFirstAndTheRestIsUntouched(t *testing.T) {
	original := "{\r\n  // the client reads this\r\n  \"theme\": \"dark\"\r\n}\r\n"
	updated, out := jsonmember.Ensure([]byte(original), "model", "x/y")
	if out.Status != jsonmember.Added {
		t.Fatalf("status = %v, want Added (%s)", out.Status, out.Reason)
	}
	for _, want := range []string{"// the client reads this", `"theme": "dark"`} {
		if !strings.Contains(string(updated), want) {
			t.Errorf("the result lost %q:\n%s", want, updated)
		}
	}
	if bytes.Contains(bytes.ReplaceAll(updated, []byte("\r\n"), nil), []byte("\n")) {
		t.Errorf("the edit introduced a bare LF into a CRLF file: %q", updated)
	}
	if strings.Index(string(updated), "model") > strings.Index(string(updated), "theme") {
		t.Errorf("model did not land first:\n%s", updated)
	}
	parse(t, updated)
}

// TestAnExistingValueWithTheSameBytesIsKept: a rerun is not a rewrite.
func TestAnExistingValueWithTheSameBytesIsKept(t *testing.T) {
	original := []byte("{\n  \"model\": \"x/y\"\n}\n")
	updated, out := jsonmember.Ensure(original, "model", "x/y")
	if out.Status != jsonmember.Already {
		t.Fatalf("status = %v, want Already", out.Status)
	}
	if !bytes.Equal(updated, original) {
		t.Errorf("the file changed:\n%s", updated)
	}
}

// TestReplacingAValueLeavesEveryOtherByte: the replacement touches the string's
// content and nothing else - no re-encode, no reordered keys, no lost comment.
func TestReplacingAValueLeavesEveryOtherByte(t *testing.T) {
	original := "{\n  // who runs it\n  \"theme\": \"dark\",\n  \"model\": \"old/name#a\",\n  \"tail\": [1, 2]\n}\n"
	updated, out := jsonmember.Ensure([]byte(original), "model", "new/name#b")
	if out.Status != jsonmember.Replaced {
		t.Fatalf("status = %v, want Replaced (%s)", out.Status, out.Reason)
	}
	if out.Previous != "old/name#a" {
		t.Errorf("previous = %q, want old/name#a", out.Previous)
	}
	want := strings.Replace(original, `"old/name#a"`, `"new/name#b"`, 1)
	if string(updated) != want {
		t.Errorf("the edit moved bytes it did not have to:\n%q\nwant\n%q", updated, want)
	}
}

// TestANonObjectRootIsRefused: an array has no member to set.
func TestANonObjectRootIsRefused(t *testing.T) {
	original := []byte(`["model"]`)
	updated, out := jsonmember.Ensure(original, "model", "x/y")
	if out.Status != jsonmember.Unreadable {
		t.Fatalf("status = %v, want Unreadable", out.Status)
	}
	if !bytes.Equal(updated, original) {
		t.Errorf("a non-object document was changed:\n%s", updated)
	}
}

// TestADuplicateKeyIsRefused: encoding/json would keep the last silently, and
// the human's intent cannot be guessed.
func TestADuplicateKeyIsRefused(t *testing.T) {
	original := []byte(`{"model":"a/b","model":"c/d"}`)
	updated, out := jsonmember.Ensure(original, "model", "e/f")
	if out.Status != jsonmember.Unreadable {
		t.Fatalf("status = %v, want Unreadable", out.Status)
	}
	if !bytes.Equal(updated, original) {
		t.Errorf("an ambiguous file was changed:\n%s", updated)
	}
}

// TestAMalformedDocumentIsRefused: a trailing comma is not repaired; the human's
// file is refused instead of being reformatted.
func TestAMalformedDocumentIsRefused(t *testing.T) {
	original := []byte(`{"theme": "dark",}`)
	updated, out := jsonmember.Ensure(original, "model", "x/y")
	if out.Status != jsonmember.Unreadable {
		t.Fatalf("status = %v, want Unreadable", out.Status)
	}
	if !bytes.Equal(updated, original) {
		t.Errorf("a malformed file was changed:\n%s", updated)
	}
}

// TestANonStringMemberIsRefused: a value we cannot read is not a value to replace.
func TestANonStringMemberIsRefused(t *testing.T) {
	original := []byte(`{"model": 3}`)
	updated, out := jsonmember.Ensure(original, "model", "x/y")
	if out.Status != jsonmember.Unreadable {
		t.Fatalf("status = %v, want Unreadable", out.Status)
	}
	if !bytes.Equal(updated, original) {
		t.Errorf("a refused file was changed:\n%s", updated)
	}
}

// TestAnEmptyValueIsRefused: the flag grammar and the editor agree that a model
// names something, and an empty string names nothing.
func TestAnEmptyValueIsRefused(t *testing.T) {
	original := []byte("{\n  \"theme\": \"dark\"\n}\n")
	updated, out := jsonmember.Ensure(original, "model", "  ")
	if out.Status != jsonmember.Unreadable {
		t.Fatalf("status = %v, want Unreadable", out.Status)
	}
	if !bytes.Equal(updated, original) {
		t.Errorf("an empty value was written:\n%s", updated)
	}
}

// TestNormalizeIdentifierKeepsTheModelAlphabet: provider/name#variant, with the
// dots, colons, plus and dashes model ids use.
func TestNormalizeIdentifierKeepsTheModelAlphabet(t *testing.T) {
	for _, value := range []string{"anthropic/claude#fast", "gpt-4o", "a.b:c+d@e", " openai/gpt-5 "} {
		got, err := jsonmember.NormalizeIdentifier(value)
		if err != nil {
			t.Errorf("NormalizeIdentifier(%q): %v", value, err)
			continue
		}
		if got != strings.TrimSpace(value) {
			t.Errorf("NormalizeIdentifier(%q) = %q, want it trimmed", value, got)
		}
	}
}

// TestNormalizeIdentifierRefusesWhatIsNotAnIdentifier: the value ends up inside
// a file a CLI parses, so shell punctuation and spaces are refused by name.
func TestNormalizeIdentifierRefusesWhatIsNotAnIdentifier(t *testing.T) {
	for _, value := range []string{"", "  ", "sonnet 4", "x;rm", `x"y`, "x\ny"} {
		if _, err := jsonmember.NormalizeIdentifier(value); err == nil {
			t.Errorf("NormalizeIdentifier(%q) accepted it", value)
		}
	}
}
