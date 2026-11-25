// Package changecontext resolves the one change range that check and review
// judge. It reads the trailer and its fork point before either verb snapshots
// the candidate, so the fence, diff, proof context and review cannot disagree
// about where the change began.
package changecontext

import (
	"fmt"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/change"
)

// Git is the repository evidence a change range needs. It contains no candidate
// read: resolving the range must finish before tree, paths or diff are requested.
type Git interface {
	Messages(root, base string) ([]string, error)
	MergeBase(root, ref string) (string, error)
	OnDefaultBranch(root, base string) (branch string, on bool, err error)
}

// Context is the resolved range. Base is the ref the candidate snapshot uses,
// ChangeBase is the ref whose commit messages were read, and Line is the trailer
// those messages declared. An empty ChangeBase keeps the trailer lane closed.
type Context struct {
	Base       string
	ChangeBase string
	Line       string
}

// Resolve reads a claimed change and its merge-base. Without a trailer, Base is
// unchanged: ordinary per-tree and documented-feature behavior stays as it was.
// A trailer moves Base to the exact fork point before the caller reads a tree.
func Resolve(git Git, root, base, changeBase string) (Context, error) {
	context := Context{Base: strings.TrimSpace(base), ChangeBase: strings.TrimSpace(changeBase)}
	// HEAD..HEAD is empty in a born repository and unreadable in an unborn one.
	// It is the per-tree base, not a trailer range; leaving the lane closed keeps
	// the existing empty-tree fallback available before the first commit.
	if context.ChangeBase == "" || context.ChangeBase == "HEAD" {
		context.ChangeBase = ""
		return context, nil
	}

	messages, err := git.Messages(root, context.ChangeBase)
	if err != nil {
		return Context{}, fmt.Errorf("could not read the commits since %s: %w; fetch that ref or pass --base <ref>", context.ChangeBase, err)
	}
	declared, ok, err := change.Parse(messages)
	if err != nil {
		return Context{}, err
	}
	if !ok {
		return context, nil
	}

	branch, onDefault, err := git.OnDefaultBranch(root, context.ChangeBase)
	if err != nil {
		return Context{}, fmt.Errorf("could not tell whether this change sits on the default branch: %w; fix the repository and run again", err)
	}
	if onDefault {
		return Context{}, fmt.Errorf("a change that names itself in a trailer lives on a branch, not on %s: move the commits to a branch and run again", branch)
	}

	start, err := git.MergeBase(root, context.ChangeBase)
	if err != nil {
		return Context{}, fmt.Errorf("could not find the merge-base with %s: %w; choose another range with --base <ref>", context.ChangeBase, err)
	}
	if start = strings.TrimSpace(start); start == "" {
		return Context{}, fmt.Errorf("could not find the merge-base with %s: git did not name a commit; choose another range with --base <ref>", context.ChangeBase)
	}

	context.Base = start
	context.Line = declared.Line
	return context, nil
}
