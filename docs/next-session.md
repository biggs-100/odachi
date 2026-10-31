# Next session: nothing is owed, and where the next work comes from

Written at the close of 0143, on v0.11.84. The chat is gone tomorrow; this file is
the state.

## What the last five features shipped

- **0139 (v0.11.80)**: this repository keeps the craft in the machine. `.agents/skills`
  and `.opencode/opencode.json` left the tree, the block names `~/.agents/skills/`, and
  the doctor learned the fact it was missing -the scope `init` resolves- so a block that
  names the other folder is a warning. Any question it cannot answer leaves both block
  variants current, because a failed read is not an absence.
- **0140 (v0.11.81, ADR 0042)**: `odachi review --since <tree>` reviews only what moved
  since a stored verdict, and the stored verdict carries the predecessor's base and its
  rows, filtered by blob. `check` names the open rows of `docs/review-debt.md` when a
  feature closes. The ADR records why the review gate lives in the clone.
- **0141 (v0.11.82)**: a blocker about bytes the fix moved travels as a non-causal
  follow-up -it stays visible and stops blocking-, and the base assertion that could not
  fail now can.
- **0142 (v0.11.83)**: the store has a door. `odachi review --all` lists every verdict
  with its tree, lenses, findings, date, base and reviewer; `--show` names it when this
  tree has one and the store holds others; `check` names `--since <tree>` when a
  predecessor exists. The bench grew three journeys and a `Contains`, so a journey about
  a message now fails when the message is gone.
- **0143 (v0.11.84)**: the voice styles the chat and never the artifacts, only the human
  changes the conversation language, and `senior` says what it does instead of being
  `novato` with the teaching taken out.

## What is open

- **`docs/review-debt.md` has no rows.** Every row the reviews of 0140 left is closed
  with its reason.
- No open questions in `odd/tasks`, and every feature is closed.

## What is measured, so nobody re-measures it

- **`opencode run` as `--exec` stalls to the 600s timeout on this machine**, at 59 KB and
  at 29 KB. The review path that works: `odachi review --base <sha>` to print the prompt
  into `.git/odachi/scratch/`, a subagent answers it, and
  `odachi review --base <sha> --attach <answer.json>` files the verdict. The verdict
  names the model that judged; `check` compares it with the proof's model.
- `mimo-v2.6-flash` closes 59 KB in about a minute and does **not** close 114 KB in
  600s; `odd/review.toml` carries both, with `max_prompt_kb = 120`.
- `check --run` needs `--timeout 400` for this suite on this machine.
- `bench` is **11 journeys and 29 steps with no dead ends**, and it can measure a
  sentence: `Step.Contains` is how a journey about a message is written.
- `clean --apply` after 0143: the store is about 715 KB in 592 files, and each review
  round files a prompt the next clean removes (`keep_prompts = 0`).
- The store holds one verdict per tree; `odachi review --all` is how a person reads it
  without opening `.git/odachi/reviews/`.

## What is already true, so it is not re-decided

- **Ownership is declared, never inferred** (ADR 0038).
- **A failed read is not an absence.** `clean --apply` refuses before the first file
  when a class it maintains cannot be read; the review refuses a store it cannot read;
  the doctor keeps both block variants current when it cannot resolve the scope.
- **The bound lives in `potentRows`** (ADR 0039): a row travels when its path is judged
  or when it still blocks, and a correction adds one rule -a blocker about bytes the fix
  moved becomes a follow-up-.
- **The review gate lives in the clone, and that is a decision** (ADR 0042): a proof is
  not versioned, and evidence about a tree cannot live inside that tree. Making CI check
  it needs a trusted job that runs the review, or a receipt signed outside the tree.
  Neither exists, and the README says so.
- **The fence in force protects** `internal/domain/proof/**`, `proof/**` and
  `skills/**`. A feature that needs one of those lands the policy correction first.
- **A relaxed key stays at the schema the released binary reads.** CI builds the gate
  from the base ref.
- **The changelog goes back to `## Unreleased`**, the release PR renames it to
  `## vX - date`, and the tag publishes.
- **The voice is data**: two texts in `internal/domain/persona`, published as the block's
  last section, and the doctor refuses a block whose voice is not the profile's. An
  artifact is English even when the chat is not; the human's own words are quoted as they
  were said.
- **Numbers follow the landing order**, and `check` reads the highest task document as
  the current feature.
- **Every irreversible action names its target and waits for the human.**

## The one decision that waits, and is not a polish

**CI cannot enforce the review**, and ADR 0042 says exactly what would change that: a
trusted job that runs `odachi review` in CI with its own model and key, or a receipt
signed outside the tree. Both are features, neither is a patch, and until one exists the
review is a discipline held at the close, where the store is. Do not present it as a gate
that already runs in CI.

## How to find the next real work

`docs/review-debt.md` holds the accepted rows (none today); `odachi doctor` and
`odachi bench` find friction nobody has named; the hygiene trio (`go vet`,
`go mod tidy -diff`, `gofmt -l .`) is cheap; and `odachi clean` says which class grew
back, which is a lifetime that does not fit the way the project works.

Measured on 2026-10-31 after v0.11.84: `doctor` 13/13, `bench` 11 journeys and 29 steps
with no dead ends, and the hygiene trio silent.
