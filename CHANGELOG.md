# Changelog

What changed in each released version, in the order a person reads it.
Versions before 0.10.0 were rebuilt from their tags: the repo did not have this
file yet, and a changelog that starts today erases the history of whoever arrives
tomorrow.

## Unreleased

A raised scope is written down.

- **A run that adds a lens says so, beside the verdict.** `--lens` widened what a
  run reviewed and nothing recorded that somebody asked for it: the verdict stamped
  the lenses it answered, so a later reader saw four lenses and could not tell
  whether the change asked for four or a person did, and a replay that dropped the
  flag left no trace. The run now writes `reviews/<tree>.scope.json` beside the
  verdict: the lenses the evidence asked for, the ones added, the flag as it was
  given, and the caller's words when they gave them.
- **`--show` and the prompt say it.** `--show` prints the scope line under the
  verdict, and the prompt no longer claims the evidence chose lenses somebody added:
  it names what was added, with which flag and why, so a reviewer knows the addition
  is deliberate.
- **The record is written once and only when it exists.** A second run that asks the
  same thing does not rewrite it, and a run that added nothing writes no file: the
  store is evidence, and a record of a decision nobody made is noise.

## v0.11.51 - 2026-10-29

A Markdown the binary ships is code.

- **A skill or a review prompt that changed asks for a lens.** The classifier
  decided a document by extension, README name and `docs/` segment, and excluded
  only the agent directories. A `skills/*/SKILL.md` compiled into the binary, a
  `internal/assets/review/*.md` that tells a reviewer what to look for, and the
  installed copy under `.agents/` were all `.md`, so a tree that only edited one of
  them classified as passive and asked for no lens at all — including a change to
  the prompt that decides the review.
- **One list of what ships.** `review.ShippedText` names the three prefixes and is
  asked by both gates: the classifier uses it to call shipped text code, and the
  proof lane uses it instead of the copy it had carried since 0096. Two lists that
  answer one question drift; now there is one.
- **The installed copy is still plumbing for the lens.** 0043 decided that adopting
  odachi must not cost a review before the first line of code, and that decision
  stands: `.agents/skills/**` never reaches a lens. It does owe the suite, because
  the pack test compares those bytes with their source — which is why the lane asks
  the list and the classifier asks it too.

## v0.11.50 - 2026-10-29

A release note stops costing a suite.

- **A tree that only rewrites documents after the close owes no new proof.** The
  commit that renames `## Unreleased` to `## v0.11.50` is a new tree, and `check`
  answered "changed since the proof": every release of this project landed with
  `--no-verify`, which is the habit a gate should never teach. After a close, when
  everything that moved since the proof is paperwork, the gate is ready and says
  which tree the proof belongs to and what moved since. The pull request runs the
  suite from the base ref, which is where the note is judged.
- **The proof asks the review's question.** The lane is the proof's twin of
  `theCloseOwes`: it never opens on the tree that closed the feature, never rescues
  a missing proof, and never covers a command this run watched fail. A candidate is
  paperwork when the same classifier the review already uses says it asks for no
  lens, so the two gates cannot disagree about what a document is.
- **What it does not cover, named:** a Markdown the suite reads -a skill, a review
  prompt, the installed copy of the skills- is still "documentation" by its
  classifier, so a change there after a close is paper. That gap is older than this
  feature (it also means a skill edit asks for no lens) and it is a row of its own.

## v0.11.49 - 2026-10-29

The plan commit stops being a stranger to the gate.

- **A feature with a plan and no steps is open.** `check` reads the task document
  to decide whether this tree is a feature, and answered no while the document had
  no steps, so the review gate fell through and asked a tree that writes only the
  brief and the fence for a verdict. The fence is ordinary code on purpose -it is
  a human decision, like `.gitignore`- so the plan commit was a code candidate, and
  the honest fence was the one that could not land without `--no-verify`. It is
  open now: a document with a base and no steps, the same rule the proof already
  followed, and an open feature owes its verdict at the close.
- **What it still owes.** The tree that marks the last step still owes the one
  verdict at the document's base; a step in the middle still owes nothing; and a
  release note written after the close is still paperwork.

## v0.11.48 - 2026-10-28

Three files odachi writes say which odachi wrote them.

- **The verdict, `agents.toml` and `odd/review.toml` carry a `schema`.** They decode
  strictly on purpose, and strictness cannot tell a typo from a writer that is ahead
  of it: a field that arrived after the installed binary reads as `unknown keys`, and
  the message sends the human to fix a file that is not broken. It cost three closes
  in one session (`project_agents` in 0081, `inspection` in 0085, `max_prompt_kb` in
  0087), each solved by swapping the installed binary by hand. The reader now asks
  the version before it names a key, so the next field arrives as "written by a newer
  odachi (schema 2); this binary reads schema 1 - upgrade odachi".
- **Nothing written before this moves.** A file with no `schema` reads as the first
  schema and keeps its bytes until something rewrites it; a key that is not a schema
  field is still an error naming the key. The one thing the field cannot do is help a
  binary that predates it, which cannot read a key it has never heard of: from here
  on, that is the case every close avoids.

## v0.11.47 - 2026-10-28

Two rules find the place where they can be kept: a fence rule that names nothing, and
a proof that belongs to the close.

- **`start` names a non-goal that protects nothing.** A non-goal written in free
  text became a fence rule whose path is a glob of its own words
  (`no-tocar-la-base-de-datos*`), which matches no file anybody has. The run now
  says so for each one, names `odd/fence.toml` and what a real rule looks like
  (`paths`, `deps`), and adds how to prove a fence works: touch a path it protects
  and run `odachi check` - it must refuse. A non-goal from the closed vocabulary
  already carries real paths and stays silent.

- **The proof is owed by the feature that closes, not by every step.** `check` used
  to demand a proof for any tree, so every intermediate commit cost a full run of the
  suite - and on a repository whose suite takes forty minutes that is a workflow that
  does not exist. While the feature is open, `check` now says it is ready for a step
  and names the close; the tree that closes the feature owes the proof again, and
  that tree is the one a pull request and CI judge. A proof this run watched fail
  still blocks: what nobody has paid for yet is not owed, but what just broke is. The
  hook needs no change, and its marker now says what it enforces.

## v0.11.46 - 2026-10-28

A judgment somebody made must be able to outlive the reviewer who got it wrong.

- **A person can retire a row, with evidence.** `odachi review --dismiss
  <path>:<line> --because "<why>"` writes the judgment beside this tree's verdict -
  `reviews/<tree>.dismissed.json`, bound to the row's fingerprint with the reason
  and who wrote it - and the verdict is never rewritten. The gate, `--show` and the
  inheritance apply it when the store is read: the row travels to the next candidate
  dismissed, and the next reviewer is told not to raise it. A row that is not in
  this tree's verdict is refused naming the rows it holds, and a causal BLOCKER or
  CRITICAL is refuted or fixed, never dismissed.
