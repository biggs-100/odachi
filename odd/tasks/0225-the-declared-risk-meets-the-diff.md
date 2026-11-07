# the-declared-risk-meets-the-diff
Base: a659b3abb2979c868bf2b92c4d4ec35fd6c82761

## Requirements (verbatim, from the user)
R1 · "a full-tier change cannot pass as harmless in silence: when the review classifies the change as full -it touches a path the classifier names- and no commit declares a risk of its own, `check` names the contrast in one line of the ready output and names the paths, so the reviewer reads what the change claims against what it touches; a named risk stays silent, and the exit code and the verdict never move"

## Done when
- the ready output names the contrast on a full-tier change whose commits say `Risk: none`, and on one whose commits carry no `Risk:` line at all; the line names the paths the classifier saw, capped
- a named risk, a focused change and a passive change keep the ready output silent about risk
- the line never changes the exit code, the reason or the proof
- a test seeds a full path and commits with `Risk: none` and asserts the line; another seeds the same path and `Risk: data` and asserts silence; another seeds a focused path and asserts silence
- the parser's tests cover `none`, a named risk, prose after the name, a missing line and an empty line refused
- `odachi check --run "go test ./..."` is green
- the changelog's Unreleased section names the line
- `go test ./...` is green

## Steps
- [ ] T1 · the Risk: line has a parser with its refusals   R1 · route:inline
- [ ] T2 · check names a full tier that declares no risk of its own   R1 · route:inline
- [ ] T3 · the changelog names the line   R1 · route:inline

## History
L1 · "a full-tier change cannot pass as harmless in silence: when the review classifies the change as full -it touches a path the classifier names- and no commit declares a risk of its own, `check` names the contrast in one line of the ready output and names the paths, so the reviewer reads what the change claims against what it touches; a named risk stays silent, and the exit code and the verdict never move"
