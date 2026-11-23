// Package evidence implements `odachi evidence export`: a close's evidence
// leaves the clone as one portable JSON bundle, keyed by the tree it proves.
//
// The three parts -the verdict, the verification and the proof- live in three
// different places and two of them are outside the worktree. A bundle gathers
// the raw bytes of whichever parts exist, hashes each one, and names the tree
// they all answer for. It carries no timestamp and writes no path other than
// the output the human chose, so exporting the same tree twice gives the same
// bytes and cannot surprise a commit.
package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/proof"
	"github.com/biggs-100/odachi/internal/domain/store"
	verifydomain "github.com/biggs-100/odachi/internal/domain/verify"
)

// Exit codes, the same shape as the other verbs.
const (
	CodeOK      = 0
	CodeBlocked = 1
)

// Schema is the version of the bundle this binary writes.
const Schema = 1

// proofDir is where check --run leaves the proof of a tree, inside the project.
const proofDir = "proof"

// Options is one `odachi evidence export` invocation. Store is the clone's
// store root, resolved through store.Root by the caller; Run resolves it from
// Dir and Git when the caller left it empty.
type Options struct {
	Dir   string
	Store string
	Tree  string
	Out   string
}

// Files is the disk port. Read and Exists find the parts and guard the output;
// List finds the proof; MkdirAll creates only the output's parent and Write
// puts the bundle there. The app never asks for another directory.
type Files interface {
	Read(path string) ([]byte, error)
	Exists(path string) (bool, error)
	List(dir string) ([]string, error)
	MkdirAll(dir string) error
	Write(path string, data []byte) error
}

// Result is what the CLI prints and exits with. Lines carries the one-line
// summary so the verb can print it the same way it prints every other result.
type Result struct {
	Code    int
	Message string
	Lines   []string
}

// hashes holds the digest of each part that exists. A nil pointer is omitted,
// so an absent part leaves no key instead of a null that claims a hash.
type hashes struct {
	Verdict      *string `json:"verdict,omitempty"`
	Verification *string `json:"verification,omitempty"`
	Proof        *string `json:"proof,omitempty"`
}

// bundle is the canonical file. Pointer strings emit null for a missing part:
// an empty string would say the part was there and held nothing. Struct fields
// fix the key order, and there is no clock, so two exports of one tree are the
// same bytes.
type bundle struct {
	Schema       int     `json:"schema"`
	Tree         string  `json:"tree"`
	Verdict      *string `json:"verdict"`
	Verification *string `json:"verification"`
	Proof        *string `json:"proof"`
	SHA256       hashes  `json:"sha256"`
}

// Run performs one export.
func Run(opts Options, files Files, git store.Git) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}
	if err := validateTree(opts.Tree); err != nil {
		return blocked(err.Error())
	}
	if strings.TrimSpace(opts.Out) == "" {
		return blocked("evidence export needs --out <path>: the bundle must have a place to go")
	}

	root := opts.Store
	if root == "" {
		resolved, err := store.Root(opts.Dir, git)
		if err != nil {
			return blocked(oneLine(err.Error()) + "; odachi evidence export needs the clone's store")
		}
		root = resolved
	}

	// Guard before reading: an existing output is never overwritten, even by
	// the same bytes, and even when there is nothing to export.
	exists, err := files.Exists(opts.Out)
	if err != nil {
		return blocked(fmt.Sprintf("could not check whether %s exists: %s", opts.Out, oneLine(err.Error())))
	}
	if exists {
		return blocked(fmt.Sprintf("refusing to overwrite %s: the file already exists; choose another --out", opts.Out))
	}

	verdict, err := readPart(files, filepath.Join(root, store.Reviews, opts.Tree+".json"))
	if err != nil {
		return blocked(err.Error())
	}
	verification, err := readPart(files, filepath.Join(root, verifydomain.Dir, opts.Tree+".json"))
	if err != nil {
		return blocked(err.Error())
	}
	proofText, err := findProof(files, opts.Dir, opts.Tree)
	if err != nil {
		return blocked(err.Error())
	}
	if verdict == nil && verification == nil && proofText == nil {
		return blocked(fmt.Sprintf("tree %s has no evidence to export: no verdict, verification or proof found", opts.Tree))
	}

	data, err := render(opts.Tree, verdict, verification, proofText)
	if err != nil {
		return blocked(err.Error())
	}
	if err := files.MkdirAll(filepath.Dir(opts.Out)); err != nil {
		return blocked(fmt.Sprintf("could not create the directory of %s: %s", opts.Out, oneLine(err.Error())))
	}
	if err := files.Write(opts.Out, data); err != nil {
		return blocked(fmt.Sprintf("could not write %s: %s", opts.Out, oneLine(err.Error())))
	}

	message := "exported the evidence of tree " + opts.Tree + " to " + opts.Out
	return Result{Code: CodeOK, Message: message, Lines: []string{message}}
}

// validateTree pins the two shapes git gives a tree: the 40-character SHA-1
// and the 64-character SHA-256. Anything else is a name, not a tree.
func validateTree(tree string) error {
	if len(tree) != 40 && len(tree) != 64 {
		return fmt.Errorf("the tree %q is not a hash: expected 40 or 64 hexadecimal characters", tree)
	}
	for _, r := range tree {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'f', r >= 'A' && r <= 'F':
			continue
		default:
			return fmt.Errorf("the tree %q is not a hash: expected 40 or 64 hexadecimal characters", tree)
		}
	}
	return nil
}

// readPart returns nil for a part that is not there. A part that is there but
// cannot be read is a refusal: guessing it was absent would export a bundle
// that lies about the tree.
func readPart(files Files, path string) ([]byte, error) {
	data, err := files.Read(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %s", filepath.ToSlash(path), oneLine(err.Error()))
	}
	return data, nil
}

// findProof scans proof/*.txt for the exact "# tree <hash>" header. A file that
// names another tree -or starts with "# failed"- is not this tree's proof.
func findProof(files Files, dir, tree string) ([]byte, error) {
	names, err := files.List(filepath.Join(dir, proofDir))
	if err != nil {
		return nil, fmt.Errorf("could not list %s: %s", proofDir, oneLine(err.Error()))
	}
	header := proof.TreeHeader + tree
	for _, name := range names {
		if !strings.HasSuffix(name, ".txt") {
			continue
		}
		path := filepath.Join(dir, proofDir, name)
		data, err := files.Read(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("could not read %s: %s", filepath.ToSlash(path), oneLine(err.Error()))
		}
		first, _, _ := strings.Cut(string(data), "\n")
		if strings.TrimSuffix(first, "\r") == header {
			return data, nil
		}
	}
	return nil, nil
}

// render builds the canonical bytes. Each pointer is set only when its part
// exists, and each hash is over that part's raw bytes -not over the escaped
// string the JSON file carries.
func render(tree string, verdict, verification, proofText []byte) ([]byte, error) {
	out := bundle{Schema: Schema, Tree: tree}
	if verdict != nil {
		out.Verdict = new(string(verdict))
		out.SHA256.Verdict = new(hash(verdict))
	}
	if verification != nil {
		out.Verification = new(string(verification))
		out.SHA256.Verification = new(hash(verification))
	}
	if proofText != nil {
		out.Proof = new(string(proofText))
		out.SHA256.Proof = new(hash(proofText))
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		// Unreachable: every field is a string, an int or a struct of strings.
		return nil, errors.New("could not encode the evidence bundle")
	}
	return append(data, '\n'), nil
}

func hash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func blocked(message string) Result {
	return Result{Code: CodeBlocked, Message: message, Lines: []string{message}}
}

func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
