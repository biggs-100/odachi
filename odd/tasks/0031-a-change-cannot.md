# a-change-cannot

## Requirements (verbatim, from the user)
R1 · "a change cannot loosen the fence it must obey"

## Done when
- with --fence-from the gate compares the base fence with the worktree fence and blocks a removed rule, path or dependency
- adding a rule or a path is stricter and stays green
- without --fence-from the worktree fence is the policy, as before
- a pull request that empties .odachi/fence.toml is blocked in CI
- the README and ADR 0027 say what is compared and what is not

## Steps
- [x] T1 · the base policy survives unless the brief changes, and the gate says so   R1 · 6b8bf61
- [x] T2 · check compares the base and worktree fence only when the brief is the same   R1 · 864ee30
- [x] T3 · README and ADR 0027: what is compared, and what is not   R1 · 2e80646

## History
L1 · "a change cannot loosen the fence it must obey"
L2 · OFICIO · "The fence may only be loosened when the brief changes" (porque start writes both, and a new feature must be able to replace the previous policy)
L3 · OFICIO · "Rule identity is the rule text, and the comparison is exact" (porque a renamed rule is a different promise, and a glob is a string the gate cannot reason about)
L4 · OFICIO · "The brief comparison normalizes line endings" (porque a Windows checkout is not a new vision)
L5 · OFICIO · "The fence compares the non-goals of the brief, not the brief bytes" (porque a trailing space or a comment is not a new vision, and the fence is made of the non-goals)
L6 · OFICIO · "A malformed candidate fence blocks even when the brief changed" (porque after the merge it is the policy, and a broken base fence would wedge every later run)
