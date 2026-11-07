# a-release-says-which-bytes-were-judged

## What must exist
- a release says which bytes were judged: `docs/releases.md` states that the tag carries the verdict of the feature's closing tree and that the release commit moves only documents, and `check` names the verdict's tree at the close next to the proof, so a reader stops believing the tag itself was reviewed

## Done when
- `docs/releases.md` states the binding: the tag carries the verdict of the feature's closing tree, and the release commit moves only documents
- `check`'s ready output at a close names the verdict's tree next to the proof when a verdict exists, and says nothing when none does
- a test seeds a verdict and asserts the line; another asserts the silence without one; the exit code stays ready in both
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- the changelog's Unreleased section names the line

## Not included (first version)
- no change to the release workflow or to the tag: the binding is a statement and a line, not a new gate
- no change to the review or to the verdict's shape
- no change to the exit code or to the verdict reasons
- no change to the brief's format or to `start`
- no change to the doctor or to the classifier
- no change to the stored proofs: never touched
- no change to the skills or to the voice
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the installers

## Decisions taken
- the line names the tree only: the findings and the coverage are the review's to say, and a line that repeated them would be a second verdict
- it prints on a ready close and never on a block: the block is the thing to fix, and the evidence is read when there is nothing left to fix
- a close with no verdict prints nothing: the silence is not a claim, and the review note already says a verdict is missing
- the document says what the line measures and what it does not: the tag names the merge of the release pull request, and the verdict judged the feature's closing tree, which is a different tree on purpose
