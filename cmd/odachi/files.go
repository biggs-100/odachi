package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/biggz/odachi/internal/adapters/fsx"
	"github.com/biggz/odachi/internal/adapters/profiletoml"
	"github.com/biggz/odachi/internal/app/start"
	"github.com/biggz/odachi/internal/domain/profile"
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

func (osFiles) Write(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("no pude crear el directorio de %s: %w", path, err)
	}
	return fsx.WriteAtomic(path, data, 0o644)
}

// loadProfile reads .odachi/profile.toml when it exists. The human owns that
// file: it wins over inference, and a malformed one blocks with a sentence
// instead of being silently replaced.
func loadProfile(dir string) (*profile.Profile, int, string) {
	rel := filepath.Join(".odachi", "profile.toml")
	path := filepath.Join(dir, rel)

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, start.CodeOK, ""
		}
		return nil, start.CodeBlocked, "no pude revisar " + rel + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi start"
	}

	loaded, err := profiletoml.Load(path)
	if err != nil {
		return nil, start.CodeBlocked, "no pude leer " + rel + ": " + oneLine(err.Error()) + "; corregí el archivo y volvé a correr odachi start"
	}
	if strings.TrimSpace(loaded.Name) == "" {
		return nil, start.CodeBlocked, fmt.Sprintf(
			`el perfil %s no dice name; agregá name = "novato" (o "senior", "ci") y volvé a correr odachi start`, rel)
	}

	// The human's file overrides the built-in profile of the same name, but a
	// field left out keeps the built-in default: a profile may change ceremony,
	// and "not mentioned" is not "blank".
	known := profile.Builtin()
	known[loaded.Name] = withDefaults(loaded, known)

	resolved, err := profile.Resolve(known, loaded.Name)
	if err != nil {
		return nil, start.CodeBlocked, "no pude resolver el perfil de " + rel + ": " + oneLine(err.Error()) + "; corregí el archivo y volvé a correr odachi start"
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
