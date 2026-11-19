# Next session

Written after 0334, the cycle that moved the slice to the door where it can prevent.

## Where the project is

- **0334 closed.** `odachi plan "<text>"` weighs the feature's candidate against the
  document's base before it writes the step, and refuses past 400 authored lines naming
  the number, the limit and the door. The count left `internal/app/check` for
  `internal/domain/authorship`, so the close and the plan read one number from one diff;
  `plan.Git` gained `Diff` and `cmd/odachi/files.go` implements it over `gitx.Candidate`.
  The close keeps 0331's refusal as the backstop for a step that is itself too large.
- **`v0.52.0` carries it**, a minor because a `plan` that used to open can now be
  refused: the tag lands on the merge of its release pull request, sixteen signed and
  attested assets, and this machine reinstalled from the release it names.
- The proof of the closing tree is `proof/0334-odachi-refuses-to.txt`, and the verdict
  and the verification record both name `af5964a7`.
- The tool counts this line for the tree it describes, copied from `odachi summary`:

  total: 332 features · 832 steps done of 832 · 0 open questions

## What waits for a decision

- **The 400 counts a move as work twice, and this cycle paid it again.** 0334 closed at
  772 authored lines while almost none of it was behaviour: `lines.go` and its fourteen
  tests moved, and moving counts a line as both deleted and added. 0333 closed at 868
  and the verifier read it in about two minutes. Two datapoints, one direction: the
  number bounds movement, not the work a person reads. The open question is whether a
  move should pay once, or not at all.
- **The citation guard is narrower than its comment.** It fails when a page loses *all*
  its `Source:` lines, not when it loses one. The pages are correct today; the guard is
  not.
- **`the-failures.md` overclaims at the end.** It says everything above was documented,
  while one row says "practice, not a doc" and three carry a year alone.
- **One quote, two documents.** `:ro does not replace API authorization` is attributed
  to the OWASP cheat sheet on one page and to Docker's docs on another; one is wrong.
- **The stderr test's budget rode in on a skill feature.** Ten seconds measured the
  machine; sixty keeps the defect visible. The reviewer is right that it deserved a step
  of its own.
- **The second vendor.** The review of 0334 was written by the model that produced its
  proof, and `check` says so. The doctor's check 17 still counts `1 of 4`.
- **Test adequacy has no evidence** (debt 6): nothing measures changed-line coverage or
  mutation on the candidate.

## What this cycle learned

- **A number prevents at the door before the work, not at the one that signs it.** At a
  step's close the commit exists and the work is paid; the only place the size can stop
  more of it is before the next step is written. The close keeps its refusal for the
  single step that is too large by itself, which no door can see in advance.
- **A move is the whole number.** Added plus deleted is what the reviewer reads, so a
  file that moves costs twice its length: the feature that moved the counter to make the
  plan weigh it was itself 772 lines over a 400 limit while authoring nearly nothing.
- **An `ask` after `check --run` voids the proof.** A craft note written after the run
  moves the task document, so the tree moves and the proof names the one before. The
  notes go before the proof; this cycle paid a second suite run for it.
- **A full change that declares `Risk: none` is named by the gate.** 0334 had to declare
  `contract`: a `plan` that used to exit 0 can now exit 1, and that exit code is the
  CLI's shape for an agent. The verifier is owed by that row, and it was paid.
- **The light lane cannot delegate.** It writes one step with the requirement's words and
  no route, so a light feature with a multi-file change runs inline or leaves the lane;
  a delegated step needs a brief.

## The commands that work

```sh
# the gate and the proof: two suite runs, about thirty minutes here
odachi check --run "go test ./..." --timeout 1800
```

```sh
# the reviewer: a wrapper the AGENT composes, not a command the human keeps.
# It runs the host CLI with the odachi-reviewer agent, cuts the last JSON object
# with cut.ps1, and the verdict lands in .git/odachi/reviews/<tree>.json
odachi review --base <feature base> --lens <lens> --model <provider/model> \
  --exec '"%TEMP%\odachi-rev2.bat" {model} {prompt} "%TEMP%\rev-out.txt" "%TEMP%\rev.json"'
```

```sh
# the verifier: the same wrapper shape, with the odachi-verify agent
odachi verify --timeout 1800 \
  --exec '"%TEMP%\odachi-ver2.bat" {prompt} "%TEMP%\ver-out.txt" "%TEMP%\ver.json"'
```

```sh
# the review: attach needs --base, or it records base HEAD and the gate refuses
odachi review --attach <answer.json> --base <feature base>
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
  silence budget on both. A verifier that streams is killed by the deadline; one that
  goes quiet is killed by silence, and since 0330 the run says which.
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
- **`windows-latest` is the slow gate**: today 7m41s against 57s for ubuntu and 5m01s
  for macOS.
- **One heavy gate per machine**: `go test ./...` takes about twelve minutes here, twice
  per proof.
- **The wrappers live in temp**: `odachi-rev2.bat`, `odachi-ver2.bat` and `odachi-cut.ps1`.
  If the folder is emptied, compose them again; the shape is the one above.
- **PowerShell's `Set-Content` rewrites a Go file's line endings**: write whole files with
  the file tool, and let gofmt own the bytes.

## The debts this cycle named, in one place

Nothing new is unpaid for the code: 0334 closed with proof, an approved verdict on four
lenses and a `verified` record, and CI is green on the pull request. What waits is the
section above, and the oldest item is the one this cycle fed twice: the number that
bounds movement instead of work.
