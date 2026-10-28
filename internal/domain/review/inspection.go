package review

import (
	"fmt"
	"strings"
)

// This file owns the reviewer's declaration of what it read.
//
// A row answers a (file, lens) question; it cannot say who answered it, and that
// is the hole the declaration closes. The reviewer says which paths it inspected
// and the gate compares that claim with the paths the candidate changed, in both
// directions: a changed path the declaration leaves out is a hole the caller
// could have filled, and a path nobody changed is an invention. Once the claim is
// on the record, an answer that covered 49 of 52 files is refused as incomplete
// instead of being sealed as if it had read every one.
//
// The vocabulary is the one gentle-ai's admission settled on, and it is closed:
// "completed" claims every path was read, "unavailable" claims none could be.

// InspectionStatus is the closed vocabulary of the declaration.
type InspectionStatus string

const (
	// InspectionCompleted claims every changed path was read by the reviewer.
	InspectionCompleted InspectionStatus = "completed"
	// InspectionUnavailable claims the candidate could not be read at all. It
	// demands the reason, and it names no path.
	InspectionUnavailable InspectionStatus = "unavailable"
)

// Inspection is the reviewer's own account of its reading: the status, the paths
// it read, and — only when it could not read them — why.
type Inspection struct {
	Status InspectionStatus `json:"status"`
	// Paths is the complete set of changed paths the reviewer read. It is empty
	// when the status is unavailable.
	Paths []string `json:"paths"`
	// Reason is required and non-empty only when the status is unavailable.
	Reason string `json:"reason,omitempty"`
}

// InspectionCoverage compares the declaration with the paths the candidate
// changed. It returns the changed paths the declaration leaves out, in the order
// the candidate named them, and the declared paths the candidate does not change,
// in the order the declaration named them.
//
// A verdict without a declaration covers nothing and invents nothing: an older
// verdict is not a lie, and this function is only asked about answers being
// sealed.
func (v Verdict) InspectionCoverage(paths []string) (missing, foreign []string) {
	if v.Inspection == nil {
		return nil, nil
	}

	declared := make(map[string]bool, len(v.Inspection.Paths))
	for _, path := range v.Inspection.Paths {
		declared[findingPath(path)] = true
	}
	changed := make(map[string]bool, len(paths))
	for _, path := range paths {
		changed[findingPath(path)] = true
	}

	for _, path := range paths {
		if !declared[findingPath(path)] {
			missing = append(missing, path)
		}
	}
	for _, path := range v.Inspection.Paths {
		if !changed[findingPath(path)] {
			foreign = append(foreign, path)
		}
	}
	return missing, foreign
}

// validateInspection is the part of Validate that belongs to the declaration. A
// verdict without one is older, not incomplete: every verdict stored before this
// field existed still parses and still validates.
func validateInspection(inspection *Inspection) error {
	if inspection == nil {
		return nil
	}

	switch inspection.Status {
	case InspectionUnavailable:
		if strings.TrimSpace(inspection.Reason) == "" {
			return fmt.Errorf(`the inspection says %q without a reason: a reviewer that could not read the candidate says why ("inspection": {"status": "unavailable", "reason": "..."})`,
				InspectionUnavailable)
		}
		if len(inspection.Paths) > 0 {
			return fmt.Errorf(`the inspection says %q and names %d path(s): a reviewer that could not read the candidate read no path, so "paths" is empty`,
				InspectionUnavailable, len(inspection.Paths))
		}
		return nil
	case InspectionCompleted:
	default:
		return fmt.Errorf(`the inspection says status %q; the only values are %q (every changed path was read) and %q (the candidate could not be read)`,
			inspection.Status, InspectionCompleted, InspectionUnavailable)
	}

	if len(inspection.Paths) == 0 {
		return fmt.Errorf(`the inspection says %q and names no path: list every changed path it read, one row each`,
			InspectionCompleted)
	}

	seen := make(map[string]bool, len(inspection.Paths))
	for _, path := range inspection.Paths {
		if strings.TrimSpace(path) == "" {
			return fmt.Errorf(`the inspection names an empty path: list the changed paths as the candidate names them`)
		}
		canonical := findingPath(path)
		if seen[canonical] {
			return fmt.Errorf(`the inspection names %q twice: the list is a set, and a path counted twice says nothing`, path)
		}
		seen[canonical] = true
	}
	return nil
}

// MergeInspection keeps one declaration for a verdict that grew by lens group.
// Every group was handed the same candidate, so the paths they read are the same
// set; a group that could not read it keeps its claim, because a verdict cannot
// say it inspected the tree when one of its reviewers could not.
func MergeInspection(existing, incoming *Inspection) *Inspection {
	switch {
	case existing == nil:
		return incoming
	case incoming == nil:
		return existing
	case existing.Status == InspectionUnavailable:
		return existing
	case incoming.Status == InspectionUnavailable:
		return incoming
	}

	merged := &Inspection{Status: InspectionCompleted}
	seen := make(map[string]bool, len(existing.Paths)+len(incoming.Paths))
	for _, path := range append(append([]string{}, existing.Paths...), incoming.Paths...) {
		canonical := findingPath(path)
		if seen[canonical] {
			continue
		}
		seen[canonical] = true
		merged.Paths = append(merged.Paths, path)
	}
	return merged
}
