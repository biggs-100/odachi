package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/adapters/fsx"
	"github.com/biggs-100/odachi/internal/adapters/gitx"
	"github.com/biggs-100/odachi/internal/adapters/profiletoml"
	"github.com/biggs-100/odachi/internal/adapters/shellcmd"
	"github.com/biggs-100/odachi/internal/app/start"
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

func (osFiles) Backup(path string) (string, error) {
	return fsx.Backup(path)
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

// MakeExecutable is what the pre-commit hook needs: git only runs a hook that
// is executable, and WriteAtomic writes 0644 on purpose.
func (osFiles) MakeExecutable(path string) error {
	return os.Chmod(path, 0o755)
}

func (osFiles) Write(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("no pude crear el directorio de %s: %w", path, err)
	}
	return fsx.WriteAtomic(path, data, 0o644)
}

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
		return nil, start.CodeBlocked, "no pude revisar " + rel + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr " + verb
	}

	loaded, err := profiletoml.Load(path)
	if err != nil {
		return nil, start.CodeBlocked, "no pude leer " + rel + ": " + oneLine(err.Error()) + "; corregí el archivo y volvé a correr " + verb
	}
	if strings.TrimSpace(loaded.Name) == "" {
		return nil, start.CodeBlocked, fmt.Sprintf(
			`el perfil %s no dice name; agregá name = "novato" (o "senior", "ci") y volvé a correr %s`, rel, verb)
	}

	// The human's file overrides the built-in profile of the same name, but a
	// field left out keeps the built-in default: a profile may change ceremony,
	// and "not mentioned" is not "blank".
	known := profile.Builtin()
	known[loaded.Name] = withDefaults(loaded, known)

	resolved, err := profile.Resolve(known, loaded.Name)
	if err != nil {
		return nil, start.CodeBlocked, "no pude resolver el perfil de " + rel + ": " + oneLine(err.Error()) + "; corregí el archivo y volvé a correr " + verb
	}
	return &resolved, start.CodeOK, ""
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
		return nil, start.CodeBlocked, "no pude revisar brief.md: " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr " + verb
	}
	hasRepo, err := files.Exists(filepath.Join(dir, ".git"))
	if err != nil {
		return nil, start.CodeBlocked, "no pude revisar el repositorio: " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr " + verb
	}

	name := profile.Infer(profile.Facts{HasBrief: hasBrief, HasRepo: hasRepo, Machine: machine})
	resolved, err := profile.Resolve(profile.Builtin(), name)
	if err != nil {
		return nil, start.CodeBlocked, "no pude resolver el perfil " + name + ": " + oneLine(err.Error()) + "; volvé a correr " + verb
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

// gitRepo is the Git adapter for `odachi check`: every repository fact comes
// from gitx, which runs git locally and never touches the network.
type gitRepo struct{}

func (gitRepo) TreeHash(root string) (string, error) { return gitx.TreeHash(root) }

func (gitRepo) ChangedPaths(root, base string) ([]string, error) {
	return gitx.ChangedPaths(root, base)
}

func (gitRepo) AddedDeps(root, base string) ([]string, error) {
	return gitx.AddedDeps(root, base)
}

// osShell is the Shell adapter for `odachi check`: the platform shell lives in
// internal/adapters/shellcmd, one file per platform, and refuses where there is
// no known shell.
type osShell struct{}

func (osShell) Run(dir, command string, timeoutSeconds int) (string, int, bool, error) {
	return shellcmd.Run(dir, command, timeoutSeconds)
}
