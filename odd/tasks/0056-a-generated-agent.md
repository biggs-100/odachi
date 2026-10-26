# a-generated-agent
Base: 6f00191b5805cd44883cbeec87241e1019b0d679

## Requirements (verbatim, from the user)
R1 · "a generated agent file is matched by its exact name, so the human file is never skipped"

## Done when
- the names odachi writes are listed once, and the review treats only those as plumbing
- a human file such as .opencode/agents/odachi-helper.md reaches a lens instead of disappearing
- the generated files of both clients keep the classification they had

## Steps
- [x] T1 · the review treats only the agent files odachi writes as plumbing, and a human file with the same prefix reaches a lens   R1 · ecea87c

## History
L1 · "a generated agent file is matched by its exact name, so the human file is never skipped"
L2 · OFICIO · "the generated agent names are derived from persona, the writer, not written again in the review" (porque a list written twice drifts the day a subagent is added; the writer owns the names)
L3 · OFICIO · "an agent file is code even when it is Markdown" (porque it grants tools and can carry a token in its prompt: the extension decides what a document is, never what an agent may do)
