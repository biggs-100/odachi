# the-skill-contract

## What must exist
- a skill carries depth without growing its 60-line body: the depth lives in `references/` beside its `SKILL.md`, declared by path, installed with the skill, checked by the binary, and loaded only when the situation asks for it

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

## Not included (first version)
- no rewrite of the other sixteen skills: their references, their gates and their Output are the follow-up, and `--check` does not require a section a skill has not migrated to yet
- no change to the voice or the interview's questions
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the stored proofs: never touched
- no change to the release workflow
- no new verb: references ride the verbs that already install skills
- no adoption of gentle-shell's headings, frontmatter, license text, slash commands, tool names or aesthetic bans

## Decisions taken
- The mechanism is adopted, not the prose: references, router and output contract are ours, in our grammar.
- The depth stays local: a reference ships beside its skill and is never fetched.
- The registry stays one row per skill; the references are the skill's own declaration, not a second index.
- The grammar lands in stages: the check enforces coherence for what a skill declares today, and the ADR names what the migration owes.
- gentle-shell is read-only: nothing of its files is copied into this repository.
