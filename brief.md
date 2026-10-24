# a-change-reaches

## What must exist
- a change reaches main only through a pull request whose gate passed

## Done when
- the ruleset in .github/main-ruleset.json is applied by scripts/protect-main.ps1, idempotently, and prints what it enforced
- a direct push to main is rejected by the remote
- a pull request with a red suite cannot merge; a green one can
- the ADR names what still does not hold: a pull request can edit the workflow it runs under, and CODEOWNERS only asks

## Not included (first version)
- required approvals: a solo repo cannot approve its own pull request
- signed commits, linear history, merge queue
- org-wide policy, identity, tag protection

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: what a required check really guarantees.
