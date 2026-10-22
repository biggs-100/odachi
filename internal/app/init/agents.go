package init

import (
	"path/filepath"

	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/rules"
)

// agentsFile is the file the block lives in. The renderer owns the name, the
// markers and the text; this package only decides when to write them.
const agentsFile = rules.File

// profileFile is the human's profile, relative to the project root. init reads it
// to know which voice the block gets and never writes it.
const profileFile = ".odachi/profile.toml"

// writeAgents writes the rules block with the voice of the profile at hand,
// replacing it in place when it is already there and backing up the human's file
// before touching it.
func (r *runner) writeAgents() *Result {
	voice := r.voice()

	path := filepath.Join(r.dir, agentsFile)

	exists, err := r.files.Exists(path)
	if err != nil {
		problem := blocked("no pude revisar " + agentsFile + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
		return &problem
	}

	current := ""
	if exists {
		data, err := r.files.Read(path)
		if err != nil {
			problem := blocked("no pude leer " + agentsFile + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
			return &problem
		}
		current = string(data)
	}

	updated, err := rules.WithBlock(current, voice)
	if err != nil {
		problem := blocked("no pude armar el bloque de " + agentsFile + ": " + oneLine(err.Error()) + "; volvé a correr odachi init")
		return &problem
	}

	if exists && updated == current {
		r.lines = append(r.lines, agentsFile+": sin cambios")
		return nil
	}

	if exists {
		backup, err := r.files.Backup(path)
		if err != nil {
			problem := blocked("no pude respaldar " + agentsFile + " antes de reescribirlo: " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
			return &problem
		}
		r.lines = append(r.lines, agentsFile+": bloque actualizado con la voz "+voice+" (el anterior quedó en "+relative(r.dir, backup)+")")
	} else {
		r.lines = append(r.lines, agentsFile+": creado con el bloque de odachi y la voz "+voice)
	}

	if err := r.files.Write(path, []byte(updated)); err != nil {
		problem := blocked("no pude escribir " + agentsFile + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
		return &problem
	}
	r.changed = append(r.changed, agentsFile)
	return nil
}

// voice is the voice the block gets: the one the profile at hand asks for, or
// novato when there is no profile file yet. When it had to fall back it leaves
// the sentence to say so in r.voiceNote, which the caller reports on every run:
// a broken profile is worth saying again after a rerun that changed nothing.
//
// It never infers the profile. init runs before the interview, so "there is a
// repository here" would make every new project senior — exactly backwards. The
// tool's default audience is the person who does not know yet, and the only thing
// that can say otherwise is a profile the human already accepted.
//
// A profile that does not resolve, or names something odachi has no voice for,
// does not fail the run: the block, the hook and the skills are still what the
// project needs. It falls back to novato and says in one line what to fix.
func (r *runner) voice() string {
	if r.profiles == nil {
		return persona.Novato
	}

	written, err := r.profiles.Written(r.dir)
	if err != nil {
		r.voiceNote = "la voz es la de novato porque " + profileFile + " no resuelve (" + oneLine(err.Error()) + "); corregí el archivo o borralo y volvé a correr odachi init"
		return persona.Novato
	}
	if written == "" {
		return persona.Novato
	}

	resolved, err := persona.VoiceName(written)
	if err != nil {
		r.voiceNote = "la voz es la de novato porque el perfil no es uno de los que tienen voz (" + oneLine(err.Error()) + "); usá novato, senior o ci en " + profileFile + " y volvé a correr odachi init"
		return persona.Novato
	}
	return resolved
}

// reportLeftovers names the per-client persona files v0.10.0 wrote, when they are
// still there. They are not an error: the voice lives in the block now, nothing
// reads them, and odachi never deletes a file it did not write in this version.
// The human may, and the line says so.
func (r *runner) reportLeftovers() {
	found := make([]string, 0, len(persona.Deprecated()))
	for _, file := range persona.Deprecated() {
		exists, err := r.files.Exists(filepath.Join(r.dir, filepath.FromSlash(file)))
		if err != nil {
			r.lines = append(r.lines, "no pude revisar "+file+": "+oneLine(err.Error())+"; revisá los permisos y volvé a correr odachi init")
			continue
		}
		if exists {
			found = append(found, file)
		}
	}
	if len(found) == 0 {
		return
	}

	r.lines = append(r.lines, "quedaron de v0.10.0 "+humanList(persona.DeprecatedDirs())+" (las personas por cliente): ya no se usan, la voz vive en "+
		agentsFile+"; no los toqué: borralos si querés")
}
