# a-code-change

## What must exist
- a code change committed after the close still owes the verdict: the gate measures what arrived after the verdict that judged the close, not only the worktree against HEAD

## Done when
- with a verdict at the close in the store, a committed code change after it makes the close owe the verdict at the feature's base
- a committed release note after the same close stays paperwork and the gate is ready
- a close with no verdict in the store keeps the answer it gives today
- a test drives each one, and one of them fails without the change

## Not included (first version)
- no new verb and no new flag
- no change to the verdict shape
- no change to the proof format, and the stored proof is never rewritten

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: this week.
- Wants to learn: how a gate stays honest when the evidence it compares is committed instead of pending.
