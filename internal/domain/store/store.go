// Package store names the clone's state: the part of odachi's own state that
// never travels with the project.
//
// The split is ownership, not taste. A rule the team decided — the fence, the
// profile, the plan — is committed and belongs to the repository. Evidence that
// only this checkout produced — a verdict, an undo backup, a lock — belongs to
// the clone and stays out of the worktree, so a teammate opening the repo sees
// the rules and the plan and nothing else.
//
// It hangs from the git common directory for three reasons. It needs no key: git
// already knows which clone this is, so there is no path hash to invent and no
// orphan to repair when a repository moves. It is shared by every worktree of the
// same clone, which is what one plan wants. And it dies with the clone: deleting
// the checkout deletes the evidence, so nothing accumulates invisibly in the
// human's home.
//
// One limitation follows from the same choice, and it is accepted: two clones of
// the same repository do not share their evidence.
package store

import (
	"fmt"
	"path/filepath"
	"strings"
)

// DirName is the folder inside the git common directory. It is the single
// spelling of the field; nothing outside this package should build the path.
const DirName = "odachi"

// Entries the store holds, spelled once so a message and a test name the same
// path.
const (
	// Reviews holds one verdict per tree.
	Reviews = "reviews"
	// Backups holds the undo copies of the artifacts this tool rewrites.
	Backups = "backups"
	// Lock is the write lock the multi-file verbs hold.
	Lock = "lock"
	// Agents is the client selection: which clients receive the native agent.
	// It is the person's choice, not the team's, so it does not travel.
	Agents = "agents.toml"
)

// Git is the one fact the store needs from the repository: where its common
// directory is. It answers absolute, or the store cannot be built from it.
type Git interface {
	CommonDir(dir string) (string, error)
}

// Root returns the absolute store root for the repository that contains dir.
//
// It does not create anything. Reading a verdict, listing backups or checking
// whether a feature has a proof must never leave a directory behind, so only the
// verbs that write create the entries they need.
func Root(dir string, git Git) (string, error) {
	common, err := git.CommonDir(dir)
	if err != nil {
		return "", fmt.Errorf("could not find the repository's git directory: %s", oneLine(err.Error()))
	}
	return filepath.Join(common, DirName), nil
}

// Path resolves an entry inside the store, accepted with forward slashes.
func Path(root, entry string) string {
	return filepath.Join(root, filepath.FromSlash(entry))
}

// oneLine keeps a message on one line whatever git said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
