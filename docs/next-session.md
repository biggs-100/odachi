# Next session

Written after v0.60.0 shipped 0345 and 0346: the cycle that closed the last hub
debt, reverted a costly one, and published the release.

**Standing rule.** When a problem or a design question shows up, read
`C:\Users\USER\Desktop\herramientas\gentle-shell` first, read-only: it may already
answer it, and 0339 is the proof -the answer was the opposite of the assumption and it
saved a day. Reading is evidence, not permission to edit that folder or run odachi there.

## Where the project is

- **main is `a377f5c`, clean and synchronized.** v0.58.0 shipped 0340 on merge `64b610a`;
  v0.59.0 shipped 0341-0344 on merge `2de2122`. Then 0345 (`80fffac`, #699) made `ask`
  validate answers against a stored choice envelope, and 0346 (`f174bdf`, #700) gave the
  wizard a section to edit a declared check's name, command and remedy. v0.60.0 shipped
  both on merge `a377f5c` (PR #702, tag `v0.60.0`, 16 signed assets, release workflow
  green). `## Unreleased` does not exist on main right now: the release renamed the
  heading, and the next feature writes under a new one.
- **The installed binary is 0.60.0 and matches main.** It knows 0345 and 0346. For a
  future feature, still build that tree's binary into a temp folder and lead PATH before
  any command that must know its behavior: a release binary is an artifact of a tag, not
  a moving pointer to main.
- **0347 was opened, reverted and never published.** It added an advisory coverage row
  that ran a second `go test ./... -coverprofile` on every close. `cmd/odachi` alone costs
  4-5 minutes here; with instrumentation it ran past an hour and had to be stopped.
  The branch was deleted locally; no PR, no remote branch, no traces in the tree.
- **This handover travels as PR #701**, still open when last checked. Its first version
  said v0.60.0 waited for a decision; this revision corrects that.
- The tool counts this line for the tree it describes, copied from `odachi summary`:

  total: 344 features · 858 steps done of 858 · 0 open questions

## What waits for a decision

- **The second vendor.** Verdicts of 0335-0346 were written by models near the ones that
  produced their proofs; doctor's check 17 still counts `1 of 4`. 0346 used a different
  provider (`opencode/deepseek-v4.1-flash-free`) when `opencode-go` hit its monthly usage
  limit: a second model, not a second vendor.
- **Test adequacy stays unresolved (debt 6).** The undo row reverts behavior and the
  mutation row probes a hand-named string; neither reports line execution. A coverage
  row cost over an hour here, so the debt now needs a decision about cost before it is
  reopened. Do not rebuild 0347 without solving that first.
- **The `[[check]]` editor is done but narrow.** 0346 edits name, command and remedy;
  severity stays hand-edited because it fixes the verdict. Editing the severity, or
  adding/removing a check, is still ADR territory.
- **The undo row can be fooled by a fixture** whose output contains literal
  `[build failed]`: `Classify` reads it as a build failure and answers `unknown`, which
  never blocks. A pre-existing limit of 0300, named in 0341's own close.
- **The macOS TUI frame test is racy (pre-existing, 0342).** `limitsTerminalFrame` in
  `cmd/odachi/tui_teatest_test.go:122-128` waits for the word `límites`, which is already
  a row in the hub menu, so it can capture the menu frame before the section renders.
  It failed once in v0.60.0's gate, passed on rerun, and never touched the release's
  one-line change. Fixing the test is a feature of its own; until then a rerun is the
  workaround. Decide whether to spend a feature on it.
- **Merging this handover PR #701** is still the human's call.

## What this cycle learned

- **An advisory signal that costs an hour is not advisory.** 0347 ran a full instrumented
  suite on every close. The right question before building such a row is who pays for it
  on each run, not whether the number is interesting.
- **A green proof and an approved verdict are not evidence that a row worked.** In 0347
  the proof carried no `adequacy:` line and no `.adequacy.txt`; both the suite and the
  reliability reviewer passed. Only looking at the evidence file by hand found it. The
  audit the human asked for must include checking the artifact a feature claims to write.
- **One stale binary fooled two features.** The temp binary from 0344 was reused for 0345,
  0346 and 0347. It did not contain the newer code, so its output could not describe those
  trees. Rebuild the binary per feature; never reuse a temp build by name.
- **Correct the wording before writing code.** In 0346 the parent claimed "no re-approval
  forced" in a choice label; verification showed name and command are in the approval
  fingerprint (`check.go:318-321`). The task doc was corrected before the worker started,
  and the apply notice now distinguishes fingerprint edits from remedy edits.
- **Reviewers find real defects; do not Wave them through.** 0346's first verdict was
  `changes_requested` on a CRITICAL finding: resolving edit identities against a slice
  being mutated let a rename redirect a later edit to the wrong row. Fixed against the
  original list, re-reviewed, then merged. A refusal on declaration grounds can still
  contain that finding -read it before rebuilding the review.
- **Answer the human's question before acting.** `AGENTS.md` now says it explicitly: a
  question in the message is answered in chat before any tool call or file write, and the
  answer is information, not permission. PR #698.
- **A red gate is not automatically a defect of the candidate.** v0.60.0's macOS gate
  failed on a test race that existed since 0342. The proof artifact named the failing
  test and the frame it captured; reading it showed the word the test waited for was
  already in the menu. Re-running that one job passed, and the release went out with a
  one-line changelog change. Read the artifact and separate harness flake from product
  defect before touching code in a release PR.
- **A workflow run must finish before a failed job can be rerun.** GitHub answered 403 to
  a rerun while `windows-latest` was still in progress. Wait for every job to conclude,
  then rerun the single job by id.
- **A release commit needs its file staged.** `git commit -m "..." --no-verify` without
  `git add` left `CHANGELOG.md` uncommitted; the push carried no commits and
  `gh pr create` failed with "No commits between main and release/v0.60.0". Stage first,
  or use `-am`.

## The commands that work

```sh
# a close, then ONE proof on that tree, then the commit -in that order
odachi plan --done T1 --commit <sha>
odachi check --run "go test ./..." --timeout 1800
git commit -m "..." -m "Risk: none"
```

```sh
# build the tree binary per feature and lead PATH before anything that must know this tree
go build -o "$env:TEMP\odachi-next-bin\odachi.exe" ./cmd/odachi
$env:PATH = "$env:TEMP\odachi-next-bin;$env:PATH"
```

```sh
# the reviewer; if one provider hits its usage limit, switch PROVIDER, not the method
odachi review --base <feature base> --lens reliability --model <provider/model> \
  --exec '"%TEMP%\odachi-rev2.bat" {model} {prompt} "%TEMP%\rev-out.txt" "%TEMP%\rev.json"'
```

```sh
# the release: a branch that renames Unreleased, a PR, then the tag on its merge
git switch -c release/vX.Y.Z && git commit -am "release vX.Y.Z" --no-verify && gh pr create --base main
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
irm https://github.com/biggs-100/odachi/releases/download/vX.Y.Z/install.ps1 | iex
```

```sh
# rerun one failed job, only after every job in the workflow run has concluded
gh api -X POST repos/biggs-100/odachi/actions/jobs/<job id>/rerun
```

## The limits to remember

- **The gate runs the proof and its reverted copy** (`internal/app/check/check.go:2361-2398`).
  `cmd/odachi` alone costs 212-307s here; one `check --run` is about ten minutes. Never add
  a second full suite to that path.
- **The reviewer's clock is 600s and the verifier's is 1800s**, with a 120-second silence
  budget on both. `deepseek-v4-pro` cut out on a monthly usage limit; a flash model from
  another provider completed the review.
- **A feature's base is a commit the remote has**, and a squash erases a branch-only base:
  rebase before retargeting. `git rebase --onto` takes the old base as the upstream.
- **The fence is read from the base**: a policy change lands first, in its own pull request.
- **`## Unreleased` is re-created after every release**, and a version is never moved or
  reused: the tag lands on the merge of the release pull request.
- **A document commit after the close leaves local `check` red**, on purpose: CI
  regenerates its own proof from a clean checkout. Release commits use `--no-verify`.
- **The proof names the index tree**, and `proof/` is gitignored, so the closing tree never
  carries it: the link is the `# tree` header.
- **`windows-latest` is the slow CI gate**: 4-12 minutes in this cycle.
- **The wrappers live in temp**: `odachi-rev2.bat`, `odachi-ver3.bat` and `odachi-cut.ps1`.
  If the folder is emptied, compose them again.
- **A first red gate deserves its artifact before its fix**: the macOS failure in v0.60.0
  was a race in an older test, and the rerun passed. A rerun is cheaper than a new feature.
