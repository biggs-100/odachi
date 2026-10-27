# the-history-keeps
Base: 0475dae821e7b9ee08cf7b5eb306ff627ffac0cd

## Requirements (verbatim, from the user)
R1 · "the history keeps the words the human gave when they carry a quote or the separator, and a test holds that promise"

## Done when
- a table test proves a record round-trips when its words carry a quote, a middle dot, a "(porque " and a trailing ")"
- ask --change with a quote or a middle dot stores the words and level counts the record as a change
- the format does not change: the lines an older binary wrote still parse

## Steps
- [x] T1 · the round trip with quotes and separators is pinned in the domain and through ask, and the map records it   R1 · tree:d9f19ac438e7efdd46d67984c28423d2ff6b82f6

## History
L1 · "the history keeps the words the human gave when they carry a quote or the separator, and a test holds that promise"
L2 · OFICIO · "the format stays; the invariant gets a test instead of an escape" (porque the probe showed quotes and middle dots round-trip already, and escaping would trade a real reading for a cosmetic one)
