package probe

import "os"

// maxSize is the biggest answer this service returns, in bytes.
const maxSize = 1024

var last []byte

// data reads the file, remembers it and returns the size it accepts.
func data(path string) int {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	last = b
	n := len(b)
	if n > maxSize {
		n = maxSize
	}
	return n
}
