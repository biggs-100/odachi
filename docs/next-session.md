# Next session

Written after 0338, the cycle that gave the wizard a look and a picker.

## Where the project is

- **0338 closed, delivered as a stack of four.** The terminal interface draws from
  a layer with named pieces -title, subtitle, cursor glyph, row, value, warning,
  failure, help- in odachi's own palette, and every list windows with `↑ más` and
  `↓ más` instead of growing. Every model row -the verifier's, each client's, each
  lens's- opens the list the host answers: `hostmodel.Models`, once per run, off the
  UI thread, with `escribir a mano` one row away. Claude never shows OpenCode's
  catalogue: its three aliases are declared in one constant, with the date they were
  checked.
- The four PRs were #683 (the layer), #684 (the window), #685 (the picker) and #686
  (the three states, the single cursor and the changelog). Each merged on its own CI
  run, with the next branch rebased onto `main` first.
- The proof of the closing tree is `proof/0338-visyalmente-la-tui.txt`; the verdict is
  `approved` on `5aacf50`, first round, after the screen render found what the suite
  had encoded as correct.
- The tool counts this line for the tree it describes, copied from `odachi summary`:

  total: 336 features · 848 steps done of 848 · 0 open questions

## What waits for a decision

- **The 400 now has two data points and no answer.** 0337 closed at 3678 authored lines
  and 0338 at 1801, and both closes refused to owe a verifier with the same sentence,
  because both plans were written whole before the code existed. `plan`'s own refusal
  never sees a candidate when the steps arrive in one sitting. The question is whether a
  step's close names the accumulated number -the only moment it can still stop
  something- or whether the close's note is the honest answer. This is the next
  candidate, and it is about the method, not the product.
- **The declared fields of `odd/review.toml`** -`max_prompt_kb`, `checks`,
  `check_timeout`, `keep_verdicts`, `keep_prompts`- which every write preserves and no
  command sets. 0336 named it as the other half of the hub, and it is still open.
- **The footer does not change while a picker is open.** It still says `enter cambiar ·
  esc volver` although the keys under an open list are `↑↓`, `enter` and `esc`. A footer
  that follows the screen is a small change with a wide surface; 0338 left it as a
  separate decision.
- **The second vendor.** The verdicts of 0335, 0336, 0337 and 0338 were written by the
  model that produced their proofs, and `check` says so on every close. The doctor's
  check 17 still counts `1 of 4`.
- **Test adequacy has no evidence** (debt 6): nothing measures changed-line coverage or
  mutation on the candidate.

## What this cycle learned

- **A green suite can encode a defect as correct.** The picker drew two cursor glyphs -
  the row and the first entry- and a test asserted exactly two. Reading the rendered
  screens in plain text found it in a minute. Look at the output before trusting the
  assertion that describes it.
- **One owner for the keys, on screen too.** When a list opens under a row, the row has
  to drop its glyph, or the person cannot tell what `↑` moves. The profile editor had
  already solved it; the rule is now the same in three places.
- **A picker must say which list it shows.** The host's catalogue is OpenCode's; putting
  it on a Claude row names models that client does not serve. One function answers which
  list a row shows, and the renderers read it.
- **`git rebase --onto` takes the old base as the upstream, not the new one.** Rebasing a
  stacked branch onto a moved base with the wrong argument replays the slices already
  merged and conflicts on the commits that landed as squashes.

## The commands that work

```sh
# the gate and the proof: two suite runs, about thirty minutes here
odachi check --run "go test ./..." --timeout 1800
```

```sh
# the reviewer, and a correction round against the approved tree
odachi review --base <feature base> --lens reliability --model <provider/model> \
  --exec '"%TEMP%\odachi-rev2.bat" {model} {prompt} "%TEMP%\rev-out.txt" "%TEMP%\rev.json"'
odachi review --since <the approved tree> --model <provider/model> \
  --exec '"%TEMP%\odachi-rev2.bat" {model} {prompt} "%TEMP%\rev-out.txt" "%TEMP%\rev.json"'
```

```sh
# the stack: merge, then move the next branch onto main with the OLD base as upstream
gh pr merge <n> --squash
git fetch origin && git rebase --onto origin/main <the old base> feature/<next>
git push --force-with-lease origin feature/<next> && gh pr edit <n> --base main
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
  base: rebase before retargeting, or CI answers `bad object`.
- **The fence is read from the base**: a policy change lands first, in its own pull request.
- **`## Unreleased` is re-created after every release**, and a version is never moved or
  reused: the tag lands on the merge of the release pull request.
- **A document commit after the close leaves the local `check` red**, on purpose: CI
  regenerates its own proof from a clean checkout. The release commit is made with
  `--no-verify` for that reason.
- **The proof names the index tree**, and `proof/` is gitignored, so the closing tree never
  carries it: the link is the `# tree` header.
- **`odachi` on PATH is the installed binary.** A schema bump makes it refuse the tree
  that writes it: build the tree's binary into the temp folder and prepend it to PATH.
- **`windows-latest` is the slow gate**: today 6m19s against 58s for ubuntu and 3m24s for
  macOS.
- **One heavy gate per machine**: `go test ./...` takes about twelve minutes here, twice
  per proof.
- **The wrappers live in temp**: `odachi-rev2.bat`, `odachi-ver3.bat` and
  `odachi-cut.ps1`. If the folder is emptied, compose them again.
- **PowerShell's `Set-Content` rewrites a Go file's line endings**: write whole files with
  the file tool, and let gofmt own the bytes. A `--done` value containing quotes breaks
  `odachi start` in PowerShell; write the sentence without them.

## The debts this cycle named, in one place

0338 paid what it opened with: the look, the picker, the three states, and the two defects
the render found. What remains is the section above -the 400 that cannot see a plan
written in advance, the declared fields of the review file, a footer that does not follow
its screen, one model judging in four costumes, and no measurement of test adequacy.
