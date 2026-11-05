# the-grammar-reaches-every-skill

## What must exist
- every skill the curriculum ships says what it returns, and `odachi skills --check` refuses one that does not, so the contract ADR 0050 opened is enforced instead of half-applied

## Done when
- `odachi skills --check` refuses a skill whose `## Output` is missing or empty, whether or not the skill declares references, and every one of the nineteen passes
- `drive` and `slices` keep the depth that does not fit under `references/`, each declared in `## References`, and no other skill gains a reference
- every skill body stays within the sixty-line cap, and no rule a skill teaches is lost to make room
- `odachi skill new` leaves a template that `odachi skills --check` accepts unchanged: the scaffold says what the skill hands back
- `odachi init` writes the new bytes in the machine copy, a second run writes nothing, and `odachi doctor` compares the nineteen skills with no drift
- the changelog's Unreleased section says the output contract is required of every skill
- `go test ./...` is green

## Not included (first version)
- no reference for a skill whose body fits the cap once its Output lands: the depth of the other fifteen stays inline
- no rewrite of a rule a skill teaches: a section is added, or a block moves verbatim, so no instruction is lost to make room
- no change to `odachi role` or to the skill it renders: the contract is enforced over `skills/`, and a role's skill keeps the grammar it has
- no change to the stored proofs: never touched
- no change to the voice or the interview's questions
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow
- no new verb, and no change to how a skill is installed or compared

## Decisions taken
- the contract closes: `## Output` is required of every skill, not only of one that carries references
- the cap decides depth: a body that would cross sixty lines moves one block to `references/`, and the block moves word for word
- the wording belongs to each skill: a router returns the route it chose, a writer the file it leaves, a check its verdict
- the check and the scaffold land together, so `odachi skill new` cannot create a skill the gate refuses
