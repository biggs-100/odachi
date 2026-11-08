# Next session

Written while closing 0243, `se-va-a` (the human's sentence: "se va a eliminar de los repos y solo se
usara el global"), after `v0.32.0` was published. The tool no longer keeps a project agent file alive:
`init` writes none, and neither `start` nor `level` refreshes one.

## Where the project is

- `main` has `v0.32.0` published (tag on `9dab0ed`); the release workflow ran the suite, goreleaser
  signed and attested, and the release has its assets.
- The feature branch `feat/0243-start-leaves-the-leftover` has one commit (`58e2663`), the task doc,
  its proof (`proof/0243-se-va-a.txt`) and an approved reliability verdict for tree `3be1b26`.
- `Unreleased` names 0243 as a contract change, minor: the release owed is `v0.33.0`.
- The installed binary on this machine is still `0.31.0`; the release is what a person installs.

## What this session did

- **0242, an agent lives only in the home**: `init` home-only, `skill new --scope`, `doctor` names a
  leftover as a warning; `level` stopped refreshing a project mirror (the review found it).
- **`v0.32.0`**: merged #449 and pushed the annotated tag; the workflow published.
- **0243, only the global**: `start` rewrites the voice of the block and leaves a project agent file
  alone, like `level`. The fence dropped 0242's non-goal about `start`, which the requirement
  superseded. Approved with no findings.

## Next: release v0.33.0

- The same flow as `v0.32.0`: a release PR that renames `Unreleased`, a merge, an annotated tag on the
  merge commit, and the workflow publishes. The release is the human's call; the notes already carry
  the upgrade line.
- Install the release afterwards: the installed binary is the gate, and the new behavior is invisible
  until `0.32.0`/`0.33.0` is installed.

## The questions that wait

- **The block a project already has.** This repository's `AGENTS.md` predates `0.31.0`, and nothing
  tells a project its block is old: `doctor` names agent files, not a stale block. Decide: `doctor`
  names a stale block, or the upgrade refreshes it. Warning 4 is the evidence.
- **The registry repair of an existing user skill is unpinned** (`docs/review-debt.md`): the branch of
  `skill new` that repairs `SKILLS.md` when the skill exists has no test that reads the rewritten row.
- **A second opinion.** The reviews used the model that wrote the code; `check` names it. Another model
  is the way the store gets one.
- **The app's project-write path has no CLI caller.** `internal/app/init` can still write project
  agents when a library caller asks (`NoProjectAgents: false`); no CLI path does. Deleting it is a
  cleanup, not a behavior change.

## The limits to remember

- The fence now protects `skills/**`, the review and the classifier, bench, the proofs, the design
  canvas and the installers, and the release workflow. `start` is no longer protected.
- The cmd package's `TestMain` isolates the home. Any test that runs `init` writes the home, so those
  tests are not `t.Parallel`.
- One heavy gate per machine; `go test ./...` takes about 700 s.
- `odachi review --exec` gives the reviewer 600 s; a prompt over roughly 100 KB does not fit (run the
  reviewer directly and `--attach`, or use `--since` for a delta).
- `odachi check --run` writes `proof/` (ignored); the proof and the verdict name the worktree tree.

## The commands that work

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Git\cmd;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..." --timeout 1800
```

```sh
odachi review --base 9dab0ed3ba17f8ed34eac99f0a03603b100df186 \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```
