# Next session

Written at the close of 0220, `design-lets-the-human-tune`. The design comparison with gentle-shell
is paid: the four skill gaps and the tuning rule are in the curriculum.

## What shipped

- **The artboard carries the tuning controls.** When the human tunes spacing, colour or type scale,
  `design` adds sliders bound to the artboard's CSS custom properties and a `Copy as prompt` button
  that copies the chosen values as text. The button copies text and never calls a network, so the
  artboard stays self-contained.
- **The canvas did not need a change.** A focused board hides the shield and receives the pointer;
  the rule was the whole gap. The comparison had called this a canvas feature; exploration showed
  gentle-shell implements it as a skill rule and odachi's canvas already allows it.
- **The gate.** `TestDesignLetsTheHumanTune` was red before the text landed. Proof
  `proof/0220-design-lets-the-human-tune.txt`; the review approved the tree on the reliability lens
  with no findings. Released as v0.25.0 and installed.

## What was measured

- **The comparison is closed.** Design delivery, product UX, per-artboard read, another approach,
  comment screenshots and tuning are all in the pair now; nothing else from gentle-shell's design
  skills is missing.
- **The plan rode the feature PR again.** The fence did not protect `skills/design/SKILL.md`, so the
  plan and the feature shared PR #392; the two-PR split remains for a path the base fence protects.
- **The install was clean.** The release installer replaced one skill and left 19 matched.

## The next feature

- **The installer names the process that holds the binary**, open from 0218: when `odachi.exe` is in
  use, say which process and what to do, instead of a `Copy-Item` IOException. Two installs were
  blocked by another project's gate; the message is the defect.
- **Pin `all:*`** with a repo test, open from 0216: a regression to `*` would pass the suite for
  dotfile assets.
- One writer per worktree; one heavy gate per machine.

## The commands that work

```sh
odachi review --base 1f1709fad33c6cef3f67f31f1d630f87e475517d \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..."
```

No `--timeout` is needed; the default is 900. A build goes to
`%TEMP%\opencode\odachi\<feature>\`, never to the clone.
