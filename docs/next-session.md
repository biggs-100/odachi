# Next session

Written at the close of 0215, `publishing-is-read-back`. A create command is not done at the URL
it prints: the result is read back and compared with what was sent.

## What shipped

- **The readback.** `pull-request` reads a created PR back with
  `gh pr view <url> --json title,body,baseRefName,headRefName` and compares it with the file that
  was posted; an empty body, a wrong base or a title that lost its point is a publish failure,
  fixed before the review is requested. `writing` says any published text, an issue included, is
  compared with the file that was written, not with memory, and names `pull-request` as the owner
  of the command.
- **The gate.** `TestPublishingIsReadBack` pins both skills and was red before the text landed.
  Proof `proof/0215-publishing-is-read-back.txt`; the review approved the tree on the reliability
  lens with no findings. Released as v0.21.0 and installed.

## What was measured

- **The plan-first flow held.** Plan PR #372 merged, the feature branch started from the new main
  (#373), the release followed (#374). No gate read an old fence this time: the CI lesson from
  0214 is now the routine, not a repair.
- **The close is what turns doctor green.** Check 14 warned until this file named 0215; check 9
  compared 19 skills with no drift because the tree's binary ran `init` before the gate.
- **The install found the home equal.** The release installer reported 19 unchanged: the tree's
  binary had already written the two skills, so the release only confirmed them.

## The next feature

- **`assets/` in the skill contract**, only when a skill needs a template or a schema. It is the
  last gap from the comparison, and it touches `internal/domain/skill/**` (the contract) and its
  tests; the fence has to let that path change.
- One writer per worktree; one heavy gate per machine.

## The commands that work

```sh
odachi review --base 96ca86f8dc4e0ffb1fdcd3905a95695dfaeac3c5 \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..."
```

The default fits the suite. A build goes to `%TEMP%\opencode\odachi\<feature>\`, never to the clone.
