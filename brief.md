# the-timeout-and

## What must exist
- the timeout and proof tests are deterministic under load

## Done when
- the timeout test's child signals that it started, and the test's verdict no longer depends on how long taskkill takes
- the proof test clears ODACHI_MODEL and OPENCODE_SESSION_ID before its two runs, so the model header is the same whatever host runs the suite
- both tests pass in three consecutive full parallel runs: go test ./... -count=1
- go test ./... is green

## Not included (first version)
- no change to the production kill path
- no change to the proof format or the gate
- no new flag and no new verb
- no change to the voice or the interview
- no release: tests ride the next one

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how to make a test about a race deterministic.
- Wants to learn: why a test that inherits the session environment is not hermetic.
