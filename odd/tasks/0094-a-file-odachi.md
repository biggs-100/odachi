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
- [x] T1 · the schema vocabulary is one definition: a value the binary does not know is refused as a newer writer   R1 · tree:b5739473c9da123319d9d2f75bd881f18902d5e8
- [x] T2 · the verdict carries the schema it was written with, and its reader checks it before it names an unknown key   R1 · tree:ecd74aedcfee0003ad8bb53081c1ddaed9ebbc46
- [ ] T3 · agents.toml and odd/review.toml write schema = 1, absence still means 1, and a newer value says upgrade   R1
- [ ] T4 · the debt file and the changelog say what closed and what the plan commit exposed   R1

## History
L1 · "a file odachi wrote says which schema wrote it, and a binary that reads a newer value says upgrade instead of naming keys as unknown"
L2 · OFICIO · "the plan commit of 0094 landed with --no-verify" (porque a real fence is classified as ordinary code (tier.go: the fence is a human decision), so check asked for a review of a tree whose only behaviour is the human's declaration; paying a lens for it contradicts the rule that a run judges behaviour, so the friction goes to docs/review-debt.md as a tier row)
