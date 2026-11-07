# the-gate-names-what-the-change-adds

## What must exist
- when the feature closes, `check` names what the diff adds —the new files and the dependency files— in one line of the ready output, so the reviewer sees the scaffolding before reading the code; the exit code and the verdict do not change

## Done when
- the ready output names the new files (capped) and the dependency files the diff adds, and stays silent when there are none
- the line never changes the exit code, the reason or the proof
- a test seeds a diff with an added file and a `go.mod` change and asserts the line; another asserts silence on a diff with neither
- `odachi check --run "go test ./..."` is green
- the changelog's Unreleased section names the line
- `go test ./...` is green

## Not included (first version)
- no new verb and no new flag
- no change to the fence, the proof or the verdict reasons
- no parsing of abstractions: they are not machine-detectable; the line names files and dependency files only
- no change to the review or to the doctor
- no change to the skills or to the voice
- no change to the stored proofs: never touched
- no change to `odachi role` or to the skill it renders
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow

## Decisions taken
- the line lives in the ready output, not in the hook's `Note`: the reviewer reads it at the close, not at every commit
- new files come from the diff already read (`--- /dev/null` then `+++ b/<path>`), not from a second git call
- dependency files are a named list of manifests and lockfiles; the list is data, easy to extend
- it is a note, never a block: the gate cannot know the requirement, only report what was added
