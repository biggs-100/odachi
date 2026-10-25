# each-lens-can

## Requirements (verbatim, from the user)
R1 · "each lens can name its model: the review runs one reviewer per model group and the verdict records who judged each lens, so a cheap model can carry the cheap lenses"

## Done when
- review.toml accepts a model per lens and refuses a lens name that is not one of the four
- review --show prints which model will judge each lens, without running anything
- review runs one reviewer per distinct model and stores one verdict that maps each lens to its model; a second run adds the missing lenses and never replaces an answered one
- check --require-review refuses only the lenses judged by the model in the proof
- the README says how to set it and what it costs

## Steps
- [x] T1 · review.toml names a model per lens, and review --show prints the plan without running anything   R1 · b0a9e7e
- [x] T2 · review runs one reviewer per distinct model and completes the verdict lens by lens, never rewriting an answered one   R1 · 6cf02c1
- [x] T3 · the identity gate refuses only the lenses the builder judged, and the README says how to set it   R1 · cf08bb4

## History
L1 · "each lens can name its model: the review runs one reviewer per model group and the verdict records who judged each lens, so a cheap model can carry the cheap lenses"
L2 · OFICIO · "The switch and the lens table share one file and one decoder" (porque review mode off rewrote the file from the mode alone, and two writers of one file would erase each other's half)
L3 · OFICIO · "The plan lists only the lenses that will run, and says who decides when nothing is assigned" (porque a line for a lens the evidence did not ask is a cost nobody approved, and claiming the session's model would be a promise the tool does not keep)
L4 · OFICIO · "A run stamps the lenses it judged, not the ones the candidate's class implies" (porque a completion run would otherwise claim lenses another run already paid for, and the store would refuse its own work)
L5 · OFICIO · "The store grows by lens and refuses a shared one" (porque the invariant is not weakened: an answer that was paid for is never replaced, only joined)
L6 · OFICIO · "A model is an identifier: NormalizeModel refuses anything a shell could read as code" (porque the model replaces {model} inside a reviewer command, and a cloned review.toml must not be able to run anything; quoting would depend on the shell, and the alphabet is knowable)
L7 · OFICIO · "The three findings are fixed together: the legacy model survives the merge, two spellings of a lens are refused, and an unreadable verdict blocks before any reviewer runs" (porque each one either lost an identity the gate needs or paid for a review that could not be stored)
L8 · OFICIO · "The identity gate only judges lenses the verdict answered, and a lens assigned to the session's model shares its group" (porque an uncovered lens must surface as coverage, not as an identity refusal; and two groups for one model would charge twice for one answer)
