# the-ci-gate

## What must exist
- the CI gate stops asking to be excused from the review: the base ref it is built from names the review instead of refusing the tree

## Done when
- the workflow runs check without --no-review, and a pull request that closes a feature says the review was not verified there
- the changelog says the flag left once the base ref could read the new rule

## Not included (first version)
- no change to the proof format, and the stored proof is never rewritten
- no change to the bundled skills

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: por que una regla de CI espera al release que la entiende.
