# the-suite-is

## What must exist
- the suite is deterministic: no test reads state another test writes

## Done when
- go test ./cmd/odachi -count=20 -shuffle=on is green three times in a row
- the version surfaces and the ldflags seam are pinned without mutating the package variable
- the full suite is green on Windows and Linux CI

## Not included (first version)
- what odachi --version prints
- a go build inside the suite

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: parallel tests and shared state.
