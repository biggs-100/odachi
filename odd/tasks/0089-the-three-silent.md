# the-three-silent
Base: 2062dff5cc36f2f3af5960f519e24aa9718ef222

## Requirements (verbatim, from the user)
R1 · "the three silent regressions are paid: the ledger's reads, the target list and pack's bytes each have a test that fails without the fix"

## Done when
- TreeOf, Has and CommitWithTree run against a real repository, including Has's exit-1 path and CommitWithTree's oldest match
- role builds its target list once, and the existing tests stay green
- pack's installed skill is compared byte for byte with its source
- each new test is observed failing when the behavior it pins is broken

## Steps
- [x] T1 · the ledger's adapter functions run against a real repository   R1
- [x] T2 · role builds the target list once   R1
- [x] T3 · pack's installed bytes are compared with their source   R1

## History
L1 · "the three silent regressions are paid: the ledger's reads, the target list and pack's bytes each have a test that fails without the fix"
L2 · OFICIO · "the gitx row was a false positive: internal/adapters/gitx/ledger_test.go already drives TreeOf, Has and CommitWithTree against a real repository" (porque the reviewer looked at gitx_test.go and missed ledger_test.go; a debt row becomes work, so it is verified before it is believed - the third false positive from the same reviewer in two sessions)
