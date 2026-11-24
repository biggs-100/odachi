package archtest_test

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/skill"
	"github.com/biggs-100/odachi/skills"
)

// contractsBaseline pins the content of every text odachi hands a model. A byte
// count catches a text that grew; a hash catches one rewritten to the same size.
const contractsBaseline = "contracts.baseline"

// TestTheContractBytesArePinned pins every model-facing text by content, not only
// by size. A rule can be rewritten to the same number of bytes, and a size pin
// calls that no change at all. The sweep runs both ways: a surface with no pin and
// a pin with no surface are the same unreviewed edit as a hash that differs.
func TestTheContractBytesArePinned(t *testing.T) {
	measured, err := contractHashes()
	if err != nil {
		t.Fatalf("measuring the contract bytes: %v", err)
	}
	pinned, err := readContractBaseline(contractsBaseline)
	if err != nil {
		t.Fatalf("reading %s: %v", contractsBaseline, err)
	}

	for _, issue := range contractIssues(measured, pinned) {
		t.Error(issue)
	}
}

// contractHashes hashes each named surface and the embedded skill payload. The
// payload is not a text read at the session's start, but it is bytes odachi hands
// to a model: the digest names the exact embedded payload, which a size pin at
// the session's start cannot see.
func contractHashes() (map[string]string, error) {
	named, err := surfaces()
	if err != nil {
		return nil, err
	}

	out := make(map[string]string, len(named)+1)
	for _, surface := range named {
		sum := sha256.Sum256(surface.Bytes)
		out[surface.Artifact] = hex.EncodeToString(sum[:])
	}

	payload, err := skills.Digest()
	if err != nil {
		return nil, err
	}
	out["skills:payload"] = payload
	return out, nil
}

// contractPin is one artifact and the hex digest it is frozen at.
type contractPin struct {
	Artifact string
	Hash     string
}

// readContractBaseline reads artifact<TAB>hex-sha256 rows. Comment lines say why
// the file exists and carry no pin.
func readContractBaseline(path string) ([]contractPin, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	pins := make([]contractPin, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("%s: %q is not an artifact and a hash separated by a tab", path, line)
		}
		pins = append(pins, contractPin{Artifact: parts[0], Hash: strings.TrimSpace(parts[1])})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return pins, nil
}

// contractIssues compares measured hashes with the pins in both directions. Each
// issue names the artifact and the act that fixes it: review the diff, then raise
// the pin on purpose, so a reviewer sees the change instead of a green suite.
func contractIssues(measured map[string]string, pinned []contractPin) []string {
	frozen := make(map[string]string, len(pinned))
	for _, pin := range pinned {
		frozen[pin.Artifact] = pin.Hash
	}

	issues := make([]string, 0)
	for artifact, hash := range measured {
		want, known := frozen[artifact]
		switch {
		case !known:
			issues = append(issues, fmt.Sprintf("%s hashes to %s and no pin names it in %s: review the diff with git diff, then add its row on purpose",
				artifact, hash, contractsBaseline))
		case hash != want:
			issues = append(issues, fmt.Sprintf("%s hashes to %s but the pin in %s says %s: review the diff with git diff, then raise the pin on purpose",
				artifact, hash, contractsBaseline, want))
		}
	}
	for _, pin := range pinned {
		if _, ok := measured[pin.Artifact]; !ok {
			issues = append(issues, fmt.Sprintf("the pin in %s names %s and nothing measured it: review the diff with git diff, then remove or update its row on purpose",
				contractsBaseline, pin.Artifact))
		}
	}
	sort.Strings(issues)
	return issues
}

// TestTheContractPinNamesBothDirections: a changed hash is one edit seen from two
// sides. The message has to name the artifact and the fix for each, or the reader
// has to open the baseline to find out what moved.
func TestTheContractPinNamesBothDirections(t *testing.T) {
	measured := map[string]string{
		"block:novato": strings.Repeat("a", 64),
		"block:senior": strings.Repeat("b", 64),
	}
	pinned := []contractPin{
		{Artifact: "block:senior", Hash: strings.Repeat("c", 64)},
		{Artifact: "artifact:ghost", Hash: strings.Repeat("d", 64)},
	}

	joined := strings.Join(contractIssues(measured, pinned), "\n")
	for _, phrase := range []string{
		"block:novato", "no pin names it",
		"block:senior", "raise the pin",
		"artifact:ghost", "nothing measured it",
		"git diff",
	} {
		if !strings.Contains(joined, phrase) {
			t.Errorf("contractIssues does not say %q:\n%s", phrase, joined)
		}
	}
}

// regenerateRegistry reads the skills on disk and renders SKILLS.md the way
// `odachi skills --write` does: parse each SKILL.md, then let the domain render
// the table. It starts from the source files, never from the committed output.
func regenerateRegistry(t *testing.T) string {
	t.Helper()
	dir := filepath.Join("..", "..", "skills")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read skills/: %v", err)
	}

	parsed := make([]skill.Skill, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name(), "SKILL.md")
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			// A folder without SKILL.md is not a skill; app/skills.load skips it too.
			continue
		}
		if err != nil {
			t.Fatalf("read skills/%s/SKILL.md: %v", entry.Name(), err)
		}
		rel := "skills/" + entry.Name() + "/SKILL.md"
		parsed = append(parsed, skill.Parse(entry.Name(), rel, string(data)))
	}
	return skill.Registry(parsed)
}

// readCommittedRegistry reads the registry as it lives in the tree.
func readCommittedRegistry(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", skill.RegistryFile))
	if err != nil {
		t.Fatalf("read %s: %v", skill.RegistryFile, err)
	}
	return string(data)
}

// TestTheRegistryRegeneratesByteIdentical: SKILLS.md is generated, so a hand edit
// is a drift from the skills it indexes. Regenerate from skills/*/SKILL.md and
// demand the exact committed bytes, or the index lies about what ships.
func TestTheRegistryRegeneratesByteIdentical(t *testing.T) {
	committed := readCommittedRegistry(t)
	generated := regenerateRegistry(t)
	if generated != committed {
		t.Errorf("%s is not byte-identical to what `odachi skills --write` regenerates from skills/*/SKILL.md; run odachi skills --write:\nwant (regenerated):\n%s\ngot (committed):\n%s",
			skill.RegistryFile, generated, committed)
	}
}
