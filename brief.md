# the-environment-a-command-inherits

## What must exist
- `odachi doctor` names the PATH this process hands to a command when it passes what Windows can pass to a child, with the length, the limit and the repair, so a proof that fails for an inherited environment is not read as a broken suite

## Done when
- check 15 exists, appended, `CheckCount` says 15, and the usage text and `docs/impact-map.md` say fifteen
- on Windows with a PATH of 8192 characters or more the check is a warning that names the length, the 8191 limit and the repair; at 8191 or less it is ok
- on a platform that does not cut the environment, the check is ok and says which platform it measured
- nothing trims, rewrites or persists the human's PATH, and the command a proof runs receives the same environment it received before
- `go test ./...` is green
- the changelog's Unreleased section names the check

## Not included (first version)
- no change to how a command is spawned, or to the environment a child receives
- no trimming, rewriting or persisting of the human's PATH
- no new verb and no new flag
- no change to the curriculum or the contract: no file under `skills/` moves
- no change to the stored proofs: never touched
- no change to the voice or the interview's questions
- no change to `odachi role` or to the skill it renders
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow

## Decisions taken
- it is a warning, not a problem: nothing is broken until the next command fails, and the doctor does not fail for an environment odachi does not own
- the limit is 8191 characters, cmd's documented ceiling per variable; the session that opened this row measured 13,888 and the child saw an empty PATH
- the measurement is the process PATH, because that is exactly what a child inherits
- the platform is passed into the check, so both branches are tested on any machine
