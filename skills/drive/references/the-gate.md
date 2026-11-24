# What the gate answers

## What a block means

- `0` ready: continue. `1` you can act: the `next:` line names the exact command; run it.
- `2` only the human can unblock: do not retry, do not go around, ask with `ask`.

## The local close

With `odachi review mode on`, a non-passive local close blocks without a valid approved
review verdict for this tree. A named `Risk:` also blocks without independent verification.
On non-passive code, declare `Risk: none - <written reason>`; missing, unknown or bare `none` is invalid.

1. `odachi review --show` inspects this tree's verdict and lenses. It runs no reviewer.
2. `odachi review --exec "<command>"` runs the real reviewer and stores its verdict.
   Printing a prompt or showing an old verdict is not approval. A causal `BLOCKER`
   blocks: fix it or `--refute` citing the diff. One correction owes one re-review;
   a `WARNING` or `SUGGESTION` goes to the task doc and earns no round.
3. `odachi verify --exec "<command>"` runs the independent verifier and stores its
   record for this tree. A named risk cannot close without that record when mode is on.
4. `odachi check` reads the verdict, required verification and proof before the commit.

The verdict and verification belong to that tree: a correction invalidates them.
Mode off leaves review and verification optional. `odachi check --no-review` bypasses one run;
CI has no clone store and makes no model calls.
Review approval never authorizes push, pull request, merge or release.
