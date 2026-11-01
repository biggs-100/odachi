# 0045 - The local run reads the base CI reads

Status: accepted - Date: 2026-11 - Session: after v0.11.90

## Context

The gate has two sources of policy and two ways to name them. Without a flag it obeys the
worktree's own `odd/fence.toml` and compares the change to `HEAD`. CI obeys another heap of
facts: it builds odachi from the base ref, reads the fence from that commit, and compares the
changed paths against it, because the candidate cannot be trusted to edit the policy that
judges it. `--fence-from <ref>` and `--base <ref>` are how the binary is told that, and the
workflow spells both of them.

A recipe is not an answer. On v0.11.90 the gate went red on a pull request that a local run
had passed: the branch had loosened its own copy of the fence in an earlier commit, and
`--fence-from HEAD` read that copy. The rule the project already had -a policy change lands
before the feature that needs it- was verifiable only in CI, after a push, because the local
command that looked like CI's was three flags written from memory.

## Decision

- **One flag names the question, not the argument.** `check --ci` resolves the base from the
  clone's remote and uses it for both halves: the fence is read there and the compared paths
  are measured from it. The order is `origin/HEAD`, the symref a clone writes, then
  `origin/main`, then `origin/master`, the spellings that outlive a lost symref.
- **It refuses `--base` and `--fence-from` beside it.** A flag that exists to remove an
  ambiguity cannot carry one: two names for the base would be resolved by precedence, which is
  the silence this flag was added to end.
- **It fetches nothing, and says which ref it read.** The git edge is offline on purpose -a
  gate must never wait on the network- so freshness cannot be measured here. The run prints
  `fence: <ref>` and the proof carries `# fence <ref>`; a stale base is a green about an older
  tree, visible to the human, and `git fetch origin` is the command that fixes it. Guessing at
  freshness would need the network the gate refuses to use.
- **A clone with no remote base blocks.** It names the command that would create the ref
  instead of falling back to `HEAD`, because falling back is the behaviour that lied.

## Consequences

- The local answer to CI becomes one string: `odachi check --ci --run "go test ./..."`.
- The workflow keeps `--fence-from "$BASE" --base "$BASE"`, and must: its base is the pull
  request's base commit, or `github.event.before` on a push, neither of which is necessarily
  the remote's default branch. `--ci` is the local approximation of that rule, not a
  replacement for it.
- `--ci` promises the same *ref*, not the same *age*. A run whose `origin/main` is behind the
  remote passes where CI will fail, and the output says which ref judged it, so the remedy is
  the fetch nobody ran.
- Nothing about the review changes: CI still has no store, and a verdict is still evidence
  about one tree held inside the clone (0042).
