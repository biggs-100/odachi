# a-professional-persona

## What must exist
- a professional persona is created by an interview and research, and odachi writes it as a selectable agent and a skill from one source

## Done when
- `.odachi/roles/<slug>.md` is the only source, with its eight sections: who it is, what it asks before starting, how it works, what it refuses, how it knows it worked, what it does not know, where this comes from, and a probe
- `odachi role new <slug>` writes that skeleton, refuses a bad or repeated slug, and the file parses back into the same sections
- `odachi role <slug>` renders the agent file for each chosen client and the role's skill from the source; a second run is byte-identical, and a changed source rewrites both naming what changed
- the rendered skill is the craft: `When` from the trigger, `Rules` from what it asks and how it works, `Never` from what it refuses; a role whose rules do not fit in 60 lines is refused, not truncated
- the rendered agent carries the same rules block as odachi plus the role section, so a role never bypasses the fence, the proof or the review
- `role` with no slug lists each role and whether its artifacts match the source
- the `role-craft` skill runs the interview and the research, drafts the source, and ends in `role new`: the human approves the words and the binary writes them
- tests pin the parse, the render, the rewrite, the 60-line refusal and the skill contract; `skills --check` is green
- the impact map records the third axis: role is not voice and not profile

## Not included (first version)
- shipping a role of our own: marketing is created with the skill as the acceptance proof, not carried in the binary
- managing several roles at once, or a `--role` flag on other verbs
- internet access from the binary: the agent researches, the binary renders
- touching novato/senior, the profile, the subagents, or the review lenses
- a knowledge base, a corpus or a vector store: the role is prose with sources, not an index
- a `seniority = 10` field: the years are not verifiable, and what the role asks and refuses is

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: this week, not today.
- Wants to learn: how a profession is written down without pretending the agent has it.
