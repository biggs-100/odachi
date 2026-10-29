# a-file-odachi

## What must exist
- a file odachi wrote says which schema wrote it, and a binary that reads a newer value says upgrade instead of naming keys as unknown

## Done when
- a verdict, agents.toml and odd/review.toml written today carry schema = 1, and a file without it decodes exactly as today
- a value the binary does not know is refused as written by a newer odachi, naming the value it knows and asking for an upgrade
- strictness stays: a key that is not a schema field is still an error naming the key
- a test drives each of the three decoders

## Not included (first version)
- no migration: a file written before this change keeps its bytes until something rewrites it
- no new verb and no new flag
- no change to the proof format
- odd/profile.toml, dismissed.json and refute.json are left out of this first version

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: where compatibility is paid.
