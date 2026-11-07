# the-harness-says-what-it-costs
Base: db3599e59006723fb23df6cd03db4d08d58f8229

## Requirements (verbatim, from the user)
R1 · "the harness says what it costs: `doctor` reports the bytes of the always-on block and the bytes of `SKILLS.md` beside the installed skill count it already names, so the context a project pays on every turn is a number and not an estimate"

## Done when
- the doctor's skills line names the bytes of the odachi block in `AGENTS.md` and the bytes of `SKILLS.md`, in the detail it already prints, in KB with one decimal or B under a KB
- it still names the registry state, the installed count and the drift it named before
- a block or a registry that cannot be read keeps the check's own problem instead of a made-up zero: a size nobody could read is not a small one
- `--json` carries the same detail, which it already mirrors from the line
- a test in a configured project asserts both sizes; the JSON test and the drift tests stay green
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- the changelog's Unreleased section names the measurement

## Steps
- [ ] T1 · the doctor names the block and the registry bytes   R1 · route:inline
- [ ] T2 · the changelog names the measurement   R1 · route:inline

## History
L1 · "the harness says what it costs: `doctor` reports the bytes of the always-on block and the bytes of `SKILLS.md` beside the installed skill count it already names, so the context a project pays on every turn is a number and not an estimate"
