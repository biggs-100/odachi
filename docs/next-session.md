# Next session

Written after 0335, the cycle that made the catalog say what it can prove.

## Where the project is

- **0335 closed.** Four debts of one class, paid together: the handover says what the
  400 actually measured (git detects a move; the 521 lines that dominated 0334 were a
  test file rewritten past the similarity threshold), the citation guard pins all 24
  `Source:` lines of `container-craft` in `skills/container_craft.sources` and fails in
  both directions, the docker-socket quote has one owner (OWASP Docker Security Cheat
  Sheet, Rule #1) with the page that repeated it naming that page, and
  `the-failures.md` closes saying how each row is sourced instead of promising that all
  of them are documented.
- **Not released when this note was written.** The next tag carries it: documents and
  one guard, so a patch.
- The proof of the closing tree is `proof/0335-el-catalogo-deja.txt`, and the verdict
  names the same tree.
- The tool counts this line for the tree it describes, copied from `odachi summary`:

  total: 333 features · 836 steps done of 836 · 0 open questions

## What waits for a decision

- **The stderr test's budget rode in on a skill feature.** Ten seconds measured the
  machine; sixty keeps the defect visible. The reviewer is right that it deserved a step
  of its own.
- **The second vendor.** The verdict of 0335 was written by the model that produced its
  proof, and `check` says so. The doctor's check 17 still counts `1 of 4`.
- **Test adequacy has no evidence** (debt 6): nothing measures changed-line coverage or
  mutation on the candidate.
- **The 400's own datapoint is still unmeasured.** 0333 closed at 868 authored lines and
  a verifier read it in about two minutes -the number the limit was never measured
  against-. The move question is closed: the count stays, and the reason is in the
  section above.

## What this cycle learned

- **A guard that asks "is there one?" does not protect a set.** The citation test asked
  each page whether it kept *a* `Source:` line, so a page that lost one of eight stayed
  green. The fix is a pin of every line and both directions: a line the page lost, and a
  line nobody pinned.
- **A citation with two owners is a citation nobody can check.** The same sentence was
  attributed to OWASP on one page and to Docker's docs on another; only one carries it,
  and the other now points at that one instead of inventing a source.
- **A handover is part of the close, not a note after it.** The review of this feature
  found that it had corrected the 400 paragraph and left the header, the totals and three
  paid debts in the same file. `odachi doctor` recomputes that totals line, and the next
  session would have re-opened work already done.
- **A number prevents at the door before the work, not at the one that signs it.** At a
  step's close the commit exists and the work is paid; the only place the size can stop
  more of it is before the next step is written (0334).
- **A move is not the whole number, and a rewrite is.** Git already recognizes a rename;
  a file rewritten past the similarity threshold pays for both sides because a reader
  reads both.

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
- **A note written after `check --run` voids the proof**, because it moves the task
  document. The notes go before the proof.
- **`windows-latest` is the slow gate**: today 6m04s against 1m18s for ubuntu and 5m25s
  for macOS.
- **One heavy gate per machine**: `go test ./...` takes about twelve minutes here, twice
  per proof.
- **The wrappers live in temp**: `odachi-rev2.bat`, `odachi-ver2.bat` and `odachi-cut.ps1`.
  If the folder is emptied, compose them again; the shape is the one above.
- **PowerShell's `Set-Content` rewrites a Go file's line endings**: write whole files with
  the file tool, and let gofmt own the bytes.

## The debts this cycle named, in one place

0335 paid what it opened with and left the catalogue's three paid rows out of this list.
What remains is the section above: a test that measures the runner, one model judging in
four costumes, no measurement of test adequacy, and the 868-line datapoint nobody has
answered.
