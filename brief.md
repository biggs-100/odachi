# the-declared-risk-meets-the-diff

## What must exist
- a full-tier change cannot pass as harmless in silence: when the review classifies the change as full -it touches a path the classifier names- and no commit declares a risk of its own, `check` names the contrast in one line of the ready output and names the paths, so the reviewer reads what the change claims against what it touches; a named risk stays silent, and the exit code and the verdict never move

## Done when
- the ready output names the contrast on a full-tier change whose commits say `Risk: none`, and on one whose commits carry no `Risk:` line at all; the line names the paths the classifier saw, capped
- a named risk, a focused change and a passive change keep the ready output silent about risk
- the line never changes the exit code, the reason or the proof
- a test seeds a full path and commits with `Risk: none` and asserts the line; another seeds the same path and `Risk: data` and asserts silence; another seeds a focused path and asserts silence
- the parser's tests cover `none`, a named risk, prose after the name, a missing line and an empty line refused
- `odachi check --run "go test ./..."` is green
- the changelog's Unreleased section names the line
- `go test ./...` is green

## Not included (first version)
- no change to the review or to the doctor
- no change to the classifier: `internal/domain/review` is read, not edited
- no new verb and no new flag: the line rides the ready output of `check`
- no change to the exit code or to the verdict reasons: it is a note, never a block
- no change to the stored proofs: never touched
- no change to the skills or to the voice
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow

## Decisions taken
- the note names the contrast, never a block: the gate measures paths and reads words, and it cannot call a risk; a full change with a named risk is the reviewer's question, not the gate's
- a missing `Risk:` line counts as `none`: silence and "none" are the same claim of harmlessness, and treating an absent line as a declaration would leave the failure the feature exists to see
- the tier comes from `Classify(paths)`, the same function the gate already asks whether a change owes a review, so the note and the gate cannot disagree and the note needs no diff and no new port method
- the note fires on the feature's own range -the base in the task doc, or the merge-base in the document-less lane- because a note about paths must measure the diff the review will judge
