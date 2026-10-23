# 0014 — go-tests and skill-improve skills

Status: accepted · Date: 2026-10 · Session: v0.11.2

## Context

After `triage` and `pull-request`, two gaps remained that were not about flow but
about craft: the concrete Go test pattern and the audit of a skill that grew or
stopped being used. Odachi already had `tests` (the TDD discipline) and
`skill-craft` (how a skill is born), but no skill chose the Go pattern or said how
to cut an existing skill without breaking the format.

Copying `go-testing` and `skill-improver` from gentle-ai as-is does not fit: the
first is tied to Bubbletea, golden files and `teatest`, which are useful but not
the whole discipline; the second audits against a 1,000-token style guide odachi
does not have. Both skills have to complement what already exists, not repeat it.

## Decision

- **`go-tests`**: loads after `tests` and chooses the Go pattern. Table with
  `t.Run`, `t.TempDir`, `testing.Short`, `teatest`, golden files, boundary mocks
  and the focal package before the suite. It does not repeat "one test per
  behavior" or "watch it fail": that is `tests`.
- **`skill-improve`**: audits a skill that is not used or grew. Starts with the
  trigger, requires observable rules, cuts at 60 lines, leaves one owner per rule
  and sends to `skills --check`/`--write`. The binary still validates the format;
  the skill does not repeat the rules of `internal/domain/skill`.
- **Nothing new in the binary.** They are two skills, not two verbs: they install
  with `init`, validate with `skills --check` and count with `doctor`.
- **Each one has a content test** in `skills/embed_test.go`: if someone deletes
  `t.TempDir` or the 60-line rule, the test fails.

## Not taken

- `go-testing` whole: the pattern was adapted, not the Bubbletea inventory.
- `skill-improver` whole: its 1,000-token style guide is not odachi's contract;
  the contract is the validator and the 60-line cap.
- `judgment-day`, `rdd-*`, `cognitive-doc-design` and `comment-writer`: they stay
  out because odachi already has `review`, `writing`, or decided not to have RDD.

## Consequences

- The catalog goes from twelve to fourteen skills; the project installs thirteen
  and the home one.
- A Go project has a concrete pattern without loading a 1,000-token skill.
- A skill that grew has an audit path that ends in the validator, not in a manual
  rewrite.
- The context cost grows a little: 78 lines between the two, each under the cap.
- The next skill that does not fit the cap goes to a document, not to a longer
  body.
