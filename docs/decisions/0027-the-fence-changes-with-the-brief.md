# 0027 - The fence changes with the brief

Status: accepted · Date: 2026-10 · Session: after v0.11.9

## Context

The gate reads the fence from the base ref (`--fence-from`), so the change that
asks to pass is judged by the policy it must obey. A review of feature 0030 found
what that does not cover: the candidate's own `.odachi/fence.toml` is never
compared against the base's. In a scratch copy, emptying the fence kept the gate
green. After the merge, the empty fence would be the policy of the next change.

The first design was monotonicity: every base rule must survive in the candidate
with at least its paths and dependencies. It is wrong for this project, and the
reason is the workflow: `odachi start` rewrites `brief.md` and the fence together
for each feature. A new feature legitimately replaces the previous non-goals, so
monotonicity would block every feature after the first. Feature 0030's fence has
three rules; 0031's has four, and none of the three survived.

## Decision

- **The fence may only be loosened by a change that also changes the non-goals of
  the brief.** When `--fence-from` is set, the gate compares the non-goals of both
  briefs. If the lists are the same, every base rule must survive with at least
  its paths and dependencies (adding rules or paths is stricter and passes). If
  the non-goals changed, the fence may change with them.
- **The brief's non-goals decide, not the fence.** A new vision is a new policy;
  the same non-goals keep the policy they had. `odachi start` writes both, which
  is why a feature passes and a fix cannot move the fence it must obey.
- **The comparison is exact.** A glob is a string here: a changed pattern is a
  changed policy, not a smarter one. Rule identity is the rule's own text, so a
  renamed rule is a removed rule.
- **An unreadable candidate fence blocks.** With `--fence-from` the candidate's
  copy is not the policy today, but it is the policy after the merge. A file the
  tool cannot read must stop the change instead of travelling silently.
- **Whitespace, comments and a changed outcome are not a new policy.** The brief
  parser trims each line and ignores unknown sections; the comparison sorts the
  non-goals as a set. Order, whitespace, comments and a changed outcome never
  decide a policy.
- **A ref with a fence and no brief is not a licence.** The comparison fails
  closed: without a brief to compare, the fence is still compared and cannot be
  loosened. Only a ref with no rules at all has nothing to protect.
- **Without `--fence-from` nothing is compared.** The worktree fence is the
  policy, exactly as before.

## Alternatives rejected

- **Plain monotonicity.** Blocks every new feature, because `start` rewrites the
  fence; the gate would be red for the normal case.
- **A cumulative fence.** `start` would merge the non-goals of every feature and
  old entries would never expire; it needs a migration and grows a list nobody
  reads.
- **Comparing the fence alone, with no brief.** That is the hole: emptying the
  fence is a loosening no rule can name.
- **Trusting `CODEOWNERS` and the merge decision.** They are a request, and with
  zero required approvals they are not even that (ADR 0024). The gate is the part
  that can be enforced.
- **Signing the fence.** It proves who wrote the bytes, not that the policy is
  the same one the brief describes.

## Consequences

- A bug-fix pull request cannot touch the fence: its brief is main's brief, and
  the gate refuses a dropped rule.
- A feature pull request can rewrite both, which is the workflow. A hostile
  change could rewrite the non-goals too; that diff is the vision, it is the
  maintainer's to read, and no mechanical rule can replace that reading.
- A maintainer who edits the fence by hand after `start` can still add rules
  (stricter) but not remove one without touching the brief.
- The check adds one read of the base brief and one of the candidate's per run
  when `--fence-from` is set; both are small files.
- `fence.Loosened` is a pure function, so the rule is testable without git.

## Sources

- Feature 0030's review: the empty-fence probe that stayed green.
- The per-feature rewrite in `start` and the fences of 0030 and 0031.
