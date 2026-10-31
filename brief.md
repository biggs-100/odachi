# the-project-decides

## What must exist
- the project decides how much review evidence survives, in odd/review.toml

## Done when
- keep_verdicts and keep_prompts live in odd/review.toml, and a project that never writes them keeps today's numbers: ten records and no prompt
- odachi clean reads them from the file, uses the defaults when the file does not name them, refuses a file it cannot read by naming it, and says which numbers it used
- an older odachi refuses the new file as newer, not as an unknown key
- odachi review mode on rewrites the file and keeps the two numbers

## Not included (first version)
- no change to the proof format, and the stored proof is never rewritten
- no change to the bundled skills

## Decisions taken
- It is used from the phone and the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: por que un default se escribe en el codigo y una eleccion en el archivo del proyecto.
