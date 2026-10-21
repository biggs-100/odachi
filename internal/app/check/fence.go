package check

import (
	"fmt"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/biggz/odachi/internal/domain/fence"
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
	path := filepath.Join(r.opts.Dir, fenceFile)

	exists, err := r.files.Exists(path)
	if err != nil {
		problem := blocked(fmt.Sprintf("no pude revisar %s: %s; revisá los permisos y volvé a correr odachi check", fenceFile, oneLine(err.Error())))
		return nil, &problem
	}
	if !exists {
		return nil, nil
	}

	data, err := r.files.Read(path)
	if err != nil {
		problem := blocked(fmt.Sprintf("no pude leer %s: %s; revisá los permisos y volvé a correr odachi check", fenceFile, oneLine(err.Error())))
		return nil, &problem
	}

	var document fenceDocument
	if _, err := toml.Decode(string(data), &document); err != nil {
		problem := blocked(fmt.Sprintf("no pude leer %s: %s; corregí el archivo (lo escribe odachi start) y volvé a correr odachi check", fenceFile, oneLine(err.Error())))
		return nil, &problem
	}

	rules := make([]fence.Rule, 0, len(document.Rules))
	for _, rule := range document.Rules {
		rules = append(rules, fence.Rule{Not: rule.Not, Paths: rule.Paths, Deps: rule.Deps})
	}
	return rules, nil
}
