# Next session

Written at the close of 0211, `writing-carries-the-doc-and-the-comment`. The curriculum learned
the two texts the comparison with gentle-ai found missing, and both travel as references beside
`writing` instead of growing its body.

## What shipped

- **`writing` routes two texts.** `references/the-doc.md` carries the shape a guide, an RFC or
  an onboarding page needs: answer first, quick path, details, checklist, next step, and what a
  reviewer reads first. `references/the-comment.md` carries the comment's formula (observation,
  why, next action), its tone (warm and direct, no pile-ons, the language of the thread, no em
  dashes) and three examples.
- **The body stayed under the cap.** It gained a `## Decision gates` table and `## References`;
  its eight rules are the same ones it had.
- **The installation is equal.** A binary built on the tree left `~/.agents/skills` with 19
  matched skills and no drift (`doctor` check 9).
- **The gate.** Proof `proof/0211-writing-carries-the-doc-and-the-comment.txt`; the review
  approved the tree on the reliability lens with no findings.

## What the review left

Nothing. `docs/review-debt.md` has no open rows.

## The next feature

- **`assets/` in the skill contract**, only if a skill needs a template or a schema: today
  `--check` accepts `.md` under `references/` and nothing else. It is a contract change (ADR
  0050) and it should wait for the need, not anticipate it.
- **Two rules `triage` still lacks**, both from the comparison: a test written against the
  report's mechanism that passes on the untouched tree means the report is right and the
  diagnosis is not; and a thread with two failure modes closes neither.
- **One writer per worktree; one heavy gate per machine**, as before.
- The store is 14 MB after the cleanup; the lifetimes in `odd/review.toml` still own it.

## The command that works

```sh
odachi review --base 7d9645d86371f5869fa8aaa91685b35b3479e3bf \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..."
```

The default fits the suite; the short PATH only repairs what the session inherits, and the
doctor's check 15 says so when it does not.
