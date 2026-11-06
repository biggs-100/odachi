# Next session

Written at the close of 0214, `verification-is-proportional`. The curriculum no longer asks every
change for the same ceremony: the blast radius decides who verifies.

## What shipped

- **Three tiers, one ladder.** `tests` names them: a passive change is verified by reading the diff,
  a medium one keeps the writer's own verification, and only a change that names one of the six
  risks in `commits` adds `odachi-verify`. `slices/references/delegation.md` keeps the verifier and
  routes to the same ladder, so the delegation policy and the test discipline cannot disagree.
- **Two comparisons a red owes.** Before a red is blamed on the change it is run on the base (a
  test that already failed there is named as pre-existing, with its output), and a red that comes
  from the machine is named with its condition, never reported as green and never silently skipped.
- **The gate.** `skills/embed_test.go` pins the three tiers and both comparisons; it was red before
  the text landed. Proof `proof/0214-verification-is-proportional.txt`; the review approved the tree
  on the reliability lens with no findings. Released as v0.20.0 and installed.

## What was measured

- **The plan must land first when the feature touches a fenced path.** PR #367 failed all three gate
  jobs with `out of scope: no change to the skill contract or the curriculum: skills/...`: CI reads
  the fence from `origin/main`, and 0213's fence still protected `skills/**`. The fix was the
  established split: plan PR #368 (brief, fence, task doc) merged first, the feature branch rebased
  onto it, force-pushed, and then the gate read the new fence.
- **The review is a second reading, not a second opinion.** `check` now says it when the verdict and
  the proof name the same model (`opencode-go/deepseek-v4.1-flash`). It is a known limitation, not a
  finding; switching the review model is the human's call.
- **A leftover gate held the binary.** The installer failed once with `odachi.exe is being used by
  another process`: a `check --run` from an earlier run was still alive. It exited on its own and the
  install went through. If it happens again, find the PID (`Get-Process odachi`) before installing.

## The next feature

- **The readback after publishing.** From the same comparison: `gh pr create` and `gh issue create`
  are never verified by reading the result back. A small rule in `pull-request` and `writing` would
  close it; the fence has to let those two skills change.
- **`assets/` in the skill contract**, only when a skill needs a template or a schema.
- One writer per worktree; one heavy gate per machine.

## The commands that work

```sh
odachi review --base 6827dfd2b8c09792beac5c62c1b6b2ed25494d94 \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..."
```

The default fits the suite. A build goes to `%TEMP%\opencode\odachi\<feature>\`, never to the clone.
