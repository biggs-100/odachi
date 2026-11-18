// Package scope carries the one configuration that turns "read only this tree"
// from a sentence in a prompt into a rule the client enforces.
//
// The prompt still says it, because a reviewer that knows the boundary does not
// waste a call probing it. But a model answers to its tools, not to its
// instructions: the field run of 0323 watched a reviewer told to read one
// directory glob the whole temp folder and cite a file it was never given. The
// rule below is the boundary, and it is why the run's working directory has a
// parent odachi owns.
package scope

// FileName is the configuration the opencode client discovers when it starts: it
// reads it from the location it runs in and from every ancestor, so a file in
// the parent of the reviewer's directory governs the whole run.
const FileName = "opencode.json"

// Render is the configuration. Every path outside the location is refused before
// a tool touches it, which is what stops `read`, `glob` and `grep` together: the
// last two take a pattern and a query instead of a path, so no per-path rule can
// bind them, but the directory a path resolves to is checked for all three.
//
// The bytes are fixed and pinned by a test: a project whose boundary drifts is a
// project that measures what a model felt like doing.
func Render() []byte {
	return []byte(`{
  "$schema": "https://opencode.ai/config.json",
  "permissions": [
    { "action": "external_directory", "resource": "*", "effect": "deny" }
  ]
}
`)
}
