# Contributing to odachi

odachi is built with odachi: the brief and the fence belong to the maintainer,
the plan and the code belong to the agent, and the gate decides when something is
done. A contribution follows the same loop, by hand or with your agent.

## Before you write code

- **A bug:** open an issue with the bug form. A small fix can go straight to a
  pull request; a fix that changes a contract or a file format should be
  discussed first.
- **A feature:** open an issue first. The maintainer writes `brief.md` (the
  vision and what is *not* included) and `.odachi/fence.toml`; nobody else edits
  them, and the gate refuses a change that loosens the fence. Once the brief
  exists, the implementation is a normal pull request.
- **No CLA.** Contributions are MIT, like the project.

## The loop

1. Fork, clone, and build: `go build ./cmd/odachi` (the Go version is in
   `go.mod`).
2. Branch from `main`: `git switch -c fix/the-thing`.
3. Run what CI runs: `go test ./...`, `gofmt -l .` (it must print nothing),
   `go vet ./...`. On Linux, also `go test -race ./internal/...`.
4. If you have odachi installed, the same gate locally is
   `odachi check --run "go test ./..."`. The proof is generated in `proof/`,
   which is ignored by git: never commit it and never write it by hand.
5. One step, one commit. The title says what changed for the person using
   odachi, in the imperative; the body says how and why; the last line is
   `Risk: none` or one of `data`, `security`, `contract`, `concurrency`,
   `delivery`, `no-test`.
6. Open the pull request with the template. It asks for the proof, the risk and
   the rollback because those are the three things a reviewer cannot guess.

## What the gate checks

`main` only takes a pull request whose four checks passed: `gate` on ubuntu,
macos and windows, and `checks` on Linux. The gate is built from the base
branch, not from your change, and it reads the fence from the base branch too.

- The brief is complete, the fence is respected, and there are no open questions.
- The proof is the real output of the command on *this* tree. CI runs it again;
  a proof that was edited or reused for another tree is refused.
- `checks` runs `gofmt`, `go vet`, the race detector on `./internal/...` and
  `goreleaser check`.

A red job blocks the merge. If the suite fails, the gate keeps the whole output
in `proof/<feature>.failed.txt` and uploads it as an artifact: open it before
asking why in the pull request.

## Size

- A pull request that does not read in twenty minutes is too big. Split it: the
  `chained-pr` and `work-unit-commits` skills in `.agents/skills` explain how.
- Tests and documentation travel with the code they describe. There is no later
  pull request called "tests" or "docs".
- A commit that could not be reverted alone is not a commit: it is two.

## Documentation

- A user-visible change updates `README.md` and `CHANGELOG.md`.
- A decision that constrains the future gets an ADR in `docs/decisions/`, in the
  same pull request. A decision that is obvious in six months needs no ADR.
- `docs/impact-map.md` says what each package and file is verified by; if you
  add a surface, add its row.

## AI-assisted work

Allowed, and expected: this project is built that way. Read
[AI_POLICY.md](AI_POLICY.md). The short version: disclose it, keep the evidence
real, and own the result. A person reads the diff, the proof and the findings;
"the model wrote it" is not an answer to a review.

## Reporting a vulnerability

Never in a public issue. See [SECURITY.md](SECURITY.md).
