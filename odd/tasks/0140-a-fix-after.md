# a-fix-after
Base: 48157fa50456721b42b34012d354b50aa18b2da7

## Requirements (verbatim, from the user)
R1 · "a fix after a verdict is reviewed as the delta, and closing a feature names the review debt it leaves"

## Done when
- a fix after a verdict is reviewed as the delta, not as the whole candidate
- the stored verdict still answers every changed path after a correction
- closing a feature names the open rows of docs/review-debt.md
- the ADR and the changelog say why the review gate lives in the clone

## Steps
- [ ] T1 · review --since <tree> builds the correction prompt from the delta since that verdict   R1
- [ ] T2 · the correction's verdict carries the predecessor's rows and still answers every changed path   R1
- [ ] T3 · check names the open rows of docs/review-debt.md when the feature closes   R1
- [ ] T4 · the ADR and the changelog say why the gate is local and what a correction costs   R1

## History
L1 · "a fix after a verdict is reviewed as the delta, and closing a feature names the review debt it leaves"
L2 · OFICIO · "T1 and T2 land in one commit: the delta and the predecessor's rows are one behaviour" (porque a correction that reviews only the delta without carrying the verdict it completes leaves paths unanswered, and a gate that refuses it is not a feature; the two are tested together and reverted together)
