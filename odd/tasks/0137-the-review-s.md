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
- [x] T2 · storeIn asks the lenses this run landed, per lens, instead of a repeat flag keyed on the split   R1 · tree:bf07a9ec785fd012be54f0aa744db880e7541623
- [x] T3 · the changelog and the review-debt row say what the guard asks now, and that a second run was already refused   R1 · tree:c4270329c6e84f17e74d184e12bfe9b3e2b5a370

## History
L1 · "the review's repeat guard asks for the lenses this run already landed, never for the shape of the candidate"
L2 · OFICIO · "T1 is a pin, not a red: the second run over a split candidate is already refused one layer up, before any reviewer runs" (porque exec counts the lenses still owed before it resolves a group, so the double payment the row feared cannot happen today; the guard asks the wrong question but nothing pays twice, and the test makes the promise un-returnable)
L3 · OFICIO · "no ADR for 0137: the rule is one condition inside storeIn and the changelog carries it" (porque an ADR records a policy the project must keep; this narrows a guard 0136 opened and its shape is already in ADR 0041)
L4 · OFICIO · "the review's one warning earns no round: it goes to docs/review-debt.md after the close" (porque the code is right and the gap is evidence, not behaviour; fixing it now would change the bytes the verdict judged and pay a second review for the same work)
