# 0040 - The project decides how much evidence survives

Status: accepted — Date: 2026-10 — Session: after v0.11.74

## Context

`odachi clean` shipped with two lifetimes written in the code: ten review records
and, with them, their prompts. The number was a defensible default, and it was
still the tool deciding for the project.

After 0133 a prompt is the biggest thing in the store. Measured in this clone on
2026-10-31: eleven prompts, 597 KB, the largest 114 KB -the prompt of a feature
whose verdict is now 2.4 KB-. Nothing the gate reads touches a prompt; it is kept
so a human can see what the reviewer was asked, and it is paid for again on every
review.

So the two numbers were not the same decision, and neither was the tool's:

- a record is the evidence the gate reads, and the store must keep at least the
  newest one, because that is the verdict `check` asks for;
- a prompt is the question, and a project may reasonably keep none, one, or the
  last twenty.

## Decision

- **`odd/review.toml` carries `keep_verdicts` and `keep_prompts`**, counted in
  trees: a prompt belongs to the tree whose review it served, so "the prompts of
  the newest three trees" and "the newest three records" read the same order.
- **Zero is the absence of a choice** and means the default odachi ships: ten
  records, no prompt. There is no "keep no records": the gate reads the newest
  one, and a store that forgot it makes the project pay for the same review twice.
- **`Render` writes both numbers**, so a project that opens the file, or runs
  `odachi review mode on`, reads what is in force instead of a default nothing
  names.
- **A number nobody reaches is a typo**: more than a thousand is refused, and so is
  a negative one, in the same words as the other keys of this file.
- **The report says where the numbers came from**: "from odd/review.toml" or "from
  odachi's defaults", beside the table. A lifetime nobody remembers choosing is
  never a mystery.
- **The file that carries the keys follows the release that reads them.** CI builds
  the gate from the base ref, and a base binary that reads schema 2 refuses a
  candidate that says schema 3 before the suite runs. This repository's own file
  stays at schema 2 until v0.11.75 is on the machine; the numbers it will carry are
  the defaults, so nothing changes when it does.

## Alternatives rejected

- **A flag on the command** (`odachi clean --keep 5`). A retention is a decision
  about the project's evidence, not about one run: a flag would let the number
  drift with whoever typed the command, and `--apply` would delete bytes nobody
  agreed to lose.
- **A knob for the undo backups too.** Their lifetime is the way back from a
  mistake, not a policy: five versions are the same answer in every project, and a
  third key would invite a project to make undo useless.
- **A stem at the top of the store** (`.git/odachi/retention.toml`). A second file
  to read, with its own schema and its own reader, for two numbers that live next
  to the switch that already governs the review.
- **Keeping the prompts out of the store** (writing them beside the run). The store
  is where a refused answer and its question survive together; moving one half is
  how 0088's "the prompt that was sent is filed" stops being true.

## Consequences

- A project can now empty 597 KB of prompts in one command, or keep the question of
  the last three reviews.
- The schema moves to 3. One counter serves the three strict files, which 0094
  accepted on purpose: a v0.11.74 binary refuses a verdict whose shape did not
  change, and its answer -upgrade- is still true.
- `clean` refuses a `review.toml` it cannot parse instead of falling back to the
  defaults, because the numbers are a decision and half of one is worse than an
  answer nobody chose.
- A project that never opens the file keeps exactly what it had before this
  feature: ten records and no prompt.
