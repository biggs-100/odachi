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
- [ ] T1 · review.toml names a model per lens, and review --show prints the plan without running anything   R1
- [ ] T2 · review runs one reviewer per distinct model and completes the verdict lens by lens, never rewriting an answered one   R1
- [ ] T3 · the identity gate refuses only the lenses the builder judged, and the README says how to set it   R1

## History
L1 · "each lens can name its model: the review runs one reviewer per model group and the verdict records who judged each lens, so a cheap model can carry the cheap lenses"
