// Package budget pins the bytes of the text an agent reads, so a surface nobody
// looks at cannot grow in silence.
//
// The block odachi writes into AGENTS.md, the four lens prompts and the skills
// index are paid for on every turn a model takes; a ceiling nobody measures is a
// number that permits anything, which is why the pin is exact and not a maximum.
// One list, two directions, the same shape the refusal ratchet uses: a text that
// grows fails until the pin is raised on purpose, and one that shrinks fails until
// the pin follows it down -a pin that only grows is a number that lies upwards
// (0295).
package budget

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Pin is one artifact and the bytes it is frozen at.
type Pin struct {
	// Artifact names the surface the way a person reads it: the block's voice,
	// the lens file, the index.
	Artifact string
	// Bytes is the measurement the pin accepts, in UTF-8 bytes.
	Bytes int
}

// Read reads a baseline file: `artifact<TAB>bytes` rows, and `#` comments that say
// why the file exists so nobody deletes it as noise.
func Read(path string) ([]Pin, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	pins := make([]Pin, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("%s: %q is not an artifact and a count separated by a tab", path, line)
		}
		count, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("%s: %q does not end in a count", path, line)
		}
		pins = append(pins, Pin{Artifact: parts[0], Bytes: count})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	sort.Slice(pins, func(i, j int) bool { return pins[i].Artifact < pins[j].Artifact })
	return pins, nil
}

// Issues compares what was measured with what is pinned, in both directions, and
// words each disagreement as the act it asks for. Silence means the pin still
// describes the text.
//
// A measurement with no pin is a surface nobody watches, and a pin with no
// measurement is a number about a text that moved or disappeared: both are named,
// because deleting either side is how the list starts describing nothing.
func Issues(measured, pinned []Pin) []string {
	frozen := make(map[string]int, len(pinned))
	for _, pin := range pinned {
		frozen[pin.Artifact] = pin.Bytes
	}

	out := make([]string, 0)
	seen := make(map[string]bool, len(measured))
	for _, m := range measured {
		seen[m.Artifact] = true
		want, known := frozen[m.Artifact]
		switch {
		case !known:
			out = append(out, fmt.Sprintf("%s is %d bytes and no pin names it: add it on purpose", m.Artifact, m.Bytes))
		case m.Bytes > want:
			out = append(out, fmt.Sprintf("%s is %d bytes and the pin says %d: it grew by %d bytes; cut the text, or raise the pin on purpose",
				m.Artifact, m.Bytes, want, m.Bytes-want))
		case m.Bytes < want:
			out = append(out, fmt.Sprintf("%s is %d bytes and the pin says %d: it shrank by %d; lower the pin so the number follows the text",
				m.Artifact, m.Bytes, want, want-m.Bytes))
		}
	}
	for _, pin := range pinned {
		if !seen[pin.Artifact] {
			out = append(out, fmt.Sprintf("the pin names %s and nothing measured it: the surface is gone or renamed, and the pin should go with it", pin.Artifact))
		}
	}
	sort.Strings(out)
	return out
}
