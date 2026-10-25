# a-change-cannot

## What must exist
- a change cannot loosen the fence it must obey

## Done when
- with --fence-from the gate blocks a fence that drops a rule, path or dependency when the non-goals of the brief did not change
- adding a rule or a path is stricter and stays green
- without --fence-from the worktree fence is the policy, as before
- a pull request that empties .odachi/fence.toml is blocked in CI
- the README and ADR 0027 say what is compared and what is not

## Not included (first version)
- a rule language, glob algebra or subset matching
- signing the fence
- comparing against anything other than --fence-from
- changing the fence file format

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how to make a gate enforce its own policy.
