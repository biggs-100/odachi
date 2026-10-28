package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/biggs-100/odachi/internal/adapters/agentstoml"
	"github.com/biggs-100/odachi/internal/adapters/fsx"
	"github.com/biggs-100/odachi/internal/adapters/gitx"
	"github.com/biggs-100/odachi/internal/adapters/hostmodel"
	"github.com/biggs-100/odachi/internal/adapters/profiletoml"
	"github.com/biggs-100/odachi/internal/adapters/reviewtoml"
	"github.com/biggs-100/odachi/internal/adapters/shellcmd"
	"github.com/biggs-100/odachi/internal/app/start"
	"github.com/biggs-100/odachi/internal/domain/layout"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/profile"
	"github.com/biggs-100/odachi/internal/domain/review"
	"github.com/biggs-100/odachi/internal/domain/store"
)

// osFiles is the Files adapter: the only place the verbs reach the disk. Writes
// are atomic and an existing file is backed up first.
//
// project and store are the two roots a backup needs: the artifact lives under
// project, the version under store. They are separate because the worktree
// travels with the repository and the store does not.
type osFiles struct {
	project string
	store   string
}

// storeRoot resolves the root of the clone's state, or the empty string when the
// folder is not a git repository. The verbs that only read evidence treat an
// empty root as "nothing on file"; the ones that write it refuse before writing.
func storeRoot(dir string) string {
	if root, err := store.Root(dir, gitRepo{}); err == nil {
		return root
	}
	return ""
}

// disk is the Files adapter for a verb that writes: the project root, and the
// clone's store when the folder is a git repository. A folder with no repository
// has no store, and a backup there stays beside the artifact - the one case
// where the old shape survives, because there is nowhere else to put it.
func disk(dir string) osFiles {
	return osFiles{project: dir, store: storeRoot(dir)}
}

func (f osFiles) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	switch {
	case err == nil:
		return true, nil
	case os.IsNotExist(err):
		return false, nil
	default:
		return false, err
	}
}

func (osFiles) IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	switch {
	case err == nil:
		return info.IsDir(), nil
	case os.IsNotExist(err):
		return false, nil
	default:
		return false, err
	}
}

func (osFiles) Read(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// List returns the entries of dir; a missing directory has no entries, which is
// exactly right for the first task document.
func (osFiles) List(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// Windows reports a file opened as a directory as "not found", so
			// "missing" and "it is a file" have to be told apart before saying
			// the directory is empty: an empty catalog and an unreadable one
			// are not the same thing.
			if info, statErr := os.Stat(dir); statErr == nil && !info.IsDir() {
				return nil, fmt.Errorf("%s is not a directory", dir)
			}
			return nil, nil
		}
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}

// Backup keeps a version before a rewrite. The version belongs to the clone, so
// it goes under the store's backups directory with the artifact's path
// preserved; a file outside the project keeps the backup beside it, because it
// has no project to belong to. With no store resolved — the binary always
// resolves one inside a repository, so this is a test or a folder with no git —
// the backup lands beside the file rather than in an invented location.
func (f osFiles) Backup(path string) (string, error) {
	if f.store == "" {
		return fsx.Backup(path)
	}
	return fsx.BackupIn(f.project, f.store, path)
}

func (osFiles) Remove(path string) error {
	return os.Remove(path)
}

// Rename moves a backup into the store without rewriting it: the file keeps its
// bytes and its mtime, so `undo` still knows which version is the newest.
func (osFiles) Rename(from, to string) error {
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return fmt.Errorf("could not create the directory of %s: %w", to, err)
	}
	return os.Rename(from, to)
}

// ModTime is what `odachi undo` chooses by: the date the filesystem recorded
// for a backup, so the newest one is the last thing odachi wrote and not a
// counter odachi would have to keep in sync with the disk.
func (osFiles) ModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// homeDir is where the Home adapter reads the human's home from. It is a
// variable so a test can point it at a directory it owns: `pack` writes outside
// the project, and no test may touch the real home of whoever runs the suite.
var homeDir = os.UserHomeDir

// osHome is the Home adapter for `odachi pack`: the one place the command asks
// where the human lives. The destination inside it comes from the domain table.
type osHome struct{}

func (osHome) Home() (string, error) {
	home, err := homeDir()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(home) == "" {
		return "", fmt.Errorf("the system reported no home")
	}
	return home, nil
}

