# the-review-s
Base: b678011a82e2082c7a272cb4b7333fd3ab0c1a8a

## Requirements (verbatim, from the user)
R1 · "the review's repeat guard asks for the lenses this run already landed, never for the shape of the candidate"

## Done when
- a second run over the same split candidate is refused and starts no reviewer
- a split run still merges its parts into one verdict, one reviewer per part
- the bypass reads the run's landed lenses, and a part of the same run is still allowed to add its rows

## Steps
- [ ] T1 · a test drives a second run over the same split candidate and sees it refused, with no reviewer started   R1
- [ ] T2 · storeIn asks the lenses this run landed, per lens, instead of a repeat flag keyed on the split   R1
- [ ] T3 · the changelog and the review-debt row say what the guard asks now, and that a second run was already refused   R1

## History
L1 · "the review's repeat guard asks for the lenses this run already landed, never for the shape of the candidate"
