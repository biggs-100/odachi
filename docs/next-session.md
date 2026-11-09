# Next session

Written while closing the session after `v0.36.0` and 0248. One row waits in `docs/review-debt.md`;
nothing else is owed.

## Where the project is

- `v0.36.0` is published (tag on `68a8ec0`); the workflow ran the suite, signed and attested.
- `main` carries 0248 and the release; the installed binary is `0.36.0`.
- This clone ran `odachi init` on `0.36.0`: the block in `AGENTS.md` now says neutral Spanish, and the
  gate is still `.git/hooks/commit-msg`. `odachi doctor` is all clear.
- `Unreleased` is empty: the release renamed it, and the next feature writes under it.

## What this session did

- **0246, a project writes nothing** (#458, `v0.35.0`): the app's project half is deleted; a stored
  `project_agents` key is carried through untouched, and a plain `init` records no decision.
- **0247, the inert key is pinned through the rewrite** (#460): a second reviewer found that the 0246
  test never rewrote the store, so its assertion was skipped. The test now fails if the key is dropped.
- **0248, the voice speaks neutral Spanish** (#462, `v0.36.0`): both voices say no regional slang, no
  voseo and no local expressions -neutral Spanish for every region- while the artifacts stay English.
- **The second opinion is the habit now.** Muse Spark 1.3 Contributor reviewed trees another model wrote,
  and found a real defect both times: the weak test in 0246 and, in 0248, the legacy heading.

## The questions that wait

- **`docs/review-debt.md` has one open row**: `rules.WithVoice` replaces the body under the heading, not
  the heading, so a block written before the English migration keeps `## Cómo hablás` after `level`.
  `init` regenerates it. The fix and its test are described there.
- Candidates, not debts: `odachi clean --apply` frees about 417 KB of the store; the bench has not
  crossed the new gate; a third model could give a third opinion on a future feature.

## The limits to remember

- **The proof is measured when `check --run` starts.** An edit during the twelve-minute suite leaves the
  proof naming the earlier tree while the run still prints "ready". Finish the tree, then run.
- **`main` is protected.** Every change lands through a pull request whose gate runs on the three
  platforms; a local commit on `main` has to move to a branch before it can ship.
- The trailer lane reads the message at `commit-msg`; a message with no trailer keeps the answer the
  `pre-commit` gate gave, and a repository with no remote keeps the lane closed.
- One heavy gate per machine; `go test ./...` takes about 700 s here.
- `odachi review --exec` gives the reviewer 600 s; a prompt over roughly 100 KB does not fit.
- The voice rules apply to the chat only: code, identifiers, comments, documents, commits and pull
  requests stay English.

## The commands that work

```sh
odachi check --run "go test ./..." --timeout 1800
```

```sh
odachi review --base <feature-base> --model opencode-go/muse-spark-1.3-contributor#high \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

```sh
odachi review --since <the-tree-the-verdict-judged> \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```