// ensureHistory creates the version history odachi needs, because owning a project
// does not require knowing what git is. Refusing with "run git init" in front of
// somebody who has never heard of git is the opposite of this tool: creating the
// history is dirty work, and dirty work is the agent's job.
//
// It returns whether it created one, so the caller can say so in plain English and
// name git only in parentheses, next to what it is for.
func ensureHistory(dir string) (bool, error) {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		return false, nil
	}
	git, err := exec.LookPath("git")
	if err != nil {
		return false, errors.New("I need a program that keeps the project history and I could not find it: install git (git-scm.com) and run odachi init again")
	}
	cmd := exec.Command(git, "init", dir)
	cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("could not create the project history: %s", oneLine(err.Error()))
	}
	return true, nil
}

// hookExe is how the pre-commit hook will reach this binary: the name when it is
// on PATH, otherwise the absolute path of the running executable. A downloaded
// binary is the normal case for a beginner, and a hook that cannot find it
// blocks every commit without explaining why.
func hookExe() string {
	if _, err := exec.LookPath("odachi"); err == nil {
		return "odachi"
	}
	self, err := os.Executable()
	if err != nil {
		return "odachi"
	}
	return filepath.ToSlash(self)
}

// pathMissing reports whether a shell cannot resolve `odachi` by name. The rules
// block names that command, so an unreachable name leaves the agent unable to
// obey a single rule — which is what happened in a real run, where the agent
// spent seven shell calls looking for the binary.
func pathMissing() bool {
	_, err := exec.LookPath("odachi")
	return err != nil
}

// MakeExecutable is what the pre-commit hook needs: git only runs a hook that
// is executable, and WriteAtomic writes 0644 on purpose.
func (osFiles) MakeExecutable(path string) error {
	return os.Chmod(path, 0o755)
}

func (osFiles) Write(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("could not create the directory of %s: %w", path, err)
	}
	return fsx.WriteAtomic(path, data, 0o644)
}

// agentSelection is the store adapter for init's --agents and for the verbs that
// read the choice: the app owns the names, this type owns the TOML and the disk.
//
// The git port is a field so a test can drive both answers the adapter has to
// tell apart - "this folder has no repository" and "git failed" - without
// breaking the machine it runs on. The binary always leaves it nil, and nil means
// the real repository.
type agentSelection struct {
	git store.Git
}

// repository is the git port this adapter asks: the real one unless a test put
// another here.
func (a agentSelection) repository() store.Git {
	if a.git != nil {
		return a.git
	}
	return gitRepo{}
}

// Written returns the stored selection, or empty when the project never chose.
// The file lives in the clone's store: the choice belongs to the person, and a
// teammate clones the rules, not somebody else's clients.
func (a agentSelection) Written(dir string) (persona.Store, error) {
	path, _, err := a.Target(dir)
	switch {
	case errors.Is(err, gitx.ErrNoRepository):
		// No repository, no store: a project that never chose, which is both.
		return persona.Store{}, nil
	case err != nil:
		return persona.Store{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return persona.Store{}, nil
		}
		return persona.Store{}, err
	}
	return agentstoml.Decode(data)
}

// Target names the file the choice lives in, and the label a human reads. A
// folder with no repository has no store, and that is ErrNoRepository, not an
// empty answer: the caller that reads treats it as "the project never chose", and
// the caller that writes names the store it could not find.
func (a agentSelection) Target(dir string) (path, label string, err error) {
	root, err := store.Root(dir, a.repository())
	if err != nil {
		return "", "", err
	}
	path = filepath.Join(root, filepath.FromSlash(persona.SelectionPath))
	return path, storeLabel(dir, path), nil
}

// storeLabel names a file inside the store the way a human reads it: relative to
// the project when it can be, so the usual answer is `.git/odachi/agents.toml`,
// and absolute when the store lives outside the worktree.
func storeLabel(dir, path string) string {
	if out, err := filepath.Rel(dir, path); err == nil && out != ".." && !strings.HasPrefix(out, ".."+string(filepath.Separator)) {
		return filepath.ToSlash(out)
	}
	return filepath.ToSlash(path)
}

func (agentSelection) Render(store persona.Store) ([]byte, error) { return agentstoml.Render(store) }

// storedAgents is the choice a verb that refreshes the mirror reads. No file
// means both; a file the tool cannot read is named instead of guessed, because
// guessing which client the human chose is what the switch exists to prevent.
func storedAgents(dir, verb string) (string, int, string) {
	stored, err := agentSelection{}.Written(dir)
	if err != nil {
		return "", start.CodeBlocked, "could not read " + persona.SelectionPath + ": " + oneLine(err.Error()) +
			"; fix the file or run odachi init --agents both, and run " + verb + " again"
	}
	return stored.Agents, start.CodeOK, ""
}

