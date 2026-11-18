# Next session

Written after 0329, the cycle that gave the skills their loop.

## Where the project is

- **0326 closed.** The probe's clean arm pays what the blind reviewer found in it:
  a test for the non-retryable 4xx branch, and a pause that stops at the last
  attempt and answers the caller's cancellation. The fixture is not compiled by the
  suite -it lives under `testdata`- and it does not compile alone: the arm calls
  what the base defines, so it is checked in a module of its own, base and arm
  together.
- **0327 closed.** The Windows streaming test writes from cmd instead of opening a
  PowerShell that has to start first: the gate had failed once at 11m41s with "a
  reviewer that kept writing was killed", and the two seconds were the runner's
  load, not the command's silence.
- **0328 closed.** The probe's two resilience defects sit in different functions,
  so one finding can no longer claim both while the reader found two.
- **0329 closed.** A skill's exam runs with one command and leaves its score:
  `odachi skills eval tests --exec "<command>" --runs 3` builds two arms -the skill
  in the project's `.agents/skills`, and nothing there-, runs the command in each,
  reads the event stream and decides the two behaviours the order shows. The
  control arm takes the home copy out under `--move-home`, with a journal written
  before the move and a restore that fails the run if it fails. The record lives in
  `.git/odachi/evals/`, under the identity of the skill text and the task.
- **`v0.50.0` carries it**: tag on the merge of release PR #663, sixteen signed and
  attested assets, and this machine reinstalled. Four features, one minor: the
  verb is a surface a person installs.
- The proof of the closing tree is `proof/0329-skills-evals.txt`, and the verdict
  is the correction `a233842f` that completes `e018c5d`.
- The tool counts this line for the tree it describes, copied from `odachi summary`:

  total: 327 features · 824 steps done of 824 · 0 open questions

## What waits for a decision

- **The verifier does not finish a large tree.** GLM and then the session's model
  each ran thirty minutes on 0329 without answering: it works by reading and
  probing, and the close had to proceed without its record. That is the next
  feature: a budget, a smaller question, or a verifier that answers a slice.
- **The second vendor.** The review of 0329 was written by the model that produced
  its proof, and `check` says so: "a second reading, not a second opinion". The
  doctor's check 17 still counts `1 of 4`.
- **The loop has no number.** The first paid run of an exam will leave the first
  record, and `docs/skill-evals.md` will quote it with the model, the draws and the
  decisive events.
- **One exam proves nothing about the form.** `tests` is the only one; a second
  skill is the second caller the shape waits for.
- **Test adequacy has no evidence** (debt 6): nothing measures changed-line
  coverage or mutation on the candidate.
- **The board look is unproven by a piece**: 0306 wrote the text, no video was
  rendered or watched.
- **The ledger:** the doctor counts the rules without an enforcer; the note used to
  say three, and the count is the tool's, not this page's.
- **`schema.Current` is 10**, and the verify record has its own counter
  (`verify.Schema = 1`).

## What this cycle learned

- **A batch is a sentence, not a list.** The proof and the verdict are owed by the
  closing tree, so debts that fit one promise cost one proof and one verdict; only
  the commit stays per debt. 0325 paid three that way, and this cycle paid the
  instrument four times without arguing about it.
- **The exam is code, and the suite never sees it.** `testdata` is invisible to
  `go test ./...`, and the fixture does not compile without the base it calls. Two
  instruments, one rule: check it in a module of its own.
- **A nested `go.mod` leaves the embed without a word.** The go tool will not embed
  a subtree that looks like another module, and it fails silently: the exam's
  project ships its module as `go.mod.exam` and a test reads the tree the binary
  carries.
- **A copied home does not authenticate.** Without its database the client answers
  401; with the database linked it keeps the real state, so the copy isolates
  nothing. The control arm is the folder absent, moved with permission.
- **A documented shape starts with the noun.** The verifier found it: the page and
  the verb's own `next:` line print `skills eval tests --exec ...`, the flag parser
  stops at the first word that is not a flag, and the shape was refused. A test
  pins it now.
- **The verifier's silence is a result.** Two thirty-minute attempts, no answer,
  and no record was invented: the close says so in the commit and the pull request.
- **The review signs with the model it is given.** Same model on both sides is a
  second reading; the quorum is what makes it a second opinion.
- **A test about a clock must not depend on another process's startup.** The
  Windows flake was PowerShell's, not the watchdog's.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
# the two arms of a skill, three draws each, and the record it leaves
odachi skills eval tests --exec '<the agent command, with {prompt}>' --runs 3 --move-home
odachi skills eval --recover        # what a dead run left out, and nothing else
```

```sh
# the review of a closing tree, and a correction that measures the delta
odachi review --base <feature base> --exec '<the reviewer command, with {model} and {prompt}>'
odachi review --since <tree prefix> --runs 1 --because "<why>" --exec '<the same command>'
```

```sh
# the verifier: one command, the prompt on stdin, the answer as one JSON object
odachi verify --exec 'C:\Users\USER\AppData\Local\Temp\odachi-ver2.bat {prompt} C:\Users\USER\AppData\Local\Temp\verify.json C:\Users\USER\AppData\Local\Temp\verify-cut.json'
```

```sh
# the release: a branch that renames Unreleased, a PR, then the tag on its merge
git switch -c release/vX.Y.Z && git commit -am "release vX.Y.Z" && gh pr create --base main
git tag -a vX.Y.Z <merge commit> -m "vX.Y.Z: <the sentence the notes carry>" && git push origin vX.Y.Z
irm https://github.com/biggs-100/odachi/releases/download/vX.Y.Z/install.ps1 | iex
odachi --version
```

## The limits to remember

- **The reviewer's clock is 600s and the verifier's is 1800s.** A candidate of a
  thousand lines is beyond both: the review timed out on 0326, and the verifier
  twice on 0329. A close that big needs a slice or a longer budget, and the gate
  carries the tree without either.
- **A feature's base is a commit the remote has.** A squash erases a branch-only
  base, and CI answers `bad object`. The squash is the moment to check the base.
- **The fence is read from the base**: a policy change lands first, in its own pull
  request.
- **`## Unreleased` is re-created after every release**, and a version is never
  moved or reused: the tag lands on the merge of the release pull request.
- **`windows-latest` is the slow gate**: five to twelve minutes against one to four
  for the others.
- **One heavy gate per machine**: `go test ./...` takes about twelve minutes here.
- **A document commit after a close leaves the local `check` red**, on purpose: CI
  regenerates its own proof from a clean checkout.
- **The proof names the index tree** (`git write-tree`).
- **PowerShell's `Set-Content` rewrites a Go file's line endings**, and gofmt then
  refuses it: write whole files with the file tool, and let gofmt own the bytes.
