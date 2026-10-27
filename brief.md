# the-subagents-live

## What must exist
- the subagents live in the home too, so a project that never writes .opencode keeps the worker, the verifier and the reviewer

## Done when
- init --global and odachi setup write odachi-worker, odachi-verify and odachi-reviewer in the home of each chosen client, next to the door
- without --global the home gets nothing, and a second run reports the subagents unchanged
- doctor compares the home copies and names a stale one
- the tests pin the paths, the write and the comparison, and the map and README record it

## Not included (first version)
- removing the project copies
- a new flag
- changing the subagent prompts

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how a file that lives in two roots stays one thing.
