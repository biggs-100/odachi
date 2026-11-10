# Next session

Written while closing 0261. Nothing is owed.

## Where the project is

- `v0.38.4` is published (tag on `c6b6592`); the session shipped eleven releases: `0.36.1` through
  `0.36.5`, `0.37.0`, `0.38.0`, `0.38.1`, `0.38.2`, `0.38.3` and `0.38.4`, and `main` carries 0249 to 0261.
- The installed binary is `0.38.4`; the installer replaced it in one run, and `doctor` is all clear.
- 21 skills bundled, 61 in `~/.agents/skills`, 21 that match the payload.
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
- **0256, a wrapped bullet arrives whole** (`v0.38.1`): `brief.Parse` reads a bullet as a list item and
  joins the lines after it until a blank line, a header or another bullet, so `start --brief` records
  the whole requirement and no longer drops the tail of every decision. The join stops at a blank line,
  so a note written under a bullet stays a note.
- **0257, every stop has its test** (tests only, no release): the header and empty-bullet resets that end
  the join are pinned, each observed red by deleting its branch.
- **0258, the hand-off can be inspected** (`v0.38.2`): `skills --resolve <step>` prints each skill a step
  would receive with its exact path and its scope, reusing `craftPath`; a name no scope carries blocks.
  The policy landed first because the base fence still protected `main.go`. The review found four gaps
  -the precedence not pinned, the empty flag falling through, `--json` ignored and two untested
  branches- and all four were paid before the merge, with a correction review of the delta.
- **The ninth probe** (docs only, no release): with a requirement that names the contract but not
  the rejected shapes, `tests` put all five edges in the test in 3 of 3 runs; without it, 2, 4 and 4
  edges, and one run shipped an implementation that accepted an inverted range. The probe is in
  `docs/skill-evaluation.md`; the order and the red did not move, the coverage did.
- **0259, the installation TUI** (`v0.38.3`): `odachi tui` detects the project, chooses the clients, the
  skills scope, the voice and the review switch and lens models, and applies with `init`/`setup`,
  `level --to`, `review mode` and the new `review.SetConfig`. A folder without the block is asked
  setup or init; without a terminal the verb refuses and names `odachi setup`. ADR 0051 supersedes
  0001's no-TUI for this surface. The verifier found the setup/init choice missing, the voseo, the
  untidy module and T4 unrecorded; the review approved with one warning (an empty table discarded) paid
  and two recorded in `docs/review-debt.md`.
- **The installer measurement** (docs only, no release): a controlled downgrade to `v0.36.2` and one
  upgrade to `v0.38.2` replaced the old `ui-craft` in a single run; the `v0.38.0` report did not
  reproduce.
- **The cleanup** (no feature): `clean --apply` freed 566 KB in 215 files, and the 89 `SKILL.md.bak-*`
  the installer had left in the home (242 KB) are gone. The installer still writes one on each upgrade:
  a slow leak, not a debt.
- **0260, the apply normalizes once** (no release, internal): `normalize` returns the canonical state
  and its problems, and `Validate` and `Apply` share it; the lens table reaches the runner as the
  validation accepted it, and a nil table is never cleared. The two review-debt rows of 0259 are paid.
- **0261, the upgrade stops leaving backups** (`v0.38.4`): `pack` backs up a replaced file only inside
  the project, where the clone's store can hold it and `undo` reads it; the home and a `--dest` outside
  are replaced and reported as written. The first proof found the other half: the setup/init report
  counted only backed-up skills, so a home skill written without a backup read as unchanged; `SkillDest`
  carries `Written` now. The delivery review approved two lenses and the verifier confirmed all four
  behaviors on real files.

## The questions that wait

- **The routing table stays out of the block** (decided while closing 0258): the plan already routes a
  step (`--skill`, the `skills:` tail) and `skills --resolve` makes that list inspectable. The third
  probe found the block is not the lever and the ninth found the step is, so a table there would add
  text without evidence. Reopen when a case appears that a step's own list cannot route.

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
- **A home skill is readable by the worker.** Measured on 2026-11-09: `opencode run --agent
  odachi-worker` read `C:\Users\USER\.agents\skills\tests\SKILL.md` headless, with no permission
  prompt. `delegate` prints `./.agents/skills` when the project carries the craft and the home path
  otherwise, and OpenCode did not refuse the absolute path. The hand-off needs no fix.
- **A flag given empty is not a flag absent.** `flag.Visit` tells them apart; `--resolve "$STEP"` with
  an unset variable would otherwise fall through to the plain listing and report success.
- **A review warning can be paid before the merge.** `review --since <verdict's tree>` reviews the
  delta, merges the predecessor's rows and costs one round; the four warnings of 0258's first verdict
  were paid that way. A warning still earns no round on its own.
- **The installer updates a changed skill in one run.** Measured on 2026-11-10: installed `v0.36.2` (the
  old `ui-craft`), then `v0.38.2` once -`setup` reported "4 replaced skills, 17 unchanged" and the
  `ui-craft` hash returned to the new text. The `v0.38.0` report did not reproduce under a controlled
  downgrade and upgrade; if it recurs, the in-use binary and the project's own `.agents/skills` copy
  are the two places to look first.
- **The installer refuses while another process holds `odachi.exe`** and names the PID. Wait for that
  run; killing it aborts somebody's gate.
- **A measurement needs an artifact the prompt does not name.** When the requirement listed the five
  rejected shapes, both arms covered them and `tests` looked inert; with the shapes unlisted, the arm
  without the skill missed edges and one contract failed. The lever was the prompt, not the text.
- **A dependency marks the change full.** `go.mod` is a signal: the review asks four lenses, and the
  module owes `go mod tidy -diff` clean before the proof, not after.
- **The verifier earns its keep before the review.** 0259's verifier found a step never recorded (T4),
  the setup/init screen missing, the voseo and the untidy module; the review would have found them a
  round later, one at a time.
- **A step's tree comes from the commit that carried its code.** When two steps share a commit,
  `plan --done T4 --commit <sha>` names the tree; the doc commit after the proof is what the gate can
  compare.
- **A large candidate splits the review.** 0259's prompt ran in two parts and the correction reviewed
  only the delta; the `--since <verdict's tree>` lane is the cheap one for fixes.
- **A backup is undo material, and undo lives inside a project.** Outside it (the home, a `--dest`)
  there is no store to read one from: the file is written without a sidecar, because a `.bak-*` beside
  it is a version nobody can restore. 0261 removed the leak at that rule.
- **A report counts every state its writer produces.** The setup summary mapped only the pack's backup
  state, so the written state -a home replacement, a gained file- read as unchanged; when a writer gains
  a state, the roll-up and its words move with it.
- One heavy gate per machine; `go test ./...` takes about 700 s. `review --exec` gives the reviewer 600 s.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
# no --model: each lens uses the model odd/review.toml names (deepseek-v4.1-flash today)
odachi review --base <feature-base> \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

```sh
# paying the review's warnings before the merge: the delta, one round
odachi review --since <the verdict's tree> \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

```sh
# the chain a protected path costs: the policy first, the work on top
git switch main && git pull --ff-only
git switch -c <feat>/<nnnn>-fence && git cherry-pick <the open commit>
# push, PR, merge; then rebase the work branch on main and push --force-with-lease
```
