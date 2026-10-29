# a-run-that
Base: 3de974936cd6915308410a42ace55366f7d1eb41

## Requirements (verbatim, from the user)
R1 · "a run that asks for a lens the evidence did not require writes it beside the verdict, so a later reader sees the scope was raised and a replay cannot drop it"

## Done when
- a run whose --lens added a lens the evidence did not ask writes reviews/<tree>.scope.json beside the verdict, naming what the evidence asked, what was added and the flag as it was given
- a run that added nothing writes no record, and a second run for the same tree does not rewrite the store
- --show and the prompt say the scope was raised, and the record is read by tree, so a verdict never inherits another tree's scope
- a test drives each answer, and one of them fails without the change

## Steps
- [x] T1 · the domain names the scope record: what the evidence asked, what was added, the flag and the instant, and its file   R1 · tree:ef4a6fe731cee6b0f524e1f706c714b481920530
- [x] T2 · a run that adds a lens writes the record beside the verdict, and a run that adds none writes nothing   R1
- [x] T3 · review --show and the prompt say the scope was raised, and the record belongs to the tree that asked   R1 · tree:bad9ac6267f466f293d6d31992d5fc22e809dbb3
- [x] T4 · the debt row leaves the file, and the changelog says why   R1 · tree:6958bb1c65ff3596d702ba92d8eabee03fd2a656

## History
L1 · "a run that asks for a lens the evidence did not require writes it beside the verdict, so a later reader sees the scope was raised and a replay cannot drop it"
L2 · OFICIO · "the step about the report was written by hand: plan refused text that starts with a flag name" (porque odachi plan reads a step that starts with -- as its own flag; the step now says 'review --show' so the next call cannot be misread, and the tool still owns the numbering)
L3 · OFICIO · "the verdict filed with the wrong base was deleted, not kept: it judged the diff since the plan commit instead of the diff since the feature's base, and the gate named it as describing another diff" (porque a verdict is evidence about the bytes it judged; mine judged the wrong range, so keeping it would preserve a file the gate refuses as if it were an answer)
