# a-checkout-is

## What must exist
- a checkout is not an edit: a block with CRLF line endings reads the same as one with LF, so doctor does not call it hand-edited and the parser folds them once

## Done when
- a CRLF block answers the same as an LF block: BodyIn, VoiceIn and HasBlock agree
- doctor stops folding on its own, because the domain does it once, and its CRLF test stays green
- a genuinely edited block is still detected as edited
- the debt row closes and the changelog says one line

## Not included (first version)
- no change to the verdict shape
- no change to the proof format, and the stored proof is never rewritten

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: dónde se pliega un fin de línea, una vez o en cada lector.
