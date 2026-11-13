// Package refusal is the ratchet over the sentences a command refuses with.
//
// A refusal that names no way out is a dead end for the agent that reads it: it
// knows what went wrong and not what to run. The ratchet freezes the ones a package
// has today in a baseline file and stops the next one from landing, which is the
// difference between debt somebody reads and debt nobody can add.
//
// Two directions, and both fail:
//
//   - a sentence that names no way out and is not in the baseline is new debt;
//   - a sentence in the baseline that the walk no longer finds is paid debt, and the
//     file has to follow it down or the number stops meaning anything.
//
// What it cannot do, and the package says so out loud: it reads sentences, not
// intentions. A refusal that points at a command which runs and resolves nothing
// passes. The walk also cannot judge a message built from a value, so those are
// counted per file and the count may not grow either.
package refusal

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
)

// wayOut is the shape of a continuation a person can type: an odachi verb or a
// flag. A sentence that names neither is a dead end.
var wayOut = regexp.MustCompile(`(?i)odachi\s+[a-z-]+|--[a-z][a-z-]+`)

// Found is what one walk found in a directory.
type Found struct {
	// Sentences are the refusals that name no way out, as "file<TAB>sentence",
	// sorted, one row per call site.
	Sentences []string
	// Unreadable counts, per file, the calls whose message is a value no walk can
	// follow.
	Unreadable map[string]int
}

// Baseline is the frozen debt: the sentences the walk did not judge kind, and the
// counts of what it could not read.
type Baseline struct {
	Sentences []string
	Counts    map[string]int
}

// Walk reads a package's production sources and collects every refusal built for
// `blocked(...)`: the sentence when it can be read statically, a count when it
// cannot. Test files are skipped: a fixture that builds a refusal on purpose is not
// a refusal a human reads.
func Walk(dir string) (Found, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return Found{}, err
	}
	found := Found{Sentences: make([]string, 0), Unreadable: make(map[string]int)}
	fset := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			return Found{}, err
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
				found.Unreadable[name]++
			case !wayOut.MatchString(sentence):
				found.Sentences = append(found.Sentences, name+"\t"+sentence)
			}
			return true
		})
	}
	slices.Sort(found.Sentences)
	return found, nil
}

// ReadBaseline reads the frozen list: `sentence<TAB>file<TAB>text` rows and
// `count<TAB>file<TAB>n` rows. A comment says why the file exists, so nobody
// deletes it as noise.
func ReadBaseline(path string) (Baseline, error) {
	file, err := os.Open(path)
	if err != nil {
		return Baseline{}, err
	}
	defer func() { _ = file.Close() }()

	baseline := Baseline{Sentences: make([]string, 0), Counts: make(map[string]int)}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			return Baseline{}, fmt.Errorf("%s: %q is not three tab-separated fields", path, line)
		}
		switch parts[0] {
		case "sentence":
			baseline.Sentences = append(baseline.Sentences, parts[1]+"\t"+strings.TrimSpace(parts[2]))
		case "count":
			n, err := strconv.Atoi(parts[2])
			if err != nil {
				return Baseline{}, fmt.Errorf("%s: %q does not end in a count", path, line)
			}
			baseline.Counts[parts[1]] = n
		default:
			return Baseline{}, fmt.Errorf("%s: %q starts with an unknown kind; use sentence or count", path, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return Baseline{}, err
	}
	slices.Sort(baseline.Sentences)
	return baseline, nil
}

// Grew is what the walk found and the baseline does not freeze: a refusal that
// arrived without a way out.
func Grew(found, baseline []string) []string {
	out := make([]string, 0)
	for _, row := range found {
		if !slices.Contains(baseline, row) {
			out = append(out, row)
		}
	}
	return out
}

// Shrunk is what the baseline froze and the walk no longer sees: the debt paid, and
// the file must follow it down.
func Shrunk(found, baseline []string) []string {
	out := make([]string, 0)
	for _, row := range baseline {
		if !slices.Contains(found, row) {
			out = append(out, row)
		}
	}
	return out
}

// staticSentence is the sentence a refusal carries when it can be read statically:
// a literal, a concatenation, or the format string of a Sprintf/Errorf. The second
// value says the sentence is only part of the message.
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
