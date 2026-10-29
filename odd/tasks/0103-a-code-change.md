# a-code-change
Base: 293be6b8adbd30148aaff788ad129b784d2c680f

## Requirements (verbatim, from the user)
R1 · "a code change committed after the close still owes the verdict: the gate measures what arrived after the verdict that judged the close, not only the worktree against HEAD"

## Done when
- with a verdict at the close in the store, a committed code change after it makes the close owe the verdict at the feature's base
- a committed release note after the same close stays paperwork and the gate is ready
- a close with no verdict in the store keeps the answer it gives today
- a test drives each one, and one of them fails without the change

## Steps
- [x] T1 · the close measures what arrived after the verdict, so committed code still owes   R1 · tree:90896b38cc15fa4c7a797e1b993ceb93d1f68018
- [x] T2 · the debt row leaves the file and the changelog says why   R1 · tree:3837d722e6c6bf4b75772f0570c265dea6a7725e

## History
L1 · "a code change committed after the close still owes the verdict: the gate measures what arrived after the verdict that judged the close, not only the worktree against HEAD"
L2 · OFICIO · "the anchor is the verdict's tree, not the close commit: the store already names the bytes a review judged, so what arrived after them is the question, and finding the close commit would mean walking history for a fact the store already keeps" (porque a verdict's tree is evidence the tool wrote and can read anywhere; the close commit would have to be guessed from the document's history, and a guess is what this gate exists to avoid)
