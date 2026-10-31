# the-review-s
Base: b678011a82e2082c7a272cb4b7333fd3ab0c1a8a

## Requirements (verbatim, from the user)
R1 · "the review's repeat guard asks for the lenses this run already landed, never for the shape of the candidate"

## Done when
- a second run over the same split candidate is refused and starts no reviewer
- a split run still merges its parts into one verdict, one reviewer per part
- the bypass reads the run's landed lenses, and a part of the same run is still allowed to add its rows

## Steps
- [x] T1 · a test drives a second run over the same split candidate and sees it refused, with no reviewer started   R1 · tree:21f42da32ba49835d63c0c45c4e3462a9b2c7ed7
- [ ] T2 · storeIn asks the lenses this run landed, per lens, instead of a repeat flag keyed on the split   R1
- [ ] T3 · the changelog and the review-debt row say what the guard asks now, and that a second run was already refused   R1

## History
L1 · "the review's repeat guard asks for the lenses this run already landed, never for the shape of the candidate"
L2 · OFICIO · "T1 is a pin, not a red: the second run over a split candidate is already refused one layer up, before any reviewer runs" (porque exec counts the lenses still owed before it resolves a group, so the double payment the row feared cannot happen today; the guard asks the wrong question but nothing pays twice, and the test makes the promise un-returnable)
