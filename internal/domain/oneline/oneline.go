// Package oneline holds the single implementation of the whitespace collapse
// every odachi message uses before it reaches a person.
package oneline

import "strings"

// Line folds every run of whitespace in s into one space and trims the ends.
// A wrapper named oneLine in another package delegates here, so the message a
// person reads is built the same way everywhere in the tree.
func Line(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
