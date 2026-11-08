# Next session

Written while closing 0242, `an-agent-lives-only-in-the-home`, after `v0.31.0`. The feature is
closed: `init`, `skill new` and `doctor` agree that the agent lives only in the home, and the
capability is not in the installed binary until a release.

## Where the project is

- `main` at `c49408f` (the review correction). The feature's closing tree is `449dbb9e`; its proof is
  `proof/0242-an-agent-lives-only-in-the-home.txt` and its verdict is stored for that tree
  (approved, four lenses).
- `odachi 0.31.0` is installed. The installed binary is the gate: the new `init`, `skill new` and
  `doctor` behavior is invisible until a release installs it.
- 240 features, 589 of 589 steps, no open questions in the plan.
- `Unreleased` names the change as a contract change, minor (0242).

## What this session did

- **T1, init.** `odachi init` writes no agent file in the project. It installs or refreshes the home
  door of every chosen client, its subagents and OpenCode's default; the craft defaults to
  `~/.agents/skills`, `--skills project` is the explicit pin, and a project that already carries
  `.agents/skills` keeps it. `--global` is accepted and changes nothing; `ODACHI_NO_SETUP=1` refuses
  the run; a project agent an older release left is named by file, never deleted.
- **T2, skill new.** `odachi skill new <slug>` refuses without `--scope`; `--scope project` writes
  `.agents/skills/<slug>/SKILL.md` and `--scope home` writes `~/.agents/skills/<slug>/SKILL.md`,
  each with that folder's `SKILLS.md`. The curriculum (`skills/`) is untouched.
- **T3, doctor.** Any odachi agent file in the project is a leftover, whatever its content: check 11
  names each path and turns the line into a warning. The check count stays at fifteen.
- **The review caught one real defect**: `level --to` still refreshed a project agent file the doctor
  called unread. Level now rewrites only the voice section and never touches a leftover.
- Two review rounds on the feature and one on the correction. The correction is approved; one
  warning stays in `docs/review-debt.md` (the registry repair is unpinned by a test).

## Next: release v0.32.0

- Cut a minor release: `docs/releases.md` says the tag carries the verdict of the feature's closing
  tree, and the last three steps moved only documents. The installer is what puts the new behavior
  on a machine.
- The changelog entry is already written under `Unreleased`; the release is a heading and a tag.

## The questions that wait

- **The block a project already has.** This repository's `AGENTS.md` predates `0.31.0`, and nothing
  tells a project its block is old: `doctor` names agent files, not a stale block. Decide: `doctor`
  names a stale block, or the upgrade refreshes it. Warning 4 is the evidence.
- **`start` still refreshes a project agent file it finds.** 0242 excluded `start`; the same
  incoherence the review found in `level` remains there. Decide whether it stops.
- **A second opinion.** The closing review used the same model that wrote the code, and `check`
  names it: a second reading, not a second opinion. Running the reviewer on another model is the
  way the store gets one.

## The limits to remember

- The fence protects `skills/**` and the review and the classifier. Doctor was released for 0242.
- The cmd package's `TestMain` isolates the home. Any test that runs `init` writes the home, so those
  tests are not `t.Parallel`.
- One heavy gate per machine; `go test ./...` takes about 700 s.
- `odachi review --exec` gives the reviewer 600 s. A prompt over roughly 100 KB does not fit: run the
  reviewer directly with the OpenCode CLI and `odachi review --attach <file>`; for a delta, attach
  with `--since <tree>` through `odachi review --since <tree> --exec ...`, which merges the rows.
- `odachi check --run` writes `proof/` (ignored); the proof and the verdict name the worktree tree.

## The commands that work

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Git\cmd;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..." --timeout 1800
```

```sh
odachi review --base d79d4226061a9e62fd1ba19c5b9b0ce37a68ba9d \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```
