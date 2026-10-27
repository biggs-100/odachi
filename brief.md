# the-folder-keeps

## What must exist
- the folder keeps only the rules and the plan, and the agent files of a project can live in the home

## Done when
- init --global --no-project-agents writes no agent file in the project, and the flag without --global is refused naming the fix
- start and level do not recreate the project agents when they are absent
- doctor accepts the home as the source, and an empty project agents folder is the clean mode
- the tests pin the flag, the refusal, the no-recreate and the doctor note, and the docs record it

## Not included (first version)
- changing the default: init still writes the project agents unless asked
- a scope on other verbs
- deleting project files that already exist

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how a tool says where its files live without guessing.
