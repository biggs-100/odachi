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
		if got.Mode != want {
			t.Errorf("Decode(%q) = %q, want %q", data, got.Mode, want)
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

// The lens table says who judges each lens; it is read back in the vocabulary's
// order, not in the order the human happened to write it.
func TestDecodeReadsALensTable(t *testing.T) {
	data := []byte("mode = \"on\"\n\n[lenses]\nreliability = \"cheap/model\"\nRisk = \"sharp/model\"\n")
	got, err := reviewtoml.Decode(data)
	if err != nil {
		t.Fatalf("Decode = %v", err)
	}
	if got.Lenses["risk"] != "sharp/model" || got.Lenses["reliability"] != "cheap/model" {
		t.Fatalf("Lenses = %v, want both entries with the lens lowercased", got.Lenses)
	}

	rendered, err := reviewtoml.Render(got)
	if err != nil {
		t.Fatalf("Render = %v", err)
	}
	risk := strings.Index(string(rendered), "risk = ")
	reliability := strings.Index(string(rendered), "reliability = ")
	if risk < 0 || reliability < 0 || risk > reliability {
		t.Errorf("Render =\n%s\nwant the lens table in the fixed order", rendered)
	}

	back, err := reviewtoml.Decode(rendered)
	if err != nil {
		t.Fatalf("Decode(Render(x)) = %v", err)
	}
	if back.Lenses["risk"] != "sharp/model" || back.Lenses["reliability"] != "cheap/model" {
		t.Errorf("the table did not survive the round trip: %v", back.Lenses)
	}
}

// A lens the vocabulary does not know, or an entry with no model, is refused:
// a plan for a lens nobody asks would be a model the human pays for nothing.
func TestDecodeRefusesABrokenLensTable(t *testing.T) {
	cases := map[string]string{
		"unknown lens":  "mode = \"on\"\n[lenses]\nperformance = \"m\"\n",
		"no model":      "mode = \"on\"\n[lenses]\nrisk = \"\"\n",
		"shell code":    "mode = \"on\"\n[lenses]\nrisk = \"x; rm -rf ~\"\n",
		"command break": "mode = \"on\"\n[lenses]\nrisk = \"a && b\"\n",
	}

	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := reviewtoml.Decode([]byte(data)); err == nil {
				t.Fatal("Decode accepted a broken lens table")
			}
		})
	}
}

// TestRenderWritesACanonicalSwitch: the comment tells the human whose switch it
// is, and the value is the normalized one.
func TestRenderWritesACanonicalSwitch(t *testing.T) {
	data, err := reviewtoml.Render(review.ProjectConfig{Mode: " ON "})
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
	if decoded.Mode != review.ModeOn {
		t.Errorf("Decode(Render(on)) = %q, want %q", decoded.Mode, review.ModeOn)
	}
}
