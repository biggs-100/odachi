package profiletoml_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggz/odachi/internal/adapters/profiletoml"
	"github.com/biggz/odachi/internal/domain/profile"
)

// example is the canonical profile artifact, relative to this package:
// internal/adapters/profiletoml -> module root -> testdata/examples.
func example(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "..", "testdata", "examples", "profile.toml")
}

func readExample(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(example(t))
	if err != nil {
		t.Fatalf("read example profile: %v", err)
	}
	return data
}

func TestDecodeCanonicalProfile(t *testing.T) {
	got, err := profiletoml.Decode(readExample(t))
	if err != nil {
		t.Fatalf("Decode(profile.toml) error: %v", err)
	}

	want := profile.Profile{
		Name:        "novato",
		Extends:     "",
		BriefSource: "interview",
		Questions:   "one-by-one",
		Assumptions: "never",
		Output:      "line",
		Proof:       "command",
		Explain:     "why",
	}
	if got != want {
		t.Errorf("Decode(profile.toml) = %+v, want %+v", got, want)
	}
	if err := profile.Validate(got); err != nil {
		t.Errorf("decoded profile does not validate: %v", err)
	}
}

func TestDecodeRejectsUnknownKeysNamingThem(t *testing.T) {
	// gates_disabled is the field Contract B forbids a profile from holding:
	// it must be rejected by name, not silently ignored.
	data := []byte(readExample(t)) // valid keys
	data = append(data, []byte("gates_disabled = true\n")...)

	_, err := profiletoml.Decode(data)
	if err == nil {
		t.Fatalf("Decode accepted an unknown key, want an error naming gates_disabled")
	}
	if !strings.Contains(err.Error(), "gates_disabled") {
		t.Errorf("error %q does not name the offending key gates_disabled", err)
	}
}

func TestDecodeRejectsUnknownKeyAlongsideValidOnes(t *testing.T) {
	data := []byte("name = \"novato\"\nfence = \"off\"\n")
	_, err := profiletoml.Decode(data)
	if err == nil {
		t.Fatalf("Decode accepted the unknown key fence, want an error")
	}
	if !strings.Contains(err.Error(), "fence") {
		t.Errorf("error %q does not name the offending key fence", err)
	}
}

func TestDecodeSurfacesValidateErrors(t *testing.T) {
	data := []byte("name = \"novato\"\noutput = \"verbose\"\n")

	_, err := profiletoml.Decode(data)
	if err == nil {
		t.Fatalf("Decode accepted output = \"verbose\", want a validation error")
	}
	// profile.Validate is the single source of the allowed values and of the
	// message; the adapter must not re-implement it.
	if !strings.Contains(err.Error(), "no es un valor permitido") {
		t.Errorf("error %q is not profile.Validate's error", err)
	}
	if !strings.Contains(err.Error(), "output") {
		t.Errorf("error %q does not name the field output", err)
	}
}

func TestDecodeRejectsUnparseableTOML(t *testing.T) {
	if _, err := profiletoml.Decode([]byte("name = \"sin cerrar\n")); err == nil {
		t.Fatalf("Decode accepted malformed TOML, want an error")
	}
	if _, err := profiletoml.Decode(nil); err != nil {
		t.Fatalf("Decode(nil) error: %v; an empty file is an empty profile, inherited field by field", err)
	}
}

func TestLoadReadsTheFileAndDelegatesToDecode(t *testing.T) {
	got, err := profiletoml.Load(example(t))
	if err != nil {
		t.Fatalf("Load(profile.toml) error: %v", err)
	}

	want, err := profiletoml.Decode(readExample(t))
	if err != nil {
		t.Fatalf("Decode(profile.toml) error: %v", err)
	}
	if got != want {
		t.Errorf("Load = %+v, want Decode = %+v", got, want)
	}
}

func TestLoadReportsAMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.toml")

	_, err := profiletoml.Load(path)
	if err == nil {
		t.Fatalf("Load(%s) succeeded, want an error", path)
	}
	if !strings.Contains(err.Error(), "profile.toml") {
		t.Errorf("error %q does not name the missing file", err)
	}
}
