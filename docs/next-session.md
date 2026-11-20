# Next session

Written after 0337, the cycle that gave the wizard its models.

## Where the project is

- **0337 closed, delivered as a stack.** The hub has its third section, `modelos`: one row
  for the verifier and one per installed client, showing what each file carries and
  writing only what changed. `odd/review.toml` declares `verify`, and `odachi verify`
  resolves it in three tiers -`--model`, the key, the session- then fills `{model}`.
  `init` and `setup` take `--model client=model` and merge the key into `opencode.json`
  (or the `.jsonc` that already exists) and `~/.claude/settings.json`, one member at a
  time, never re-encoding the human's file. ADR 0053 supersedes ADR 0051's exclusion of
  per-client pickers.
- **Four PRs, not one.** 0337 measured 3678 authored lines from its base, so the close
  refused to owe a verifier and the delivery was cut into #678 (the verifier), #679 (the
  client writers), #680 (the section, the ADR and the changelog) and #681 (the review's
  corrections). Each merged on its own proof and its own CI run.
- The proof of the closing tree is `proof/0337-el-tui-elige.txt`; the verdict is
  `approved` on tree `201203c`, four lenses, after one correction round.
- The tool counts this line for the tree it describes, copied from `odachi summary`:

  total: 335 features · 844 steps done of 844 · 0 open questions

## What waits for a decision

- **The other half of the hub is decided and unbuilt**: the declared fields of
  `odd/review.toml` -`max_prompt_kb`, `checks`, `check_timeout`, `keep_verdicts`,
  `keep_prompts`- which every write preserves and no command sets. That is the next
  feature, and 0336's brief already named it.
- **The 400 has a second data point and still no answer.** 0337 was planned whole in four
  steps, so `plan`'s refusal never had a candidate to weigh and the close's note fired
  instead, at 3678. The question from the previous note stands: does a step's close name
  the accumulated number, or is the close's own note the only honest answer?
- **The schema lines are not in this repository's files yet.** `schema.Current` is 11 in
  the code; `odd/review.toml` and `odd/rules.toml` still declare 10, because the gate is
  built from the base and a candidate declaring the newer number is unreadable to it.
  The release that carries the new reader is where they move -and where this repository's
  own `verify` line lands. Check both landed before closing the session.
- **The second vendor.** The verdicts of 0335, 0336 and now 0337 were written by the model
  that produced their proofs, and `check` says so on every close. The doctor's check 17
  still counts `1 of 4`.
- **Test adequacy has no evidence** (debt 6): nothing measures changed-line coverage or
  mutation on the candidate.

## What this cycle learned

- **The gate is built from the base, so a schema bump cannot travel with its own key.**
  CI compiled the old binary and refused a candidate that declared schema 11. The fix is
  the rule 0312 left: the project's files keep their base bytes, and they move once a
  release has put the reader in the base. A feature that adds a key pays for it in the
  release, not in the feature.
- **A stack is cut at commit boundaries, and the hook does not follow it.** The four
  slices needed no new commits until the schema lines had to come back out; from there,
  the two rebases each wanted a proof for a tree that is not the close.
- **The review found the one combination the tests did not.** Every refusal of `verify`
  had a test; "the project declares a model and the command has no `{model}`" did not, and
  that is the state every existing project reaches on the day it writes the key.
- **The reviewer's clock can stop the process after the verdict is written.** Read
  `.git/odachi/reviews/<tree>.json`, not the process output: the first round wrote its
  verdict and then hung until the shell killed it.

## The commands that work

```sh
# the gate and the proof: two suite runs, about thirty minutes here
odachi check --run "go test ./..." --timeout 1800
```

```sh
# the reviewer: a wrapper the AGENT composes, not a command the human keeps
odachi review --base <feature base> --lens reliability --model <provider/model> \
  --exec '"%TEMP%\odachi-rev2.bat" {model} {prompt} "%TEMP%\rev-out.txt" "%TEMP%\rev.json"'
```

```sh
# a correction round: only what moved
odachi review --since <the approved tree> --model <provider/model> \
  --exec '"%TEMP%\odachi-rev2.bat" {model} {prompt} "%TEMP%\rev-out.txt" "%TEMP%\rev.json"'
```

```sh
# the verifier, now that the verb fills {model} from odd/review.toml
odachi verify --timeout 1800 \
  --exec '"%TEMP%\odachi-ver3.bat" {model} {prompt} "%TEMP%\ver-out.txt" "%TEMP%\ver.json"'
```

```sh
# the stack: merge, then move the next base onto main
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
- **A feature's base is a commit the remote has.** A squash erases a branch-only base, and
  CI answers `bad object`.
- **The fence is read from the base**: a policy change lands first, in its own pull request.
- **`## Unreleased` is re-created after every release**, and a version is never moved or
  reused: the tag lands on the merge of the release pull request.
- **A document commit after the close leaves the local `check` red**, on purpose: CI
  regenerates its own proof from a clean checkout.
- **The proof names the index tree**, and `proof/` is gitignored, so the closing tree never
  carries it: the link is the `# tree` header.
- **`odachi` on PATH is the installed binary.** A schema bump makes it refuse the tree that
  writes it: build the tree's binary into the temp folder and prepend it to PATH before
  the commit hook runs.
- **`windows-latest` is the slow gate**: today 6m58s against 1m08s for ubuntu and 3m43s for
  macOS.
- **One heavy gate per machine**: `go test ./...` takes about twelve minutes here, twice
  per proof.
- **The wrappers live in temp**: `odachi-rev2.bat`, `odachi-ver3.bat` (takes `{model}` as
  its first argument) and `odachi-cut.ps1`. If the folder is emptied, compose them again.
- **PowerShell's `Set-Content` rewrites a Go file's line endings**: write whole files with
  the file tool, and let gofmt own the bytes.

## The debts this cycle named, in one place

0337 paid what it opened with: the review's critical regression, the wizard that closed on
an unreadable file, and the comment that taught the wrong rule. What remains is the section
above -the 400 that cannot see a plan written in advance, a test that measures the runner,
one model judging in four costumes, no measurement of test adequacy, and the schema lines
that wait for the release.
