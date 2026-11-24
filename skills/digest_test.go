package skills_test

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/biggs-100/odachi/skills"
)

// canonicalBytes is the digest contract written independently of Digest: skills
// in All() order, files in Path order, and each record name + NUL + path + NUL +
// 8-byte big-endian length + content. A change to the implementation's format
// fails here even when the implementation and the test share the same input.
func canonicalBytes(t *testing.T) []byte {
	t.Helper()
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}

	stream := []byte{}
	for _, s := range all {
		for _, file := range s.Files {
			stream = append(stream, []byte(s.Name)...)
			stream = append(stream, 0)
			stream = append(stream, []byte(file.Path)...)
			stream = append(stream, 0)
			var length [8]byte
			binary.BigEndian.PutUint64(length[:], uint64(len(file.Content)))
			stream = append(stream, length[:]...)
			stream = append(stream, file.Content...)
		}
	}
	return stream
}

func TestDigestPinsTheEmbeddedSkillBytes(t *testing.T) {
	got, err := skills.Digest()
	if err != nil {
		t.Fatalf("Digest(): %v", err)
	}
	sum := sha256.Sum256(canonicalBytes(t))
	want := hex.EncodeToString(sum[:])
	if got != want {
		t.Fatalf("Digest() = %q, want %q", got, want)
	}
	if len(got) != 64 {
		t.Errorf("Digest() has %d characters, want 64 hex characters: %q", len(got), got)
	}
	for _, char := range got {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			t.Fatalf("Digest() is not lowercase hex: %q", got)
		}
	}
}

func TestDigestIsStableAcrossCalls(t *testing.T) {
	first, err := skills.Digest()
	if err != nil {
		t.Fatalf("first Digest(): %v", err)
	}
	second, err := skills.Digest()
	if err != nil {
		t.Fatalf("second Digest(): %v", err)
	}
	if first != second {
		t.Fatalf("Digest() changed between calls: %q then %q", first, second)
	}
}
