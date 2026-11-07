# a-release-says-which-bytes-were-judged
Base: 8bb33bb2806e93195311df673c0990a3573898ca

## Requirements (verbatim, from the user)
R1 · "a release says which bytes were judged: `docs/releases.md` states that the tag carries the verdict of the feature's closing tree and that the release commit moves only documents, and `check` names the verdict's tree at the close next to the proof, so a reader stops believing the tag itself was reviewed"

## Done when
- `docs/releases.md` states the binding: the tag carries the verdict of the feature's closing tree, and the release commit moves only documents
- `check`'s ready output at a close names the verdict's tree next to the proof when a verdict exists, and says nothing when none does
- a test seeds a verdict and asserts the line; another asserts the silence without one; the exit code stays ready in both
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- the changelog's Unreleased section names the line

## Steps
- [x] T1 · check names the verdict's tree next to the proof   R1 · route:inline · tree:4cc9f6adb273f3c08fa844db2694887715976cee
- [x] T2 · the releases document states the binding   R1 · route:inline · tree:9f0df0fb2bf09db9a992f286a6b7c250f079edab

## History
L1 · "a release says which bytes were judged: `docs/releases.md` states that the tag carries the verdict of the feature's closing tree and that the release commit moves only documents, and `check` names the verdict's tree at the close next to the proof, so a reader stops believing the tag itself was reviewed"
