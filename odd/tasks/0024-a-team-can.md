# a-team-can

## Requirements (verbatim, from the user)
R1 · "a team can run the gate in CI with a trusted binary and a fence the change cannot edit"

## Done when
- check --fence-from <ref> reads the fence from that git ref, and a worktree fence that loosens it is ignored
- the proof names the ref the fence came from, and the local default without the flag is byte-identical
- init, setup and both installers honor ODACHI_NO_SETUP=1, so CI gets only the binary
- the CI recipe builds the gate from the base ref and uploads the proof even when the gate fails
- the suite stays green and the proof runs once per tree

## Steps
- [x] T1 · check --fence-from reads the fence from a ref and names it in the proof   R1 · 5d612c3
- [x] T2 · init, setup and the installers honor ODACHI_NO_SETUP=1   R1 · 5d612c3
- [x] T3 · the CI recipe builds the gate from the base ref and uploads the proof   R1 · 39cc9a3

## History
L1 · "a team can run the gate in CI with a trusted binary and a fence the change cannot edit"
L2 · OFICIO · "check --fence-from reads the policy from a git ref, not the worktree" (porque the change cannot be the policy it must obey)
L3 · OFICIO · "A missing file at a valid ref is no fence; a ref that does not exist blocks" (porque a typo must never turn the gate off)
L4 · OFICIO · "The proof names the fence ref in an optional third header" (porque evidence must say which policy it obeyed, and old proofs without it still parse)
L5 · OFICIO · "The CI gate is built from the base ref, not from the candidate" (porque gate code shipped by the change is advisory; code the candidate cannot edit is the gate)
L6 · OFICIO · "ODACHI_NO_SETUP=1 makes the CLI refuse setup and init --global, not only the installer skip" (porque a script on a shared runner must not mutate the home even if it calls the verb directly)
