package doctor

import (
	domaindoctor "github.com/biggs-100/odachi/internal/domain/doctor"
	"github.com/biggs-100/odachi/internal/domain/review"
	"github.com/biggs-100/odachi/internal/domain/store"
	"github.com/biggs-100/odachi/internal/domain/verify"
)

// store is check 18. The evidence store -the reviews and verifications the clone
// keeps outside the worktree- holds the verdicts doctor's own proof relies on. A
// folder it cannot read is not an empty store: the human has to be told where the
// unreadable bytes are and what to fix (0349).
//
// It is read-only, like every check: List on the root and on each record folder
// is the whole question. A store that is not there yet has no entries and is ok;
// reading it must never create it.
func (r *runner) store() domaindoctor.Check {
	if r.opts.Git == nil {
		// The CLI always wires the repository port; this is a run that has no way
		// to locate a store, so it cannot call one broken.
		return domaindoctor.OK(domaindoctor.CheckStore,
			"no repository port: the evidence store was not located in this run")
	}

	root, err := store.Root(r.opts.Dir, r.opts.Git)
	if err != nil {
		return domaindoctor.Broken(domaindoctor.CheckStore,
			"could not locate the evidence store: "+oneLine(err.Error()),
			"check that this folder is a git repository and run odachi doctor again")
	}

	for _, dir := range []string{root, store.Path(root, review.Dir), store.Path(root, verify.Dir)} {
		if _, err := r.files.List(dir); err != nil {
			return domaindoctor.Broken(domaindoctor.CheckStore,
				"could not read the evidence store at "+dir+": "+oneLine(err.Error()),
				"check the permissions on "+dir+" and run odachi doctor again")
		}
	}

	return domaindoctor.OK(domaindoctor.CheckStore,
		"the evidence store at "+root+" is readable")
}
