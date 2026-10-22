package init

import (
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

// PersonaDest is one client's persona directory and what this run did there. It
// mirrors SkillDest: one row a human reads in a line and a machine reads in
// --json.
type PersonaDest struct {
	// Client is the name the skills table uses for the same program.
	Client string
	// Dir is the project-relative directory the client reads personas from.
	Dir string
	// Written counts the files this run wrote; Kept the ones already right.
	Written int
	Kept    int
	// Files is every persona path this run owns in that directory.
	Files []string
}

// writePersonas writes both personas into every client that has a native
// mechanism for one, and reports what it wrote.
//
// Both, always. `odachi init` cannot know who the human is, and choosing from the
// profile would push the decision back into the TOML this exists to get it out
// of: the client's own switch is the chooser, and the binary stays the only owner
// of the gates. A client odachi does not know gets no file, because there is
// nowhere verified to put it.
//
// The files are generated, so an identical one is not rewritten and a differing
// one is backed up before the rewrite: the human reads that the file is generated
// on its first line, and a hand edit is lost work either way — the backup is what
// keeps that honest.
func (r *runner) writePersonas() *Result {
	for _, client := range persona.Clients() {
		files, ok := persona.Files(client.Agent)
		if !ok {
			// Unreachable: the client came from the same table. Skipping instead
			// of writing a file to a directory nobody verified.
			continue
		}

		dest := PersonaDest{Client: client.Agent, Dir: client.Dir}
		var written, kept []string

		for _, file := range files {
			path := filepath.Join(r.dir, filepath.FromSlash(file.Path))
			name := filepath.Base(path)

			exists, err := r.files.Exists(path)
			if err != nil {
				return personaBlocked("no pude revisar "+file.Path, err)
			}
			if exists {
				current, err := r.files.Read(path)
				if err != nil {
					return personaBlocked("no pude leer "+file.Path, err)
				}
				if string(current) == file.Content {
					dest.Kept++
					kept = append(kept, name)
					dest.Files = append(dest.Files, file.Path)
					continue
				}
				if _, err := r.files.Backup(path); err != nil {
					return personaBlocked("no pude respaldar "+file.Path+" antes de reescribirlo", err)
				}
			}

			if err := r.files.Write(path, []byte(file.Content)); err != nil {
				return personaBlocked("no pude escribir "+file.Path, err)
			}
			dest.Written++
			written = append(written, name)
			dest.Files = append(dest.Files, file.Path)
			r.personaFiles = append(r.personaFiles, file.Path)
		}

		r.personaDests = append(r.personaDests, dest)
		r.personaLabels = append(r.personaLabels, client.Dir)
		r.lines = append(r.lines, personaLine(client, written, kept))
	}
	return nil
}

// personaBlocked is the persona half's refusal: the file, the reason the adapter
// gave, and the command that retries the run.
func personaBlocked(what string, err error) *Result {
	result := blocked(what + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
	return &result
}

// personaLine is one client in one line: what arrived, what was already there, and
// how the human selects the persona in that client. The last part is the point of
// the whole job — a file nobody knows how to pick is a file nobody uses.
func personaLine(client persona.Client, written, kept []string) string {
	parts := make([]string, 0, 2)
	if len(written) > 0 {
		parts = append(parts, "escribí "+humanList(written))
	}
	if len(kept) > 0 {
		parts = append(parts, humanList(kept)+" sin cambios")
	}
	line := client.Dir + ": " + strings.Join(parts, ", ")
	if len(written) > 0 {
		line += "; " + client.Switch
	}
	return line
}
