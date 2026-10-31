# a-review-whose

## What must exist
- a review whose prompt does not fit the project's limit is split into parts that do, and one verdict answers the tree

## Done when
- a candidate whose prompt exceeds max_prompt_kb is sent as parts under the limit, and the store holds one verdict for the tree
- every changed path is answered by exactly one part, and the verdict's coverage is complete
- a single file whose diff alone exceeds the limit is refused, naming the file
- a candidate under the limit runs as before: one reviewer, one prompt, nothing extra written
- the prompt says which part it is, and the changelog and the ADR say why

## Not included (first version)
- no change to the proof format, and the stored proof is never rewritten
- no change to the bundled skills

## Decisions taken
- It is used from the phone and the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: por que partir por el limite medido es mejor que pedirle al revisor que no se cuelgue.
