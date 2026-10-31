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
		"two spellings": "mode = \"on\"\n[lenses]\nrisk = \"a\"\nRisk = \"b\"\n",
	}

	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := reviewtoml.Decode([]byte(data)); err == nil {
				t.Fatal("Decode accepted a broken lens table")
			}
		})
	}
}

// TestDecodeReadsTheRetention: the file says how much review evidence survives.
// A project that never wrote the keys keeps the numbers odachi shipped, and a
// zero on keep_verdicts is that same absence -the gate reads the newest record,
// so "keep none" is not a lifetime a project can choose-.
func TestDecodeReadsTheRetention(t *testing.T) {
	cases := []struct {
		data   string
		keep   int
		prompt int
	}{
		{data: "mode = \"on\"\n", keep: review.DefaultKeepVerdicts, prompt: review.DefaultKeepPrompts},
		{data: "mode = \"on\"\nkeep_verdicts = 3\nkeep_prompts = 1\n", keep: 3, prompt: 1},
		{data: "mode = \"on\"\nkeep_verdicts = 0\n", keep: review.DefaultKeepVerdicts, prompt: review.DefaultKeepPrompts},
		{data: "mode = \"on\"\nkeep_prompts = 0\n", keep: review.DefaultKeepVerdicts, prompt: 0},
	}

	for _, c := range cases {
		got, err := reviewtoml.Decode([]byte(c.data))
		if err != nil {
			t.Fatalf("Decode(%q) = %v", c.data, err)
		}
		if got.KeepVerdicts != c.keep || got.KeepPrompts != c.prompt {
			t.Errorf("Decode(%q) = %d/%d, want %d/%d", c.data, got.KeepVerdicts, got.KeepPrompts, c.keep, c.prompt)
		}
	}
}

// TestDecodeRefusesARetentionThatCannotBeOne: a negative number is a typo, not a
// lifetime, and a number nobody would ever reach is one too.
func TestDecodeRefusesARetentionThatCannotBeOne(t *testing.T) {
	cases := map[string]string{
		"negative records": "mode = \"on\"\nkeep_verdicts = -1\n",
		"negative prompts": "mode = \"on\"\nkeep_prompts = -2\n",
		"absurd records":   "mode = \"on\"\nkeep_verdicts = 100000\n",
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := reviewtoml.Decode([]byte(data))
			if err == nil {
				t.Fatal("Decode accepted a lifetime that cannot be one")
			}
			if !strings.Contains(err.Error(), "keep_") {
				t.Errorf("error = %q, want the key named", err)
			}
		})
	}
}

// TestRenderKeepsTheRetention: the numbers survive the round trip, and a project
// that never chose gets them written, so the file says what is in force instead
// of leaving it to a default nobody reads.
func TestRenderKeepsTheRetention(t *testing.T) {
	rendered, err := reviewtoml.Render(review.ProjectConfig{Mode: review.ModeOn, KeepVerdicts: 3, KeepPrompts: 1})
	if err != nil {
		t.Fatalf("Render = %v", err)
	}
	back, err := reviewtoml.Decode(rendered)
	if err != nil {
		t.Fatalf("Decode(Render(x)) = %v", err)
	}
	if back.KeepVerdicts != 3 || back.KeepPrompts != 1 {
		t.Errorf("the retention did not survive the round trip: %d/%d", back.KeepVerdicts, back.KeepPrompts)
	}

	empty, err := reviewtoml.Render(review.ProjectConfig{Mode: review.ModeOff})
	if err != nil {
		t.Fatalf("Render(off) = %v", err)
	}
	back, err = reviewtoml.Decode(empty)
	if err != nil {
		t.Fatalf("Decode(Render(off)) = %v", err)
	}
	if back.KeepVerdicts != review.DefaultKeepVerdicts || back.KeepPrompts != review.DefaultKeepPrompts {
		t.Errorf("a file that never chose writes %d/%d, want the defaults %d/%d",
			back.KeepVerdicts, back.KeepPrompts, review.DefaultKeepVerdicts, review.DefaultKeepPrompts)
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

// TestTheSizeLimitSurvivesARoundTrip: the limit is the human's measurement of
// what their reviewer takes, and `review mode on|off` rewrites this file. A writer
// that dropped the key would erase the measurement the first time somebody flipped
// the switch.
func TestTheSizeLimitSurvivesARoundTrip(t *testing.T) {
	data, err := reviewtoml.Render(review.ProjectConfig{Mode: review.ModeOn, MaxPromptKB: 96})
	if err != nil {
		t.Fatalf("Render = %v", err)
	}
	if !strings.Contains(string(data), "max_prompt_kb = 96") {
		t.Errorf("Render = %q, want the limit written", data)
	}

	decoded, err := reviewtoml.Decode(data)
	if err != nil {
		t.Fatalf("Decode(Render) = %v", err)
	}
	if decoded.MaxPromptKB != 96 {
		t.Errorf("MaxPromptKB = %d, want 96", decoded.MaxPromptKB)
	}

	// A file that never wrote the key is no limit, and Render writes no line for
	// it: absence and zero are the same decision.
	plain, err := reviewtoml.Render(review.ProjectConfig{Mode: review.ModeOn})
	if err != nil {
		t.Fatalf("Render(no limit) = %v", err)
	}
	if strings.Contains(string(plain), "max_prompt_kb") {
		t.Errorf("Render = %q, want no limit line for zero", plain)
	}
}

// TestTheSizeLimitRefusesANegative: a negative size is a typo, not a rule.
func TestTheSizeLimitRefusesANegative(t *testing.T) {
	if _, err := reviewtoml.Decode([]byte("mode = \"on\"\nmax_prompt_kb = -3\n")); err == nil {
		t.Fatal("a negative limit passed")
	}
}
