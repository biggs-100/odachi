# Next session

Written after 0340, the cycle that answered what the limit counts.

**Standing rule.** When a problem or a design question shows up, read
`C:\Users\USER\Desktop\herramientas\gentle-shell` first, read-only: it may already
answer it, and 0339 is the proof -the answer was the opposite of the assumption and it
saved a day. Reading is evidence, not permission to edit that folder or run odachi there.

## Where the project is

- **0340 closed, one PR.** #690 squashed as `60e88da`. A task document may carry
  `Slice: <lines> - <reason>` in its preamble, where `Base:` and the waivers already
  live: the close stops refusing the size, answers with the record the verifier still
  owes, and prints the accepted number and the reason the human wrote. What counts did
  not change -docs, skills and tests are authored lines- and ADR 0055 says why: an
  exclusion is the code-golf the method forbids, and the second answer gentle-shell
  already had -record the reason- is now ours.
- **The measurement is what decided it.** 0339's 438 lines were 221 tests, 113 code, 59
  the decision record, 39 the brief and the task document, and 6 changelog and skills.
  The note blamed documentation; the diff said the tests the method demands were half
  the slice. Measure before designing, even when the note writes the design for you.
- The proof of the closing tree is `proof/0340-el-cierre-con.txt` on `a3a5d4c0`; the
  verdict is `approved` on `99efaa19`, the tree the code lives in, reliability, first
  round; the independent verification names both halves of R1 on `a3a5d4c0`. The last
  commit only names the change in the worklog, and `check` asks no verdict for it.
- The tool counts this line for the tree it describes, copied from `odachi summary`:

  total: 338 features · 852 steps done of 852 · 0 open questions

## What waits for a decision

- **The declared fields of `odd/review.toml`** -`max_prompt_kb`, `checks`,
  `check_timeout`, `keep_verdicts`, `keep_prompts`- which every write preserves and no
  command sets. 0336 named it as the other half of the hub; it is still open.
- **The footer does not change while a picker or an editor is open.** It still says
  `enter cambiar · esc volver` although the keys under an open list are the arrows,
  `enter` and `esc`. 0338 left it as a separate decision.
- **The second vendor.** The verdicts of 0335 to 0340 were written by the model that
  produced their proofs, and `check` says so on every close. The doctor's check 17 still
  counts `1 of 4`.
- **Test adequacy has no evidence** (debt 6): nothing measures changed-line coverage or
  mutation on the candidate.
- **The undo row answers `unknown` for a change whose behavior is a struct field.**
  `undone.Revertable` reverts production files and keeps `_test.go`, so the reverted copy
  fails to build -the right outcome- but `declaredName` reads `func`, `type`, `const`
  and `var` and not fields, so the build that names `SliceAccepted` finds no added name
  and the row stays unknown. It never blocks, and it is a defect of 0300, not of 0340.

## What this cycle learned

- **The note is a hypothesis, and the diff is the measurement.** 0339's handover asked
  "what does the number count" as if the answer were "not the docs"; one `git diff
  --numstat -M` moved the question to the tests, and that changed the answer.
- **`Render` must write every decision `Parse` reads.** The package promises
  `Parse(Render(d)) == d`, and a line read but never written is deleted by the next
  `plan`. Only a round-trip test caught it, and the reviewer did not.
- **A stale binary lies quietly.** `odachi.exe` in the repository root is a `dev` build
  and did not print 0339's closing sentence that `main` already carries: build the tree's
  binary into the temp folder before any command that has to know the tree's behavior.
- **The close can land on the same tree.** Committing `plan --done`'s check-off produced
  exactly the tree the proof, the verdict and the verification named, so the second suite
  run was the only cost of writing the worklog late -not a third.
- **The worklog line belongs to the step's commit.** Written after the close, it moved
  the tree and cost a second proof, a second verdict and a second verification.

## The commands that work

```sh
# the gate and the proof: two suite runs, about thirty minutes here
odachi check --run "go test ./..." --timeout 1800
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
# the step's new line, on a real repository
odachi plan --done T1 --commit <sha>
```

```sh
# the release: a branch that renames Unreleased, a PR, then the tag on its merge
git switch -c release/vX.Y.Z && git commit -am "release vX.Y.Z" && gh pr create --base main
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
irm https://github.com/biggs-100/odachi/releases/download/vX.Y.Z/install.ps1 | iex
```

## The limits to remember

- **The reviewer's clock is 600s and the verifier's is 1800s**, with a 120-second silence
  budget on both. A reviewer reading a large candidate can be cut at 600s and write
  nothing; the run says which clock stopped it.
- **A feature's base is a commit the remote has**, and a squash erases a branch-only
  base: rebase before retargeting, or CI answers `bad object`. `git rebase --onto` takes
  the **old** base as the upstream.
- **The fence is read from the base**: a policy change lands first, in its own pull request.
- **`## Unreleased` is re-created after every release**, and a version is never moved or
  reused: the tag lands on the merge of the release pull request.
- **A document commit after the close leaves the local `check` red**, on purpose: CI
  regenerates its own proof from a clean checkout. The release commit is made with
  `--no-verify` for that reason.
- **The proof names the index tree**, and `proof/` is gitignored, so the closing tree never
  carries it: the link is the `# tree` header.
- **`odachi` on PATH is the installed binary.** Build the tree's binary into the temp
  folder and prepend it to PATH before a command that has to know the tree's behavior.
- **`windows-latest` is the slow gate**: today 8m38s against 1m35s for ubuntu and 3m57s
  for macOS.
- **One heavy gate per machine**: `go test ./...` takes about twelve minutes here, twice
  per proof.
- **The wrappers live in temp**: `odachi-rev2.bat`, `odachi-ver3.bat` and
  `odachi-cut.ps1`. If the folder is emptied, compose them again.
- **PowerShell's `Set-Content` rewrites a Go file's line endings**, and a bash-built
  binary that lands in `bin/` is a file the next commit stages by accident: never build
  into the repository.
- **A skill body is capped at 60 lines and a rule at 300 characters**, checked by
  `odachi skills --check`.

## The debts this cycle named, in one place

0340 answered the question 0339 left open -what the number counts- and paid it with a
decision record and a door a human can open on purpose. What remains is the section
above: the declared fields of the review file, a footer that does not follow its screen,
one model judging in six costumes, no measurement of test adequacy, and the undo row that
cannot name a struct field.
