# a-team-can

## What must exist
- a team can run the gate in CI with a trusted binary and a fence the change cannot edit

## Done when
- check --fence-from <ref> reads the fence from that git ref, and a worktree fence that loosens it is ignored
- the proof names the ref the fence came from, and the local default without the flag is byte-identical
- init, setup and both installers honor ODACHI_NO_SETUP=1, so CI gets only the binary
- the CI recipe builds the gate from the base ref and uploads the proof even when the gate fails
- the suite stays green and the proof runs once per tree

## Not included (first version)
- identity, author vs reviewer: a later feature
- org policy beyond the fence
- monorepo scopes
- signing, provenance or SBOM
- the PR trigger becoming a required check: it lands after main carries the flag
- changing the proof format beyond one comment line

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: no rush.
- Wants to learn: how a gate stays a gate when the change can edit the gate.
