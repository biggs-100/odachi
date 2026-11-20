# Next session

Written after 0339, the cycle that gave the limit a third moment.

## Where the project is

- **0339 closed, one PR.** `odachi plan --done T1` now reports the authored lines
  since the feature's base, the base it counted from and the 400 a close that owes a
  verifier may carry. Past the limit it repeats the gate's sentence -the rest belongs
  to a feature of its own- and the step still closes: blocking the record would hide
  work that is already written. The measurement is one function shared with `plan`'s
  refusal, and a document with no base or a caller with no repository says the count
  was not taken instead of inventing one.
- ADR 0054 records why: the step measures, only the feature refuses, the split is the
  human's decision at publish time. It also records what gentle-shell answers to the
  same question -a number after every commit, a hard refusal only at delivery, an
  overage paid with a rationale and never by shrinking code- because that is where the
  alternative was measured, not guessed.
- The proof of the closing tree is `proof/0339-el-cierre-de.txt`; the verdict is
  `approved` on `d1f8387`, one lens, first round.
- The tool counts this line for the tree it describes, copied from `odachi summary`:

  total: 337 features · 851 steps done of 851 · 0 open questions

## What waits for a decision

- **The limit now fires on everything, including small features.** 0339 was a small
  change -one measurement, one decision record, two sentences in two skills- and it
  closed at 438 authored lines, because documentation and skills count as authored
  lines. 0337 and 0338 were genuinely large; this one was not. So the next question is
  not where the number is named -that is answered- but **what the number counts**: a
  decision record and a skill body are what makes a change legible, and charging them
  to the slice makes the limit a constant the project can never satisfy. The honest
  options are to exclude `docs/` and `skills/` from the authored count, to raise the
  limit for a feature that carries a decision record, or to accept that every feature
  states its overage -which is what this one now does, out loud.
- **The declared fields of `odd/review.toml`** -`max_prompt_kb`, `checks`,
  `check_timeout`, `keep_verdicts`, `keep_prompts`- which every write preserves and no
  command sets. 0336 named it as the other half of the hub; it is still open.
- **The footer does not change while a picker or an editor is open.** It still says
  `enter cambiar · esc volver` although the keys under an open list are `↑↓`, `enter`
  and `esc`. 0338 left it as a separate decision.
- **The second vendor.** The verdicts of 0335 to 0339 were written by the model that
  produced their proofs, and `check` says so on every close. The doctor's check 17 still
  counts `1 of 4`.
- **Test adequacy has no evidence** (debt 6): nothing measures changed-line coverage or
  mutation on the candidate.

## What this cycle learned

- **A plan written whole cannot be weighed while it is a plan.** The only moment a diff
  exists is after a step's code does; that is why the note belongs there and why the
  plan's refusal will keep missing the features that needed it most.
- **Ask the project that already answered the question.** gentle-shell has the same
  limit and had already chosen where it fires; reading it took one exploration and
  removed a design argument that would have cost a day. The answer was not the one
  expected -it never refuses at a close- and that is the value.
- **A skill has a size too.** Two sentences added to `commits` broke
  `odachi skills --check` twice: a body of 63 lines against 60, and a rule of 311
  characters against 300. The rule is a paragraph when it needs more than one line.
- **The note names its own author.** Each of this feature's three step closes printed
  the sentence, past the limit, before the feature had a changelog.

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
- **`windows-latest` is the slow gate**: today 8m02s against 1m10s for ubuntu and 3m26s
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

0339 paid what it opened with: the third moment, the decision record, and the two skills
that now say where the number appears. What remains is the section above -what the number
counts, the declared fields of the review file, a footer that does not follow its screen,
one model judging in five costumes, and no measurement of test adequacy.
