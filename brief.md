# the-harness-says-what-it-costs

## What must exist
- the harness says what it costs: `doctor` reports the bytes of the always-on block and the bytes of `SKILLS.md` beside the installed skill count it already names, so the context a project pays on every turn is a number and not an estimate

## Done when
- the doctor's skills line names the bytes of the odachi block in `AGENTS.md` and the bytes of `SKILLS.md`, in the detail it already prints, in KB with one decimal or B under a KB
- it still names the registry state, the installed count and the drift it named before
- a block or a registry that cannot be read keeps the check's own problem instead of a made-up zero: a size nobody could read is not a small one
- `--json` carries the same detail, which it already mirrors from the line
- a test in a configured project asserts both sizes; the JSON test and the drift tests stay green
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- the changelog's Unreleased section names the measurement

## Not included (first version)
- no sixteenth check and no change to the "fifteen" the usage, the summary and the JSON name
- no change to what the check verifies: it is a measurement, not a gate
- no change to the block, its markers or the voice
- no change to the registry's shape or to the skills' text
- no change to the review or to the classifier
- no change to the stored proofs: never touched
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow or to the installers
- no change to the brief's format or to `start`

## Decisions taken
- the cost goes into the skills line, not a sixteenth check: the line already names the count and the registry, and a new check would change the number the usage, the summary and four tests name for one measurement
- the block is measured as the odachi block inside `AGENTS.md`, not the whole file: that is what `init` writes and what every agent reads on every turn
- a failed read is the check's problem, never 0 B: the doctor already refuses to turn an unreadable file into a clean answer
- the size is formatted once, with the same units a human reads in `clean`: bytes under a kilobyte, kilobytes with one decimal above it
