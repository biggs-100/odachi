# a-delegated-step-is-handed-with-delegate

## What must exist
- a delegated step is handed with `odachi delegate`, which composes the prompt from the task doc: the step records its skills at plan time, and the prompt always carries `## Skills to load before work` with their exact paths or an explicit none

## Done when
- `odachi plan --route delegated --agent odachi-worker --skill drive --skill tests` records the two skills in the step and validates each name against the installed skills
- `odachi delegate <taskdoc> --step T1 --surface <path> --command <cmd>` prints the task doc path, the step, its requirement verbatim, the surfaces, the commands and the skills section with exact paths
- a delegated step with no skills prints the section with an explicit none, and an unknown skill fails naming it and the known ones
- the generated block names the verb, and the other skills are untouched
- go test ./... is green and odachi check --run "go test ./..." is green
- Unreleased names the verb and the task-doc field: a contract change, minor

## Not included (first version)
- no change to the skills: drive and slices keep their text, and the block names the verb
- no change to the worker refusal or the door: their prompts stay as they are
- no auto-delegation: the verb prints the prompt, and the parent still calls the tool
- no change to odachi bench and no new verb beyond delegate
- no change to the brief format or to start
- no change to the review, the doctor or the classifier
- no change to the stored proofs: never touched
- no change to the design canvas, its token or the installers
- no change to the release workflow

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: this week.
- Wants to learn: whether a composed delegation makes the parent pass the skills, measured with the same three runs.