## v0.11.45 - 2026-10-28

A review is evidence, and evidence that is overwritten is not evidence.

- **A failed run keeps the answer it paid for.** When a reviewer writes into the
  `{verdict}` file and then fails - a timeout, a bad exit code - the bytes are
  copied under a name their content decides, the refusal names the file, and the
  scaffolding is removed so the next run on the same tree cannot empty it. The
  same path already kept an answer the tool could not read; now it keeps the ones
  it never got to read.
- **`review --exec` files the prompt it sends.** The printed form wrote
  `reviews/<tree>.prompt.md`; a run that handed the prompt to a reviewer kept it
  nowhere, so the bytes that produced a verdict could only be recovered by
  regenerating them and hoping the base had not moved. The question is filed once
  it has been sent, under the same name, and a run split across models files one
  per group. A store that cannot hold it says so and keeps the verdict: the
  answer is the paid thing.

## v0.11.44 - 2026-10-28

A gate is only as strong as the least it lets you say, and a reviewer is only as
strong as the prompt it can take.

- **A review refuses a prompt its reviewer cannot take.** `odd/review.toml` carries
  `max_prompt_kb`: the largest prompt this project pays a reviewer for. A model
  that stalls at a size does not answer late - it does not answer at all - and the
  size is the one thing odachi knows before spending. `review --exec` refuses
  before starting anything, naming the size, the limit and the two ways out; the
  printed form still prints, because it spends nothing, and says the size where a
  person reading the run can see it. A file without the key is no limit.
- **An older binary refuses the new key, on purpose.** `review.toml` carrying
  `max_prompt_kb` is an error to a binary that predates it (`unknown keys`): the
  key and the binary land together.

## v0.11.43 - 2026-10-28

A gate is only as strong as the least it lets you say.

- **A verdict that says nobody judged it is refused.** `check` compares the model
  in a verdict with the one in the proof, and a verdict that named none passed
  that comparison by saying less. It now refuses it, names the lenses, the file
  the verdict is in and the model the proof carries, and offers the two honest
  roads: re-run the review naming the model that judged
  (`odachi review --model <provider/model>`), or record the human's decision that
  there is no second model (`odachi start --waive-review`). The prompt now asks
  the answer to name its model, and `--help` lists the waiver lane it always had.

## v0.11.42 - 2026-10-28

Three refusals learn to say the true thing: a name is not an identity, a gap goes
back to the reviewer, and an answer that cannot say what it read is not a verdict.

- **`start` never reuses a closed task document.** A document whose steps are all
  done is the memory of a finished feature. A new feature whose slug collides with
  it now gets the next number and its own file, instead of planning into the old
  one and reporting done without writing anything; a document that still has work
  is still reused, so one feature never leaves two files behind. The run names the
  document it kept.
- **The refusal asks the reviewer for the rows it left out.** `review --attach`
  already names every missing (path, lens) row and keeps the answer, but its last
  line said "fix the rows and attach it again", which is an instruction the caller
  can satisfy by inventing them. It now sends the gap back to the reviewer and
  says the answer is not edited to pass the gate; the kept file and the command
  that attaches it again are unchanged.
- **The reviewer declares what it inspected.** Every answer to a review that asks
  for lenses now carries `inspection`: `{"status": "completed", "paths": [...]}`
  with every changed path it read, or `{"status": "unavailable", "reason": "..."}`
  when it could not read the candidate at all. `review --attach` and the `--exec`
  path refuse an answer whose declaration leaves a changed path out, names a path
  the candidate does not change, or says unavailable, and the refusal says where
  the answer is kept so the reviewer is asked for it, not the caller. The prompt
  asks for the field with both statuses, and the schema it prints carries it. A
  verdict stored before this field existed still parses.
- **An older binary refuses the new key, on purpose.** A verdict carrying
  `inspection` is an error to a binary that predates it (`unknown field`), the
  same rule that refuses a typo: upgrade the installed odachi before the store
  holds a verdict the old binary cannot read.

## v0.11.41 - 2026-10-28

Two decisions stop being improvised: where the project's agent lives, and where a
review's question and answer live.

