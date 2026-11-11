package probe

// maxSize is the biggest answer this service returns, in bytes.
const maxSize = 1024

// limit caps the size a caller asked for, and refuses a negative one.
func limit(n int) int {
	if n < 0 {
		return 0
	}
	if n > maxSize {
		return maxSize
	}
	return n
}
