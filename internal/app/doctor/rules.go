package doctor

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/adapters/fencetoml"
	"github.com/biggs-100/odachi/internal/adapters/reviewtoml"
	"github.com/biggs-100/odachi/internal/adapters/rulebooktoml"
	domaindoctor "github.com/biggs-100/odachi/internal/domain/doctor"
	"github.com/biggs-100/odachi/internal/domain/layout"
	"github.com/biggs-100/odachi/internal/domain/rulebook"
)

// rulesLedger is check 16. Three answers on one line: the rules nothing enforces
// are debt a human can read, a row that names a mechanism the project does not
// have is a problem, and a table that is not there is neither - a project that
// keeps no table has no debt to count.
//
// A row is resolved against something that exists on disk, because a row that
// names a fence path, a check or a test nothing answers is worse than no row at
// all: the project would believe it is protected (0312).
func (r *runner) rulesLedger() domaindoctor.Check {
	content, exists, err := r.read(layout.Rules)
	if err != nil {
		return r.unreadable(domaindoctor.CheckRules, layout.Rules, err)
	}
	if !exists {
		return domaindoctor.OK(domaindoctor.CheckRules, "no "+layout.Rules+": a project that keeps no table has no debt to count")
	}

	rules, err := rulebooktoml.Decode([]byte(content))
	if err != nil {
		return domaindoctor.Broken(domaindoctor.CheckRules, oneLine(err.Error()),
			"fix the row the message names, or delete "+layout.Rules+" when this project keeps no table")
	}

	fences, problem := r.fencedNames()
	if problem != nil {
		return *problem
	}
	checks, problem := r.declaredChecks()
	if problem != nil {
		return *problem
	}
	tests, err := r.testNames()
	if err != nil {
		return domaindoctor.Broken(domaindoctor.CheckRules,
			"could not read the suite's test names: "+oneLine(err.Error()),
			"check the permissions of internal and cmd, then run odachi doctor again")
	}

	var unresolved []string
	for _, rule := range rules {
		kind, name := rulebook.Parse(rule.Enforces)
		var found bool
		var mechanism, source string
		switch kind {
		case rulebook.KindFence:
			found, mechanism, source = fences[name], "fence", layout.Fence
		case rulebook.KindCheck:
			found, mechanism, source = checks[name], "check", layout.Review
		case rulebook.KindTest:
			found, mechanism, source = tests[name], "test", "the suite"
		default:
			continue
		}
		if found {
			continue
		}
		unresolved = append(unresolved, fmt.Sprintf("%s names %s:%s, which %s does not have", oneLine(rule.Statement), mechanism, name, source))
	}

	if len(unresolved) > 0 {
		return domaindoctor.Broken(domaindoctor.CheckRules,
			fmt.Sprintf("%s resolve to nothing: %s", count(len(unresolved), "enforcer", "enforcers"), strings.Join(capped(unresolved, 3), "; ")),
			"name a mechanism the project has, or write enforces = \"none\" and its because")
	}

	debt := rulebook.Debt(rules)
	if len(debt) > 0 {
		return domaindoctor.Warning(domaindoctor.CheckRules,
			fmt.Sprintf("%s without an enforcer: %s", count(len(debt), "rule", "rules"), rulebook.DebtLine(debt, 4)))
	}
	return domaindoctor.OK(domaindoctor.CheckRules, fmt.Sprintf("%s, every one enforced", count(len(rules), "rule", "rules")))
}

// fencedNames is what the fence already keeps out: the paths and the dependencies
// a rule forbids. A malformed fence is this check's problem too, because a table
// that points at a policy nobody can read is a row nobody can resolve.
func (r *runner) fencedNames() (map[string]bool, *domaindoctor.Check) {
	names := map[string]bool{}
	content, exists, err := r.read(layout.Fence)
	if err != nil {
		problem := r.unreadable(domaindoctor.CheckRules, layout.Fence, err)
		return names, &problem
	}
	if !exists {
		return names, nil
	}
	rules, err := fencetoml.Decode([]byte(content))
	if err != nil {
		problem := domaindoctor.Broken(domaindoctor.CheckRules,
			"could not read "+layout.Fence+": "+oneLine(err.Error()),
			"fix the fence, or delete it when the project keeps none: odachi check names the same problem")
		return names, &problem
	}
	for _, rule := range rules {
		for _, one := range rule.Paths {
			names[one] = true
		}
		for _, one := range rule.Deps {
			names[one] = true
		}
	}
	return names, nil
}

// declaredChecks is what the review runs before any lens. A table that points at
// a project with no review.toml has no check to resolve.
func (r *runner) declaredChecks() (map[string]bool, *domaindoctor.Check) {
	names := map[string]bool{}
	content, exists, err := r.read(layout.Review)
	if err != nil {
		problem := r.unreadable(domaindoctor.CheckRules, layout.Review, err)
		return names, &problem
	}
	if !exists {
		return names, nil
	}
	config, err := reviewtoml.Decode([]byte(content))
	if err != nil {
		problem := domaindoctor.Broken(domaindoctor.CheckRules,
			"could not read "+layout.Review+": "+oneLine(err.Error()),
			"fix the review table, or delete it when the project keeps none: odachi review names the same problem")
		return names, &problem
	}
	for _, one := range config.Checks {
		names[one.Name] = true
	}
	return names, nil
}

// testNames is the set of Test functions the suite runs, read through the Files
// port. It walks the Go tree only, because that is where `go test ./...` looks,
// and it skips a folder it cannot read instead of failing the check for one
// directory: the name it cannot confirm is reported, which is the honest answer.
func (r *runner) testNames() (map[string]bool, error) {
	names := map[string]bool{}
	for _, root := range []string{"internal", "cmd"} {
		found, err := r.isDir(root)
		if err != nil || !found {
			continue
		}
		if err := r.walkTests(root, names); err != nil {
			return names, err
		}
	}
	return names, nil
}

// walkTests reads one folder of the Go tree, recursing into its subfolders.
func (r *runner) walkTests(rel string, names map[string]bool) error {
	entries, err := r.files.List(filepath.Join(r.opts.Dir, filepath.FromSlash(rel)))
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		child := path.Join(rel, entry)
		dir, err := r.isDir(child)
		if err != nil {
			continue
		}
		if dir {
			if strings.HasPrefix(entry, ".") || entry == "testdata" || entry == "node_modules" {
				continue
			}
			if err := r.walkTests(child, names); err != nil {
				return err
			}
			continue
		}
		if !strings.HasSuffix(entry, "_test.go") {
			continue
		}
		data, err := r.files.Read(filepath.Join(r.opts.Dir, filepath.FromSlash(child)))
		if err != nil {
			continue
		}
		for _, name := range testFuncNames(string(data)) {
			names[name] = true
		}
	}
	return nil
}

// testFuncNames reads the names Go's test runner answers to: a top-level
// `func TestXxx(` is a test, and nothing else in the file is.
func testFuncNames(source string) []string {
	var names []string
	for _, line := range strings.Split(source, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "func Test") {
			continue
		}
		name, _, ok := strings.Cut(strings.TrimPrefix(line, "func "), "(")
		if ok {
			names = append(names, name)
		}
	}
	return names
}

// capped keeps a doctor line one line: the first n items and how many it hid.
func capped(items []string, n int) []string {
	if len(items) <= n {
		return items
	}
	return append(append([]string{}, items[:n]...), fmt.Sprintf("and %d more", len(items)-n))
}
