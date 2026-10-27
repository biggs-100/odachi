# the-history-keeps

## What must exist
- the history keeps the words the human gave when they carry a quote or the separator, and a test holds that promise

## Done when
- a table test proves a record round-trips when its words carry a quote, a middle dot, a "(porque " and a trailing ")"
- ask --change with a quote or a middle dot stores the words and level counts the record as a change
- the format does not change: the lines an older binary wrote still parse

## Not included (first version)
- changing the record format or escaping the quotes
- touching start or plan
- rewriting old task docs

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: when a suspected bug is only a rendering, the cheapest fix is a test.
