# a-change-reaches

## Requirements (verbatim, from the user)
R1 · "a change reaches main only through a pull request whose gate passed"

## Done when
- the ruleset in .github/main-ruleset.json is applied by scripts/protect-main.ps1, idempotently, and prints what it enforced
- a direct push to main is rejected by the remote
- a pull request with a red suite cannot merge; a green one can
- the ADR names what still does not hold: a pull request can edit the workflow it runs under, and CODEOWNERS only asks

## Steps
- [x] T1 · T1 - .github/main-ruleset.json and scripts/protect-main.ps1, idempotent   R1 · faabb59
- [x] T2 · T2 - applied: a direct push is rejected and a red pull request cannot merge   R1
- [x] T3 · T3 - README and ADR 0024: what protects main, and what still does not   R1 · b4f3e38

## History
L1 · "a change reaches main only through a pull request whose gate passed"
L2 · OFICIO · "The ruleset requires zero approvals" (porque a solo maintainer cannot approve their own pull request; one required approval would block the only merge)
L3 · OFICIO · "The policy is a JSON file applied by name" (porque a settings change outside the repository cannot be reviewed or repeated)
L4 · OFICIO · "The required checks are the job names in gate.yml" (porque a required check with no run behind it keeps every pull request blocked, so the duplicate is on purpose)
