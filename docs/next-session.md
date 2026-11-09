# Next session

Written while closing the session after `v0.37.0` and 0254. One row waits in `docs/review-debt.md`;
nothing else is owed.

## Where the project is

- `v0.37.0` is published (tag on `2149f0c`); the session shipped six releases: `0.36.1` through
  `0.36.5` and `0.37.0`.
- `main` carries 0249 to 0254; the installed binary is `0.37.0`.
- 21 skills bundled, 61 in `~/.agents/skills`, 21 that match the payload. `doctor` is all clear except
  the note this file fixes.
- `Unreleased` is empty: the release renamed it, and the next feature writes under it.

## What this session did

- **0249, the legacy heading** (`v0.36.1`): `level` renames `## Cómo hablás` to `## How you speak`;
  `locateVoice` returns the heading it matched instead of discarding it.
- **0250, motion graphics** (`v0.36.2`): a new skill for motion-led pieces up to ~60 s, with six
  references (the piece, the principles, the machine-made tells, type, formats, the render) and a scene
  plan. HyperFrames by default, Remotion named; the eyes are a captured frame.
- **0251, ui-craft adopts three** (`v0.36.3`): the user's job before the appearance, the copy of each
  state, and the accessibility figure with its source. The comparison with `gentle-shell` found them.
- **0252, tui-design** (`v0.36.4`): the terminal floor, a scored critique, twelve playbooks, and the
  capture loop at 180x50, the breakpoint and 80x24. The method is Apache-2.0; the text is ours.
- **0253, the NOTICE that did not ship** (`v0.36.5`): the installer copies `SKILL.md`, `references/` and
  `assets/` only, so a file at a skill's root never arrives. `skills --check` now names it.
- **0254, the review is on at birth** (`v0.37.0`): `init` writes `odd/review.toml` with `mode = "on"`
  when the project has none, and never replaces one. A contract change.

## The questions that wait

- **`docs/review-debt.md` has one open row**: `odachi start --brief` keeps a bullet's first physical
  line and drops the rest without a word. One line per bullet is the format the parser respects; the fix
  is to join continuations or refuse the draft, with a test.
- Candidates, not debts: `odachi clean --apply` frees about 509 KB; the installer leaves a
  `SKILL.md.bak-*` beside a skill it replaced, and they accumulate; the `odachi check` we stopped at
  17:07 left its own tree without a proof.

## The limits to remember

- **CI reads the fence from the base ref.** A feature that touches a path `main` protects cannot be
  judged until the policy lands: PR that carries only brief and fence, merge, rebase the work on it.
  Both 0250 and 0254 paid that chain, and both times the cause was a path a non-goal never named.
- **A fence rule protects paths, not phrases.** 0252 put `cmd/odachi/main.go` under the runtime rule
  reading "no new verb" from another brief; it blocked the CLI wiring two features later.
- **The proof names the tree when `check --run` starts**, and `proof/` lives in the clone: an edit
  during the suite leaves a proof naming an older tree while the run prints ready.
- **`plan --done` must be committed before the merge.** Left uncommitted, it aborts the next pull and
  the release branch has to be rebased by hand (0253).
- **The installer's path is what a project gets.** A file that exists in the source and not in
  `references/` or `assets/` is a file nobody installs (0253: it worked, 0252).
- **The installer refuses a running binary and names the PID** (0227). It worked at 0.37.0 and the
  holder was a `check` from another terminal.
- One heavy gate per machine; `go test ./...` takes about 700 s. `review --exec` gives the reviewer 600 s.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
odachi review --base <feature-base> --model opencode-go/muse-spark-1.3-contributor#high \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

```sh
# the chain a protected path costs: the policy first, the work on top
git switch main && git pull --ff-only
git switch -c <feat>/<nnnn>-fence && git cherry-pick <the open commit>
# push, PR, merge; then rebase the work branch on main and push --force-with-lease
```
