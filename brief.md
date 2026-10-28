# the-reviewer-declares

## What must exist
- the reviewer declares what it inspected, and the gate refuses an answer whose declaration does not cover every changed path

## Done when
- review --attach refuses an answer with no inspection declaration when the verdict names lenses, and names the field it is missing
- a declaration that leaves a changed path out is refused, and the refusal names every path it left out
- a declaration that names a path outside the changed paths is refused, and the refusal names it
- inspection.status unavailable refuses the answer and echoes the reviewer's reason instead of sealing it
- the prompt asks for the declaration and the printed schema shows it, pinned by a test
- a test drives each refusal, observed red first

## Not included (first version)
- no new verb and no new flag
- no change to where the verdicts live
- no refusal of a verdict stored before this feature, which still parses

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how a declaration the tool can check beats a rule the reviewer can ignore.
