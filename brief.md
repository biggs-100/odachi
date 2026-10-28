# a-closed-task

## What must exist
- a closed task document is never reused, so a new feature with the same three words gets its own number

## Done when
- start picks the next number when the only document that matches the slug is closed, and writes a new one
- a re-run of start for the same open feature still reuses its document, so one feature never gets two
- the run says which document it wrote or reused
- a test drives both answers with two documents that share a slug, observed red first

## Not included (first version)
- no new verb and no new flag
- no change to how a number is chosen for a slug nobody used
- no rewrite of an existing task document

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how a name decides identity: reuse by intent, not by string.
