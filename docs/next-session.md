# Next session

Written while closing 0243, `only-the-global` (the human's sentence: "se va a eliminar de los repos y
solo se usara el global"), after `v0.32.0` was published. No command writes or refreshes a project
agent file any more: the agent is the door in the home.

## Where the project is

- `v0.32.0` is published (tag on `9dab0ed`); the workflow ran the suite, signed and attested.
- `main` has the plan of 0243 (`0eadec0`): the brief and the fence without 0242's non-goal about
  `start`, with the other six rules and their paths.
- The feature branch `feat/0243-only-the-global` has the code (`3caf441`), the tree record
  (`5add161`), the proof (`proof/0243-only-the-global.txt`) and an approved reliability verdict for
  tree `799e361`.
- `Unreleased` names 0243 as a contract change, minor: the release owed is `v0.33.0`.
- The installed binary is still `0.31.0`; the release is what a person installs.

## What this session did

- **0242, an agent lives only in the home**: `init` home-only, `skill new --scope`, `doctor` names a
  leftover as a warning; `level` stopped refreshing a project mirror (the review found it).
- **`v0.32.0`**: merged #449, pushed the annotated tag, the workflow published.
- **0243, only the global**: `start` rewrites the voice of the block and leaves a project agent file
  alone, like `level`.
- **A lesson the hard way**: the light lane cannot change a path the fence protects. CI reads the
  fence from the base, so the plan (brief + fence) merges first and the code lands on the new main.
  The first code PR (#450) was closed for that reason; the plan is #451.

## Next: release v0.33.0

- Merge the code PR, then the same flow as `v0.32.0`: a release PR that renames `Unreleased`, a merge,
  an annotated tag on the merge commit, and the workflow publishes.
- Install the release afterwards: the installed binary is the gate.

## The questions that wait

- **The block a project already has.** This repository's `AGENTS.md` predates `0.31.0`, and nothing
  tells a project its block is old: `doctor` names agent files, not a stale block. Decide: `doctor`
  names a stale block, or the upgrade refreshes it.
- **The registry repair of an existing user skill is unpinned** (`docs/review-debt.md`).
- **A second opinion.** The reviews used the model that wrote the code; `check` names it. Another model
  is the way the store gets one.
- **The app's project-write path has no CLI caller.** `internal/app/init` can still write project
  agents when a library caller asks (`NoProjectAgents: false`); no CLI path does. Deleting it is a
  cleanup.

## The limits to remember

- **The fence is read from the base in CI.** A change to a protected path needs a plan PR that merges
  first; the light lane is only for a change the fence does not protect.
- The fence protects `skills/**`, the review and the classifier, bench, the proofs, the design canvas
  and the installers, and the release workflow. `start` is no longer protected.
- The cmd package's `TestMain` isolates the home; tests that run `init` are not `t.Parallel`.
- One heavy gate per machine; `go test ./...` takes about 700 s.
- `odachi review --exec` gives the reviewer 600 s; a prompt over roughly 100 KB does not fit.
- `odachi check --run` writes `proof/` (ignored); the proof and the verdict name the worktree tree.

## The commands that work

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Git\cmd;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..." --timeout 1800
```

```sh
odachi review --base 9dab0ed3ba17f8ed34eac99f0a03603b100df186 \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```
