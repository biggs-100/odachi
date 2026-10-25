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
	return decodeFence(data, found)
}

// worktreeFenceRules reads the fence the candidate would publish, the one that
// becomes the policy after the merge. With --fence-from the gate obeys the ref's
// fence, so this is the copy that has to be compared against it.
func (r *runner) worktreeFenceRules() ([]fence.Rule, *Result) {
	data, found, problem := r.worktreeFenceBytes()
	if problem != nil {
		return nil, problem
	}
	return decodeFence(data, found)
}

// decodeFence turns the file into rules. One decoder for both copies: the base
// fence and the candidate's cannot be read by two different grammars.
func decodeFence(data []byte, found bool) ([]fence.Rule, *Result) {
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

	return r.worktreeFenceBytes()
}

// worktreeFenceBytes reads the candidate's own copy. It is the policy when there
// is no --fence-from, and the copy to compare when there is one.
func (r *runner) worktreeFenceBytes() ([]byte, bool, *Result) {
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

// fenceLoosenings is the rule the review of feature 0030 found missing: a
// candidate could empty the fence with the brief untouched and the gate still
// said ready. The fence may only be loosened by a change that also changes the
// brief, because the brief and the fence are the human's and `odachi start`
// writes them together. A new brief is a new policy; the same brief keeps the
// policy it had.
func (r *runner) fenceLoosenings() ([]string, *Result) {
	ref := strings.TrimSpace(r.opts.FenceFrom)
	if ref == "" {
		return nil, nil
	}

	changed, problem := r.briefChanged(ref)
	if problem != nil {
		return nil, problem
	}
	if changed {
		return nil, nil
	}

	base, problem := r.fenceRules() // with the flag, this is the ref's fence
	if problem != nil {
		return nil, problem
	}
	if len(base) == 0 {
		return nil, nil
	}

	candidate, problem := r.worktreeFenceRules()
	if problem != nil {
		return nil, problem
	}

	loosenings := fence.Loosened(base, candidate)
	lines := make([]string, 0, len(loosenings))
	for _, loosening := range loosenings {
		lines = append(lines, looseningText(loosening))
	}
	return lines, nil
}

// briefChanged compares the brief of the ref with the worktree's. The comparison
// normalizes line endings and trailing whitespace: a Windows checkout is not a
// new vision.
func (r *runner) briefChanged(ref string) (bool, *Result) {
	base, found, err := r.git.FileAt(r.opts.Dir, ref, briefFile)
	if err != nil {
		problem := blocked(fmt.Sprintf("could not read %s from %q: %s; fix --fence-from or drop the flag to use the worktree's fence", briefFile, ref, oneLine(err.Error())))
		return false, &problem
	}
	if !found {
		// No brief at the ref means there was no policy to loosen.
		return true, nil
	}

	candidate, err := r.files.Read(filepath.Join(r.opts.Dir, briefFile))
	if err != nil {
		// A missing brief is the first gate's business, not the fence's: there is
		// nothing to compare a policy against when the vision is gone.
		return true, nil
	}
	return normalizeText(string(base)) != normalizeText(string(candidate)), nil
}

// normalizeText is the comparison the brief and the fence use: line endings and
// trailing whitespace do not change what the human decided.
func normalizeText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return strings.TrimRight(text, " \t\n")
}

// looseningText is one line per piece of policy the candidate dropped.
func looseningText(loosening fence.Loosening) string {
	switch loosening.Kind {
	case fence.KindRule:
		return loosening.Rule + ": the rule is gone"
	case fence.KindDep:
		return loosening.Rule + ": no longer forbids the dependency " + loosening.What
	default:
		return loosening.Rule + ": no longer forbids " + loosening.What
	}
}
