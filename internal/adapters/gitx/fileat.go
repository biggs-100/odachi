package gitx

import (
	"fmt"
	"strings"
)

// FileAt answers the bytes of path at rev, and whether that revision has it.
//
// It is how a gate reads the policy it must obey from a ref the candidate cannot
// edit: the base branch, not the worktree. A missing file at a valid rev is
// "this revision has no such file" (found false), not an error; a rev that does
// not exist is an error, because silently treating a typo as "no rules" would
// turn the gate off.
func FileAt(root, rev, path string) ([]byte, bool, error) {
	ref := strings.TrimSpace(rev)
	if ref == "" {
		return nil, false, fmt.Errorf("git: there is no revision to read %s from", path)
	}

	if _, err := run(root, "cat-file", "-e", ref+":"+path); err != nil {
		if _, revErr := run(root, "rev-parse", "--verify", ref); revErr != nil {
			return nil, false, fmt.Errorf("git: %q is not a revision: %w", ref, revErr)
		}
		return nil, false, nil
	}

	out, err := run(root, "show", ref+":"+path)
	if err != nil {
		return nil, false, err
	}
	return []byte(out), true, nil
}
