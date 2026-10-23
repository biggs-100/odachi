# odachi-writes-artifacts

## What must exist
- odachi writes artifacts, docs, skills and CLI messages in English; the agent chat stays in the human's language

## Done when
- the 14 skills and their format use When/Rules/Never and skills --check is green
- brief, task doc, AGENTS.md and profile are generated in English; human quotes stay verbatim
- README, CHANGELOG, ADRs and impact map are English
- CLI messages and their tests are English
- go test ./... and odachi check are green

## Not included (first version)
- do not translate closed task docs: they are history
- do not change the agent chat language
- do not change behavior, only language

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: this week.
- Wants to learn: how to migrate the language of a contract without breaking the tests.
