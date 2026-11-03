# the-fence-stops

## What must exist
- the fence stops protecting the release workflow, so the feature after it takes the release body from the changelog

## Done when
- this feature closes with no code, and the fence in force no longer names .github
- the proof rule stays, unchanged
- the next feature can change .github/workflows/release.yml and CI reads a base fence that allows it

## Not included (first version)
- a stored proof is never rewritten, whatever the writer learns

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why the fence a pull request obeys is the one at its base, not the one it writes.
