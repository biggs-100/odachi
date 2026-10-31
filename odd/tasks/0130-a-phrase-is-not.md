# a-phrase-is-not
Base: fc4d0fa9c0a8084718ff897890a0f762857ae3a7

## Requirements (verbatim, from the user)
R1 · "a rule names paths, not a phrase: a glob of the phrase's own words matches files named like the phrase and protects nothing"

## Done when
- a non-goal that arrives as free text writes no rule in odd/fence.toml, and the file carries it as a template the human completes
- a preset non-goal keeps its rule with its catalog paths and deps
- the run says the phrase is a template and where to write its paths, and no longer claims the fence wrote a glob
- a project with no non-goals keeps the file that teaches what a rule looks like

## Steps
- [x] T1 · a phrase writes a template, not a rule, and the run says so   R1 · tree:1eb29ccae848060c4629c04f710496035374a719
- [ ] T2 · the ADR and the changelog say a phrase is not a rule   R1

## History
L1 · "a rule names paths, not a phrase: a glob of the phrase's own words matches files named like the phrase and protects nothing"
L2 · OFICIO · "a rule is written only when somebody gave it paths, and a phrase becomes a template in the fence" (porque the run invented a glob of the phrase's own words, the exact thing the fence header forbids, and the file carried a policy that could not refuse anything)
