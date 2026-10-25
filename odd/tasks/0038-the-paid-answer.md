# the-paid-answer

## Requirements (verbatim, from the user)
R1 · "the paid answer survives a later failure: each model group's verdict is stored as it lands, and the gate still refuses a lens nobody judged"

## Done when
- a run stores each model group's answer as soon as it lands: a later group's failure leaves the earlier answers on disk
- the run that fails says how many lenses of how many were saved and which ones are missing, and names the command that completes them
- review --show calls the verdict partial when it covers only some of the lenses this tree requires
- check --require-review still blocks on a lens that no stored verdict answered
- a stored verdict still never replaces an answered lens, and a test pins that refusal
- the README says what a partial verdict is and how to finish it

## Steps
- [x] T1 · each group's answer is stored as it lands, so a later failure keeps what was paid for, and the run says saved X of Y   R1 · f9b6ca0
- [ ] T2 · review --show names a partial verdict, the store's refusal is pinned by a test, and the README explains both   R1

## History
L1 · "the paid answer survives a later failure: each model group's verdict is stored as it lands, and the gate still refuses a lens nobody judged"
L2 · OFICIO · "Each group's answer is written before the next group starts, and the store reads the tree again before each write" (porque what a model answered is already paid for, and a reviewer may move the worktree: filing the answer under the bytes it left is the only attribution the gate can check)
L3 · OFICIO · "--show says partial, with the count, only when the file covers fewer lenses than the tree requires" (porque a verdict that answered half the lenses must not read as a review: the message is what a human sees before trusting the file, and the gate will still ask for the rest)
