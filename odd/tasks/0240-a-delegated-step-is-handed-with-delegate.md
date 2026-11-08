# a-delegated-step-is-handed-with-delegate
Base: ec3036f6fd451f6afac84290fb688e497ebcff14

## Requirements (verbatim, from the user)
R1 · "a delegated step is handed with `odachi delegate`, which composes the prompt from the task doc: the step records its skills at plan time, and the prompt always carries `## Skills to load before work` with their exact paths or an explicit none"

## Done when
- `odachi plan --route delegated --agent odachi-worker --skill drive --skill tests` records the two skills in the step and validates each name against the installed skills
- `odachi delegate <taskdoc> --step T1 --surface <path> --command <cmd>` prints the task doc path, the step, its requirement verbatim, the surfaces, the commands and the skills section with exact paths
- a delegated step with no skills prints the section with an explicit none, and an unknown skill fails naming it and the known ones
- the generated block names the verb, and the other skills are untouched
- go test ./... is green and odachi check --run "go test ./..." is green
- Unreleased names the verb and the task-doc field: a contract change, minor

## Steps
- [x] T1 · the plan records the skills a delegated step owes, and validates each name   R1 · route:inline · tree:7d6b8693ba0dd3e00838618630691dd3ad6581a2
- [x] T2 · delegate prints the prompt with the section, and the block names the verb   R1 · route:inline · tree:8d4c28658405b8e7f088c269644efc2971a0addc

## History
L1 · "a delegated step is handed with `odachi delegate`, which composes the prompt from the task doc: the step records its skills at plan time, and the prompt always carries `## Skills to load before work` with their exact paths or an explicit none"
