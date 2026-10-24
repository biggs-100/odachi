# installing-the-door

## Requirements (verbatim, from the user)
R1 · "installing the door with init --global also makes odachi the default agent in OpenCode"

## Done when
- odachi init --global sets default_agent in the global OpenCode config, creating the file when it is missing
- an existing different default agent is preserved and named with its file and value
- comments, key order and CRLF survive the edit, and the result stays valid OpenCode config
- init without --global writes nothing in the home
- a malformed or ambiguous config is refused with the file named, and nothing is written

## Steps
- [x] T1 · the config edit adds default_agent without rewriting comments, order or CRLF   R1 · a350ea1
- [x] T2 · init --global writes the default through the editor and reports what it found   R1 · f285428
- [x] T3 · doctor names a door that is not the default, and the docs tell the whole story   R1 · 0066e47

## History
L1 · "installing the door with init --global also makes odachi the default agent in OpenCode"
L2 · OFICIO · "The config edit inserts one member after the root brace instead of re-encoding the document" (porque comments, key order and line endings are the human's, not formatting odachi may replace)
L3 · OFICIO · "Another default_agent is named, never overwritten" (porque installing a door is not hijacking a selector the human picked)
L4 · OFICIO · "A config that cannot be read safely is pending, not a failed init" (porque the door is installed and the default is one hand edit away; a broken config is not odachi's to repair)
L5 · OFICIO · "JSONC comments are stripped only for validation; the written bytes keep them" (porque OpenCode publishes .jsonc, and refusing it would fail exactly the configs its docs invite)
L6 · OFICIO · "Doctor reads the default with CurrentAgent and never writes" (porque a question is not a reason to touch the human's config)
