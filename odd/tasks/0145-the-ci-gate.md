# the-ci-gate
Base: 2e19007c4f329facdecd09de9f39eb56e7eaa649

## Requirements (verbatim, from the user)
R1 · "the CI gate stops asking to be excused from the review: the base ref it is built from names the review instead of refusing the tree"

## Done when
- the workflow runs check without --no-review, and a pull request that closes a feature says the review was not verified there
- the changelog says the flag left once the base ref could read the new rule

## Steps
- [x] T1 · the workflow runs check without --no-review and every closing pull request says so   R1 · tree:dee4551a7bac92047cb59bbebf812b82be2af6bf
- [x] T2 · the changelog says the flag left once the base ref could read the new rule   R1

## History
L1 · "the CI gate stops asking to be excused from the review: the base ref it is built from names the review instead of refusing the tree"