// reviewModeCodec is the format adapter for `odachi review mode`: the app owns
// the actions and the disk port, this type owns TOML.
type reviewModeCodec struct{}

func (reviewModeCodec) Decode(data []byte) (review.ProjectConfig, error) {
	return reviewtoml.Decode(data)
}
func (reviewModeCodec) Render(config review.ProjectConfig) ([]byte, error) {
	return reviewtoml.Render(config)
}

// loadProfile reads odd/profile.toml when it exists. The human owns that
// file: it wins over inference, and a malformed one blocks with a sentence
// instead of being silently replaced. The verb is the command the human has to
// run again, so the message never sends them to the wrong one.
func loadProfile(dir, verb string) (*profile.Profile, int, string) {
	rel := filepath.FromSlash(layout.Profile)
	path := filepath.Join(dir, rel)

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, start.CodeOK, ""
		}
		return nil, start.CodeBlocked, "could not check " + rel + ": " + oneLine(err.Error()) + "; check the permissions and run " + verb + " again"
	}

	loaded, err := profiletoml.Load(path)
	if err != nil {
		return nil, start.CodeBlocked, "could not read " + rel + ": " + oneLine(err.Error()) + "; fix the file and run " + verb + " again"
	}
	if strings.TrimSpace(loaded.Name) == "" {
		return nil, start.CodeBlocked, fmt.Sprintf(
			`the profile %s does not say name; add name = "novato" (or "senior", "ci") and run %s again`, rel, verb)
	}
	// A name nobody shipped and that inherits from nothing resolves to a profile
	// with every field empty: it validates field by field and configures nothing.
	// That is the broken-ceremony case, and it is refused here because this is the
	// one place that knows which profiles exist.
	if _, shipped := profile.Builtin()[loaded.Name]; !shipped && strings.TrimSpace(loaded.Extends) == "" {
		return nil, start.CodeBlocked, fmt.Sprintf(
			`the profile %s says name = %q, which does not exist and inherits from nobody: use "novato", "senior" or "ci", or set extends to point at one of them, and run %s again`,
			rel, loaded.Name, verb)
	}

	// The human's file overrides the built-in profile of the same name, but a
	// field left out keeps the built-in default: a profile may change ceremony,
	// and "not mentioned" is not "blank".
	known := profile.Builtin()
	known[loaded.Name] = withDefaults(loaded, known)

	resolved, err := profile.Resolve(known, loaded.Name)
	if err != nil {
		return nil, start.CodeBlocked, "could not resolve the profile of " + rel + ": " + oneLine(err.Error()) + "; fix the file and run " + verb + " again"
	}
	return &resolved, start.CodeOK, ""
}

// writtenProfile answers init's one question: the profile the human wrote down,
// or none. It is not the profile in force — init must not infer one, because
// inferring a voice is guessing who the person is, and the tool's default
// audience is the person who does not know yet.
type writtenProfile struct{}

func (writtenProfile) Written(dir string) (string, error) {
	loaded, code, message := loadProfile(dir, "odachi init")
	switch {
	case code != start.CodeOK:
		return "", errors.New(message)
	case loaded == nil:
		return "", nil
	default:
		return loaded.Name, nil
	}
}

// currentProfile is the profile in force for a verb that does not ask anything:
// the human's file when there is one, otherwise what `odachi start` would have
// inferred from what the tool can see.
func currentProfile(dir, verb string, machine bool) (*profile.Profile, int, string) {
	loaded, code, message := loadProfile(dir, verb)
	if code != start.CodeOK || loaded != nil {
		return loaded, code, message
	}

	files := osFiles{}
	hasBrief, err := files.Exists(filepath.Join(dir, "brief.md"))
	if err != nil {
		return nil, start.CodeBlocked, "could not check brief.md: " + oneLine(err.Error()) + "; check the permissions and run " + verb + " again"
	}

	name := profile.Infer(profile.Facts{HasBrief: hasBrief, Machine: machine})
	resolved, err := profile.Resolve(profile.Builtin(), name)
	if err != nil {
		return nil, start.CodeBlocked, "could not resolve the " + name + " profile: " + oneLine(err.Error()) + "; run " + verb + " again"
	}
	return &resolved, start.CodeOK, ""
}

