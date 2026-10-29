# the-fence-stops

## What must exist
- the fence stops protecting the CLI surface, so the feature after it can change a default: the policy moves in its own commit, before the code that depends on it

## Done when
- this feature closes with no code, and the fence in force no longer names cmd/odachi/main.go
- the two rules that stay are the verdict shape and the proof format, unchanged
- the next feature can change the CLI default and CI reads a base fence that allows it
- a test drives nothing new here: the change is the policy, and check proves it by passing

## Not included (first version)
- no change to the verdict shape
- no change to the proof format, and the stored proof is never rewritten

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how a fence moves without pretending the code did.
