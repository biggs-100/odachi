package check

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/biggs-100/odachi/internal/domain/layout"
)

// byteOrderMark is what a console writes in front of UTF-8 when it is told to save
// "as UTF-8". The bytes are valid UTF-8, and the first line stops being the line
// the parser reads, so it is refused with its name (0274).
var byteOrderMark = []byte{0xEF, 0xBB, 0xBF}

// artifactsDecode reads the files this gate is about to interpret and refuses the
// ones that are not text. It is asked before the fence and the brief because
// everything after reads them: a record that does not decode is a record nobody can
// trust, and the gate would otherwise judge a tree it cannot see (0274).
func (r *runner) artifactsDecode() *Result {
	paths := []string{briefFile, fenceFile, layout.Profile}

	names, err := r.files.List(filepath.Join(r.opts.Dir, filepath.FromSlash(layout.Tasks)))
	if err != nil {
		problem := blocked(fmt.Sprintf("could not read %s: %s; check the permissions and run odachi check again", layout.Tasks, oneLine(err.Error())))
		return &problem
	}
	for _, name := range names {
		if strings.HasSuffix(name, ".md") {
			paths = append(paths, filepath.ToSlash(filepath.Join(layout.Tasks, name)))
		}
	}

	for _, rel := range paths {
		data, found, err := r.readArtifact(rel)
		if err != nil {
			problem := blocked(fmt.Sprintf("could not read %s: %s; check the permissions and run odachi check again", rel, oneLine(err.Error())))
			return &problem
		}
		if !found {
			continue
		}

		if bytes.HasPrefix(data, byteOrderMark) {
			problem := blocked(fmt.Sprintf("%s starts with a BOM: %s is text this tool reads line by line, so save it as UTF-8 without BOM and run odachi check again", rel, rel))
			return &problem
		}
		if offset, ok := firstInvalidByte(data); !ok {
			problem := blocked(fmt.Sprintf("%s is not UTF-8 at byte %d: save it as UTF-8 and run odachi check again", rel, offset))
			return &problem
		}
	}
	return nil
}

// readArtifact is one artifact, or the fact that it is not there: a project before
// its first feature has no task document, and a light feature has no brief by
// design (0166).
func (r *runner) readArtifact(rel string) ([]byte, bool, error) {
	path := filepath.Join(r.opts.Dir, filepath.FromSlash(rel))
	exists, err := r.files.Exists(path)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, nil
	}
	data, err := r.files.Read(path)
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

// firstInvalidByte names where the bytes stop being UTF-8, because "not UTF-8" with
// no place sends a human looking at a whole file for one byte (0274).
func firstInvalidByte(data []byte) (int, bool) {
	for offset := 0; offset < len(data); {
		r, size := utf8.DecodeRune(data[offset:])
		if r == utf8.RuneError && size == 1 {
			return offset, false
		}
		offset += size
	}
	return 0, true
}
