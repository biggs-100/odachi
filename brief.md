# a-test-pins

## What must exist
- a test pins the branch where keepFailed cannot write the .failed.txt: the gate promises not to name a file that does not exist, and nobody verifies it

## Done when
- a test makes only the .failed.txt write fail and the gate names no output path
- the same run still fails with its one line and its next command
- go test ./... stays green and check --run writes this tree's proof

## Not included (first version)
- changing keepFailed's behavior: it stays best effort
- a retry or fallback for a failed write
- changing the failed file format

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why an error swallowed on purpose still needs a test.
