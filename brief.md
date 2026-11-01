# the-fence-lets-a-lens-grow

## What must exist
- the fence stops protecting the four lens prompts, so a feature that improves what a lens looks at can pass CI, and the review's own engine stays fenced

## Done when
- odd/fence.toml protects the stored proofs and the review's own question -the prompt builder and the refutation rule- and names no lens asset
- a change under internal/assets/review/** passes the gate with this tree as the base, which is what the pull request that follows needs

## Not included (first version)
- no change to the lens content itself: a lens is what a later feature edits, and this lands only the policy that lets it
- no change to the gate: the fence is data the gate reads, not code
- no new verb and no new flag

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why a policy lands before the change that needs it, and not in the same pull request.
