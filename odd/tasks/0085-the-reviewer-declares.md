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
- [x] T2 · the review prompt asks for the declaration and the printed schema shows it   R1

## History
L1 · "the reviewer declares what it inspected, and the gate refuses an answer whose declaration does not cover every changed path"
L2 · OFICIO · "the declaration is judged on the answer, never on the merged verdict" (porque the store lends a verdict its rows; the moment the merge runs a second lens group inherits the first group's inspection, and checking the union seals an answer that declared nothing - the review of this feature found it)
L3 · OFICIO · "the review model is the one odd/review.toml names; the closing verdict of 0085 names none, so check cannot see it was written by the model that built the tree" (porque the human pays for the tokens, and a second model for one review was not authorized; the price is a gate that cannot attribute this verdict, written down in docs/review-debt.md)
