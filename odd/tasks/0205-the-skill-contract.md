# the-skill-contract
Base: 5332d6f6ee2b696de540f0c428d99b40fa126fb3

## Requirements (verbatim, from the user)
R1 · "a skill carries depth without growing its 60-line body: the depth lives in `references/` beside its `SKILL.md`, declared by path, installed with the skill, checked by the binary, and loaded only when the situation asks for it"

## Done when
- `odachi skills --check` refuses a `references/` path the skill names and does not have, a file under `references/` the skill never names, and a reference that is not a `.md`; every line names the skill and the file
- a skill with more than one reference has a `## Decision gates` table whose rows name every reference, and a skill with references declares `## Output` and `## References`
- `odachi init` and `odachi pack` install a skill as its folder -`SKILL.md` and `references/`- with the binary's exact bytes, in the project copy and the home copy; a second run writes nothing, and a differing file is backed up before it is replaced
- `odachi doctor` compares the installed folder with the binary: a missing or changed reference is drift, named by file, and the comparison counts skills, not files
- the doctor row `docs/review-debt.md` leaves open is paid with its test: the opencode default note appends instead of returning, so the design plugin note is never skipped
- the grammar is written where a person writing a skill reads it (`skill-craft`): `## Decision gates` when a skill routes, `## Output` what it returns, `## References` what it reads, and one ADR records the target grammar and the migration debt
- `ui-craft` is the first router: its body keeps the order and the first move, four references carry the catalogue, the numbers, the motion and the final read, and each gate row names its reference
- `design` declares what it returns and keeps its 60 lines: the slug, each artboard with its tradeoff, the files changed, the comments resolved and the canvas URL, with the on-disk format of a design folder moved to a reference
- `go test ./...` is green

## Steps
- [x] T1 · the check reads a skill's references: each named path exists, each file is named and listed, and a router has gates   R1 · route:inline · tree:8623cecda484f9987489871c6bafe2cd04e7b83e
- [ ] T2 · the binary carries a skill's folder: embed reads references/ and pack writes the tree byte for byte   R1 · route:inline
- [x] T3 · init installs the folder and doctor compares it file by file; the design-plugin review-debt row is paid   R1 · route:inline · tree:6737d19ebec98f0ffc216549906505062b2534d1
- [ ] T4 · ui-craft becomes a router: four references carry the catalogue, the numbers, the motion and the final read   R1 · route:inline
- [ ] T5 · design declares its output and moves the on-disk format of a design folder to a reference   R1 · route:inline
- [ ] T6 · the grammar is written in skill-craft, ADR 0050 records the debt, and the changelog names it   R1 · route:inline

## History
L1 · "a skill carries depth without growing its 60-line body: the depth lives in `references/` beside its `SKILL.md`, declared by path, installed with the skill, checked by the binary, and loaded only when the situation asks for it"
L2 · OFICIO · "T2 was inserted by hand and the plan renumbered before any step was done" (porque plan appends and undo restores one backup, so a step written out of order cannot be moved by the verb)
