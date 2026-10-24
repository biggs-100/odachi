# installing-the-door

## What must exist
- installing the door with init --global also makes odachi the default agent in OpenCode

## Done when
- odachi init --global sets default_agent in the global OpenCode config, creating the file when it is missing
- an existing different default agent is preserved and named with its file and value
- comments, key order and CRLF survive the edit, and the result stays valid OpenCode config
- init without --global writes nothing in the home
- a malformed or ambiguous config is refused with the file named, and nothing is written

## Not included (first version)
- the project-level .opencode/opencode.json
- any OpenCode key other than default_agent
- overwriting an existing default: no force flag in the first version
- re-encoding the file or dropping comments
- a new dependency

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: no rush.
- Wants to learn: how to edit a file you do not own without rewriting it.
