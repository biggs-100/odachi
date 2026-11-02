# the-review-of
Base: 99a1f75e4c16453b00fd23097a9b96dfd591cee3

## Requirements (verbatim, from the user)
R1 · "the review of a light feature is judged against that feature's own line, not against the previous feature's brief"

## Done when
- odachi review --base <sha> on a light feature writes a prompt whose vision section quotes that feature's requirement and names the task document it came from
- a normal feature's prompt carries brief.md exactly as before
- a light document that is already closed still names its own line, so the closing tree is judged against it
- a task document that cannot be read is named as such: the review never falls back to a brief it did not read
- no lens, verdict or store entry changes: the same tree still gets the same lenses and the same JSON

## Steps
- [ ] T1 · a light feature's review carries its own line, found by the base the document records   R1 · route:inline
- [ ] T2 · the prompt says where the vision came from, and a normal feature still carries brief.md   R1 · route:inline
- [ ] T3 · a task document that cannot be read is named as such, never read as an absence   R1 · route:inline

## History
L1 · "the review of a light feature is judged against that feature's own line, not against the previous feature's brief"
