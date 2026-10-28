# the-reviewer-declares
Base: ea439e8183d49c6d90ea2b96ba9766756cc717ff

## Requirements (verbatim, from the user)
R1 · "the reviewer declares what it inspected, and the gate refuses an answer whose declaration does not cover every changed path"

## Done when
- review --attach refuses an answer with no inspection declaration when the verdict names lenses, and names the field it is missing
- a declaration that leaves a changed path out is refused, and the refusal names every path it left out
- a declaration that names a path outside the changed paths is refused, and the refusal names it
- inspection.status unavailable refuses the answer and echoes the reviewer's reason instead of sealing it
- the prompt asks for the declaration and the printed schema shows it, pinned by a test
- a test drives each refusal, observed red first

## Steps
- [x] T1 · the verdict carries the inspection declaration, and attach refuses one that does not cover every changed path   R1
- [ ] T2 · the review prompt asks for the declaration and the printed schema shows it   R1

## History
L1 · "the reviewer declares what it inspected, and the gate refuses an answer whose declaration does not cover every changed path"
