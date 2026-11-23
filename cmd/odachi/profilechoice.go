package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/biggs-100/odachi/internal/adapters/profiletoml"
	"github.com/biggs-100/odachi/internal/app/start"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/profile"
	"github.com/biggs-100/odachi/internal/domain/rules"
)

// globalProfileRel is the machine's default, outside every repository. A
// project's odd/profile.toml still wins: a default must never rewrite a choice
// somebody committed in another folder (0357).
const globalProfileRel = ".odachi/profile.toml"

func globalProfilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "", fmt.Errorf("could not find the home folder for %s: check the environment and run the command again", globalProfileRel)
	}
	return filepath.Join(home, filepath.FromSlash(globalProfileRel)), nil
}

// loadGlobalProfile answers the name the human chose for this machine, if any.
func loadGlobalProfile() (string, bool, error) {
	path, err := globalProfilePath()
	if err != nil {
		return "", false, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("global profile: could not read %s: %w", path, err)
	}
	loaded, err := profiletoml.Decode(data)
	if err != nil {
		return "", false, fmt.Errorf("global profile: %w", err)
	}
	if _, ok := profile.Builtin()[loaded.Name]; !ok {
		return "", false, fmt.Errorf("global profile: %q is not a profile this binary ships", loaded.Name)
	}
	return loaded.Name, true, nil
}

// saveGlobalProfile writes the machine's default. It touches no repository.
func saveGlobalProfile(name string) error {
	if _, ok := profile.Builtin()[name]; !ok {
		return fmt.Errorf("%q is not a profile this binary ships", name)
	}
	path, err := globalProfilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("global profile: could not create %s: %w", filepath.Dir(path), err)
	}
	body := "# odachi: the default profile for this machine; odd/profile.toml wins in a project\n" +
		"name = " + strconv.Quote(name) + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return fmt.Errorf("global profile: could not write %s: %w", path, err)
	}
	return nil
}

// profileNameForVoice maps the text in AGENTS.md back to a shipped profile by
// comparing the one canonical source, the voice, instead of keeping a second
// table of names.
func profileNameForVoice(text string) string {
	for name := range profile.Builtin() {
		voice, err := persona.Voice(name)
		if err == nil && strings.TrimSpace(voice) == strings.TrimSpace(text) {
			return name
		}
	}
	return ""
}

// resolveProfile is the one owner of the answer "which profile is in force".
// Order: the project's file, then a voice already published in AGENTS.md, then
// the machine's default, then the old inference. The bool says whether the
// project itself chose; doctor uses it to decide whether to compare the block.
func resolveProfile(dir, verb string, machine bool) (*profile.Profile, bool, error) {
	loaded, code, message := loadProfile(dir, verb)
	if code != start.CodeOK {
		return nil, false, errors.New(message)
	}
	if loaded != nil {
		return loaded, true, nil
	}

	if content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md")); err == nil {
		if voice, has := rules.VoiceIn(string(content)); has {
			if name := profileNameForVoice(voice); name != "" {
				resolved, err := profile.Resolve(profile.Builtin(), name)
				return &resolved, true, err
			}
		}
	}

	if name, found, err := loadGlobalProfile(); err != nil {
		return nil, false, err
	} else if found {
		resolved, err := profile.Resolve(profile.Builtin(), name)
		return &resolved, false, err
	}

	exists, err := os.Stat(filepath.Join(dir, "brief.md"))
	hasBrief := err == nil && !exists.IsDir()
	if err != nil && !os.IsNotExist(err) {
		return nil, false, fmt.Errorf("could not check brief.md: %w", err)
	}
	name := profile.Infer(profile.Facts{HasBrief: hasBrief, Machine: machine})
	resolved, err := profile.Resolve(profile.Builtin(), name)
	return &resolved, false, err
}
