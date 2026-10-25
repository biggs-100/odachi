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
L2 · OFICIO · "The fake denies one write by path predicate, not by breaking the port" (porque a broken filesystem would fail every operation, so the test would prove nothing about the one branch that matters)
L3 · OFICIO · "The test ships with a manual RED: the path is claimed before the write, the test fails, the guard comes back" (porque a characterization test that never failed is a test nobody proved)
