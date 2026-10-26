# the-review-prompt
Base: e60b31cdd83f2665932d7dce1660fe2da6fa89aa

## Requirements (verbatim, from the user)
R1 · "the review prompt carries each row of the store once, not once per verdict that repeats it"

## Done when
- the carried section lists each row once
- two rows that differ in any part both survive
- the order is deterministic: two runs print the same prompt
- a store with one verdict prints the same prompt as today
- tests pin the duplicate, the merge and the order

## Steps
- [x] T1 · the carried section lists each row once, and two runs print the same prompt   R1 · fff095c
- [x] T2 · two rows that differ in any part both survive: the dedupe never merges an answer   R1 · 9b46c65

## History
L1 · "the review prompt carries each row of the store once, not once per verdict that repeats it"
