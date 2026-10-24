# the-suite-is

## Requirements (verbatim, from the user)
R1 · "the suite is deterministic: no test reads state another test writes"

## Done when
- go test ./cmd/odachi -count=20 -shuffle=on is green three times in a row
- the version surfaces and the ldflags seam are pinned without mutating the package variable
- the full suite is green on Windows and Linux CI

## Steps
- [x] T1 · versionLine is pure: the tests stop mutating the package variable   R1 · 8e7daa7

## History
L1 · "the suite is deterministic: no test reads state another test writes"
L2 · OFICIO · "versionLine is pure: the test passes the version in instead of writing the package variable" (porque a parallel test that mutates state other tests read makes the suite depend on who wins the race)
L3 · OFICIO · "The ldflags seam is pinned by reading main.go, not by mutating the variable" (porque a go build inside the suite is out (the fence), and a guard cannot ride on scheduling)
L4 · OFICIO · "Review artifacts are written outside the tree they describe" (porque a verdict stored in the candidate changes the candidate hash and stops matching)
