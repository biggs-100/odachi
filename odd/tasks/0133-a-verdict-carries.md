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
- [x] T1 · the inherited rows are bounded by the candidate's paths: a row about an untouched path does not travel   R1 · tree:35f1003e2c26a7a9ac3b38d8dfd52301f0ccdb5a
- [x] T2 · the gate still answers every changed path, and three reviews in a row leave verdicts that do not grow   R1 · tree:2c055c5dbfa0f4d4055b3681a87e0e0db7f65616
- [x] T3 · the changelog and the ADR say the bound, and the impact map names the package   R1 · tree:7b1b81e11003eb2106bc8042e3b51904bc8d2b5f

## History
L1 · "a verdict carries the rows of the paths it judged and nothing else, so its size does not grow with the store"
L2 · OFICIO · "the review ran with the tree's own binary, not the installed release" (porque this feature changes what a review writes, and the installed v0.11.73 is one version behind; judging the new store with the old code would have measured nothing)
L3 · OFICIO · "the verdict the old binary wrote was deleted and attached again by the new one" (porque the attach inherited 767 rows with the release's logic and stored 125 KB, the exact defect this feature removes; the reviewer's answer did not change, so nothing was paid twice)
L4 · OFICIO · "the fenced answer was attached after its JSON was extracted" (porque the reviewer answered the verdict inside a markdown fence; the answer is the JSON, and the tool keeps a refused answer so it never has to be paid again)
