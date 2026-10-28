# a-closed-task
Base: 95fb7d8e5c7234bca6dd9a56d49c2d6979c397eb

## Requirements (verbatim, from the user)
R1 · "a closed task document is never reused, so a new feature with the same three words gets its own number"

## Done when
- start picks the next number when the only document that matches the slug is closed, and writes a new one
- a re-run of start for the same open feature still reuses its document, so one feature never gets two
- the run says which document it wrote or reused
- a test drives both answers with two documents that share a slug, observed red first

## Steps

## History
L1 · "a closed task document is never reused, so a new feature with the same three words gets its own number"
