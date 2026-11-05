# 0050 - A skill's depth travels beside it

Status: accepted - Date: 2026-11 - Session: after v0.16.2

## Context

A skill is one `SKILL.md` capped at sixty lines. The cap is what makes it loadable: a page the agent reads every session has to stay short, and a rule that did not fit was read as "it is a document, not a skill". The consequence was quiet: the depth that did not fit stayed in the repository, in prose nobody loaded, or it was squeezed into a paragraph that lost its reasons.

Feature 0205 studied gentle-shell's skill system -read-only, never copied- and found the mechanism that answers this: a skill is a folder, its body routes, and the depth lives in `references/*.md` that ship with it. The study is the source of the grammar below; its headings, frontmatter, prefixed names, slash commands and aesthetic bans are not.

## Decision

- **A skill is its folder.** `skills/<name>/SKILL.md` and `skills/<name>/references/*.md` are one unit: `init` and `pack` install every file, `doctor` compares every file, and a skill with a reference whose bytes differ is drift. The folder holds markdown references only; a binary or a tree under `references/` is refused.
- **The body declares; the reference carries.** Every reference is named in the body with its exact path, and the inventory is `## References`. A skill with more than one reference routes them in `## Decision gates`: one row per situation, the path in backticks. A reference never restates the trigger or the output contract; it is depth, not a second skill.
- **A skill that carries depth says what it returns.** `## Output` is required when a skill has references: the agent has to know what to hand back after reading them.
- **The check enforces what can be checked, today.** `odachi skills --check` refuses a path the body names and the folder does not have, a file under `references/` nobody names in `## References`, a reference that is not a `.md`, a router whose gates omit a reference, and an empty `## Output`. The 60-line body cap stays.
- **The registry stays one row per skill.** `SKILLS.md` indexes the entry point, not the depth: references are discovered from the `SKILL.md` that names them, and a second index would drift.
- **The grammar lands in stages.** `When`, `Rules` and `Never` were required before this feature and stay required. The other sixteen skills do not carry references yet, so the check does not demand `## Output` from a skill that declares none; the migration is named in the task document and is the next feature's work. A rule that demands what nobody wrote would be red for a reason this feature did not create.
- **gentle-shell is a reference, not a template.** Its section names, license and metadata frontmatter, `gentle-ai-` prefix, slash commands, tool names, `.atl` registry and aesthetic bans stay out. Our grammar is `When`, `Rules`, `Never` plus the three above.

## Consequences

- A skill can keep teaching the reasons without growing the page every session loads. The depth arrives with the skill on every machine, and the binary is the only source.
- The counts stay human: `pack` reports a skill once, `doctor` says how many skills it compared, and a drift names the file a person has to open.
- The study of a foreign repository leaves a mechanism and an ADR, not a copy. What we adopt is the routing idea and the shipped folder; what we reject is its packaging.
- The next feature owes the migration: sixteen skills, their `## Output`, and their references where the body exceeds the cap. Until then `--check` accepts a skill with no references, which is the honest state of the curriculum.

## Amendment - the contract is required of every skill (0206, 2026-11-05)

The stage this record named landed. The seventeen bodies that declared no `## Output` declare it
now, `drive` and `slices` moved the block that no longer fit to `references/`, and
`odachi skills --check` refuses a skill without the section whether or not it carries references.
The scaffold `odachi skill new` writes it, so a new skill cannot be born without it. The sentence
above that left the demand to "the next feature" is history, not the rule.
