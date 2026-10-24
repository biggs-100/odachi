package check

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/biggs-100/odachi/internal/domain/fence"
)

// fenceDocument is the TOML shape `odachi start` writes: one [[rule]] per
// non-goal, each with a name and the paths and dependencies it forbids.
//
// The file is small, canonical and owned by this gate, so it is decoded here and
// nowhere else: one owner for the shape, and the fence rules stay pure domain
// values.
type fenceDocument struct {
	Rules []fenceRule `toml:"rule"`
}

type fenceRule struct {
	Not   string   `toml:"not"`
	Paths []string `toml:"paths"`
	Deps  []string `toml:"deps"`
}

// fenceRules reads the human's fence. A missing file means no rules — nothing
// was excluded, so nothing is fenced — and a malformed one blocks with the file
// named, because silently running with an empty fence would mean the gate
// stopped protecting anything without saying so.
func (r *runner) fenceRules() ([]fence.Rule, *Result) {
	data, found, problem := r.fenceBytes()
	if problem != nil {
		return nil, problem
	}
	if !found {
		return nil, nil
	}

	var document fenceDocument
	if _, err := toml.Decode(string(data), &document); err != nil {
		problem := blocked(fmt.Sprintf("could not read %s: %s; fix the file (odachi start writes it) and run odachi check again", fenceFile, oneLine(err.Error())))
		return nil, &problem
	}

	rules := make([]fence.Rule, 0, len(document.Rules))
	for _, rule := range document.Rules {
		rules = append(rules, fence.Rule{Not: rule.Not, Paths: rule.Paths, Deps: rule.Deps})
	}
	return rules, nil
}

// fenceBytes reads the fence the gate obeys. With --fence-from it comes from
// that git ref, not the worktree: CI reads the base branch's policy, which the
// candidate cannot loosen in the same change that asks to pass it. Without the
// flag the worktree's own file is the policy, exactly as before.
func (r *runner) fenceBytes() ([]byte, bool, *Result) {
	if ref := strings.TrimSpace(r.opts.FenceFrom); ref != "" {
		data, found, err := r.git.FileAt(r.opts.Dir, ref, fenceFile)
		if err != nil {
			problem := blocked(fmt.Sprintf("could not read %s from %q: %s; fix --fence-from or drop the flag to use the worktree's fence", fenceFile, ref, oneLine(err.Error())))
			return nil, false, &problem
		}
		return data, found, nil
	}

	path := filepath.Join(r.opts.Dir, fenceFile)
	exists, err := r.files.Exists(path)
	if err != nil {
		problem := blocked(fmt.Sprintf("could not check %s: %s; check the permissions and run odachi check again", fenceFile, oneLine(err.Error())))
		return nil, false, &problem
	}
	if !exists {
		return nil, false, nil
	}

	data, err := r.files.Read(path)
	if err != nil {
		problem := blocked(fmt.Sprintf("could not read %s: %s; check the permissions and run odachi check again", fenceFile, oneLine(err.Error())))
		return nil, false, &problem
	}
	return data, true, nil
}
