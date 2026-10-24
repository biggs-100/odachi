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
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/profile"
)

// osFiles is the Files adapter for `odachi start`: the only place that command
// reaches the disk. Writes are atomic and an existing file is backed up first.
type osFiles struct{}

func (osFiles) Exists(path string) (bool, error) {
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

// Backup keeps a version before a rewrite. When path belongs to a project, the
// version goes to that project's .odachi/backups store; a file outside a
// project keeps the backup beside it, because it has no store to join.
func (osFiles) Backup(path string) (string, error) {
	if root, ok := projectRoot(path); ok {
		return fsx.BackupIn(root, path)
	}
	return fsx.Backup(path)
}

// projectRoot finds the nearest ancestor that owns a .odachi directory: the
// same boundary the rest of the verbs use. It starts at the file's directory,
// so a file inside .odachi walks up to the project root and not to itself.
func projectRoot(path string) (string, bool) {
	dir := filepath.Dir(path)
	for {
		info, err := os.Stat(filepath.Join(dir, ".odachi"))
		if err == nil && info.IsDir() {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
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
type agentSelection struct{}

// Written returns the stored selection, or empty when the project never chose.
func (agentSelection) Written(dir string) (string, error) {
	path := filepath.Join(dir, filepath.FromSlash(persona.SelectionPath))
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return agentstoml.Decode(data)
}

func (agentSelection) Render(selection string) ([]byte, error) { return agentstoml.Render(selection) }

// storedAgents is the choice a verb that refreshes the mirror reads. No file
// means both; a file the tool cannot read is named instead of guessed, because
// guessing which client the human chose is what the switch exists to prevent.
func storedAgents(dir, verb string) (string, int, string) {
	selection, err := agentSelection{}.Written(dir)
	if err != nil {
		return "", start.CodeBlocked, "could not read " + persona.SelectionPath + ": " + oneLine(err.Error()) +
			"; fix the file or run odachi init --agents both, and run " + verb + " again"
	}
	return selection, start.CodeOK, ""
}

// reviewModeCodec is the format adapter for `odachi review mode`: the app owns
// the actions and the disk port, this type owns TOML.
type reviewModeCodec struct{}

func (reviewModeCodec) Decode(data []byte) (string, error) { return reviewtoml.Decode(data) }
func (reviewModeCodec) Render(mode string) ([]byte, error) { return reviewtoml.Render(mode) }

// loadProfile reads .odachi/profile.toml when it exists. The human owns that
// file: it wins over inference, and a malformed one blocks with a sentence
// instead of being silently replaced. The verb is the command the human has to
// run again, so the message never sends them to the wrong one.
func loadProfile(dir, verb string) (*profile.Profile, int, string) {
	rel := filepath.Join(".odachi", "profile.toml")
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

// Candidate is the one read both verbs share: the tree, the changed paths and
// the diff of the same worktree, from one throwaway index.
func (gitRepo) Candidate(root, base string) (string, []string, string, error) {
	return gitx.Candidate(root, base)
}

func (gitRepo) AddedDeps(root, base string) ([]string, error) {
	return gitx.AddedDeps(root, base)
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

// RunInput is the review's form: the reviewer receives the prompt the verb built.
func (osShell) RunInput(dir, command, input string, timeoutSeconds int) (string, string, int, bool, error) {
	return shellcmd.RunInput(dir, command, input, timeoutSeconds)
}

// hostModel is the Host adapter: the model of the session odachi runs in, read
// from the host's own contract and never from odachi's imagination.
type hostModel struct{}

func (hostModel) Model(dir string) string { return hostmodel.Model(dir) }
