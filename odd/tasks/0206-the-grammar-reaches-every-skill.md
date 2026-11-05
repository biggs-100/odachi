# the-grammar-reaches-every-skill
Base: eac2079a2075e7a879bad670eed4994be98fc349

## Requirements (verbatim, from the user)
R1 · "every skill the curriculum ships says what it returns, and `odachi skills --check` refuses one that does not, so the contract ADR 0050 opened is enforced instead of half-applied"

## Done when
- `odachi skills --check` refuses a skill whose `## Output` is missing or empty, whether or not the skill declares references, and every one of the nineteen passes
- `drive` and `slices` keep the depth that does not fit under `references/`, each declared in `## References`, and no other skill gains a reference
- every skill body stays within the sixty-line cap, and no rule a skill teaches is lost to make room
- `odachi skill new` leaves a template that `odachi skills --check` accepts unchanged: the scaffold says what the skill hands back
- `odachi init` writes the new bytes in the machine copy, a second run writes nothing, and `odachi doctor` compares the nineteen skills with no drift
- the changelog's Unreleased section says the output contract is required of every skill
- `go test ./...` is green

## Steps
- [x] T1 · slices states what it returns and moves the delegation policy to references/, so the body fits the cap again   R1 · route:inline · tree:59161f80858087176680545a1e0a971cd9d676b7
- [x] T2 · shape, triage, tests and go-tests state what they return   R1 · route:inline · tree:de1d467b2eef48c6d2f905a3f4e45a710c2ede92
- [x] T3 · commits, pull-request, writing and acceptance state what they return   R1 · route:inline · tree:274cdbe86ffe1e609d200ee9d54232f1399eb39e
- [x] T4 · drive states what it returns and moves the review section to references/   R1 · route:inline · tree:a3cb90cb81b1cd06340e36b0e691f92bc6e402fe
- [ ] T5 · bootstrap, ship, data, security, role-craft and skill-improve state what they return   R1 · route:inline
- [ ] T6 · the check asks every skill for ## Output, skill new writes one, and skill-craft teaches the rule   R1 · route:inline
- [ ] T7 · the machine copy carries the new bytes, and the changelog and ADR 0050 record the stage   R1 · route:inline

## History
L1 · "every skill the curriculum ships says what it returns, and `odachi skills --check` refuses one that does not, so the contract ADR 0050 opened is enforced instead of half-applied"
L2 · OFICIO · "T5 was inserted by hand and the plan renumbered before any step was done" (porque plan appends, refuses a line over 120 characters, and undo restores one backup, so a step written out of order cannot be moved by the verb (0205's lesson))
L3 · OFICIO · "the fence carries the rules somebody gave paths, written by hand after start; the sixteen-skills rule is gone and the role rule is new" (porque start writes every non-goal as a template, and this feature rewrites exactly the files the old rule protected; a rule protects paths the feature must not touch, so it cannot survive the feature it forbade)
L4 · OFICIO · "drive's reference is the gate's answer: the block codes and the review policy move together" (porque the body overran the cap by one line once its Output landed, both sections answer the same question, and a reference loaded when a check is not plainly green keeps the manual under sixty lines)
