package check

import (
	"fmt"
	"strings"
)

// mojibake are the sequences a UTF-8 file leaves behind when a console reads it as
// CP850 or CP1252 and writes the result back: the classic double encoding. They are
// valid UTF-8, so no decoder can see them, and no Spanish or English sentence uses
// them by accident (0274).
var mojibake = []string{"Ã", "Â", "â€"}

// briefSaysWhatItMeans refuses a brief or a TOML artifact whose text carries the
// signature of a double encoding. The task documents are left out on purpose: they are
// prose, and a document may have to name these very sequences to explain the defect.
func (r *runner) briefSaysWhatItMeans() *Result {
	for _, rel := range []string{briefFile, fenceFile, profileFileLayout()} {
		data, found, err := r.readArtifact(rel)
		if err != nil {
			problem := blocked("could not read " + rel + ": " + oneLine(err.Error()) + "; check the permissions and run odachi check again")
			return &problem
		}
		if !found {
			continue
		}
		for _, marker := range mojibake {
			if strings.Contains(string(data), marker) {
				problem := blocked(fmt.Sprintf("%s carries %q: the text was read as one encoding and saved as another; save it as UTF-8 read as UTF-8 and run odachi check again", rel, marker))
				return &problem
			}
		}
	}
	return nil
}

// profileFileLayout is where the profile lives, named once for this file.
func profileFileLayout() string { return "odd/profile.toml" }
