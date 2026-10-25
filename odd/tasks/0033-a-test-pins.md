# a-test-pins

## Requirements (verbatim, from the user)
R1 · "a test pins the branch where keepFailed cannot write the .failed.txt: the gate promises not to name a file that does not exist, and nobody verifies it"

## Done when
- a test makes only the .failed.txt write fail and the gate names no output path
- the same run still fails with its one line and its next command
- go test ./... stays green and check --run writes this tree's proof

## Steps
- [ ] T1 · a test fails only the .failed.txt write and pins that the gate claims no evidence it could not write   R1

## History
L1 · "a test pins the branch where keepFailed cannot write the .failed.txt: the gate promises not to name a file that does not exist, and nobody verifies it"
