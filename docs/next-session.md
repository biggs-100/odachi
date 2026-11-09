# Next session

Written while closing the session that shipped `v0.34.0` and adopted it in this clone. The project has
no open feature: 0245 closed, released, installed, and this repository now runs the gate it wrote.

## Where the project is

- `v0.34.0` is published (tag on `a1aed42`); the workflow ran the suite, signed and attested.
- `main` carries 0244, 0245 and the release; the last commit is the adoption of `0.34.0` in this clone.
- The installed binary is `0.34.0`. `.git/hooks/commit-msg` is the gate; the old `.git/hooks/pre-commit`
  was retired with a backup in the clone's store. `odachi doctor` has one warning left, which this file
  closes: it names 0245 as the newest feature.
- `Unreleased` is empty: the release renamed it, and the next feature writes under it.

## What this session did

- **0244, the registry repair is pinned** (#454, merged): `skill new` regenerating `SKILLS.md` when the
  slug exists is now tested by seeding the skill with the registry missing and stale, and reading the
  regenerated file. The row left `docs/review-debt.md`, which is now empty.
- **0245, the gate sees the message** (#455, merged): the `Odachi-Change:` lane could not make its first
  commit because `pre-commit` runs before the message file exists. The gate moved to `commit-msg`,
  `init` retires its own `pre-commit` (only when every line is odachi's, and after a backup), and
  `check --message <path>` reads the line, checks the fence, and defers the proof.
- **The review found a CRITICAL** on the first tree: a marker is not ownership. `init` tells a human to
  paste its snippet inside their own hook, so the retirement now keeps any file with a line odachi does
  not write, byte for byte.
- **The verifier found three promises** the code did not make: the fence reads the worktree (not the
  staged paths), the fork point only exists after the commit, and no remote closes the lane. All three
  sentences were corrected in the brief, the task doc and the changelog.
- **`v0.34.0`** (#456, merged, tag `a1aed42`), then installed and adopted here with `odachi init`.

## The questions that wait

- **A second opinion.** Every verdict in this clone came from the model that wrote the code; `check`
  names it. Another model is the way the store gets one.
- **The app's project-write path has no CLI caller.** `internal/app/init` can still write project agent
  files when a library caller asks (`NoProjectAgents: false`); no CLI path does. Deleting it is a
  cleanup.
- **The store grows.** `odachi clean --apply` would free about 371 KB in 159 files; nothing depends on
  it, and it is a housekeeping decision, not a feature.

## The limits to remember

- **The proof is measured when `check --run` starts.** An edit during the twelve-minute suite leaves the
  proof naming the earlier tree, and the run still prints "ready". Finish the tree first, then run.
- The trailer lane reads the message file at `commit-msg`; a message with no trailer keeps the answer the
  `pre-commit` gate gave, and a repository with no remote keeps the lane closed.
- A fused gate: the first commit after `init` runs `commit-msg`, so a change that names itself no longer
  needs the proof to exist before the commit — it is owed after it.
- One heavy gate per machine; `go test ./...` takes about 700 s here.
- `odachi review --exec` gives the reviewer 600 s; a prompt over roughly 100 KB does not fit.
- `odachi check --run` writes `proof/` (ignored); the proof and the verdict name the worktree tree.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
odachi review --base <feature-base> \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

```sh
odachi review --since <the-tree-the-verdict-judged> \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```
