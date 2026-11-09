# Next session

Written while closing the session after `v0.35.0` and 0247. No feature is open; the store's queue is
empty, and the only thing waiting is a decision, not a debt.

## Where the project is

- `v0.35.0` is published (tag on `82e2fed`); the workflow ran the suite, signed and attested.
- `main` carries 0246 and 0247, and the release; the installed binary is `0.35.0`.
- The gate is `.git/hooks/commit-msg` (since 0245), and `odachi doctor` is all clear.
- `Unreleased` is empty: the release renamed it, and the next feature writes under it.
- `docs/review-debt.md` has no open row.

## What this session did

- **0244, the registry repair is pinned** (#454): the test reads the regenerated `SKILLS.md` when the
  skill already exists, with the red observed by skipping `syncRegistry`.
- **0245, the gate sees the message** (#455, `v0.34.0`): `pre-commit` runs before the message file exists,
  so the `Odachi-Change:` lane could not make its first commit. The gate moved to `commit-msg`; `init`
  retires its own `pre-commit` only when every line is odachi's, and backs it up first; `check
  --message <path>` reads the trailer, checks the fence, and defers the proof.
- **0246, a project writes nothing** (#458, `v0.35.0`): the app's project half is deleted (117 lines in,
  585 out). A stored `project_agents` key is carried through untouched, and a plain `init` records no
  decision now that none is forced.
- **0247, the inert key is pinned through the rewrite** (#460): a second reviewer found that the 0246
  test set up a store the run never rewrote, so its assertion was skipped. The test now runs with an
  explicit scope and fails if the key is dropped (the red was observed with the pass-through removed).
- **The second opinion, paid.** Muse Spark 1.3 Contributor reviewed the tree another model wrote and
  approved it with one causal warning and one suggestion -both real, both fixed in 0247. It is the way
  a clone gets a verdict from someone other than the author: `odachi review --model <other> --exec ...`.

## The questions that wait

- **Nothing is owed.** The candidates, not debts: a third opinion from a third model on a future
  feature; `odachi clean --apply` frees about 417 KB of the store; and the bench has not crossed the
  new gate yet.

## The limits to remember

- **The proof is measured when `check --run` starts.** An edit during the twelve-minute suite leaves the
  proof naming the earlier tree while the run still prints "ready". Finish the tree, then run.
- **`main` is protected.** A direct `git push origin main` is rejected; every change, the close included,
  lands through a pull request whose gate runs on the three platforms.
- The trailer lane reads the message at `commit-msg`; a message with no trailer keeps the answer the
  `pre-commit` gate gave, and a repository with no remote keeps the lane closed.
- One heavy gate per machine; `go test ./...` takes about 700 s here.
- `odachi review --exec` gives the reviewer 600 s; a prompt over roughly 100 KB does not fit.
- The review store is keyed by tree; a docs-only commit after a close is paperwork, and `check` says so.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
odachi review --base <feature-base> --model opencode-go/muse-spark-1.3-contributor#high \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

```sh
odachi review --since <the-tree-the-verdict-judged> \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```
