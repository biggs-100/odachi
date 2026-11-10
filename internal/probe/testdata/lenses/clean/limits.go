package probe

// maxSize is the biggest answer this service returns, in bytes.
const maxSize = 1024

// limit caps the size a caller asked for.
func limit(n int) int {
	if n > maxSize {
		return maxSize
	}
	return n
}
