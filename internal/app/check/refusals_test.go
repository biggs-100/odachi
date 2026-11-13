package check_test

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// baselineFile is the frozen list of refusals this package had the day the ratchet
// was written. It only shrinks: a refusal that names no way out and is not here
// fails the suite, and one that disappears is reported so the number goes down.
const baselineFile = "refusals.baseline"

// wayOut is the shape of a continuation a person can type: an odachi verb or a
// flag. A sentence that names neither is a dead end for the agent reading it.
var wayOut = regexp.MustCompile(`(?i)odachi\s+[a-z-]+|--[a-z][a-z-]+`)

// TestEveryRefusalNamesAWayOutOrIsInTheBaseline is the ratchet over this package's
// refusals. A refusal is one message built for `blocked(...)`: every one either
// names something the reader can run -a command, a flag- or is frozen in the
// baseline with the rest of today's debt.
//
// The walk reads the sentences it can see statically. A call whose message is a
// value is counted instead, because a sentence nobody can read is a sentence nobody
// can judge; that count may not grow either (0287).
//
// What this does not prove: that the named continuation works. A sentence that
// points at a command which runs and does nothing passes here.
func TestEveryRefusalNamesAWayOutOrIsInTheBaseline(t *testing.T) {
	found, unreadable, err := walkRefusals(".")
	if err != nil {
		t.Fatalf("walking this package: %v", err)
	}
	baseline, frozen, err := readBaseline(baselineFile)
	if err != nil {
		t.Fatalf("reading %s: %v", baselineFile, err)
	}

	for _, row := range found {
		if !slices.Contains(baseline, row) {
			t.Errorf("a refusal names no way out and is not in %s: %s\nname the command in its sentence, or add it to the baseline on purpose", baselineFile, row)
		}
	}
	for _, row := range baseline {
		if !slices.Contains(found, row) {
			t.Logf("this refusal is gone; lower %s: %s", baselineFile, row)
		}
	}

	for file, count := range unreadable {
		want, ok := frozen[file]
		switch {
		case !ok:
			t.Errorf("%s has %d refusal(s) the walk cannot read and no count in %s", file, count, baselineFile)
		case count > want:
			t.Errorf("%s has %d refusal(s) the walk cannot read and %s froze %d: a message built from a value cannot be judged, and it may not grow", file, count, baselineFile, want)
		case count < want:
			t.Logf("%s now has %d refusal(s) the walk cannot read; lower %s", file, count, baselineFile)
		}
	}
	for file := range frozen {
		if _, ok := unreadable[file]; !ok {
			t.Logf("%s no longer builds a refusal from a value; remove its count from %s", file, baselineFile)
		}
	}
}

// walkRefusals reads the package's production sources and returns the sentences
// that name no way out, as "file<TAB>sentence" rows, plus how many calls per file
// build their message from a value this walk cannot follow.
func walkRefusals(dir string) ([]string, map[string]int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}
	rows := make([]string, 0)
	unreadable := make(map[string]int)
	fset := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			return nil, nil, err
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			fun, ok := call.Fun.(*ast.Ident)
			if !ok || fun.Name != "blocked" {
				return true
			}
			sentence, _ := staticSentence(call.Args[0])
			sentence = strings.TrimSpace(sentence)
			switch {
			case sentence == "":
				unreadable[name]++
			case !wayOut.MatchString(sentence):
				rows = append(rows, name+"\t"+sentence)
			}
			return true
		})
	}
	slices.Sort(rows)
	return rows, unreadable, nil
}

// staticSentence is the sentence a refusal carries when it can be read statically:
// a literal, a concatenation, or the format string of a Sprintf/Errorf. The second
// value says the sentence is only part of the message, which matters to nobody here
// and to a reader who wonders why the text looks cut.
func staticSentence(expr ast.Expr) (string, bool) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind != token.STRING {
			return "", true
		}
		s, err := strconv.Unquote(e.Value)
		if err != nil {
			return "", true
		}
		return s, false
	case *ast.BinaryExpr:
		if e.Op != token.ADD {
			return "", true
		}
		left, leftCut := staticSentence(e.X)
		right, rightCut := staticSentence(e.Y)
		return left + right, leftCut || rightCut
	case *ast.ParenExpr:
		return staticSentence(e.X)
	case *ast.CallExpr:
		sel, ok := e.Fun.(*ast.SelectorExpr)
		if !ok || len(e.Args) == 0 || (sel.Sel.Name != "Sprintf" && sel.Sel.Name != "Errorf") {
			return "", true
		}
		return staticSentence(e.Args[0])
	default:
		return "", true
	}
}

// readBaseline reads the frozen list: `sentence<TAB>file<TAB>text` rows and
// `count<TAB>file<TAB>n` rows. A comment says why the file exists, so nobody
// deletes it as noise.
func readBaseline(path string) ([]string, map[string]int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = file.Close() }()

	rows := make([]string, 0)
	counts := make(map[string]int)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			return nil, nil, fmt.Errorf("%s: %q is not three tab-separated fields", path, line)
		}
		switch parts[0] {
		case "sentence":
			rows = append(rows, parts[1]+"\t"+strings.TrimSpace(parts[2]))
		case "count":
			n, err := strconv.Atoi(parts[2])
			if err != nil {
				return nil, nil, fmt.Errorf("%s: %q does not end in a count", path, line)
			}
			counts[parts[1]] = n
		default:
			return nil, nil, fmt.Errorf("%s: %q starts with an unknown kind; use sentence or count", path, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	slices.Sort(rows)
	return rows, counts, nil
}
