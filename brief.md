# each-lens-can

## What must exist
- each lens can name its model: the review runs one reviewer per model group and the verdict records who judged each lens, so a cheap model can carry the cheap lenses

## Done when
- review.toml accepts a model per lens and refuses a lens name that is not one of the four
- review --show prints which model will judge each lens, without running anything
- review runs one reviewer per distinct model and stores one verdict that maps each lens to its model; a second run adds the missing lenses and never replaces an answered one
- check --require-review refuses only the lenses judged by the model in the proof
- the README says how to set it and what it costs

## Not included (first version)
- per-lens prompts: the four lens prompts already exist
- per-lens refuters or validators
- model discovery, pricing or autodetection
- changing the lens vocabulary
- a UI or a TUI

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how a review can be split by model without splitting the verdict.
