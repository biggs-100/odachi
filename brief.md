# the-review-prompt

## What must exist
- the review prompt carries each row of the store once, not once per verdict that repeats it

## Done when
- the carried section lists each row once
- two rows that differ in any part both survive
- the order is deterministic: two runs print the same prompt
- a store with one verdict prints the same prompt as today
- tests pin the duplicate, the merge and the order

## Not included (first version)
- the lens rules and the six rules
- which rows count as potent
- truncating by size or a budget
- the verdict schema

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how a repeated datum becomes a cost.
