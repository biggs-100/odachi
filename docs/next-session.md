# Next session

Written at the close of 0212, `triage-learns-the-two-lessons`. The barrido of gentle-ai is done:
two texts entered `writing` as references, `triage` learned its two lessons, and nothing else of
theirs earned a place here.

## What shipped

- **`triage` names the failed diagnosis.** A test written against the mechanism a report names
  that passes on the untouched tree means the report is right and the line is not: the test goes
  to the surface the symptom touches, and the thread says so.
- **`triage` names the two-failure thread.** A thread that carries two failure modes closes
  neither: the comment names both, what shipped for which, and what each still waits on.
- **The gate.** Proof `proof/0212-triage-learns-the-two-lessons.txt`; the review approved the
  tree on the reliability lens with no findings.
- **The comparison closed.** From gentle-ai only two texts (0211) and two rules (0212) were
  adapted; `judgment-day` was descoped against the lenses plus the refuter, and the rest is their
  repo's machinery.

## What the review left

Nothing. `docs/review-debt.md` has no open rows.

## The next feature

- **`assets/` in the skill contract**, only when a skill needs a template or a schema: today
  `--check` accepts `.md` under `references/` and nothing else. A contract change (ADR 0050),
  and it should wait for the need.
- **One writer per worktree; one heavy gate per machine.** Parallel lines belong in separate
  clones.
- The store is 27 MB; the lifetimes in `odd/review.toml` own it, and `clean --apply` frees what
  they allow.

## The command that works

```sh
odachi review --base 9d057fd52a9ddb2a2622ab369d138a005ec62632 \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..."
```

The default fits the suite; the short PATH only repairs what the session inherits, and the
doctor's check 15 says so when it does not.
