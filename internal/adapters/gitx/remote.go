package gitx

import (
	"errors"
	"fmt"
	"strings"
)

// ErrNoOrigin identifies a genuinely absent origin. Callers may keep the local
// trailer lane closed for this case; an existing origin whose default ref is
// unresolved is a different failure and must not be swallowed.
var ErrNoOrigin = errors.New("no origin remote")

// RemoteBase names the ref CI reads as its base: the remote's default branch as
// this clone sees it. CI builds the gate from that commit and reads the fence
// there, because the candidate cannot edit what it is judged by; a local run
// that wants the same answer has to read the same ref, and naming a branch in
// this code would decide for the project what its base is.
//
// The clone answers for itself, in this order: origin/HEAD, the symref a clone
// writes when it knows the remote's default branch; then origin/main and
// origin/master, the spellings that outlive a lost symref. Nothing is fetched:
// the git edge never waits on the network, so a stale ref is a base the human
// can see named in the run, not one this function guesses around. A dangling
// origin/HEAD is unresolved evidence, not permission to choose another branch.
func RemoteBase(root string) (string, error) {
	if out, err := run(root, "symbolic-ref", "--quiet", "refs/remotes/origin/HEAD"); err == nil {
		ref := strings.TrimPrefix(strings.TrimSpace(out), "refs/remotes/")
		if ref != "" {
			if _, err := run(root, "rev-parse", "--verify", ref+"^{commit}"); err == nil {
				return ref, nil
			}
			return "", fmt.Errorf("origin's default ref %s is not fetched: run git fetch origin, or pass --base <ref>", ref)
		}
	}

	for _, ref := range []string{"origin/main", "origin/master"} {
		if _, err := run(root, "rev-parse", "--verify", ref+"^{commit}"); err == nil {
			return ref, nil
		}
	}

	out, err := run(root, "remote")
	if err != nil {
		// A folder without a repository is not an unresolved origin. Preserve the
		// typed state so the caller can defer to its usual no-repository refusal.
		if _, repoErr := CommonDir(root); repoErr != nil {
			return "", repoErr
		}
		return "", fmt.Errorf("could not read git remotes: %w", err)
	}
	for name := range strings.SplitSeq(out, "\n") {
		if strings.TrimSpace(name) == "origin" {
			return "", fmt.Errorf("origin carries no fetched branch this gate can call the base: looked for origin/HEAD, origin/main and origin/master; run git fetch origin and git remote set-head origin --auto, or pass --base <ref>")
		}
	}

	return "", fmt.Errorf("%w: this clone has no origin remote to read a base from; git remote add origin <url>, then git fetch origin", ErrNoOrigin)
}
