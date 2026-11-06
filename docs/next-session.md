# Next session

Written at the close of 0216, `skills-carry-assets`. A skill folder can now carry the files it
copies, not only the text it reads.

## What shipped

- **`assets/` is part of the contract.** A skill folder may carry flat files under `assets/`, any
  extension, declared in `## Assets`; the loader reads them, the embed carries them (`all:*`),
  `pack` and the project catalog install them beside `SKILL.md` and `references/`. An orphaned
  asset, a declaration that names no file and a tree under `assets/` are each refused with their
  own line.
- **`writing` ships the first consumer.** `skills/writing/assets/doc.md` is the copyable skeleton
  of a long document, and the skill's rules send the writer to copy it.
- **The gate.** Proof `proof/0216-skills-carry-assets.txt`; the review approved the tree on the
  reliability lens with no findings. Released as v0.22.0 and installed.

## What was measured

- **The verifier earned its place.** The risk was `contract`, and `odachi-verify` found two paths
  the writer missed: `cmd/odachi/pack.go` built a project skill from `SKILL.md` and references
  only, and `//go:embed *` skips names starting with `.` or `_`. Both are fixed in T6, with a test
  and a build probe. A risk with no verifier would have shipped a silent hole.
- **The suite can exceed the gate's 600 s default.** On a cold cache after the embed change,
  `check --run "go test ./..."` timed out at 600 s with no output; `--timeout 1200` was green. The
  default is 0209's; whether the default or the suite is wrong is the next decision.
- **A RED that lands with its fix cannot live in the commit.** The pinned test was red before the
  asset existed, but test and asset share one commit, so the record cannot show it; the
  observation is a craft note in the task doc.
- **`all:*` is proven by a build probe, not by a repo test.** A regression to `*` would still pass
  the suite for dotfiles. Pinning it needs a fixture the loader and the embed share.

## The next feature

- **The gate's timeout against a cold cache.** Either the default moves above the worst measured
  suite or the gate learns to say why it was slow; the message already names `--timeout 1200`.
- **Pin `all:*`** with a repo test, or record the build probe as the measurement it is.
- One writer per worktree; one heavy gate per machine.

## The commands that work

```sh
odachi review --base 71e370e9797d80b78fc8a96ceca2d54c4fcda726 \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..." --timeout 1200
```

The default fits a warm cache; a cold one needs 1200. A build goes to
`%TEMP%\opencode\odachi\<feature>\`, never to the clone.
