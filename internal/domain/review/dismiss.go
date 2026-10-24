package review

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

// This file owns dismissal identity: the states a finding can have, the reason
// a dismissal carries, and the fingerprint that lets a later candidate
// recognize the same defect instead of re-raising it.
//
// A dismissal is not a deleted finding. The verdict keeps the row, the evidence
// that judged it and the fingerprint of what it judged: a review that cannot
// defend its silence does not stay silent, it records why.

const (
	// Open is a finding still standing. It is also the meaning of an empty
	// status, because every verdict written before dismissals existed carries
	// open findings.
	Open = "open"
	// Dismissed is a finding judged and discarded with evidence: a refutation
	// that cited why, or a human's dismissal.
	Dismissed = "dismissed"
)

// IsDismissed reports whether this row was judged and discarded. An empty
// status is an open finding, not a third state.
func IsDismissed(f Finding) bool {
	return strings.EqualFold(strings.TrimSpace(f.Status), Dismissed)
}

// Dismiss marks one finding as judged and discarded. The reason is the
// evidence: a dismissal without it is refused by Validate, because silence is
// not an argument.
func Dismiss(f Finding, reason string) Finding {
	dismissed := f
	dismissed.Status = Dismissed
	dismissed.Reason = strings.TrimSpace(reason)
	dismissed.Fingerprint = Fingerprint(f)
	return dismissed
}

// Fingerprint is the identity of the defect a finding describes: where it is,
// which line it names, which lens answered, and the evidence itself. Two
// spellings of the same evidence hash the same; a different line, path, lens or
// claim is a different defect.
//
// The line is part of it on purpose: a dismissal only travels while its blob
// matches, so the line cannot drift behind the dismissal without the receipt
// saying the bytes changed — and when the bytes change, the judgment is over.
func Fingerprint(f Finding) string {
	evidence := strings.ToLower(strings.Join(strings.Fields(f.Why), " "))
	parts := []string{
		normalizePath(f.Path),
		strconv.Itoa(f.Line),
		strings.ToLower(strings.TrimSpace(f.Lens)),
		evidence,
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])[:16]
}
