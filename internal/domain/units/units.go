// Package units formats the numbers a person reads, so two verbs cannot print
// the same size two ways: `clean` says what the store holds and `doctor` says
// what the always-on harness costs, and a reader compares them.
package units

import "fmt"

// Bytes is a size as a human reads it: bytes below a kilobyte, one decimal below
// ten of them, whole kilobytes below a megabyte, one decimal above it.
func Bytes(n int64) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 10*1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	case n < 1024*1024:
		return fmt.Sprintf("%.0f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	}
}
