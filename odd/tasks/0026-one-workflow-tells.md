# one-workflow-tells

## Requirements (verbatim, from the user)
R1 · "one workflow tells the truth about a tree: the gate runs the suite on every OS"

## Done when
- the gate checks the candidate on ubuntu, macos and windows, each reading the fence from the base ref and keeping its proof
- gofmt, vet, race and the goreleaser check run once in that workflow; no second workflow runs the suite
- a PR that breaks the suite on one OS fails there, and a release tag still runs the suite before publishing

## Steps
- [x] T1 · T1 - gate.yml runs the suite on three OSes, one proof each, static checks beside it   R1 · 5c926bf
- [x] T2 · T2 - ci.yml is gone: no second workflow tells a different truth   R1 · aa72053
- [x] T3 · T3 - README and ADR 0023: one workflow, one green light   R1 · 68b6cee

## History
L1 · "one workflow tells the truth about a tree: the gate runs the suite on every OS"
L2 · OFICIO · "The race detector and the static checks live in the gate workflow, not in a second one" (porque two workflows can disagree about the same tree, and one of them ends up ignored)
L3 · OFICIO · "The CI gate passes --base as well as --fence-from" (porque --fence-from picks where the policy is read; --base picks what the policy applies to, and a clean checkout compared to HEAD is an empty list)
L4 · OFICIO · "On a push the base is github.event.before, not origin/main" (porque origin/main is the pushed commit itself; only the previous tip is a base the change cannot edit)
L5 · OFICIO · "ci.yml is deleted instead of kept as a second opinion" (porque two files that can disagree about the same tree end with one of them ignored)
L6 · OFICIO · "A red run is read from the log, not from a proof" (porque odachi writes a proof only when the command exits 0, and a proof for a red tree would not be real output)
L7 · OFICIO · "A failed command leaves no proof and only its last line, not a log" (porque check captures the output and reports only the last line; claiming a log promises evidence that is not kept)
