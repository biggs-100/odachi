package reviewtoml_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/reviewtoml"
	"github.com/biggs-100/odachi/internal/domain/review"
)

// TestDecodeReadsTheSwitch: the one key the file carries, normalized by the
// domain.
func TestDecodeReadsTheSwitch(t *testing.T) {
	cases := map[string]string{
		"mode = \"on\"\n":  review.ModeOn,
		"mode = \"OFF\"\n": review.ModeOff,
	}

	for data, want := range cases {
		got, err := reviewtoml.Decode([]byte(data))
		if err != nil {
			t.Fatalf("Decode(%q) = %v", data, err)
		}
		if got != want {
			t.Errorf("Decode(%q) = %q, want %q", data, got, want)
		}
	}
}

// TestDecodeRefusesATypo: unknown keys and unknown modes are errors, because a
// switch the tool half-understands is a gate that may never open.
func TestDecodeRefusesATypo(t *testing.T) {
	cases := map[string]string{
		"unknown key": "mode = \"on\"\nrequired = true\n",
		"bad mode":    "mode = \"maybe\"\n",
		"empty mode":  "mode = \"\"\n",
		"malformed":   "mode = \"on\n",
	}

	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if got, err := reviewtoml.Decode([]byte(data)); err == nil {
				t.Fatalf("Decode = %q, want an error", got)
			}
		})
	}
}

// TestDecodeNamesTheUnknownKey: the human must know which line to remove.
func TestDecodeNamesTheUnknownKey(t *testing.T) {
	_, err := reviewtoml.Decode([]byte("mode = \"on\"\nrequired = true\n"))
	if err == nil {
		t.Fatal("Decode accepted an unknown key")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("error = %q, want the unknown key named", err)
	}
}

// TestRenderWritesACanonicalSwitch: the comment tells the human whose switch it
// is, and the value is the normalized one.
func TestRenderWritesACanonicalSwitch(t *testing.T) {
	data, err := reviewtoml.Render(" ON ")
	if err != nil {
		t.Fatalf("Render(ON) = %v", err)
	}
	rendered := string(data)
	for _, want := range []string{"the human's decision", `mode = "on"`} {
		if !strings.Contains(rendered, want) {
			t.Errorf("Render = %q, want %q", rendered, want)
		}
	}

	decoded, err := reviewtoml.Decode(data)
	if err != nil {
		t.Fatalf("Decode(Render(on)) = %v", err)
	}
	if decoded != review.ModeOn {
		t.Errorf("Decode(Render(on)) = %q, want %q", decoded, review.ModeOn)
	}
}
