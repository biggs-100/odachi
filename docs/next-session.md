# Next session

Written after 0333, the cycle that made the verifier an instrument.

## Where the project is

- **0330 closed.** A cut `verify` run names the clock that stopped it: the silence
  budget and the total deadline have their own sentence and their own number, and only
  the deadline offers `--timeout`. The adapter knew which one fired and was throwing it
  away in a boolean (`shellcmd.go`), so the fix was to stop discarding it.
- **0331 closed.** A close that owes a verifier weighs its candidate first: past 400
  authored lines the gate names the number, the limit and the door (`odachi start`). It
  fires only while the store holds no record for this tree at all, and it rides the same
  door as a missing record, so without `--require-review` it is a note.
- **0332 closed.** A store nobody can look at is named, and does not fire the slice.
  Two probes that could not answer answered in opposite directions; both now say the
  store could not be looked at, and the door is `odachi check` — `odachi doctor` never
  opens `.git/odachi/verifications`, so naming it was a dead end.
- **0333 closed.** `container-craft` joins the catalog: 60-line body, seven references,
  a catalogue of eighteen failures with the column that matters. Every rule carries its
  source and the date it was checked. Closed with an `approved` verdict and a `verified`
  record on the same tree.
- **`v0.51.2` carries it**, and `v0.51.0` and `v0.51.1` before it: three releases, one
  minor (0331, a close that used to pass can now be refused) and two patches. Each is a
  tag on the merge of its release pull request, sixteen signed and attested assets, and
  this machine reinstalled from the release it names.
- The proof of the closing tree is `proof/0333-una-skill-para.txt`, and the verdict and
  the verification record both name `4b5d07e`.
- The tool counts this line for the tree it describes, copied from `odachi summary`:

  total: 331 features · 831 steps done of 831 · 0 open questions

## What waits for a decision

- **The 400 is measured, and it points against itself.** 0333 closed at 868 authored
  lines and the verifier read and verified it in about two minutes. The number was
  copied from two documents, not from a run, and the first run says it is too strict.
  The open question is not "another number": it is whether the slice should measure per
  step (where it can prevent) instead of at the close (where it can only refuse).
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
- **The second vendor.** The review of 0333 was written by the model that produced its
  proof, and `check` says so. The doctor's check 17 still counts `1 of 4`.
- **Test adequacy has no evidence** (debt 6): nothing measures changed-line coverage or
  mutation on the candidate.

## What this cycle learned

- **A verifier with a working command is a different instrument from a review.** It
  refuted three trees in about two minutes each — twice for the same defect, a page
  stating external behaviour with a document and no date — by running the tests and
  reading line by line. Six review rounds had passed those pages. The reviews found what
  it could not: a guard that measured nouns, a body rule that lost its scope, a `next:`
  that pointed at a verb which never reads the store.
- **The order of a verdict is not negotiable.** A verdict is attached while standing on
  its tree; fixing before attaching loses the `changes_requested` record and turns a
  delta into a full re-review. It was paid twice in this cycle.
- **The gate you interrogate must be the one of the tree.** `odachi` on PATH is the
  installed binary: asking it about a rule the tree adds produced a false green, and the
  slice "did not fire" because the code was not there to fire.
- **A guard measures the line, not a sample.** A test that pins two markers out of
  eighteen rows stays green while the property disappears; a test that pins a document
  instead of a date stays green while every date is deleted.
- **The rule that loses its scope first is the one you know best.** Four times in one
  feature: the healthcheck, `--pull`, the Podman defaults, and the pasta bullet. The
  skill was teaching itself it was wrong.
- **One `Source:` line per section, and the bullets as its consequences.** Repeating the
  citation in thirty bullets is worse than not having one; naming the document once is
  what a reader can check.
- **A test that fails only inside the suite measured the machine.** 10.87s in the suite
  against 1.24s alone: the same class 0327 paid for the watchdog.
- **A skill declares `Risk: contract`.** A full change refuses `Risk: none` outright,
  and the payload is a shape an agent consumes, the same declaration `db-craft` carried.
- **After a squash merge, local `main` is reset, never pulled.** The document commit
  made before branching makes a fast-forward impossible.

## The commands that work

```sh
# the gate and the proof: two suite runs, about thirty minutes here
odachi check --run "go test ./..." --timeout 1800
```

```sh
# the verifier: a wrapper the AGENT composes, not a command the human keeps.
# It runs the host CLI with the odachi-verify agent and cuts the last JSON object.
# Five and a half minutes of work, two of wall clock; the record lands in
# .git/odachi/verifications/<tree>.json
odachi verify --exec '"C:\Users\USER\AppData\Local\Temp\odachi-ver2.bat" {prompt} <out.json> <cut.json>' --timeout 1800
```

```sh
# the review: attach needs --base, or it records base HEAD and the gate refuses
odachi review --base <feature base>            # prints the prompt, also stored under .git/odachi/reviews/
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
  regenerates its own proof from a clean checkout. Verified in practice on #667.
- **The proof names the index tree** (`git write-tree`), and `proof/` is gitignored, so
  the closing tree never carries it: the link is the `# tree` header.
- **`odachi` on PATH is the installed binary.** Build the tree's binary into the temp
  folder before asking it anything about a change the tree makes.
- **`windows-latest` is the slow gate**: five to twelve minutes against one to four for
  the others.
- **One heavy gate per machine**: `go test ./...` takes about twelve minutes here, twice
  per proof.
- **The verifier command is a transport that lives in temp.** If the temp folder is
  emptied, compose it again: `opencode.exe run --agent odachi-verify --model <model>
  --file {prompt}`, then cut the last JSON object from the answer.
- **PowerShell's `Set-Content` rewrites a Go file's line endings**: write whole files with
  the file tool, and let gofmt own the bytes.

## The debts this cycle named, in one place

The four unpaid findings of 0333 are quoted in pull request #669 and in the changelog
entry; the slice's datapoint is the section above. Nothing else is pending: the catalog
is 26 skills, every feature closed with proof and a verdict, and CI is green on the
release.
