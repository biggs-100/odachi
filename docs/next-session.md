# Next session

Written after 0336, the cycle that made the wizard a hub.

## Where the project is

- **0336 closed.** `odachi tui` opens on the state plus two sections -`instalar y
  configurar`, `perfil`- and `esc` returns to the menu from either one. Navigation is a
  table (forward, back, whether the screen exists for this run), so a new screen is a
  row. The profile section edits the six axes of the profile in force through `level.Set`,
  which backs the file up, leaves the voice alone and is silent when the value already
  holds.
- **`v0.54.0` carries it**, a minor because the first screen and its keys changed: the tag
  lands on the merge of its release pull request, sixteen signed and attested assets, and
  this machine reinstalled from the release it names. The hub ships with two sections; the
  other two are the next feature.
- The proof of the closing tree is `proof/0336-odachi-tui-abre.txt`, and the verdict and
  the verification record both name `367bb50`.
- The tool counts this line for the tree it describes, copied from `odachi summary`:

  total: 334 features · 840 steps done of 840 · 0 open questions

## What waits for a decision

- **The next work is decided, and it carries one decision of its own.** The hub ships with two
  sections and is missing two: the verifier model and the per-client models, and the declared
  fields of `odd/review.toml` (`max_prompt_kb`, `checks`, `check_timeout`, `keep_*`), which every
  write preserves and no command sets. `docs/decisions/0051-the-installation-tui-is-an-adapter.md`
  keeps per-client model pickers out of the installation TUI: the human asked for them, so that
  feature either supersedes the ADR or says why the picker belongs to the clients it writes.
- **The 400 cannot see a plan written in advance.** 0336 closed at 1012 authored lines:
  the four steps were planned before the code existed, so `plan`'s refusal -which weighs
  the candidate already on disk- never had a moment to fire. The number still bounds the
  feature that grows step by step; a feature planned whole escapes it. The decision is
  whether the close's own note (which fired here) is the only answer, or a step's close
  should name the accumulated number too.
- **The stderr test's budget rode in on a skill feature.** Ten seconds measured the
  machine; sixty keeps the defect visible. The reviewer is right that it deserved a step
  of its own.
- **A second test measures the machine, and it failed on windows-latest.** On the PR that
  only moved this note, `TestPowerShellItselfIsNotRefused` ran
  `powershell -NoProfile -Command "exit 0"` with a 30-second budget and took 30.38s on a
  loaded runner, while the same tree passed ubuntu and macOS. The product is fine; the
  budget is a measurement of the runner.
- **The second vendor.** The verdicts of 0335 and 0336 were written by the model that
  produced their proofs, and `check` says so on every close. The doctor's check 17 still
  counts `1 of 4`.
- **Test adequacy has no evidence** (debt 6): nothing measures changed-line coverage or
  mutation on the candidate.

## What this cycle learned

- **A guard can be worse than the defect it guards.** The review found that the hub let
  one section apply the voice while another edited the profile, so an axis written by
  hand was reverted in silence. The first fix compared the chosen voice against the one
  the wizard found: choosing a voice disarmed it. The fix that survived has no guard -
  the axes written in this run are recorded per write and re-applied after the voice, so
  the explicit edit is the last word.
- **A hub creates an order between sections, and the order is a contract.** The voice
  operation rewrites the whole profile from the builtin; anything explicit has to be
  re-applied after it, and a write that landed before a failure is still on disk.
- **The third correction needs its reason in the record** (0296). It was paid here with
  `--because`, and the reason is what a reader sees beside the verdict.
- **A `-q` commit that fails looks like a commit.** The hook refused once, the rest of a
  chained command ran, and the proof was taken on the staged bytes; nothing was lost
  because the index tree was the one the records named. Check `git log`, not the last
  exit code of a chain.

## The commands that work

```sh
# the gate and the proof: two suite runs, about thirty minutes here
odachi check --run "go test ./..." --timeout 1800
```

```sh
# the reviewer: a wrapper the AGENT composes, not a command the human keeps.
# It runs the host CLI with the odachi-reviewer agent and cuts the last JSON object
# with cut.ps1; the verdict lands in .git/odachi/reviews/<tree>.json
odachi review --base <feature base> --lens <lens> --model <provider/model> \
  --exec '"%TEMP%\odachi-rev2.bat" {model} {prompt} "%TEMP%\rev-out.txt" "%TEMP%\rev.json"'
```

```sh
# the correction of a verdict: only what moved, and a third round needs --because
odachi review --since <tree> --because "<why this round exists>" --model <provider/model> \
  --exec '"%TEMP%\odachi-rev2.bat" {model} {prompt} "%TEMP%\rev-out.txt" "%TEMP%\rev.json"'
```

```sh
# the verifier: the same wrapper shape, with the odachi-verify agent
odachi verify --timeout 1800 \
  --exec '"%TEMP%\odachi-ver2.bat" {prompt} "%TEMP%\ver-out.txt" "%TEMP%\ver.json"'
```

```sh
# the release: a branch that renames Unreleased, a PR, then the tag on its merge
git switch -c release/vX.Y.Z && git commit -am "release vX.Y.Z" && gh pr create --base main
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
irm https://github.com/biggs-100/odachi/releases/download/vX.Y.Z/install.ps1 | iex
odachi --version
```

```sh
# after the squash, and before the next feature
git fetch origin && git switch main && git reset --hard origin/main
```

## The limits to remember

- **The reviewer's clock is 600s and the verifier's is 1800s**, with a 120-second
  silence budget on both. A reviewer reading a large candidate can be cut at 600s and
  writes nothing; the run says which clock stopped it.
- **A feature's base is a commit the remote has.** A squash erases a branch-only base,
  and CI answers `bad object`.
- **The fence is read from the base**: a policy change lands first, in its own pull
  request.
- **`## Unreleased` is re-created after every release**, and a version is never moved or
  reused: the tag lands on the merge of the release pull request.
- **A document commit after the close leaves the local `check` red**, on purpose: CI
  regenerates its own proof from a clean checkout.
- **The proof names the index tree**, and `proof/` is gitignored, so the closing tree
  never carries it: the link is the `# tree` header.
- **`odachi` on PATH is the installed binary.** Build the tree's binary into the temp
  folder before asking it anything about a change the tree makes.
- **A note written after `check --run` voids the proof**, because it moves the task
  document. The notes go before the proof.
- **`windows-latest` is the slow gate**: today 6m49s against 1m01s for ubuntu and 5m13s
  for macOS.
- **One heavy gate per machine**: `go test ./...` takes about twelve minutes here, twice
  per proof.
- **The wrappers live in temp**: `odachi-rev2.bat`, `odachi-ver2.bat` and `odachi-cut.ps1`.
  If the folder is emptied, compose them again; the shape is the one above.
- **PowerShell's `Set-Content` rewrites a Go file's line endings**: write whole files with
  the file tool, and let gofmt own the bytes.

## The debts this cycle named, in one place

0336 paid what it opened with: the review's two causal findings and the missing test of
the write path. What remains is the section above -a limit that cannot see a plan written
in advance, a test that measures the runner, one model judging in four costumes, and no
measurement of test adequacy.
