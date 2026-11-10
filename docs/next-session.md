# Next session

Written while closing the session after `v0.38.0` and 0255. One row waits in `docs/review-debt.md`;
nothing else is owed.

## Where the project is

- `v0.38.0` is published (tag on `a12ea3d`); the session shipped seven releases: `0.36.1` through
  `0.36.5`, `0.37.0` and `0.38.0`.
- `main` carries 0249 to 0255; the installed binary is `0.38.0`.
- 21 skills bundled, 61 in `~/.agents/skills`, 21 that match the payload. `doctor` is all clear except
  the note this file fixes.
- `Unreleased` is empty: the release renamed it, and the next feature writes under it.

## What this session did

- **0249, the legacy heading** (`v0.36.1`): `level` renames `## Cómo hablás` to `## How you speak`;
  `locateVoice` returns the heading it matched instead of discarding it.
- **0250, motion graphics** (`v0.36.2`): a skill for motion-led pieces up to ~60 s, six references and a
  scene plan, HyperFrames by default and the eyes as a captured frame.
- **0251, ui-craft adopts three** (`v0.36.3`): the user's job, the copy of each state, and the
  accessibility figure with its source. The comparison with `gentle-shell` found them.
- **0252, tui-design** (`v0.36.4`): the terminal floor, a scored critique, twelve playbooks, and the
  capture loop at 180x50, the breakpoint and 80x24.
- **0253, the NOTICE that did not ship** (`v0.36.5`): the installer copies `SKILL.md`, `references/` and
  `assets/` only, so a file at a skill's root never arrives; `skills --check` now names it.
- **0254, the review is on at birth** (`v0.37.0`): `init` writes `odd/review.toml` with `mode = "on"`
  when the project has none, and never replaces one.
- **0255, a rule is one bullet** (`v0.38.0`): `skills --check` refuses prose in `## Rules` and a rule over
  300 characters; the catalog and the fixtures became bullets. The review caught the brief saying 140
  while the code measured 300, and the change was recorded with `ask --change`.

## The questions that wait

- **`docs/review-debt.md` has one open row**: `odachi start --brief` keeps a bullet's first physical
  line and drops the rest without a word. One line per bullet is the format; the fix is to join
  continuations or refuse the draft, with a test.
- **A reproduction, not a diagnosis**: the first install of `v0.38.0` left the old `ui-craft` text and
  the second wrote the new one, while `install.ps1` replaces the binary (line 215) before `setup`
  (line 258), which should take one run. Reproduce it on the next upgrade before calling it a defect.
- Candidates, not debts: `odachi clean --apply` frees about 555 KB; the installer leaves a
  `SKILL.md.bak-*` beside a skill it replaced and they accumulate (five for `ui-craft` alone); an
  `odachi check` stopped at 17:07 left its own tree without a proof.

## The limits to remember

- **CI reads the fence from the base ref.** A feature that touches a path `main` protects cannot be
  judged until the policy lands: PR that carries only brief and fence, merge, rebase the work on it.
  Both 0250 and 0254 paid that chain, and both times the cause was a path a non-goal never named.
- **A fence rule protects paths, not phrases.** 0252 put `cmd/odachi/main.go` under the runtime rule
  reading "no new verb" from another brief; it blocked the CLI wiring two features later.
- **A blocker travels with the blob it names, not with the argument.** 0255's review approved the tree
  while the previous verdict's BLOCKER still pointed at `skill.go`, so the gate derived
  `changes_requested` and the tool wrote nothing. The way out is to move the bytes the row cites, or to
  refute it on the same tree.
- **`check` says ready with a `changes_requested` verdict.** Review mode asks for *a* verdict, not an
  approving one; refusing to ship on a standing blocker is the agent's judgment, not the gate's.
- **The proof names the tree when `check --run` starts**, and `proof/` lives in the clone.
- **`plan --done` must be committed before the merge.** Left uncommitted, it aborts the next pull.
- **The installer's path is what a project gets.** A file that exists in the source and not in
  `references/` or `assets/` is a file nobody installs.
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
