package agentstoml_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/agentstoml"
	"github.com/biggs-100/odachi/internal/domain/persona"
)

// TestDecodeReadsTheSelection: the one key the file carries, normalized by the
// domain.
func TestDecodeReadsTheSelection(t *testing.T) {
	cases := map[string]string{
		"agents = \"opencode\"\n": persona.SelectionOpenCode,
		"agents = \"BOTH\"\n":     persona.SelectionBoth,
	}

	for data, want := range cases {
		got, err := agentstoml.Decode([]byte(data))
		if err != nil {
			t.Fatalf("Decode(%q) = %v", data, err)
		}
		if got != want {
			t.Errorf("Decode(%q) = %q, want %q", data, got, want)
		}
	}
}

// TestDecodeRefusesATypo: unknown keys and unknown values are errors, because a
// store the tool half-understands would silently pick a client.
func TestDecodeRefusesATypo(t *testing.T) {
	cases := map[string]string{
		"unknown key": "agents = \"opencode\"\nrequired = true\n",
		"bad value":   "agents = \"maybe\"\n",
		"empty value": "agents = \"\"\n",
		"malformed":   "agents = \"opencode\n",
	}

	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if got, err := agentstoml.Decode([]byte(data)); err == nil {
				t.Fatalf("Decode = %q, want an error", got)
			}
		})
	}
}

// TestDecodeNamesTheUnknownKey: the human must know which line to remove.
func TestDecodeNamesTheUnknownKey(t *testing.T) {
	_, err := agentstoml.Decode([]byte("agents = \"opencode\"\nrequired = true\n"))
	if err == nil {
		t.Fatal("Decode accepted an unknown key")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("error = %q, want the unknown key named", err)
	}
}

// TestRenderWritesACanonicalStore: the comment says the choice disables writing
// a client, not deleting what is already there, and the value is normalized.
func TestRenderWritesACanonicalStore(t *testing.T) {
	data, err := agentstoml.Render(" OPENCODE ")
	if err != nil {
		t.Fatalf("Render(OPENCODE) = %v", err)
	}
	rendered := string(data)
	for _, want := range []string{"no lo borra", `agents = "opencode"`} {
		if !strings.Contains(rendered, want) {
			t.Errorf("Render = %q, want %q", rendered, want)
		}
	}

	decoded, err := agentstoml.Decode(data)
	if err != nil {
		t.Fatalf("Decode(Render(opencode)) = %v", err)
	}
	if decoded != persona.SelectionOpenCode {
		t.Errorf("Decode(Render(opencode)) = %q, want %q", decoded, persona.SelectionOpenCode)
	}
}
