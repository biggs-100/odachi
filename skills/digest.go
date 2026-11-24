package skills

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

// Digest is the SHA-256 of the canonical bytes of the embedded skill payload:
// skills in All() order, files in Path order, and each record name + NUL + path
// + NUL + 8-byte big-endian length + content. The length prefix keeps the stream
// unambiguous, so moving a content boundary cannot produce the same bytes. It
// reads only the embedded filesystem: no checkout, no clock and no map order.
func Digest() (string, error) {
	all, err := All()
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	var length [8]byte
	for _, skill := range all {
		name := []byte(skill.Name)
		for _, file := range skill.Files {
			hash.Write(name)
			hash.Write([]byte{0})
			hash.Write([]byte(file.Path))
			hash.Write([]byte{0})
			binary.BigEndian.PutUint64(length[:], uint64(len(file.Content)))
			hash.Write(length[:])
			hash.Write(file.Content)
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
