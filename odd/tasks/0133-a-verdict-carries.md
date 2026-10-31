# a-verdict-carries
Base: d787d15600230d1699eb242b15d221b50300c376

## Requirements (verbatim, from the user)
R1 · "a verdict carries the rows of the paths it judged and nothing else, so its size does not grow with the store"

## Done when
- a review of a candidate carries at most one row per changed path per lens, so a verdict stops growing with the store
- the gate still finds a current answer for every changed path, and an earlier verdict's row is still inherited when its path is in the candidate
- the prompt still names the rows of the changed paths, and a store written by the previous release is read unchanged: no verdict is rewritten
- three reviews in a row over one file leave three verdicts whose size does not grow

## Steps
- [ ] T1 · the inherited rows are bounded by the candidate's paths: a row about an untouched path does not travel   R1
- [ ] T2 · the gate still answers every changed path, and three reviews in a row leave verdicts that do not grow   R1
- [ ] T3 · the changelog and the ADR say the bound, and the impact map names the package   R1

## History
L1 · "a verdict carries the rows of the paths it judged and nothing else, so its size does not grow with the store"
