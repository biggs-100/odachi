package rulebooktoml

import (
	"strconv"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/rulebook"
	"github.com/biggs-100/odachi/internal/domain/schema"
)

// TestDecodeReadsTheTwoShapes: an enforced row and a debt row are both legal, and
// the debt row's because is what makes it a decision instead of a silence.
func TestDecodeReadsTheTwoShapes(t *testing.T) {
	// A file with no schema line reads as the first schema, which is what a
	// project that wrote the table before the key existed has.
	data := []byte(`
[[rule]]
rule     = "the proof is this tree's real output"
skill    = "drive"
enforces = "fence:proof/**"

[[rule]]
rule     = "the voice is the profile's, not the session's"
skill    = "drive"
enforces = "none"
because  = "no command can read a register"
`)
	rules, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("Decode read %d rows, want 2", len(rules))
	}
	if rules[0].Enforces != "fence:proof/**" || rules[1].Enforces != "none" || rules[1].Because == "" {
		t.Errorf("Decode = %+v, want the enforcer and the reason as written", rules)
	}
}

// TestDecodeRefusesATypo: the table is what a session counts, so an unknown key
// is named instead of dropped and a bad row is refused by its number.
func TestDecodeRefusesATypo(t *testing.T) {
	_, err := Decode([]byte("[[rule]]\nrule = \"x\"\nskill = \"drive\"\nenforce = \"none\"\n"))
	if err == nil || !strings.Contains(err.Error(), "enforce") {
		t.Fatalf("Decode = %v, want it to name the misspelled key", err)
	}

	_, err = Decode([]byte("[[rule]]\nrule = \"x\"\nskill = \"drive\"\nenforces = \"lint:yes\"\n"))
	if err == nil || !strings.Contains(err.Error(), "row 1") {
		t.Fatalf("Decode = %v, want it to name the row", err)
	}
}

// TestDecodeRefusesANewerFile: a table written by a newer odachi is not broken,
// and the message has to say so or the human edits a file that is fine.
func TestDecodeRefusesANewerFile(t *testing.T) {
	data := []byte("schema = " + strconv.Itoa(schema.Current+1) + "\n")
	_, err := Decode(data)
	if err == nil || !strings.Contains(err.Error(), "newer odachi") {
		t.Fatalf("Decode = %v, want the upgrade message", err)
	}
}

// TestRenderRoundTrips: the writer and the reader are one grammar, so the bytes
// of a rendered table decode to the same rows.
func TestRenderRoundTrips(t *testing.T) {
	in := []rulebook.Rule{
		{Statement: "a", Skill: "drive", Enforces: "fence:proof/**"},
		{Statement: "b", Skill: "writing", Enforces: "none", Because: "a register cannot be held by a command"},
	}

	rendered, err := Render(in)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(string(rendered), "schema = ") {
		t.Errorf("Render wrote no schema:\n%s", rendered)
	}
	if !strings.Contains(string(rendered), "because  = ") {
		t.Errorf("Render dropped the because of the debt row:\n%s", rendered)
	}

	decoded, err := Decode(rendered)
	if err != nil {
		t.Fatalf("Decode(Render(x)): %v", err)
	}
	if len(decoded) != 2 {
		t.Fatalf("round trip read %d rows, want 2", len(decoded))
	}
	if decoded[0] != in[0] || decoded[1] != in[1] {
		t.Errorf("round trip = %+v, want %+v", decoded, in)
	}
}

// TestRenderRefusesADebtWithNoReason: the writer enforces the same rule the
// reader does, because a file written by the tool cannot be one the tool refuses.
func TestRenderRefusesADebtWithNoReason(t *testing.T) {
	_, err := Render([]rulebook.Rule{{Statement: "a", Skill: "drive", Enforces: "none"}})
	if err == nil || !strings.Contains(err.Error(), "row 1") {
		t.Fatalf("Render = %v, want it to refuse the row", err)
	}
}
