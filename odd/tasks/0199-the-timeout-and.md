# the-timeout-and
Base: 7c81decea310e192b39a2357aa75d3ea409bb600

## Requirements (verbatim, from the user)
R1 · "the timeout and proof tests are deterministic under load"

## Done when
- the timeout test's child signals that it started, and the test's verdict no longer depends on how long taskkill takes
- the proof test clears ODACHI_MODEL and OPENCODE_SESSION_ID before its two runs, so the model header is the same whatever host runs the suite
- both tests pass in three consecutive full parallel runs: go test ./... -count=1
- go test ./... is green

## Steps
- [x] T1 · the timeout test's child announces that it started and the verdict stops depending on taskkill's speed   R1 · route:inline · tree:f3169fc21e5b11e365dd0450fbe222af8a08076b
- [x] T2 · the proof test clears ODACHI_MODEL and OPENCODE_SESSION_ID before its two runs   R1 · route:inline · tree:5fa132daea2d4b5781304e33a5b1e1d5ce026420

## History
L1 · "the timeout and proof tests are deterministic under load"
