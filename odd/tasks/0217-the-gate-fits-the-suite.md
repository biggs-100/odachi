# the-gate-fits-the-suite
Base: 1212e58f0ee1307a6b55e8ede06d1b33224db531

## Requirements (verbatim, from the user)
R1 · "the gate's default bound fits the suite it actually runs: `go test ./...` takes about 700 seconds on this machine (686 for cmd/odachi under contention, 207 alone), so 600 kills a green suite and the default moves to 900"

## Done when
- `check.DefaultTimeout` is 900, and `TestTheDefaultTimeoutFitsARealSuite` pins it and the seconds the shell receives
- the comment names the measurement that decided 900, and the help text and `docs/odachi-bootstrap.md` say 900
- the changelog's Unreleased section names the bound
- `odachi check --run "go test ./..."` is green with no `--timeout` on this machine
- `go test ./...` is green

## Steps
- [x] T1 · the default bound is 900, measured on this machine   R1 · route:inline · tree:44bd203591b5e456feaa789d25ed81e282f4a1b3
- [x] T2 · the changelog names the bound   R1 · route:inline · tree:016d880533e67ea06f0ca84ac8302e2b7d78209d

## History
L1 · "the gate's default bound fits the suite it actually runs: `go test ./...` takes about 700 seconds on this machine (686 for cmd/odachi under contention, 207 alone), so 600 kills a green suite and the default moves to 900"
