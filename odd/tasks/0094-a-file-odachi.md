# a-file-odachi
Base: 9a46e05c2e1b4dc8585289bf5ce077ffb1a3ffd0

## Requirements (verbatim, from the user)
R1 · "a file odachi wrote says which schema wrote it, and a binary that reads a newer value says upgrade instead of naming keys as unknown"

## Done when
- a verdict, agents.toml and odd/review.toml written today carry schema = 1, and a file without it decodes exactly as today
- a value the binary does not know is refused as written by a newer odachi, naming the value it knows and asking for an upgrade
- strictness stays: a key that is not a schema field is still an error naming the key
- a test drives each of the three decoders

## Steps

## History
L1 · "a file odachi wrote says which schema wrote it, and a binary that reads a newer value says upgrade instead of naming keys as unknown"
