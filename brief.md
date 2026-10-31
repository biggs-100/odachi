# a-verdict-carries

## What must exist
- a verdict carries the rows of the paths it judged and nothing else, so its size does not grow with the store

## Done when
- a review of a candidate carries at most one row per changed path per lens, so a verdict stops growing with the store
- the gate still finds a current answer for every changed path, and an earlier verdict's row is still inherited when its path is in the candidate
- the prompt still names the rows of the changed paths, and a store written by the previous release is read unchanged: no verdict is rewritten
- three reviews in a row over one file leave three verdicts whose size does not grow

## Not included (first version)
- no change to the proof format, and the stored proof is never rewritten
- no change to the bundled skills

## Decisions taken
- It is used from the phone and the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: por que acotar por la pregunta que se hace sale mas barato que deduplicar.
