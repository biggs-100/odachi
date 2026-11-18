package skilleval

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
)

// RecordSchema is the shape this binary writes. A record from another schema is
// read by the binary that wrote it, and an older one says so instead of guessing.
const RecordSchema = 1

// Record is what one exam leaves in the clone: which skill, which text, and what
// each draw showed. It is filed under the identity of what was measured -the
// skill's files and the exam's task- so the next run can say "this was already
// paid for" instead of paying again, and a skill that changes answers a new name.
type Record struct {
	Schema   int    `json:"schema"`
	Skill    string `json:"skill"`
	Identity string `json:"identity"`
	At       string `json:"at"`
	Draws    []Draw `json:"draws"`
}

// Draw is one paid answer, as the record keeps it: the arm, the run number, and
// the behaviours with the event that decided each one.
type Draw struct {
	Arm        string      `json:"arm"`
	Run        int         `json:"run"`
	Loaded     bool        `json:"loaded"`
	Behaviours []Behaviour `json:"behaviours"`
	Problem    string      `json:"problem,omitempty"`
}

// Identity answers the hash of what an exam measures: the skill's files, by path,
// and the task the same files are asked about. Two runs over the same text are the
// same measurement, whatever folder they were built in.
func Identity(skill string, files map[string][]byte, task string) string {
	sum := sha256.New()
	_, _ = fmt.Fprintf(sum, "skill:%s\n", skill)
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		_, _ = fmt.Fprintf(sum, "%s:%x\n", path, sha256.Sum256(files[path]))
	}
	_, _ = fmt.Fprintf(sum, "task:%x\n", sha256.Sum256([]byte(task)))
	return hex.EncodeToString(sum.Sum(nil))[:16]
}

// Path is where a record lives: the store's evals folder, named by the skill and
// the identity.
func Path(root, skill, identity string) string {
	return filepath.Join(root, "evals", skill+"-"+identity+".json")
}

// ParseRecord reads a record the tool wrote. A file that is not one is a refusal,
// not an empty measurement: the run that finds a broken record must not guess
// what it said.
func ParseRecord(data []byte) (Record, error) {
	var record Record
	if err := json.Unmarshal(data, &record); err != nil {
		return Record{}, fmt.Errorf("the record is not JSON: %w", err)
	}
	if record.Skill == "" || record.Identity == "" {
		return Record{}, fmt.Errorf("the record names no skill or no identity: %+v", record)
	}
	return record, nil
}

// RenderRecord writes the record the way ParseRecord reads it.
func RenderRecord(record Record) ([]byte, error) {
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("the record cannot be written: %w", err)
	}
	return append(data, '\n'), nil
}

// Summarize is the one line a reader greps: the skill, the identity, and what the
// draws met, out of what they could.
func (r Record) Summarize() string {
	met, of := 0, 0
	for _, draw := range r.Draws {
		for _, behaviour := range draw.Behaviours {
			of++
			if behaviour.Met {
				met++
			}
		}
	}
	return fmt.Sprintf("%s %s: %d of %d behaviours across %d draw(s)", r.Skill, r.Identity, met, of, len(r.Draws))
}
