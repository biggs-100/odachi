# doctor-compares-the
Base: c927201329b8f08fa94c2ed0711cc5e4afef9e9d

## Requirements (verbatim, from the user)
R1 · "doctor compares the block even when the profile is inferred: a project that ran init and never ran start still learns that its block is older than the binary"

## Done when
- a project with no odd/profile.toml and an older block gets the warning and names odachi init
- a project with no profile and the current block stays ok
- the tests that already exist keep their answer, and one of them fails without the change

## Steps
- [x] T1 · the comparison runs with an inferred profile, and a test without profile.toml drives it   R1 · tree:6b2ac4b4620ddea32b3b2c4fafbbf3a66ab2f2e8
- [ ] T2 · the changelog names the case the smoke found   R1

## History
L1 · "doctor compares the block even when the profile is inferred: a project that ran init and never ran start still learns that its block is older than the binary"
