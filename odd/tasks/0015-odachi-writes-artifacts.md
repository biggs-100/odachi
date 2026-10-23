# odachi-writes-artifacts

## Requirements (verbatim, from the user)
R1 · "odachi writes artifacts, docs, skills and CLI messages in English; the agent chat stays in the human's language"

## Done when
- the 14 skills and their format use When/Rules/Never and skills --check is green
- brief, task doc, AGENTS.md and profile are generated in English; human quotes stay verbatim
- README, CHANGELOG, ADRs and impact map are English
- CLI messages and their tests are English
- go test ./... and odachi check are green

## Steps
- [x] T1 · the 14 skills and the skill format speak English   R1 · 130082c
- [x] T2 · brief, task doc, AGENTS.md and profile generators write English   R1 · 002536b
- [ ] T3 · README, CHANGELOG, ADRs and impact map are English   R1
- [ ] T4 · domain and app messages speak English   R1
- [ ] T5 · cmd messages, help and tests speak English   R1
- [ ] T6 · dogfood: init installs the English skills, docs counts and check green   R1

## History
L1 · "odachi writes artifacts, docs, skills and CLI messages in English; the agent chat stays in the human's language"
L2 · OFICIO · "T2 is split into two commits: brief and task doc first, AGENTS.md and persona after" (porque the artifact parser and the rules block are two surfaces with different tests; one commit per surface reviews better)
