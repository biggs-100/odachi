# Next session

Written after 0344, the cycle that closed four debts and made the tool guide the close.

**Standing rule.** When a problem or a design question shows up, read
`C:\Users\USER\Desktop\herramientas\gentle-shell` first, read-only: it may already
answer it, and 0339 is the proof -the answer was the opposite of the assumption and it
saved a day. Reading is evidence, not permission to edit that folder or run odachi there.

## Where the project is

- **Four features closed, four PRs, no release.** v0.58.0 (#692, tag on `64b610a`) carried
  0340. Then 0341 (`1dfb499`, #693) names a struct field in the undo row; 0342
  (`aa8077a`, #694) gives the hub a limits section for the four review scalars; 0343
  (`281b2f4`) makes the help footer follow an open picker or editor; 0344 (`7ca0acb`,
  #695) stages the task doc at the close and names the command that regenerates a
  stale proof. `## Unreleased` carries 0341-0344 and waits for a release decision.
- **The workflow defect was the point of 0344.** The human asked why every feature ran
  the full suite twice. The gate runs a proof and its reverted copy by design, but the
  agent ran both twice because `plan --done` moved the tree after the proof and `check`
  suggested `odachi check --proof <file>` -a command that reuses a proof, not one that
  regenerates it. Now the close stages its doc and names `check --run`, and the hook's
  stale hint falls back to the command named by the proof on disk.
- **The order that works, verified live.** Close the step with `plan --done`, then run
  `odachi check --run "go test ./..."` once on the tree that includes the check-off,
  then commit. The first version of 0344 missed the hook, which calls `check --message`
  without `--run`; the reliability reviewer found it, and the correction was proven when
  the hook switched from `next: odachi check --proof <file>` (installed binary) to
  `next: odachi check --run "go test ./..."`. A commit after that passed the hook without
  `--no-verify`.
- **The tool counts this line for the tree it describes, copied from `odachi summary`:**

  total: 342 features · 856 steps done of 856 · 0 open questions

## What waits for a decision

- **A release.** `## Unreleased` carries four features; v0.59.0 waits for the human.
- **The second vendor.** The verdicts of 0335 to 0344 were written by models close to
  the ones that produced their proofs, and `check` says so on every close. The doctor's
  check 17 still counts `1 of 4`. 0344's last review used `deepseek-v4-pro` for a second
  model, not a second vendor.
- **Test adequacy has no evidence** (debt 6): nothing measures changed-line coverage or
  mutation on the candidate.
- **The `[[check]]` array has no editor.** 0342 left it out on purpose: it carries the
  `CRITICAL` severity of the gate, and a picker that could lower it would change policy
  without an explicit decision. Who may edit it needs a decision and probably an ADR.
- **Two small leftovers of 0342.** `check_timeout_s` is shown even with no `[[check]]`
  declared, although the codec only writes it when checks exist. The limits write lives
  on `tuiRun`/`apply` rather than `configure.Runner`, because `configure.go` was outside
  the worker's allowed edit surface; moving it there would need a new surface.
- **The undo row can still be fooled by a fixture.** A test whose quoted output contains
  the literal `[build failed]` makes `Classify` read a test failure as a build failure;
  0341's own row stayed `unknown` for that reason. It never blocks and is a pre-existing
  limit of 0300.

## What this cycle learned

- **Audit the workflow, not the intention.** The agent had the right order in the
  handover and ignored it three times. Discipline does not survive lost context; a tool
  that makes the right order the default does. The human asked for an audit of the agent
  workflow in future sessions; its observable criteria are one local `check --run` per
  feature, a close that stages its doc and names the next command, a stale hint that
  names `--run` when the hook saw no `--run`, and no worker rerunning packages the proof
  already covers.
- **Answer a question before acting on it.** The human reported that sometimes a
  question in their message was ignored and the agent went straight to writing code.
  `AGENTS.md` now says it explicitly: when the message contains a question, answer it in
  the chat before any tool call or file write, and treat the answer as information, not
  as permission to act. The tool cannot enforce this -it never sees the chat message-
  so it depends on the auditor checking that the answer came first.
- **A reviewer's refusal can carry a real finding.** 0344's first review was refused
  because the declaration named four paths outside the candidate. Inside that refused
  file was the exact defect in the fix. Do not edit a verdict to pass a gate; fix the
  code and review the new tree, with a different model when the first one misdeclares.
- **Build the tree binary into the temp folder and put it first on PATH.** The installed
  `odachi` is the last release, not the code in the worktree; it prints the old hint and
  silently lies about the tree's behavior. This is why the hook blocked a correct commit
  until the tree binary led PATH.
- **Measure the debt before opening it.** `review.toml` has nine keys and `SetConfig` had
  a setter for three; the "other half of the hub" was a section, not a flag. Before
  that, the undo debt was one function reading the diff, not a new policy.
- **A close past 400 lines may owe no verifier.** 0342 closed at 714 authored lines and
  the gate said so explicitly: this step owes no verifier and closes anyway. Do not
  invent a `Slice:` acceptance the method does not ask for.

## The commands that work

```sh
# the close, then ONE proof on that tree, then the commit -in that order
odachi plan --done T1 --commit <sha>
odachi check --run "go test ./..." --timeout 1800
git commit -m "..." -m "Risk: none"
```

```sh
# build the tree's binary and lead PATH before anything that must know this tree
go build -o "$env:TEMP\odachi-next-bin\odachi.exe" ./cmd/odachi
$env:PATH = "$env:TEMP\odachi-next-bin;$env:PATH"
```

```sh
# the reviewer, and a correction round against the approved tree
odachi review --base <feature base> --lens reliability --model <provider/model> \
  --exec '"%TEMP%\odachi-rev2.bat" {model} {prompt} "%TEMP%\rev-out.txt" "%TEMP%\rev.json"'
```

```sh
# the verifier, whose model comes from odd/review.toml
odachi verify --exec '"%TEMP%\odachi-ver3.bat" {model} {prompt} "%TEMP%\ver-out.txt" "%TEMP%\ver.json"'
```

```sh
# the release: a branch that renames Unreleased, a PR, then the tag on its merge
git switch -c release/vX.Y.Z && git commit -am "release vX.Y.Z" --no-verify && gh pr create --base main
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
irm https://github.com/biggs-100/odachi/releases/download/vX.Y.Z/install.ps1 | iex
```

## The limits to remember

- **The gate runs the proof and its reverted copy** (`internal/app/check/check.go:2361-2398`);
  `cmd/odachi` alone costs 212-234s here, so one `check --run` is about ten minutes.
- **A worker's full-package rerun is not evidence the proof owes.** Focused RED/GREEN is
  enough for the writer; the proof covers the whole suite once.
- **The reviewer's clock is 600s and the verifier's is 1800s**, with a 120-second silence
  budget on both. A reviewer reading a large candidate can be cut at 600s and write
  nothing; the run says which clock stopped it.
- **A feature's base is a commit the remote has**, and a squash erases a branch-only
  base: rebase before retargeting, or CI answers `bad object`. `git rebase --onto` takes
  the **old** base as the upstream.
- **The fence is read from the base**: a policy change lands first, in its own pull request.
- **`## Unreleased` is re-created after every release**, and a version is never moved or
  reused: the tag lands on the merge of the release pull request.
- **A document commit after the close leaves local `check` red**, on purpose: CI
  regenerates its own proof from a clean checkout. The release commit is made with
  `--no-verify` for that reason.
- **The proof names the index tree**, and `proof/` is gitignored, so the closing tree never
  carries it: the link is the `# tree` header.
- **`windows-latest` is the slow gate**: 5-12 minutes in this cycle; 11m59s on #695.
- **One heavy gate per machine**: `go test ./...` takes about twelve minutes here, twice
  per `check --run`.
- **The wrappers live in temp**: `odachi-rev2.bat`, `odachi-ver3.bat` and `odachi-cut.ps1`.
  If the folder is emptied, compose them again.
