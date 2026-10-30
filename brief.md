# one-command-refreshes

## What must exist
- one command refreshes the machine: odachi setup reports what it would change before it changes it, writes nothing under --dry-run, and names what it skipped, so an upgraded binary leaves a current home in one run

## Done when
- --dry-run changes no byte anywhere and names every file it would write
- a second setup after the first reports every destination unchanged, and the same run twice is a no-op
- the report names the clients that are not installed here and what it left alone, so nothing is silently skipped

## Not included (first version)
- no change to the proof format, and the stored proof is never rewritten
- no change to how a skill or the door is written: the report learns to speak first

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: por que un ensayo general tiene que escribir cero bytes para ser creible.
