# the-three-silent

## What must exist
- the three silent regressions are paid: the ledger's reads, the target list and pack's bytes each have a test that fails without the fix

## Done when
- TreeOf, Has and CommitWithTree run against a real repository, including Has's exit-1 path and CommitWithTree's oldest match
- role builds its target list once, and the existing tests stay green
- pack's installed skill is compared byte for byte with its source
- each new test is observed failing when the behavior it pins is broken

## Not included (first version)
- no new verb and no new flag
- no behavior change: the three are tests and one refactor
- no model and no network

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how a test is paid for by breaking the code it claims to guard.
