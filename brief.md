# the-store-prunes

## What must exist
- the store prunes what it can regenerate and names what it keeps: a stored verdict disposes of the prompts and scaffolding its keep_prompts stopped keeping, a ready close names the store size and what clean --apply would free, and one ADR says which evidence odachi never bounds

## Done when
- with keep_prompts = 0 a stored verdict leaves no prompt of its tree in the store, and the verdict is stored
- with keep_prompts = 1 the newest tree keeps its prompt when its verdict is stored, and an older tree does not
- a review that gets no verdict removes nothing, because there is nothing stored
- a ready close prints one line with the store size and what clean --apply would free, and the gate answer and exit code do not change
- an ADR records that proof/ and odd/tasks are never bounded by odachi, and clean keeps its owner lines
- the suite pins every one of them and go test ./... is green

## Not included (first version)
- no change to the bytes of a verdict or of a prompt
- the close never deletes: clean --apply keeps the switch
- no automatic bounding of proof/ or odd/tasks
- no new verb and no new flag

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why an artifact a machine can regenerate is not evidence, and why the evidence is the one thing never bounded.
