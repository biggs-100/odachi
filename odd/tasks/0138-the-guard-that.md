# the-guard-that
Base: 5072eecccd7d3954809dcf7a773fee7aebf94746

## Requirements (verbatim, from the user)
R1 · "the guard that decides whether an answer may merge into an existing verdict is a predicate the package pins, so asking the wrong question fails a test"

## Done when
- a run that landed the lens merges its rows into that lens
- a run that landed nothing is refused, whatever shape the candidate has
- making the guard ask the candidate's shape again fails a test

## Steps
- [x] T1 · a table test names the collision rule and fails until the predicate exists   R1 · tree:a8aeee58739df8e34a693daa160bffe5cb027bb8
- [x] T2 · storeIn asks the predicate, and the split and second-run tests still pass   R1 · tree:7ecd4252f3ea2f377bf6f42021718b42874b587a
- [x] T3 · the changelog and the review-debt row say the guard is pinned   R1 · tree:a179823a9f5863eb2cad872ca2a0b2ff306af18b

## History
L1 · "the guard that decides whether an answer may merge into an existing verdict is a predicate the package pins, so asking the wrong question fails a test"
L2 · OFICIO · "T1 carries the test and the predicate in one commit: the red was observed, then the function landed" (porque a table test that names a function which does not exist leaves the suite uncompilable, and a commit with red tests is never allowed, so the step is closed by the commit that makes its own test green)
L3 · OFICIO · "the changelog goes back to '## Unreleased' and the release commit renames it" (porque 0137 wrote a versioned heading, so v0.11.78 needed no release commit and the tag pointed at the feature merge; one honest change per release PR is the habit worth keeping)
