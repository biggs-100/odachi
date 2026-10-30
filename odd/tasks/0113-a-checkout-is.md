# a-checkout-is
Base: 91f1cee771ecf4c2ffd5fde29cd30864ccd89afa

## Requirements (verbatim, from the user)
R1 · "a checkout is not an edit: a block with CRLF line endings reads the same as one with LF, so doctor does not call it hand-edited and the parser folds them once"

## Done when
- a CRLF block answers the same as an LF block: BodyIn, VoiceIn and HasBlock agree
- doctor stops folding on its own, because the domain does it once, and its CRLF test stays green
- a genuinely edited block is still detected as edited
- the debt row closes and the changelog says one line

## Steps
- [x] T1 · the domain folds CRLF once, and a CRLF block answers like an LF one   R1 · tree:8127db64000e1f1032e7d48b86182034acf33918
- [x] T2 · doctor drops its own fold, and the debt row closes   R1 · tree:ab25f98e009553000bddc198b305ebd0b029cfe4

## History
L1 · "a checkout is not an edit: a block with CRLF line endings reads the same as one with LF, so doctor does not call it hand-edited and the parser folds them once"