// withDefaults fills every field the human left empty from the built-in profile
// of the same name.
func withDefaults(loaded profile.Profile, known map[string]profile.Profile) profile.Profile {
	defaults, ok := known[loaded.Name]
	if !ok {
		return loaded
	}
	fill := func(field *string, fallback string) {
		if *field == "" {
			*field = fallback
		}
	}
	fill(&loaded.Extends, defaults.Extends)
	fill(&loaded.BriefSource, defaults.BriefSource)
	fill(&loaded.Questions, defaults.Questions)
	fill(&loaded.Assumptions, defaults.Assumptions)
	fill(&loaded.Output, defaults.Output)
	fill(&loaded.Proof, defaults.Proof)
	fill(&loaded.Explain, defaults.Explain)
	return loaded
}

// gitRepo is the Git adapter for `odachi check` and `odachi review`: every
// repository fact comes from gitx, which runs git locally and never touches the
// network. Both verbs read the same tree and the same diff, so they share one
// adapter and a review can never be produced for a candidate `check` did not see.
type gitRepo struct{}

func (gitRepo) TreeHash(root string) (string, error) { return gitx.TreeHash(root) }
func (gitRepo) Head(root string) (string, error)     { return gitx.Head(root) }

// CommonDir is where the clone's store hangs from: the directory every worktree
// of this clone shares, which is what makes the store one per clone.
func (gitRepo) CommonDir(root string) (string, error) { return gitx.CommonDir(root) }

// HeadTree is the tree the step's commit carries: the bytes plan records, because
// a rebase rewrites the commit and never the tree.
func (gitRepo) HeadTree(root string) (string, error) { return gitx.HeadTree(root) }

// WorktreeTree is the tree the worktree holds right now; plan compares it with
// HEAD so a step is only marked with bytes a commit can back.
func (gitRepo) WorktreeTree(root string) (string, error) { return gitx.TreeHash(root) }

// TreeOf, Has and CommitWithTree are how progress finds the commit a step landed
// in: a rebase rewrites the commit and keeps the tree, so the tree is the search.
func (gitRepo) TreeOf(root, commit string) (string, error) { return gitx.TreeOf(root, commit) }

func (gitRepo) Has(root, commit string) (bool, error) { return gitx.Has(root, commit) }

func (gitRepo) CommitWithTree(root, tree string) (string, error) {
	return gitx.CommitWithTree(root, tree)
}

// Resolve is the gate's question "which commit does this spelling name": a base
// is a commit, and two spellings of it are one diff.
func (gitRepo) Resolve(root, ref string) (string, error) { return gitx.Resolve(root, ref) }

// Candidate is the one read both verbs share: the tree, the changed paths and
// the diff of the same worktree, from one throwaway index.
func (gitRepo) Candidate(root, base string) (string, []string, string, error) {
	return gitx.Candidate(root, base)
}

func (gitRepo) Between(root, from, to string) ([]string, error) { return gitx.Between(root, from, to) }

func (gitRepo) AddedDeps(root, base string) ([]string, error) {
	return gitx.AddedDeps(root, base)
}

// FileAt reads a file from a git ref: the gate uses it to obey the fence of a
// base branch the candidate cannot edit.
func (gitRepo) FileAt(root, rev, path string) ([]byte, bool, error) {
	return gitx.FileAt(root, rev, path)
}

// BlobHash gives the review its content receipts: the blob id of the bytes each
// finding judged, computed by git from the worktree file.
func (gitRepo) BlobHash(root, path string) (string, error) { return gitx.BlobHash(root, path) }

// Ignored is how both verbs ask whether the project already keeps its own
// evidence out of the tree: `odachi init` writes that rule into .gitignore, and
// the lazy per-directory file is only the fallback for a project that never ran it.
func (gitRepo) Ignored(root, path string) (bool, error) { return gitx.Ignored(root, path) }

// osShell is the Shell adapter for `odachi check` and `odachi review`: the
// platform shell lives in internal/adapters/shellcmd, one file per platform, and
// refuses where there is no known shell.
type osShell struct{}

func (osShell) Run(dir, command string, timeoutSeconds int) (string, string, int, bool, error) {
	return shellcmd.Run(dir, command, timeoutSeconds)
}

// Quote is how this platform's shell spells one path, so a token the verb
// substitutes is quoted exactly like {prompt}.
func (osShell) Quote(path string) string { return shellcmd.Quote(path) }

// RunInput is the review's form: the reviewer receives the prompt the verb built.
func (osShell) RunInput(dir, command, input string, timeoutSeconds int) (string, string, int, bool, error) {
	return shellcmd.RunInput(dir, command, input, timeoutSeconds)
}

// hostModel is the Host adapter: the model of the session odachi runs in, read
// from the host's own contract and never from odachi's imagination.
type hostModel struct{}

func (hostModel) Model(dir string) string { return hostmodel.Model(dir) }
