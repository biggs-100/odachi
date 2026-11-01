package gitx

import (
	"fmt"
	"strings"
)

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
// can see named in the run, not one this function guesses around.
func RemoteBase(root string) (string, error) {
	if out, err := run(root, "symbolic-ref", "--quiet", "refs/remotes/origin/HEAD"); err == nil {
		return strings.TrimPrefix(strings.TrimSpace(out), "refs/remotes/"), nil
	}

	for _, ref := range []string{"origin/main", "origin/master"} {
		if _, err := run(root, "rev-parse", "--verify", ref+"^{commit}"); err == nil {
			return ref, nil
		}
	}

	if _, err := run(root, "remote", "get-url", "origin"); err != nil {
		return "", fmt.Errorf("this clone has no origin remote to read a base from: git remote add origin <url>, then git fetch origin")
	}
	return "", fmt.Errorf("origin carries no branch this gate can call the base: looked for origin/HEAD, origin/main and origin/master; run git fetch origin and git remote set-head origin --auto")
}
