# What rsc-harness taught us

Read on 2026-11-06 from `C:\Users\USER\Desktop\herramientas\rsc-harness`, v4.0.5: 274 skills, 42
target files, 24 scripts, 162 test files, a 250 KB generated manifest, about 17 MB. Read-only:
nothing there was run or changed. This page keeps the analysis and the four improvements it
suggests; the next session decides which one opens first.

## What rsc is

A productised agent-harness builder: an onboarding that proposes a plan with dispositions
(selected / deferred / excluded), acceptance by hash, a post-apply verification, three lanes
(answer / FTD / ASDD), a risk classifier from paths and content, review inside every task, and a
local knowledge wiki. It is the large cousin of what odachi does: breadth instead of a fence.

## What is genuinely strong

- **The plan is accepted by hash.** Every component carries a reason and a `reevaluateWhen`; if the
  project evidence changes between preview and acceptance, the run returns `RSC_PLAN_CHANGED` and
  writes nothing; after applying, it verifies and prints `RSC_ONBOARDING_READY` or `INCOMPLETE` with
  each missing path.
- **Risk is not the agent's word.** `rsc risk` classifies from paths and content patterns (auth in
  `src/app.js` is still auth); tier 2 needs a three-refuter panel **and human approval**; an agent's
  approval never counts; self-merge only tier 0-1 with green CI and no blocking entry.
- **Review converges.** A ledger caps the loop at three rounds, a repeated finding is never raised
  twice, outcomes close each finding, and a revert becomes a review case. The `finding-verifier`
  starts from the position that the finding is false.
- **Park, don't block.** A decision queue with `blocking` and `review-after`; the run continues with
  independent tasks and the autopilot never hands the turn back.
- **Evidence over claims.** The manifest is generated and CI fails when stale; the README's numbers
  are checked by tests; `drift:check` resolves every prose link; `eval-lint` validates the cases.
- **The sello is honest.** It binds a verdict to bytes and says so: "it proves what ships is what
  was reviewed, never that the review was any good; it is self-attested, not tamper-evidence."
- **Limits are declared.** Hooks run only in Claude Code; Windows and OpenCode 2 are less tested;
  memory is local; the VSL pack's eval status says only `offer-design` beat the baseline.

## What worries (and should not be copied)

- **The combinatorial surface.** 274 skills × 18 targets × hooks × 3 OS; the support matrix has many
  "unsupported" cells and the behavioural evals cover a fraction. The long tail is the risk.
- **A 1,680-line CLI and a 1,003-line installer.** Organised, but the next features make it worse.
- **Auto-update by default** in a tool that writes hooks, and a `rscFixes` history that is a list of
  hook regressions. Each hook is a failure mode.
- **Knowledge sync commits `📝 docs(auto)` on your branch and pulls before each message.** Powerful,
  scoped, and still surprising; it is a project-wide switch.
- **Catalog breadth as a quality signal.** 274 skills across software, legal, marketing and finance
  cannot be uniformly good; the badge sells volume.
- **Skills of 185-350 lines.** Progressive loading helps, but a 300-line skill dilutes the
  instruction when it fires.

## What to take, in order

### 1. The declared `Risk:` meets the diff (0225)

The strongest gap. Odachi already classifies a diff with `internal/domain/review/tier.go` (signals:
auth, money, contracts, concurrency, delivery), but the `Risk:` line in the commit body is the
agent's word and nothing contrasts it. The check already reads the commit messages for the
`Odachi-Change:` trailer; it can read `Risk:` too.

- A parser for `Risk:` in `internal/domain/change`, next to the trailer, with its refusals.
- A note in the ready output, never a block: "the review classifies this change as full (it touches
  <paths>), and the commits declare Risk: none: name the risk or say why it is none."
- Tests: a full-tier diff with `Risk: none` warns; a full-tier diff with a named risk does not; a
  passive diff stays silent; the verdict never moves.

Why it pays: it closes the loop between what the agent claims and what the diff touches. It is the
measurement half of the risk discipline, like 0224 was for the scaffolding.

### 2. A non-goal says when to revisit it (0226)

rsc's deferred components carry `reevaluateWhen`; odachi's "Not included" is rewritten every feature
and the next reader cannot tell whether an exclusion still holds.

A convention, not code: every non-goal in the brief names the evidence that would include it, for
example "no `assets/` until a skill needs a template or a schema". It goes in the brief (the parser
keeps text) and in `shape`/`drive` as a writing rule, pinned in `skills/embed_test.go`.

Why it pays: the fence stops being a dead list and becomes a decision with a trigger.

### 3. The verdict and the proof are bound to the published tree (documental)

Odachi's proof and verdict are per tree, but the release step does not say it: the tag is cut on the
release PR's merge, and the verdict belongs to the feature's closing tree. The sello's lesson is not
the mechanism, it is stating the binding.

- `docs/releases.md`: the tag carries the verdict of the feature's closing tree; the release commit
  moves only documents.
- A line in `check` at the close naming the tree the verdict belongs to, next to the proof.

Why it pays: a reader stops believing the tag itself was reviewed.

### 4. The harness cost is measured (documental or a doctor line)

rsc has a `context-budget`; odachi does not say what its always-on costs. `doctor` can report the
bytes of the block plus `SKILLS.md` and the installed skill count, the way it already reports the
store. It is the same rule as the overengineering investigation: measure, do not estimate.

## Deferred: behavioural evaluation

rsc separates structural checks (frontmatter, links, manifest) from behavioural evaluation (does
loading the skill improve the result?) and publishes each skill's status. Odachi pins strings in
`embed_test.go`: that proves the skill says something, not that it helps. An `odachi bench` journey
with and without `tests` or `commits` would measure the effect. It is expensive and subjective; it
waits until the queue is empty.

## The decision that waits

Open **0225** first: the risk warning is code, capped, and it closes a real hole. Then **0226**,
which is a convention and a pin. The two documental pieces (release binding, harness cost) can ride
the next close. The behavioural eval stays deferred.

Open from earlier sessions, still waiting: the installer naming the process that holds the binary
(0218) and the repo test that pins `all:*` (0216).

## What was read

`README.md` (679 lines), `package.json` (v4.0.5, `rscFixes`), `scripts/rsc.js` (1,680 lines),
`scripts/install-apply.js` (1,003), `scripts/doctor.js` (603), `targets/sello.mjs`, the 42 files in
`targets/`, `skills/asdd/SKILL.md` (185 lines), `skills/orient` (71), `skills/suggest` (92),
`.github/workflows/ci.yml` and `release.yml`, `manifest.json`, and the 162 test files by name and
size. No command was run there and no file was written.