- **`agents.toml` remembers where the project's agent lives.** `project_agents`
  sits next to the client selection: `none` keeps the door in the home, `write`
  (or the key's absence) keeps the mirror in the tree. `init` honors the stored
  answer on a plain run, records it when a flag makes the choice, and an explicit
  `--agents` is the way back - no second flag to learn.
- **`doctor` names the choice instead of guessing it.** A project with no agent
  files is the clean mode when `agents.toml` says so, and a project that lost its
  files while nothing recorded why is sent to `--global --no-project-agents`; a
  home-only project with no door is sent to `--global`, not back to the mirrors.
- **The review prompt has a home beside its verdict.** `odachi review` writes it
  into the clone's store as `reviews/<tree>.prompt.md` and names it on stderr, so
  an agent hands a reviewer that path instead of inventing one. stdout stays the
  prompt alone, because a pipe into a reviewer must carry nothing else; a prompt
  in `proof/` is no longer something anyone has to improvise.
- **A reviewer can answer in a file.** A command that names `{verdict}` gets an
  empty file in the store's `answers/`, writes its JSON there, and stdout becomes
  diagnosis only: banners, spinners and logs stop being part of the answer. The
  path is quoted exactly as `{prompt}` is, the file is emptied before each lens
  group so a group cannot inherit another's answer, and it is removed once read. A
  command without the token still answers on stdout, and an empty file is refused
  instead of silently falling back to it.
- **An older binary refuses the new key, on purpose.** A store carrying
  `project_agents` is an error to a binary that predates it (`unknown keys`), the
  same rule that refuses a typo: upgrade the installed odachi before the project
  records the choice.

## v0.11.40 - 2026-10-27

The first row of the debt file closes with the test that keeps it closed.

- **The client selection's two answers are pinned end to end.** The adapter that
  reads the choice answers "both clients" only when git says the folder has no
  repository, and lets a git failure through so a verb stops instead of guessing;
  `init` names the clone's store when it cannot resolve it and writes no
  `agents.toml`. The repository port is a field with a nil default, so the test
  swaps one value instead of bending the machine's environment.

## v0.11.39 - 2026-10-27

The review debt of the store move is paid: three rows leave the debt file, each
with the test that keeps it closed.

- **The client selection tells "no repository" from "git failed".**
  `gitx.CommonDir` returns `gitx.ErrNoRepository` when a folder is not inside a
  repository, `store.Root` keeps that answer through the wrap, and the adapter
  that reads the choice answers "the project never chose" only for it. A broken
  repository now stops the verb and names the store instead of falling back to
  both clients, and `init` names the store it could not find instead of writing
  with an empty path.
- **`review` refuses a store it could not resolve.** Joining an empty root put the
  verdict relative to the process directory, which is the one place the evidence
  must never land; the verb says why and writes nothing.
- **The ledger's three git questions are pinned against a real repository**:
  `TreeOf` (full and short sha), `Has` (exit 1 is the answer "no", not a failure)
  and `CommitWithTree` (the oldest match for a tree, and a repository with no
  history).

## v0.11.38 - 2026-10-27

The store leaves the project: the rules stay with the team, the evidence stays
with the clone.

- **`.odachi/` is gone.** The rules the team decided — `odd/profile.toml`,
  `odd/fence.toml`, `odd/review.toml` and `odd/roles/` — live next to the plan in
  `odd/`, because they travel with the repository. The state that never travels —
  verdicts, versions, the write lock and the client choice — lives in the clone's
  store at `.git/odachi/`, which git does not track: no ignore rule has to keep
  evidence out of the tree it describes.
- **`odachi store move`** migrates a project from an older release: it adopts
  `.odachi/reviews` and `.odachi/agents.toml` into the store byte for byte, moves
  the old versions with the same rule as `undo --collect`, and prints the `git mv`
  lines for the tracked rules. It never moves or deletes a tracked file, and a
  second run changes nothing.
- **`review` no longer prepares an ignore file** and `init`'s `.gitignore` block
  is `/proof/` and `*.bak-*`. A verdict cannot invalidate itself because it does
  not live in the worktree.
- **`doctor` gained a thirteenth check** that names a leftover `.odachi/`: a
  problem when the rules inside it are the ones not being read, a warning when
  they are dead weight next to the new ones.

## v0.11.37 - 2026-10-27

The project keeps the rules and the plan; the agents can live in the home.

- **`odachi pack` installs the project's own skills** (`./skills`) next to the
  bundled ones, so a skill created with `odachi skill new` reaches
  `.agents/skills`. A project skill with a bundled name replaces it, and the
  replacement is named. A catalog that cannot be read is refused instead of
  looking empty.
- **`init --global` writes the subagents in the home** (`odachi-worker`,
  `odachi-verify` and `odachi-reviewer`) next to the door, so a project that
  never writes `.opencode/` keeps delegation and verification.
- **`init --no-project-agents`** skips the project's agent files; it needs
  `--global`, because skipping both halves would leave no agent at all. `start`
  refreshes a mirror it finds and never installs one.
- **`doctor` calls the clean mode a decision**: with the door and its subagents
  in the home, an empty project agents folder is reported as such, not as
  missing files.

## v0.11.36 - 2026-10-27

A professional is written as decisions, and a rule about the outside world carries its version.

- **`odachi role` creates a professional** from one source:
  `.odachi/roles/<slug>.md` holds the eight sections -who it is, what it asks
  before starting, how it works, what it refuses, how it knows it worked, what it
  does not know, where this comes from, and a probe-. `role new <slug>` writes the
  skeleton, `role <slug>` renders the agent of each chosen client and the role's
  skill from that source, and a role that would not fit in 60 lines is refused
  instead of truncated. `role` lists which roles are stale.
- **`role-craft` runs the interview and the research**: six questions, two scars,
  three to five sources with a date and a probe. The human approves the words and
  the binary writes them.
- **`skill-craft` explores and researches too**: a rule that claims how an
  outside system behaves carries the version and the date inside the rule, and
  what the model already knows needs no source.
- **A role is not a voice**: the persona stays how the agent talks (ADR 0006), the
  profile is the ceremony, and the role is what the professional knows and how
  they work. The rendered agent carries the project's block, so a role never
  skips the fence, the proof or the review.

## v0.11.35 - 2026-10-27

A correction counts.

- **`level` counts a requirement change as a vision decision**: the line shows
  `requirement changes` as its own number and the graduation sums them with the
  answered questions, so five changes and three finished features earn the senior
  proposal without a single answered question.
- **Craft decisions still buy nothing**: the agent deciding is not the human
  deciding, however many there are. The thresholds and the ladder do not move.

## v0.11.34 - 2026-10-27

The work is shaped before it is declared, and a requirement that moves says so.

- **`skills/shape` explores before the brief** and ends in
  `odachi start --brief`: what to read, which question the exploration opened, and
  no artifact of its own. The brief stays the human's, and `drive` loads the skill
  when the request does not fit in one reading.
- **The six `Risk:` items now decide the ceremony** (`commits`): `data`,
  `security`, `contract`, `concurrency`, `delivery` and `no-test` each name the
  lens to add and the verifier to run, and none of them buys a waiver. `slices`
  owns the trigger: a risk asks for `odachi-verify` even when the parent wrote
  inline.
- **`odachi ask --change "<words>" --req R1`** appends the human's words to the
  history as `CAMBIO R1` and reopens every done step that cites R1, dropping its
  tree mark, because the mark was evidence for the step the change invalidated.
  A change with no requirement is refused: `brief.md` moves only when the vision
  does, and that is a new feature.
- **ADR 0032 records Organic PDD**: the plan grows and the vision does not, and
  the ceremony is sized by what the change can break. The fence and the proof
  stay for every change.

## v0.11.33 - 2026-10-26

A mark is durable whichever form you use.

- **`plan --done T1 --commit <sha>` records the tree of that commit** (#88), the
  same durable mark the default writes: a rebase rewrites the commit and keeps the
  tree.
- **A commit this clone cannot read keeps the sha**, because inventing a tree would
  be worse than a sha that may not resolve.
- **The message names what was stored**, and `progress` resolves it to the landed
  commit.
- **Dogfooded on itself**: 0066's T1 is marked `tree:23b3085…` and `progress` prints
  `c8daa04`.

## v0.11.32 - 2026-10-26

The ledger names commits that exist.

- **`plan --done T1` records the tree of the commit** (#85), the identity a rebase
  cannot rewrite: the same one a proof and a verdict already use. It stays honest
  too: a step marked before its commit, or an older step marked after a later one,
  records no tree and says why.
- **`progress` names the landed commit**: it resolves the tree to the commit on the
  branch, resolves an older sha through its bytes when it landed with another one,
  and says `not on this branch` when the bytes never landed. The machine state
  carries `note`.
- **`--commit` still works** as the explicit form for a step whose commit you know.
- **No manual corrections**: 0064 and 0065 now report the commits that exist on
  `main` from marks that were local.

## v0.11.31 - 2026-10-26

A paid review survives a tool's refusal.

- **`review --exec` keeps the answer it cannot read** (#77): an answer the strict
  JSON refuses -an ANSI footer after the JSON, a declared verdict that contradicts
  its own findings- is written in the ignored store under a name its own bytes
  decide, and the blockage names the file and `odachi review --attach <file>`.
  Repair it and attach it instead of paying for the same review twice.
- **Refused and unreadable answers share one namespace**, because both are answers
  odachi would not seal, and the answer that lands is the one act that drops them.
- **A reviewer that never ran, timed out or exited nonzero keeps nothing**: there
  are no bytes to keep, and that half is still a plain failure.
- **The debt ledger names the waiver branch of `ReviewOwed`**: a WARNING the last
  verdict carried about unchanged bytes, written where the next person looks.

## v0.11.30 - 2026-10-26

A feature can waive its review, and the gate says so.

- **`Review: waived`** (#74): `odachi start --waive-review` records the human's
  decision in the task doc's preamble, next to `Base:`. A closed feature with the
  waiver is ready and its block says `review: waived by the human`; the proof and
  the fence still gate it.
- **The decision outranks `--require-review`**, so the lane holds in CI as well.
- **`progress` publishes `waived` as its own state**, not `not_due` and not
  `on_file`: a decision, not a silence. The contract test pins the name.
- **No size rule**: size is not risk, and a twenty-line hole can be the security
  one. A waiver is signed; a threshold is guessed.

## v0.11.29 - 2026-10-26

A feature can be chosen by its number.

- **`--feature NNNN`** (#71): `plan`, `ask`, `check` and `progress` act on the
  selected document instead of the highest, so a feature in flight can still be
  planned and judged while a higher one exists. Without the flag nothing changes.
- **An unknown number blocks and names the ones that exist**, negative numbers
  included; the gate asks the same selector instead of repeating the lookup, so a
  named number on a project with no task docs is refused like any other.
- **The flag is not `--for`**: `odachi ask --for P1` already means the question
  being answered, and one flag cannot mean two things.

## v0.11.28 - 2026-10-26

An agent file is code.

- **Only the names odachi writes are plumbing** (#61): the review matched
  `.opencode/agents/odachi*.md` by prefix, so a file the human wrote there -where
  tool grants and tokens live- disappeared from the candidate and never reached a
  lens. The generated names are a closed set derived from the writer.
- **Agent directories are code** (#61, decision 0030): their Markdown is
  instructions an agent obeys, not documentation, so a human agent file reaches
  the lens that reads what the agent may do.
- **The default is human**: a name odachi does not write is reviewed, even when it
  starts with `odachi`.

## v0.11.27 - 2026-10-26

A block arrives with its words in both readers, and a base is a commit.

- **`open_questions` is the list** (#57): `progress --json` publishes each open
  question as `{id, text}` instead of a count, so a panel in another program can
  show what blocks instead of a number. The count is the length of the list; a
  feature with none says `[]`, never `null`.
- **The human report names them too**: the same id and the same words, under the
  headline and before the steps. With no questions it prints exactly what it
  printed before, down to the line count.
- **The gate compares commits, not spellings** (#57, decision 0029): a receipt
  filed from `--base 6aeb877` answers the feature whose base is the full hash.
  Before, `check` refused it as "another diff" and `review` would not judge again,
  so the review was paid for and could not be used.
- **The plan lands first** (#58): the brief, the fence and the task doc reach main
  before the work, which is what decision 0028 asks and what a fence read from the
  base ref needs.

## v0.11.26 - 2026-10-26

A block teaches instead of pointing.

- **A blocking question arrives with its text** (#54): `check` names each open
  question with its id and its words, and its `next` line is the form that answers
  the first one - `odachi ask --answer "<answer>" --for P1` - instead of a bare
  `ask` that only prints usage. The gate still counts; the document knows the words.
- **Reporting only**: nothing is answered or written. A feature with no open
  questions prints exactly what it printed before.

## v0.11.25 - 2026-10-26

The report can be read by a machine.

- **`progress --json`** (#52): the same report as one object - `code`, `message`
  and `state` - so a renderer in another program depends on names, not on English.
- **One shape, two readers**: the state carries the feature, the steps with their
  commit, the proof state and the review state; the human lines are rendered from
  it, so the two cannot be told different facts.
- **The unreadable branches keep their cause**: when the store or the rule cannot
  be read, the report says which one and why, instead of blaming the tree.

## v0.11.24 - 2026-10-26

A rejection is not a reason to pay twice.

- **A refused review answer is kept, named and attachable again** (#49): the bytes
  the gate would not seal are written to the ignored store as
  `<tree>.refused-<digest>.json`, and the refusal says where they are and which
  `odachi review --attach` sends them back. The name is the content, so the same
  answer twice is one file and two different answers coexist - an occupied slot is
  never overwritten.
- **The answer that lands ends them**: once a review covers its rows, the sealed
  verdict is the record and the refused answers for that tree are removed.

## v0.11.23 - 2026-10-26

The report and the gate ask the same question the same way.

- **The review rule answers about the tree that stands** (#47): `ReviewOwed`
  defaulted to no base, and for the fence that means "compare against nothing" -
  every file looked like code, the close rule never fired, and `progress` claimed
  a verdict `check` never asked for. It now uses the same `HEAD` default `check`
  uses. v0.11.22 shipped that disagreement; this release is the fix.
- **The two verbs are tested together** (#47): an end-to-end test builds a real
  repository, closes a feature and requires `check` and `progress` to agree twice
  - at the close, and on the documents-only tail after it.

## v0.11.22 - 2026-10-26

The report asks the gate instead of guessing.

- **`progress` asks the gate** whether this tree owes a verdict, instead of
  deriving it from the steps: one owner, one answer. The documents written after
  the close are reported as not due, which is what `check` always said - the two
  verbs used to disagree on `main` after every release.
- **An unreadable review store is not a verdict owed**: the report says it could
  not read the store instead of sending the human to pay for a verdict that may
  already exist.

## v0.11.21 - 2026-10-25

The tool stops repeating itself and starts reporting.

- **The review prompt says each carried row once** (#42): a row travels from
  verdict to verdict as it is inherited, so the store accumulates copies of the
  same finding. The prompt now writes each row once and keeps the store's order.
  On this repository's own store the same tree went from 192 KB and 2599 rows -
  207 distinct - to 34 KB and 207 rows.
- **The dedupe never merges two answers**: two defects on one path, one line
  apart, or under another lens, both reach the reviewer.
- **`odachi progress`** (#43): the open feature, its steps with the commit each
  one landed in, the proof of this tree and the review it owes. Read-only by
  construction: its disk port has no writer to call.

## v0.11.20 - 2026-10-25

The review answers coverage before the gate asks for it.

- **`review` refuses an answer that leaves a row unanswered**: before sealing, the
  run asks the gate's own arithmetic about the lenses the answer claimed, measured
  on the bytes git can hash now. Nothing is written, and the refusal names the file
  and the lens.
- **A split run still stores**: the promise covers the lenses the run was asked
  for, so a sensitive candidate judged by two models files each group's answer, and
  the store keeps growing by lens.
- **A complete answer is untouched**: same message, same file, no new line.
- **The failure moves to where it is cheap**: the rows are named while the answer
  is in hand, instead of at `check`, after the reviewer call was paid for and the
  verdict could no longer be rewritten.

## v0.11.19 - 2026-10-25

The tool says what it does not read.

- **`doctor` names the documents odachi does not read**: the `.md` files in
  `odd/tasks` whose name is not `NNNN-slug.md`. It is a warning, not a failure: a
  project may keep another shape - ODD names a feature without a number - and the
  two memories can live in the same folder.
- **`start` and `summary` say the same sentence**: at adoption, and in the
  report - including the case where every document is foreign and the folder used
  to look empty.
- **Numbering stays**: the number is how odachi knows which feature is current,
  and the README says so. Odachi does not read, rename or migrate the foreign
  documents.

## v0.11.18 - 2026-10-25

Closing a feature is reviewed; publishing it is not.

- **The close is the change that marked the last step**: the gate compares the
  task doc at HEAD with the worktree, so the verdict is owed by the closing
  commit - even though that commit only carries the document.
- **Documents after the close do not reopen it**: a release note, or a note added
  to a closed task doc, is ready with its proof; the verdict already judged the
  same code, and asking again would charge it twice.
- **Code after the close still owes a verdict**: the exemption is only for
  paperwork, so the gate never goes blind.
- `--require-review` still forces one tree; `--no-review` still turns the mode
  off for one run.

## v0.11.17 - 2026-10-25

The review asks for an answer, not for work.

- **The reviewer reads; it does not act**: the first rule of the prompt says
  everything it needs is inside, so it must not run commands, edit files or read
  others. A reviewer that does puts its transcript in stdout, and stdout is the
  verdict; the retry asks again with no tools.
- **Nothing due is named as such**: `review` on a change with no lens says
  `this change asks for no lens: documents only` and offers `--lens all`. It no
  longer says `already has a verdict` when the store is empty: that sentence is
  only said when it is true.
- **The first close under the once-per-feature gate**: the feature's document
  recorded its base, the plan and the steps needed only their proof, and the
  closing commit demanded the one verdict for the diff the feature merges.

## v0.11.16 - 2026-10-25

The review asks once per feature.

- **The task doc records the commit the feature started from** (`Base:`): `start`
  reads HEAD and writes it, `taskdoc` parses and renders it, and a canonical
  rewrite does not report it lost.
- **With the mode on, the gate asks once**: the trees inside an open feature owe
  only their proof, and the tree that closes the feature owes the verdict. That
  verdict has to judge the diff the feature merges, not the closing document - so
  it must name the feature's base, and the message is the runnable
  `odachi review --base <sha>`.
- **Old documents keep the old rule**: a feature with no steps, or a document
  written before the base line, still asks per tree. `--require-review` asks for
  one tree; `--no-review` turns the mode off for one run.

## v0.11.15 - 2026-10-25

The review's answer belongs to the tree it judged.

- **Each group is filed under the tree the prompt named**: the run holds the
  candidate it read once, so a reviewer that edits the worktree cannot move the
  answer, split it across two trees or strand the lens that landed first.
- **A moved worktree stops the run**: the message names the file that keeps the
  answer, how many lenses landed, and the tree that exists now - those bytes ask
  for their own review. The paid answer stays as evidence of what was judged.
- **The project's own gate is on**: `.odachi/review.toml` says `on` locally, and
  CI keeps enforcing the fence and the proof with `--no-review`, because a
  verdict is evidence about one tree and lives in a store the tree ignores by
  design.

## v0.11.14 - 2026-10-25

A stopped review keeps what it already paid for.

- **Each model group's verdict is stored as it lands**: a reviewer that fails,
  times out or answers something that is not a verdict no longer throws away the
  lenses the earlier groups answered. The run says `saved 2 of 4 lenses` and names
  the ones still owed, with the command that asks for them.
- **`review --show` calls a half answer partial**: `review of <tree>: BLOCK
  (partial: 1 of 2 lenses)`, and the gate still blocks on the lens nobody judged.
  A partial verdict is not a review.
- **The store never replaces an answered lens**: a test pins the refusal, even
  when the second file also brings lenses the tree still owes.

## v0.11.13 - 2026-10-25

The review knows who judged it, and the human chooses who that is.

- **The verdict cannot come from the model that wrote the tree**: the proof
  records the model behind `--run`, the verdict records the one that judged, and
  `check --require-review` refuses the lenses the builder judged (`review_self`),
  naming the fix. It is disclosure, not identity: a model name can be lied about.
- **Each lens can name its model** in `.odachi/review.toml`: the review runs one
  reviewer per model group, so a cheap model carries the cheap lenses, and
  `review --show` prints the plan before anything is paid for. The verdict grows
  by lens — a second run adds what the tree still owes and never replaces an
  answer already stored.

## v0.11.12 - 2026-10-25

The release says which source built it.

- **Every archive carries a provenance attestation**: after the release is
  published, the workflow attests each archive named in the signed checksums
  with the repository, the commit and the tag that produced it.
  `gh attestation verify <file> --repo biggs-100/odachi` reads it without a key.
- **The chain has three links and the docs say what each proves**: the signature
  says who published, the attestation says which source built it, and the SBOM
  says what is inside. None of them says the source was good.
- **The attestation can be retried without republishing**: it runs after the
  release is public, so a failure there is fixed by running the workflow by hand
  on the tag; that run only attests.

## v0.11.11 - 2026-10-25

The release proves where it came from.

- **The checksums are signed without keys**: the tag workflow signs
  `checksums.txt` with cosign and the OIDC identity of the run, so there is no
  key to store, rotate or leak, and the release publishes one bundle
  (`checksums.txt.sigstore.json`) that holds the signature and its certificate.
  The installer still verifies `sha256` only.
- **Every archive ships an SBOM** (SPDX, by syft): what went into the file is
  published beside it.
- **The README and SECURITY say how to verify and what the chain does not
  prove**: the signature says who published, the SBOM says what is inside, and
  provenance is not published yet.

## v0.11.10 - 2026-10-25

The gate stops being ceremony: a red run says why, and a change cannot loosen the
policy it must obey.

- **A failed run keeps its whole output** in `proof/<feature>.failed.txt` —the
  tree, the command, the exit code and every line, including a timeout—, so the
  one-line verdict sends you to the evidence instead of replacing it. A green run
  removes the stale file, and a file the tool could not write is never named.
- **The fence only loosens with a new vision**: `check --fence-from` compares the
  non-goals of the base brief and every rule, path and dependency the base fence
  had. Dropping one blocks and the gate names it; adding is stricter and passes;
  a fence the tool cannot read blocks either way.
- **One workflow judges a tree**: `.github/workflows/gate.yml` runs on Linux,
  macOS and Windows, builds odachi from the base ref and uploads `proof/` even
  when it fails. `ci.yml` is gone, so no two files can disagree about the same
  tree.
- **Review pays for itself**: `review --show` answers whether this tree asks for a
  verdict and why, a documentation-only tree is not asked for ceremony, and the
  rule is one review, one correction, one re-review.
- **main only takes a gate that passed**: `.github/main-ruleset.json` requires the
  four checks, needs no approval from a second human and has no bypass list; the
  policy is a file in the repo, so a settings change can be reviewed.
- **A stranger can go from zero to a merged pull request**: CONTRIBUTING with the
  loop and the exact local command, issue forms, a PR template that asks for the
  proof, SECURITY and private vulnerability reporting, and a ruleset that does
  not demand a second person to merge.
- **The plan lands before the work**: `start` ends by telling the agent to commit
  the plan —brief, fence, profile and task doc with its steps—, and the drive
  skill makes that commit step two of the cycle. ADR 0028 records the three
  resets that paid for the rule.
- **The suite is deterministic**: no test reads state another test writes, the
  version line is pure, and the ldflags seam is pinned by reading the source
  instead of mutating the package.

## v0.11.9 - 2026-10-24

The gate can run in CI without trusting the change.

- **`check --fence-from <ref>`**: the fence is read from a git ref, not the
  worktree, so a change cannot loosen the policy it must obey. A missing file at
  a valid ref is no fence; a ref that does not exist blocks. The proof names the
  ref it used.
- **The CI recipe builds the gate from the base ref**
  (`.github/workflows/gate.yml`) and uploads `proof/` even when the gate fails.
- **`ODACHI_NO_SETUP=1` means binary only**: both installers skip `setup`, and
  `setup` and `init --global` refuse with the variable named, so a shared runner
  never gets a home door or a default agent.

## v0.11.8 — 2026-10-24

Installing odachi is enough.

- **`odachi setup` leaves the machine ready**: the bootstrap skill in
  `~/.agents/skills`, the home door for the agent clients found, and OpenCode's
  default agent. It runs from any folder and writes nothing in it.
- **The installer runs it**, so the user only installs and talks. The door and
  the bootstrap skill now say the agent runs `odachi init` and `odachi start`
  itself, asking permission before writing in a repository that is not the
  human's, instead of asking the human to type them.

## v0.11.7 — 2026-10-24

The global door is the default agent.

- **`odachi init --global` makes odachi OpenCode's default agent**: it writes
  `default_agent` in `~/.config/opencode/opencode.json` (or the `.jsonc` that
  already exists), creating the file with the published schema when there is
  none. A default the human picked before is named and left alone; a config that
  cannot be read safely is refused with the exact member to add by hand.
- **The edit never rewrites the file**: comments, key order and CRLF survive, and
  `doctor` notes when the door is installed but the config does not choose it.

## v0.11.6 — 2026-10-24

The gates read the worktree once.

- **`check` and `review` measure the candidate once**: the tree, the changed
  paths and the diff come from one git read instead of three, and the proof, the
  floor, the lens class, the prompt and the receipt reuse that value. A review
  drops from about 13 git processes to 6; `review --show` goes from 1029ms to
  351ms warm, and `check` from about 1009ms to 340ms.
- **No weaker gate and no new dependency**: the fence, the lenses and the
  coverage gate read the same bytes as before; a reviewer that edits the tree is
  still measured again before the verdict is filed. The suite wall is
  machine-bound: 23.8-24.6s in a quiet window, 25.6-36s under background load.
  The work removed is the honest signal; see ADR 0019.

## v0.11.5 — 2026-10-24

The reviewer inherits the session's model; review answers come from stdout, and
the proof keeps both streams.

- **`{model}` in `review --exec` is the model the session is using**: the
  reviewer judges with the same model you are talking to, so the verdict and the
  change come from the same eyes. The order is `--model`, then `ODACHI_MODEL`
  —the contract any host can set—, then the OpenCode session through
  `opencode session export`. A command without the placeholder runs exactly as
  before.
- **No fallback, and a flag that changes nothing is refused**: when `{model}` is
  asked for and no model resolves, the review stops before starting the reviewer
  and names `--model`; `--model` without `{model}`, or without `--exec`, is
  named instead of silently dropped. The receipt records the command that really
  ran, never the template.
- **`review --exec` reads the verdict from stdout**: a reviewer CLI's banner,
  warnings or ANSI on stderr no longer poisons the JSON answer. When the
  reviewer fails, the message shows stdout and stderr together.
- **The shell ports report the streams separately**: `check --run` still writes
  both into the proof — a failing test's stdout and its build errors are both
  evidence — while the review reads only the answer. This fixes a real failure
  found by dogfooding `opencode run`: the verdict was correct and the ANSI
  banner made it unparseable.

## v0.11.4 — 2026-10-23

The lens class can be raised by the diff's content: a concurrent or
credential-touching change to an ordinary file now asks for the four lenses.

- **`Assess` raises, never lowers**: the path class is the floor, and added
  lines that spawn concurrency (`go func`, `chan`, `sync.`, `atomic.`, `mutex`,
  `<-`) or touch secrets (`password`, `secret`, `token`, `api_key`,
  `private_key`) raise a focused candidate to full. A passive candidate (a
  document) is immune to prose, and no content lowers a class.
- **`review` and `check` share the assessment**: the prompt's lens list, the
  stored verdict's lenses, and the gate's per-file coverage demand all come from
  the same deterministic reading of the diff. The check port gained `Diff`.
- **Still deterministic and opt-in**: no model decides the class, and the
  human-owned review switch is untouched. See ADR 0017.

## v0.11.3 — 2026-10-23

Review receipts: a verdict now answers per file and per byte, and a corrected candidate pays only for what it changed.

- **Per-finding content receipts**: every finding and every clean pass carries
  the git blob id of the file it judged (path, blob, lens). The tool derives the
  blob from git; a reviewer-declared hash is overwritten. A finding whose file
  git cannot hash (deleted, or invented by the reviewer) is stored without a
  receipt, and the gate treats it as answering nothing.
- **Coverage is the gate's arithmetic**: with `--require-review`, every changed
  file whose bytes git can measure must answer each lens its path class
  requires — a finding, a clean pass, or a dismissal. A stale receipt, a
  missing row, or a verdict written before receipts existed blocks with exit 2
  naming `odachi review`; an approval the tool cannot prove is not used.
- **Potency is per blob**: storing a verdict inherits the standing rows of
  earlier verdicts whose blob still matches, so the fix loop re-reviews the two
  files you touched and inherits the rest. The prompt lists what is already
  carried, and a repeated finding is filed once.
- **Dismissals are first-class**: a finding refuted with evidence travels as a
  dismissed row with its reason and a fingerprint (path, line, lens,
  evidence). A re-raise of the same fingerprint stays dismissed; the verdict
  says what it discarded, and `--show` prints the reason. Dismissed rows answer
  coverage and never block.
- **A deterministic floor runs before any lens**: the diff names the added
  lines a finding may cite and the categories they touch (concurrency, secrets,
  network, failure, time). A causal BLOCKER or CRITICAL that cannot cite them
  is downgraded one tier before the verdict is stored — blocking is earned. The
  floor reports the path class; it does not change ADR 0010's lens selection.
- **Candidate diffs use the merge-base**: `Diff` and `ChangedPaths` compare
  against the fork point instead of the base tip, so another branch's commits
  no longer appear as deletions in the candidate. This is a behavior change for
  scripts that pinned the old two-dot diff.

## v0.11.2 · 2026-10-23

The version that adds a method for attacking a backlog, cutting delivery into reviewable PRs, writing Go tests, auditing skills, and speaking English in everything that is not chat.

- **English artifacts, docs and CLI**: the brief, task doc, `AGENTS.md` block,
  skills and CLI messages are English; the agent chat stays in the human's
  language. The old Spanish headers remain readable as aliases, so closed task
  docs and old briefs still parse, and a pre-commit hook installed before the
  migration is still odachi's: `init` backs it up and rewrites it in place.
  This is a breaking change for custom skills and scripts that pinned the old
  `Cuándo / Reglas / Nunca` sections or the Spanish CLI messages.
- **`triage`**: classifies each report by root cause, groups the ones that share a
  cause into one step, reproduces before fixing, and closes with the test that
  fails. It does not copy GitHub labels or the ODD grammar: it uses `odachi ask`,
  `odachi plan`, `tests` and `check --run`.
- **`pull-request`**: one work unit per PR, a 400-line budget, a single chain
  strategy (stacked or feature branch chain), an approved issue or `Refs`, and
  `odachi check --run` before review. The slug is `pull-request` because the skill
  format requires names of three characters or more.
- **`go-tests`**: complements `tests` with the Go pattern: a table with `t.Run`,
  `t.TempDir`, `testing.Short`, `teatest`, golden files, mocks at the edge and the
  focal package before the suite. It does not repeat the discipline of `tests`.
- **`skill-improve`**: audits a skill that is not used or grew. It starts with the
  trigger, requires observable rules, cuts at 60 lines, leaves one owner per rule,
  and sends you to `skills --check`/`--write`. The format is still validated by the
  binary.

## v0.11.1 · 2026-10-23

La versión que protege el contexto del agente principal y termina de limpiar el arranque.

- **`init --agents`**: `odachi init --agents opencode|claude|both` elige qué
  clientes reciben el agente nativo y la puerta del home. La elección queda en
  `.odachi/agents.toml`; `start`, `level` y `doctor` la respetan, y un `init`
  posterior sin flag la conserva. El default sigue siendo `both`, y los archivos
  del cliente que no elegís no se borran: sólo dejan de escribirse.
- **Subagentes nativos**: `init` escribe `.opencode/agents/odachi-worker.md` y
  `.opencode/agents/odachi-verify.md`, y sus equivalentes en `.claude/agents/`.
  El worker implementa un paso acotado; el verifier es read-only. `slices` manda
  a pasarles el task doc, el paso, las superficies de edición, los comandos y los
  paths exactos de las skills; vuelven con evidencia y ≤2k tokens. El padre
  conserva git, el gate y la visión.
- **`init --skills home`**: todas las skills en `~/.agents/skills`, sin copia en
  el proyecto. El default sigue siendo `project`. `doctor` cuenta y compara las
  dos ubicaciones, y no afirma que las skills ajenas del home coincidan.
- **`*.bak-*` ignorados**: `plan`, `ask --craft`, `start` y `level` dejan backups
  para `undo`; el bloque de `.gitignore` los mantiene fuera de git.
- **`start` sin TTY**: si el terminal muere en la primera pregunta, el mensaje
  nombra los flags que faltan en vez de "no pude leer la respuesta".
- **Review proporcional**: `check --require-review` ya no pide veredicto cuando
  el cambio es sólo documentación; el candidato pasivo no tiene lentes que
  responder. El clasificador de riesgo suma contratos públicos (`api`,
  `openapi`, `proto`, `graphql`, `sdk`), concurrencia (`lock`, `mutex`, `race`,
  `async`, `worker`, `queue`) y entrega (`.github`, `release`, `goreleaser`,
  `terraform`, `helm`, `k8s`, `pipeline`).
- **Modo de review del proyecto**: `odachi review mode status|on|off` escribe
  `.odachi/review.toml`; `check` lo lee, y `--require-review`/`--no-review` lo
  override por una corrida. Sin archivo, el default sigue siendo `off`.
- **Backups fuera de la vista**: las versiones que `undo` guarda antes de
  reescribir un artefacto viven en `.odachi/backups/`, con la ruta relativa del
  archivo. `undo` sigue leyendo los backups viejos al lado, y `undo --collect`
  los mueve con un rename que conserva bytes y fecha; la segunda corrida no hace
  nada. Los archivos fuera del proyecto conservan el backup al lado.
- **Timeout accionable**: un `check --run` que no termina sugiere duplicar
  `--timeout` en vez de repetir el mismo comando.
- **`init --global`**: la puerta del home es opt-in. `odachi init` ya no escribe
  `~/.config/opencode/agents/odachi.md` ni `~/.claude/output-styles/odachi.md`:
  instalar un agente `primary` en el selector de cualquier carpeta es una decisión
  del humano, y como efecto de `init` podía reemplazar el agente que la persona
  ya había elegido sin que lo pidiera. `odachi init --global` la instala para los
  clientes elegidos, y nada se borra: lo ya instalado sigue donde está. `doctor`
  nombra `--global` cuando la puerta falta o no coincide.
- **`doctor` ve la puerta global**: el chequeo 11 también compara
  `~/.config/opencode/agents/odachi.md` y `~/.claude/output-styles/odachi.md`.

## v0.11.0 · 2026-10-22

La voz vive en el bloque, el agente nativo la espeja, y el doctor ve lo que `init` dejó.

- **La voz vive en el bloque**: `AGENTS.md` termina con `## Cómo hablás`, y esa
  sección la escribe el binario. `odachi init` pone la voz de `novato` si todavía
  no hay perfil; `level --to <perfil>` cambia el perfil y la voz juntas, y `start`
  también, porque es el otro comando que decide el perfil.
- **El agente nativo es un espejo del bloque**: `init` escribe
  `.opencode/agents/odachi.md` y `.claude/output-styles/odachi.md` con las mismas
  reglas y la misma voz, para elegir odachi con Tab (OpenCode) o `/config` (Claude
  Code). Medido en OpenCode 2.1.3: la carpeta se lista como agente primario cuando
  el proyecto tiene `.git`; `mode/` también funciona, pero la documentada es
  `agents/`. `level --to` y `start` refrescan el espejo si ya existe; `doctor`
  avisa si dejó de coincidir. Las dos personas por perfil de la v0.10.0 quedan
  como restos: `init` las nombra y no las borra.
- **La puerta del home**: el mismo `init` deja `odachi` en
  `~/.config/opencode/agents/odachi.md` y `~/.claude/output-styles/odachi.md`, así
  que aparece en **cualquier** carpeta, no solo en la que corrió `init`. Es un
  despachador corto, no una copia de las reglas: si la carpeta tiene el bloque de
  `AGENTS.md`, lo lee y adopta su voz; si no, carga la skill `bootstrap` y guía
  para arrancar con `odachi init` y `odachi start`.
- **`doctor` pasa a once chequeos**: el 9 dice las dos verdades de las skills (el
  registro del proyecto y lo instalado en `.agents/skills`), y el 11 compara la voz
  del bloque con el perfil, y el agente nativo con el bloque.
- **Un proyecto nuevo arranca en `novato`**: `start` ya no infiere `senior` porque
  haya `.git`; `init` es justamente quien crea ese repo. Sin perfil escrito y sin
  brief, el default es `novato`. `senior` se gana con `odachi level --to`, o se
  infiere de un brief que la persona ya trajo.
- **La delegación tiene una política y un dueño**: `slices` gana sobre otras
  skills globales; `drive` manda a cargarla cuando el pedido no entra en una
  lectura; un escritor se delega solo por paralelismo real (unidades disjuntas) o
  por el context backstop, nunca por tamaño o cantidad de archivos. Se tomaron
  los triggers y el handoff acotado de gentle-ai, no su orquestador.
- **ODD-lite, sin cambiar el formato del task doc**: `drive` exige requisitos
  citados verbatim y un "Listo cuando" observable, `tests` con RED antes del
  código, el registro del paso delegado con `ask --craft` y la línea de riesgo de
  `commits`. No se adopta la gramática `Specs`/`Tasks`/`Log` de gentle-ai; la
  decisión está en el ADR 0008.
- **El instalador no depende de `Get-FileHash`**: si el cmdlet no existe —pasa en
  Windows PowerShell 5.1—, verifica el sha256 con `certutil`. Probado con las dos
  herramientas sobre la release real.

## v0.10.0 · 2026-10-22

Listo en un comando, y el agente sabe manejarlo.

- **`odachi init` deja el proyecto listo de una vez**: historial de git si faltaba,
  bloque de reglas en `AGENTS.md`, bloque de `.gitignore`, hook de pre-commit, las
  skills del proyecto y las dos personas. Ya no son dos comandos.
- **Skills en la ruta estándar**: `.agents/skills` en el proyecto y `~/.agents/skills`
  en el home, la convención cross-client que recomienda el estándar Agent Skills.
  `pack` gana `--scope project|home`, `--dest <carpeta>` y el pseudo-agente `standard`.
- **Personas generadas**: `novato` y `senior` en `.opencode/agents/` (se eligen con
  Tab) y en `.claude/output-styles/` (desde `/config`). La ceremonia la sigue
  decidiendo `.odachi/profile.toml`; la persona decide solo cómo habla el agente.
- **Skill `drive`**: el manual que el agente carga para manejar la herramienta —los
  verbos, el ciclo de una feature, qué significa cada exit— y la línea del bloque
  que lo obliga a cargarla.
- **Skill `bootstrap`** en el home: enseña a arrancar un proyecto donde todavía no
  hay nada, que es justo cuando no hay otro archivo que leer.
- **`doctor` con dos formas**: en una carpeta que no es un proyecto responde una
  situación y un paso (`init`, después `start`); en el resto, los diez chequeos.
- **El bloque de `AGENTS.md` dice qué es odachi**, quién decide qué y dónde está el
  contexto; y quedó fijado byte a byte por un test.
- `pack --json` cambia de forma: las filas son por destino e incluyen `agents: []`.
- `doctor --json` gana `lines`.

## v0.9.1 · 2026-10-21

- La línea de cierre de `init` deja de asumir que el humano commitea: el que guarda
  el avance es el agente.

## v0.9.0 · 2026-10-21

- **`init` crea el historial de git** cuando falta, y lo explica en castellano llano.
  Antes se negaba con «corré git init» a alguien que no sabe qué es git.
- Sin git en la máquina sigue bloqueando, con una frase que dice qué instalar.

## v0.8.0 · 2026-10-21

- Tres skills nuevas para cerrar el currículum del que construye: `commits`,
  `writing` y `skill-craft`. Ocho skills, todas bajo el cap de 60 líneas.
- El registro de skills pasa a preguntarse al payload y al directorio: cambiar el
  contenido ya no obliga a editar un test.

## v0.7.1 · 2026-10-21

- La regla de delegación queda con presupuesto medible (~3 lecturas o ~10k tokens)
  y retorno acotado (≤2k tokens, evidencia `path:line`), y el trigger se amplía.

## v0.7.0 · 2026-10-21

- La mitad de delegación de un orquestador vive en la skill que ya parte pedidos:
  delegar por razón, pasar el task doc por path, verificar contra el brief y no
  contra el resumen, y no delegar nunca la decisión de visión.

## v0.6.1 · 2026-10-21

- El doctor deja de mentir en dos casos medidos: un worktree (`.git` es un archivo)
  ya no se reporta como «no hay repositorio», y un perfil inventado sin `extends`
  ya no valida en verde.

## v0.6.0 · 2026-10-21

- **`doctor`**: diez chequeos de salud, de solo lectura, cada uno con el comando que
  lo arregla.
- **`bench`**: ocho recorridos que clasifican cada bloqueo (in-band, out-of-band,
  templated, interactive, dead end) con ratchet: un dead end hace fallar la corrida.

## v0.5.0 · 2026-10-21

- **Las cuatro lentes 4R** con selección por evidencia: sin lentes para documentos,
  `reliability` para lógica, las cuatro cuando el cambio roza algo que puede lastimar.
- **Refutador**: un `BLOCKER` causal solo se destraba citando la línea del diff que
  lo desmiente. Lo que no se puede citar, no se refuta.

## v0.4.0 · 2026-10-21

- **`review`**: veredicto de diseño por árbol, content-addressed en
  `.odachi/reviews/<hash>.json`. La aprobación no puede flotar porque el nombre del
  archivo *es* el estado del código.
- `check --require-review` opcional: el review como compuerta.

## v0.3.1 · 2026-10-21

- `init` avisa cuando `odachi` no resuelve desde un shell: una regla que el agente no
  puede ejecutar es peor que ninguna.

## v0.3.0 · 2026-10-21

- **OpenCode** como destino de las skills, con destinos por alcance (proyecto vs home)
  y la lista de agentes soportados.
- `undo` respalda lo que reemplaza y se niega a elegir por vos cuando hay más de un
  candidato; `plan` agrega y cierra pasos sin editar Markdown a mano.

## v0.2.0 · 2026-10-21

- **`pack`**: las skills llegan al directorio nativo de cada agente, sin borrar nada,
  con respaldo de lo que difiere.
- Registro generado de skills, validador y creador.

## v0.1.0 · 2026-10-21

- Primera versión que se instala: `start`, `ask`, `level`, `summary`, `check`, `init`,
  `--version`, hook de pre-commit, instaladores `sh` y `ps1`, y seis binarios.

