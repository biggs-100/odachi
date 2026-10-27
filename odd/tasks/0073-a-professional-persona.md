# a-professional-persona
Base: 519c797fd1753ffc3f2ca4425b0c92ebdda6005e

## Requirements (verbatim, from the user)
R1 · "a professional persona is created by an interview and research, and odachi writes it as a selectable agent and a skill from one source"

## Done when
- `.odachi/personas/<slug>.md` is the only source, with its eight sections: who it is, what it asks before starting, how it works, what it refuses, how it knows it worked, what it does not know, where this comes from, and a probe
- `odachi persona new <slug>` writes that skeleton, refuses a bad or repeated slug, and the file parses back into the same sections
- `odachi persona <slug>` renders the agent file for each chosen client and the role's skill from the source; a second run is byte-identical, and a changed source rewrites both naming what changed
- the rendered skill is the craft: `When` from the trigger, `Rules` from what it asks and how it works, `Never` from what it refuses; a role whose rules do not fit in 60 lines is refused, not truncated
- the rendered agent carries the same rules block as odachi plus the role section, so a role never bypasses the fence, the proof or the review
- `persona` with no slug lists each role and whether its artifacts match the source
- the `persona-craft` skill runs the interview and the research, drafts the source, and ends in `persona new`: the human approves the words and the binary writes them
- tests pin the parse, the render, the rewrite, the 60-line refusal and the skill contract; `skills --check` is green
- the impact map records the third axis: role is not voice and not profile

## Steps
- [x] T1 · the role source parses into its eight sections, and persona new writes the skeleton   R1 · tree:708083d529ff2a1ec4e7f4f3ec70287f4e0ecd91
- [ ] T2 · persona-craft teaches the interview and the research, and the impact map records the third axis   R1
- [ ] T3 · persona renders the agent and the skill from the source, and a role over 60 lines is refused   R1

## History
L1 · "a professional persona is created by an interview and research, and odachi writes it as a selectable agent and a skill from one source"
L2 · OFICIO · "the role source is Markdown in .odachi/personas/" (porque a role is human content with sources and a probe, not machine config; TOML would force prose into a field)
L3 · OFICIO · "a role whose rules do not fit in the 60-line skill is refused, not truncated" (porque the cap is a contract of the renderer: a role that cannot be loaded is not a role)
L4 · CAMBIO R1 · "the verb is role, not persona, and the source lives in .odachi/roles/: persona is the voice, and one word with two owners is the defect ADR 0006 fixed"
L5 · CAMBIO R1 · "the creator skill is role-craft, not persona-craft, for the same reason"
L6 · CAMBIO R1 · "the Done-when bullets still say .odachi/personas/ and persona new; brief.md and the code say .odachi/roles/ and role, and the brief is the vision"
