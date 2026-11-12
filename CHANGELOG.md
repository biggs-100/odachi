# Changelog

What changed in each released version, in the order a person reads it.
Versions before 0.10.0 were rebuilt from their tags: the repo did not have this
file yet, and a changelog that starts today erases the history of whoever arrives
tomorrow. Every `## vX.Y.Z` heading here names a tag: a version with no tag is
not a release (`docs/releases.md`). `## Unreleased` is the worklog, and it is
not a version.


## Unreleased

- **The count counts answers, not landings.** A prompt that does not fit is asked in
  parts, and the parts of one answer are one answer: every landing used to add a draw,
  so a split candidate recorded `draws: 2` for a run of one and `6` for a run of three,
  and the next run read that inflated count as already paid. Found by a second draw over
  the tree that shipped `--runs`, which the first draw had approved. The other two
  defects that draw found are paid with it: the command a failed run hands over now
  carries `--exec` and the lens it was told to judge, because pasting a hint the verb
  refuses pays nothing; and the refusal that says a lens already holds enough names the
  count the file holds plus one, not the count the refused run asked for. Patch (0276).

## v0.39.0 - 2026-11-11

- **A verdict can be paid by more than one draw.** The review was one answer per lens
  -one sample of one judge- and the store refused a second run over the same tree, so a
  defect that draw did not think of stayed invisible behind an approval. `--runs N` asks
  each lens N times and files the union as one verdict: a defect is identified by its
  file, bytes, lens, place and severity, so two draws that name it pay one row, and a row
  one draw saw stands because the other draws being quiet is one sample each, not an
  acquittal. The verdict now records how many answers paid each lens, `--show` prints it,
  and a lens paid by one draw is brought to N by a later run that pays only the missing
  answers -the union only adds, so drawing again can never turn a `changes_requested`
  into an `approved`-. A draw that does not answer leaves the ones that did, names the
  draw it lost and the top-up that pays the rest. The schema moves to 5, so an older
  binary says upgrade instead of calling a good verdict broken. Minor (0275).

## v0.38.14 - 2026-11-11

- **A record that does not decode is refused.** The brief corrupted on 0271 decoded as UTF-8 -the bytes
  were the text of one encoding read by another-, so no parser and no gate saw it and only a model did. The
  gate now refuses a file whose bytes are not UTF-8, names where the first invalid byte is, refuses a file
  saved with a BOM, and refuses the sequences a UTF-8 file leaves when a console reads it as CP850 and
  writes it back -in the brief and the TOMLs, never in a task document, because prose may have to name
  them. Patch (0274).

## v0.38.13 - 2026-11-11

- **`review --since` takes the tree a person copies.** The store names a verdict by forty characters, and
  `--since` with a prefix looked for a file named after the prefix: "there is no verdict for the tree
  --since names (bbf385ca)", about a verdict the clone held. The correction now resolves the prefix against
  the store's own listing -the one `--all` prints- refuses an ambiguous prefix naming its candidates, says
  when nothing starts with it, and tells which tree the correction completed. Patch (0272).
- **The public face says what the repository does.** `odachi --help` now names `design` and `review probe`,
  and a test reads the dispatcher's own cases so a verb cannot be answered and hidden at once. The README
  lost the counts that rot -eighteen surfaces, 44 packages-, stopped denying the provenance attestation
  `gh attestation verify` reads, stopped claiming `install.ps1` has no automated test, and opens with a
  table of contents. Patch (0273).

## v0.38.12 - 2026-11-11

- **The brief remembers every feature.** `odachi start` wrote `brief.md` from its own answers, so the
  sentence behind each fence rule died with the feature that wrote it while the rule -append-only since
  0266- survived. A run now appends a block named after the feature, the block repeats the canonical
  headers the parser already reads, and the four lists are the union of every block: outcomes, done-when,
  non-goals and decisions. A requirement the file already carries is not written twice. Patch (0271).

## v0.38.11 - 2026-11-11

- **A requirement may start with a dash.** `odachi start --light -- "--runs 0 se rechaza"` wrote nothing: the
  rule against a step starting with a dash refused the line, and without the separator the message was the
  flag package's own "invalid flag". The rule now refuses the markdown bullet that doubles the checkbox
  (`- `) and says so, and `start --light` names the separator a line that starts with a dash needs. Patch
  (0270).

## v0.38.10 - 2026-11-11

- **The Windows installer no longer trusts a cmdlet.** `scripts/install.ps1` extracted the release with
  `Expand-Archive -Force`, and a Windows PowerShell 5.1 that inherits a `PSModulePath` carrying PowerShell
  7's modules loads the modern module, whose `-Force` calls `ExtractToDirectory` with the three-argument
  overload .NET Framework does not have: the installer died before it verified anything. It now opens the
  zip with `ZipFile` and extracts each entry, which ships with the framework. A fix, patch (0269).
- **A base that came before is not another diff.** `odachi check` compared a verdict's base with the
  feature's by equality, so a review that judged a wider diff read as one that judged something else: the
  gate asked for a verdict the store already had, next to the same run saying ready, and named a
  `review --base` the store answers with "this tree already has a verdict". The gate now asks git whether
  that base came before the feature's -earlier or equal covers, later starts inside it and still refuses-
  and a close that only moved documents keeps the verdict it has. Patch (0268).

## v0.38.9 - 2026-11-11

- **The fence keeps what it protects.** `odachi start` reads `odd/fence.toml` before writing it and keeps
  every rule already there, adding only the non-goals the new run brings; a fence it cannot read is left
  alone and named instead of overwritten. The rules are read and written through one adapter, so a file a
  run writes is a file the gate can read. Patch (0266).
- **Zero draws is refused, not measured.** `odachi review probe --runs 0` no longer becomes one draw in
  silence: the count is checked against the range the measurement holds and the message names the minimum.
  The row the last review left open in that file is paid. A fix, patch (0267).

## v0.38.8 - 2026-11-11

- **The measurement repeats.** `odachi review probe --runs N` asks every lens and arm N times and reports a median with the range it moved in: on the fixture, every planted defect was found in most draws, while the findings nobody planted moved from 3 to 7 and the clean arm from 0 to 5. One draw is not a measurement. A new flag, patch (0265).

## v0.38.7 - 2026-11-11

- **The clean arm pays what the lenses find.** The fixture's arm with nothing planted was reviewed three times
  and gave 10, 1 and 4 findings: the ten it named were real and are now fixed in the code, and the five planted
  defects are still found. The lesson belongs to the instrument -one run is one draw, so precision needs
  repetition- and `docs/review-research.md` carries the reading. Patch (0264).

## v0.38.6 - 2026-11-10

- **The lenses are measured.** `odachi review probe` asks the four shipping lens prompts about a fixture
  with defects planted in it and scores the answers against a written key: found, called low, missed, and
  findings nobody planted. Without `--exec` it prints the prompts and runs nobody. The first measurement is
  in `docs/review-research.md`: five of five planted defects found, and ten findings on the arm with
  nothing planted, all of them real. A new surface,
  patch (0263).

## v0.38.5 - 2026-11-10

- **A verdict is filed under the tree `check` names.** `odachi review` fills its candidate index from
  `HEAD`, the same parent `check` reads, so a path the base tracked and a later commit untracked no
  longer reappears in the candidate: `review` and `check` name one tree for one clean worktree, and an
  ignored file -a machine-local registry, a build folder- cannot move a verdict's id. A fix, patch (0262).

## v0.38.4 - 2026-11-10

- **An upgrade stops leaving backups in the home.** `odachi pack` -and the `setup` the installer runs-
  writes a backup only when the destination is inside the project, where the clone's store can hold it
  and `undo` can read it. A home skill (or a `--dest` outside) is replaced and reported as written, so
  `SKILL.md.bak-*` no longer accumulates beside it. A fix, patch (0261).

## v0.38.3 - 2026-11-10

- **A TUI configures an installation.** `odachi tui` opens a terminal wizard that detects the project,
  chooses the clients, the skills scope, the voice and the review switch and lens models, and applies
  each choice with the verbs that already exist: a failure stops the run and names what remains, and
  cancelling before applying writes nothing. The screens are Bubbletea + Lipgloss; the domain and the
  flags are untouched, and without a terminal the verb refuses and names `odachi setup`. A new surface,
  patch (0259).

## v0.38.2 - 2026-11-09

- **The hand-off can be inspected before it runs.** `odachi skills --resolve T1` prints every skill a
  step of the current feature would receive, with the exact `SKILL.md` each name resolves to and whether
  it comes from the project or the home. It reads only: a name no scope carries blocks with both places
  it looked, and a step that declares none says so. A new read-only surface on skills, patch (0258).

## v0.38.1 - 2026-11-09

- **A wrapped bullet arrives whole.** `odachi start --brief` now reads a bullet's continuation lines
  until a blank line, a header or the next bullet, instead of keeping its first physical line: a draft
  written at 80 columns no longer loses the tail of every requirement and decision. The join stops at a
  blank line, so a note written under a bullet stays a note. A fix, patch (0256).

## v0.38.0 - 2026-11-09

- **A rule is one bullet, and a paragraph is not a rule.** `skills --check` refuses a `## Rules` section
  written as prose and a rule over 300 characters, because both hide instructions that the line count
  reads as none. 140 would have split 78 single sentences; 300 catches the one rule that carries two
  ideas. `ui-craft`'s twelve rules in one paragraph are eight bullets now; `bootstrap`, `drive`
  and `tests` lose their prose for the same reason, and a wrapped bullet still counts as one rule. A
  contract change for the skills, minor (0255).

## v0.37.0 - 2026-11-09

- **A new project is born with the review on.** `odachi init` writes `odd/review.toml` with `mode = "on"`
  when the project has none, and never replaces one: the switch is the human's, so an existing file is
  kept byte for byte and the run says it kept it. A close on a new project now asks for a verdict, and a
  project with no reachable model closes by configuring one or by `odachi review waive`. A contract
  change, minor (0254).

## v0.36.5 - 2026-11-09

- **The attribution travels with the skill, and a file that cannot travel is refused.** `odachi skills
  --check` now names every file at a skill's root besides `SKILL.md`: the installer copies `SKILL.md`,
  `references/` and `assets/` only, so such a file is embedded in the binary and never reaches a
  project. 0252 shipped a `NOTICE.md` that way, and the installed skill pointed at it. `tui-design`
  carries its attribution in the body instead. A fix, patch (0253).

## v0.36.4 - 2026-11-09

- **A terminal is read before it is judged.** `tui-design` designs, critiques and refines a terminal
  interface from a captured frame: the floor (the cell, the roles, the glyph family, the refusals), a
  critique that scores ten heuristics with P0-P3 severities, twelve playbooks, the capture loop at three
  widths, and the terminal tokens in the `design/system.md` the project already keeps. The method is
  adapted from `gentle-shell`'s `tui-design` (Apache-2.0) with original text and a `NOTICE`. A new
  capability, patch (0252).

## v0.36.3 - 2026-11-09

- **ui-craft starts from the job.** The skill now carries the five jobs a surface serves —find, compare,
  act, read, decide— with the content it may use, the label that predicts its outcome, and the six states
  with the sentence each one says. The accessibility figures cite WCAG 2.2 and say that a ratio is not
  conformance. The three gaps the comparison with gentle-ui-design found in this project. A capability of
  the skill, patch (0251).

## v0.36.2 - 2026-11-09

- **Motion graphics is a skill.** A request for a title card, a stat, a logo sting or a short video for
  YouTube, Shorts or Reels loads `motion-graphics`: one idea and a beat sheet before a frame, the craft
  of easing and weight, the tells a generated piece repeats and the fix for each, type read at arm's
  length, the destination's frame and safe zone, and a deterministic render through HyperFrames with
  Remotion as the named alternative. A new capability, patch (0250).

## v0.36.1 - 2026-11-09

- **The heading speaks the block's language.** `odachi level --to` renames a block's `## Cómo hablás`
  heading to `## How you speak` when it rewrites the voice, so a block written before the language
  migration does not keep a Spanish heading over an English voice. A fix, patch (0249).

## v0.36.0 - 2026-11-08

- **The voice speaks neutral Spanish.** Both voices say it where the rule lives: no regional slang,
  no voseo and no local expressions -neutral Spanish, the same for every region- while the artifacts
  stay English. Every project reads the sentence in the block `init` writes, so its next `init`
  refreshes it. A contract change, minor (0248).
- **Install:** `curl -fsSL https://raw.githubusercontent.com/biggs-100/odachi/main/scripts/install.sh | bash`
  or, on Windows, `irm https://raw.githubusercontent.com/biggs-100/odachi/main/scripts/install.ps1 | iex`.

## v0.35.0 - 2026-11-08

- **The app writes no project agent at all.** The project half of `odachi init` is deleted: the
  `NoProjectAgents` option, the resolver that honoured it and the writers that mirrored the block in
  `.opencode/agents` and `.claude/output-styles` are gone, with the tests that pinned them. A stored
  `project_agents` key from an older release is still read, does nothing, and is never rewritten; a
  project agent file a release left is still named by `doctor`, with its path. A plain `init` records
  no decision now that none is forced: the store gains no `project_agents` key, and a flag still
  records what it named. A contract change, minor (0246).
- **Install:** `curl -fsSL https://raw.githubusercontent.com/biggs-100/odachi/main/scripts/install.sh | bash`
  or, on Windows, `irm https://raw.githubusercontent.com/biggs-100/odachi/main/scripts/install.ps1 | iex`.

## v0.34.0 - 2026-11-08

- **The gate sees the message.** `odachi init` installs the gate at `.git/hooks/commit-msg`, the one
  hook git hands the message, and retires odachi's own `.git/hooks/pre-commit` while leaving a hook a
  project wrote alone. A commit whose message carries `Odachi-Change: <line>` now passes: the gate
  reads the line before the commit exists, measures the fence against the worktree it always reads, and
  defers the proof to the tree the commit carries. `odachi check --message <path>` is what the hook
  runs, a message with no trailer keeps today's answer, and a repository with no remote keeps the lane
  closed. A contract change, minor (0245).
- **Install:** `curl -fsSL https://raw.githubusercontent.com/biggs-100/odachi/main/scripts/install.sh | bash`
  or, on Windows, `irm https://raw.githubusercontent.com/biggs-100/odachi/main/scripts/install.ps1 | iex`.

## v0.33.0 - 2026-11-08

- **No command refreshes a project agent any more.** `odachi start` rewrites the voice of the block
  and leaves a project agent file an older release wrote untouched, the same rule `level` follows.
  With `init` writing none, the tool never keeps alive a file `doctor` names as a leftover (0243).
  A contract change, minor.
- **Install:** `curl -fsSL https://raw.githubusercontent.com/biggs-100/odachi/main/scripts/install.sh | bash`
  or, on Windows, `irm https://raw.githubusercontent.com/biggs-100/odachi/main/scripts/install.ps1 | iex`.

## v0.32.0 - 2026-11-08

- **Upgrade note: the agent lives only in the home.** A script that ran `odachi init --agents <client>`
  to put an agent file in the project has no replacement for that file: the project keeps the rules
  and the hook, and the agent is the door in the home. `odachi skill new` now asks for
  `--scope project|home`; `--global` is accepted and changes nothing; `ODACHI_NO_SETUP=1` refuses a
  plain `init` too. Run `odachi init` once to record the answer in the clone's store.
- **An agent lives only in the home.** `odachi init` writes no agent file in the project and installs
  or refreshes the door of the clients `--agents` names; the craft defaults to `~/.agents/skills` and
  `--skills project` is the explicit pin. A project copy an older release left is named by `init` and
  never deleted. `odachi skill new` writes the user skill in `.agents/skills` or `~/.agents/skills`,
  with that folder's `SKILLS.md`. `doctor` names a project agent an older release left as a warning,
  with each file's path (0242).
- **Install:** `curl -fsSL https://raw.githubusercontent.com/biggs-100/odachi/main/scripts/install.sh | bash`
  or, on Windows, `irm https://raw.githubusercontent.com/biggs-100/odachi/main/scripts/install.ps1 | iex`.

## v0.31.0 - 2026-11-08

- **A delegated step is handed with `odachi delegate`.** The plan records the skills a delegated step
  owes, and the verb prints the prompt with the step, its requirement, the surfaces, the commands and
  the exact skill paths -or an explicit none-; the block names the verb, so the parent stops writing
  the section by hand (0240).

## v0.30.0 - 2026-11-07

- **The worker refuses a delegation without the skill paths.** When the parent omits
  `## Skills to load before work`, `odachi-worker` returns `status: blocked` and names the missing
  section instead of starting; a section that says there are none lets it proceed (0238).
- **The harness says what it costs.** `doctor`'s skills check names the bytes of the odachi block in
  `AGENTS.md` and the bytes of `SKILLS.md` beside the count it already prints, so the context an agent
  reads on every turn is a number (0231).

## v0.29.0 - 2026-11-07

- **A release says which bytes were judged.** `docs/releases.md` states that the tag carries the verdict
  of the feature's closing tree and that the steps after the close move only documents, and `check`
  prints that tree beside the proof (0230).
- **The skills embed is measured.** A hidden fixture beside the skills is read through the bundled payload,
  so a regression from `all:*` to `*` fails the suite instead of shipping a folder with files missing (0229).
- **The gate reads the newest verdict by the store's write time.** A correction that declares a date
  where its predecessors declared a time no longer loses to them: `check` asks the file's modification
  time, the fact `clean` already orders the store by, so an existing verdict is never named as
  missing (0228).
- **The installer names the process that holds the binary.** Windows refuses to replace a running
  `odachi.exe`; the installer now answers with the holding process's PID and the way out —close that
  terminal or the other project's gate, or install with `ODACHI_BIN_DIR`— and installs nothing (0227).
- **A non-goal says when it comes back.** `shape` asks for the condition that reopens each entry of a
  brief's "Not included" —a measurement, a second caller, a version— and its output returns that
  condition beside the path each one protects, so the list of what waits keeps the date it was
  postponed to (0226).
- **The declared `Risk:` meets the diff.** When the review classifies a change as full —it touches a
  path the classifier names— and the commits declare `Risk: none` or carry no line at all, `check`
  names the contrast and the paths in the ready output; a named risk stays silent, and the exit code
  and the verdict do not move (0225).

## v0.28.1 - 2026-11-06

- **The gate names what the change adds.** The ready output lists the new files and the dependency
  files the diff adds (`go.mod`, `go.sum`, `package.json` and the other manifests and lockfiles), in
  one line and without moving the verdict; the plan and the proof are filtered out (0224).

## v0.28.0 - 2026-11-06

- **The smallest change, named before it is written.** `shape` asks what the change reuses, adds and
  deletes, waits for three uses for the same reason, checks what the props describe, and inlines a
  wrong abstraction back first; `pull-request` carries the list and the reviewer's three questions
  (0223).

## v0.27.0 - 2026-11-06

- **The chain says where it stands.** A child PR names its tracker, position, base, dependency,
  next PR and what it leaves out, marks itself in the diagram, and retargets a diff that shows
  another unit; a branch is named for what it delivers, and a commit says what a revert would take
  with it (0222).

## v0.26.0 - 2026-11-06

- **The writing pair measures the review path.** `the-doc` names the changed-markdown command and
  the review checklist (first, out of scope, previous and next PR, one decision per section,
  acceptance), `pull-request` measures the 400-line budget with `gh pr view`, and `writing` loads
  when a document feels long, dense or hard to scan (0221).

## v0.25.0 - 2026-11-06

- **The artboard carries the tuning controls.** When the human tunes spacing, colour or type scale,
  `design` adds sliders bound to the artboard's CSS custom properties and a `Copy as prompt` button
  that copies the chosen values as text; the artboard stays self-contained (0220).

## v0.24.0 - 2026-11-06

- **The design pair covers product UX.** `ui-craft` ships `browse-and-compare` (the representation,
  search scope, filters, the loading pattern, the return to the place), and `design` sends every
  artboard through `the-read` (and `adaptive-interfaces` when it has states), treats another
  approach as another artboard, and reads a comment's screenshot first (0219).

## v0.23.0 - 2026-11-06

- **The verifier runs the focused package, not a second gate.** The delegation request names the
  package the change touches, never a second `go test ./...`: a full suite beside the gate competes
  for the machine and made 686 seconds out of a 207-second package (0218).

## v0.22.1 - 2026-11-06

- **The gate waits fifteen minutes for a suite that takes eleven.** `check --run`'s default timeout
  moves from 600 to 900 seconds: measured on this machine, `go test ./...` takes about 700 seconds
  (686 for `cmd/odachi` under `./...`, 207 alone), so the old default killed a green suite and the
  failed proof was empty. `--timeout` still tightens it (0217).

## v0.22.0 - 2026-11-06

- **A skill folder carries the assets it copies.** `assets/` holds a template or a schema the skill
  uses, declared under `## Assets` and installed beside `SKILL.md` and `references/`; an orphaned
  file or a tree under `assets/` is refused. `writing` ships `assets/doc.md`, the copyable skeleton
  of a long document (0216).

## v0.21.0 - 2026-11-06

- **A published text is read back.** A PR is compared with the file that was posted
  (`gh pr view <url> --json title,body,baseRefName,headRefName`), and any published text,
  an issue included, with the file that was written; a mismatch is fixed before the review
  is requested (0215).

## v0.20.0 - 2026-11-06

- **Verification is proportional to the blast radius.** A passive change is verified by reading
  the diff, a medium one keeps the writer's own verification, and only a change that names one
  of the six risks adds `odachi-verify`; a red is run on the base first, and a red that comes
  from the machine is named instead of hidden (0214).

## v0.19.0 - 2026-11-05

- **The block keeps notes, not builds.** `.git/odachi/scratch/` is for a note a later turn has
  to read (a plan, a draft) and `clean` empties it; a build artifact and a one-use transport (a
  binary, a pull request body, a prompt answer) live in the temp folder and die with the run
  (0213).

## v0.18.5 - 2026-11-05

- **`triage` learns two lessons from its reports.** A test written against the mechanism a report
  names that passes on the untouched tree accuses the diagnosis, not the report, so the test goes
  to the surface the symptom touches; and a thread that carries two failure modes closes neither,
  with a comment that names both and what each waits on (0212).

## v0.18.4 - 2026-11-05

- **`writing` carries the document and the comment.** The skill routes two new references:
  `the-doc.md`, the shape a guide, an RFC or an onboarding page needs (answer first, quick
  path, details, checklist, next step), and `the-comment.md`, the formula and the tone of a
  comment in a thread. The body stays under the cap and keeps its rules (0211).

## v0.18.3 - 2026-11-05

- **The doctor names an environment that cuts a command.** Check 15 measures the PATH this
  process hands to a child; above the 8191 characters cmd can pass it warns with the length,
  the limit and the repair, and a platform that does not cut the environment says so. A proof
  that fails because a binary disappeared now has a line that explains why (0210).

## v0.18.2 - 2026-11-05

- **The gate waits ten minutes for a proof, not two.** `check --run`'s default timeout is 600
  seconds, so a real suite is not killed by the default on a slow machine; `--timeout` still
  tightens it (0209).

## v0.18.1 - 2026-11-05

- **The four rows of the review ledger are paid.** `pack` counts a backup only when one ran and
  reports a skill that only gained a file as `written`; `skills --check` reads the gates table and
  the references list instead of searching the section; the doctor's drift test covers a
  reference the installed copy lacks; and a git that cannot answer `Head` is an error instead of
  an empty history (0207).

## v0.18.0 - 2026-11-05

- **Every skill says what it returns.** `odachi skills --check` refuses a skill without
  `## Output`, whether or not it carries references: the seventeen bodies that had none declare
  it, the scaffold `odachi skill new` writes includes the section, and `drive` and `slices` keep
  the depth that no longer fit under `references/` (0206).

## v0.17.0 - 2026-11-05

- **A skill carries its depth beside it.** A skill folder may hold `references/*.md`: the body
  names every one in `## References`, a router with more than one routes them in `## Decision
  gates`, and a skill with references says what it returns in `## Output`. `odachi init` and
  `odachi pack` install the folder, `doctor` compares it file by file, and `skills --check`
  refuses a path nobody wrote, a file nobody named, and a reference that is not markdown.
  `ui-craft` is the first router, with four references, and `design` declares its output (0205).

## v0.16.2 - 2026-11-05

- **The release publishes the installer it documents.** The release assets now carry `install.ps1`
  and `install.sh`, so the README's `releases/latest/download/install.ps1` answers; v0.16.1 and
  v0.15.1 never published them and the documented command returned 404 (0204).

## v0.16.1 - 2026-11-05

- **A new project reaches its first proof, and the installed copies say what they are.** The
  `AGENTS.md` block names Proof-Driven Development and the first proof; the `bootstrap` skill is the
  first-cycle checklist: `odachi init` creates the repository when the folder has none, `odachi
  start` asks the human and passes their words, and `drive` closes one step with `odachi check
  --run`. The doctor names that cycle in an empty folder, compares the installed skills with the
  binary, and audits the design plugin beside the door. `check --run` warns when a proof has no
  output instead of calling it evidence; a `.ps1` command is refused with the PowerShell invocation
  that runs it; and `review` reads a repository with no commits as an empty history instead of a
  fatal git error. The bench has a journey from an empty folder to the first proof and uses `echo
  ok`, which both shells run, instead of `true` (0203).
- **v0.16.0 is yanked.** Its tag points at the feature merge, before the release PR merged; it has no
  release and no signed assets. Install v0.16.1 or later.

## v0.15.1 - 2026-11-05

- **The design system says when it went stale, and a design writes back to its feature.** `odachi
  design list` reports the project's design sources and, when `design/system.md` exists, the sources
  newer than it; the skill syncs before the first design and names a stale source before following
  tokens the project no longer has. `odachi design handoff` on a design linked to
  `odd/tasks/<feature>.md` appends one log entry to that document naming the chosen artboard and its
  alternatives, refuses a link outside `odd/tasks/`, and writes no document for untracked work. The
  skill loads `ui-craft` for the hierarchy and the numbers while it owns the delivery (0201).

## v0.15.0 - 2026-11-04

- **The design canvas works like a product surface.** `odachi design open` now leaves a detached
  canvas running and returns, and a second `open` moves the page already open instead of opening a
  tab. The canvas has navigation and focus: the wheel zooms, the drag pans, and a click focuses an
  artboard at exactly 100% with its own scroll and clicks; `Shift+Esc` or Back to canvas returns.
  Comments pin their element, can be sent now or collected and sent in a batch, and reach the
  OpenCode session through a plugin `odachi init --global` installs (Claude Code reads the files).
  A design records its chosen artboard and its handoff state (`sent`, `in_progress`, `implemented`,
  a loopback preview) in `design/<slug>/handoffs.json`, and Compare shows the artboard beside the
  running app. The canvas renders artboards in sandboxed iframes, serves only manifest-listed files
  and requires a per-server token; the system board comes first when it exists (0200).

## v0.14.9 - 2026-11-04

- **The agent teaches while it works.** `odachi init` writes a voice that explains every decision
  in a short paragraph -what it chose, what it costs and what the other way would have cost-, names
  and defines a concept the first time it appears, and says what an answer changes before asking a
  question. Both profiles explain now: `novato` teaches step by step with the craft ceremony, and
  `senior` leads with the decision and explains it in fewer lines instead of going to the point. The
  anti-sermon guard stays in both: no repetition, no glossary, depth is not length. `odachi start`
  prints a why of two lines per interview question, the second indented under the first. A project
  receives the new voice when it runs `odachi init` or `odachi level --to`, never silently (0197).

## v0.14.8 - 2026-11-04

- **The documented PowerShell install prints no error.** `scripts/install.ps1` opened with a UTF-8
  BOM, and PowerShell 5.1 reads that character as part of the first token when the file arrives
  through `irm | iex`: `#Requires` became `?#Requires`, so the README's own command printed an error
  and then ran anyway. The file now opens with the directive, and a test pins the bytes so an editor
  cannot put the BOM back (0189).

- **The review refuses a model the host no longer serves.** A project's lens table could name a
  model the provider had retired, and the run waited for the whole timeout before saying only that
  the command failed. `odachi review` now asks the host for its catalog (`opencode models`) before
  starting a reviewer: a provider the host serves but a name it no longer lists is refused, and the
  refusal names the lens, the model, the line that chose it, the session's model, the free names the
  host uses and the three ways to switch. `--model` and the session's model are not refused; a
  provider the host does not know is left to the command; a host that cannot answer refuses nothing.
  A reviewer that fails after the preflight names the same options, and the tool never switches by
  itself. This project's four lenses now name DeepSeek V4.1 Flash, which answered when the
  configured model stalled twice (0194).

## v0.14.7 - 2026-11-03

- **The cleanup switch is not wedged by the shape of a folder.** `odachi clean --apply` refused to
  remove anything while a class it maintains held a path it could not read, and the walk's depth
  bound counted as a read failure: a nested repository in the scratch folder made the switch
  unusable while `clean` still said it would free 34 MB. A folder past the walk is now the class's
  shape - counted and named, not a failure -, a class `--apply` empties is removed as a folder so
  the subtree the walk never reached goes with it, and the report says the switch is blocked
  instead of promising bytes. A real read failure still blocks and names the path (0188).

## v0.14.6 - 2026-11-03

- **The store disposes of the question it no longer keeps.** A review filed its prompt beside the
  verdict and left it there: 60 to 120 KB each, growing with every review, while `keep_prompts = 0`
  already said the question was disposable. A stored verdict now removes the prompts its
  `keep_prompts` stopped keeping, by the same rule `odachi clean` reports and counts in trees. A
  review that gets no verdict keeps its prompt, and a store that cannot be read prunes nothing
  (0187).
- **A close names what the store grew.** A clone's store grows without anybody noticing: prompts,
  records and scratch. The tree that closes a feature now prints one line with the store's size and
  what `odachi clean --apply` would free - or that the switch is blocked by the paths it could not
  read. The line removes nothing and changes no answer: the switch stays the human's (0187).

## v0.14.5 - 2026-11-03

- **The run records the brief it read.** `odachi start --brief <draft>` read a draft and wrote no
  `brief.md`, so a feature shaped by the `shape` skill landed without the vision the run had just
  read and `check` asked for it. The run now records the draft as `brief.md` in the canonical shape
  (the source is untouched), a draft that is already `brief.md` is never rewritten, and an existing
  `brief.md` beside a draft elsewhere needs `--force` (0186).

## v0.14.4 - 2026-11-03

- **A close owes its verdict whatever the commit order.** The review mode read a document that was
  already closed at `HEAD` as "nothing to ask", so a close committed before its verdict was answered by
  silence; only `--require-review` caught it. The question is the feature's diff: a range that carries
  code owes the verdict, a range of documents owes none, and a tree after a verdict that only moved
  documents stays paperwork (0185).

## v0.14.3 - 2026-11-03

- **A model is an identifier, and attach says so.** `odachi review --attach` stored a verdict whose
  `model` was a display name (`DeepSeek V4.1 Flash`); the next `--since` correction then refused it with
  "a model is an identifier", when nothing could be done about the stored file. Attach now refuses the
  model before anything is written, naming the shape (`provider/name#variant`); a verdict already
  stored with a display name still reads (0184).

## v0.14.2 - 2026-11-03

- **The default branch is a fact of the commits, not of a name.** `odachi check` refused an
  `Odachi-Change:` trailer on the default branch by comparing the checked-out branch's name with the
  base's; on a detached HEAD it had no name, the CLI dropped the error and the trailer was judged.
  The gate now asks the repository for the local branch the base names -`origin/main` names `main`- and
  whether HEAD is one of its commits: a detached checkout is refused for the same reason a named one is,
  a feature branch that inherited `origin/main` as its upstream is still a feature branch, CI's detached
  pull request keeps the lane working, and a repository git cannot read blocks with the reason (0183).

## v0.14.1 - 2026-11-03

- **The review asks the document the gate asks.** `odachi review` asked whether a document's base
  equaled the ref it judged: a document records a commit and the default base is `HEAD`, so the answer
  was no and an `Odachi-Change:` trailer could stand in for an open feature's vision. It now asks the
  newest document, the one `check` takes, so a light feature's line is its vision on the default base
  instead of "there is no brief.md yet". A correction keeps asking the document its predecessor judged
  against, because the newest one may belong to a later feature (0182).

## v0.14.0 - 2026-11-03

- **A release note is the changelog, printed.** The workflow takes the body of a release from the
  changelog section for the tag and refuses a tag with no section, so the notes are the same text a
  person reads and nobody writes them twice by hand (0180).

- **A change can name itself in its commit.** `Odachi-Change: <line>` in a commit of the branch is the
  vision when no document is open: `check` measures the change from the merge-base, `review` judges
  it against that line and says the diff is the record, and the proof is named after the line. A
  trailer on the default branch, an empty one or two different ones are refused by name (0181).

## v0.13.1 - 2026-11-02

- **The design skill says again what it knows.** The screenshot rule of 0176 fit by compressing two
  bullets nobody asked to change, and the review named the loss: "or reorder", "saying why",
  "component folders" and "propose a starter system" are back inside sixty lines (0177).

## v0.13.0 - 2026-11-02

- **An artboard names the screenshot it was read against.** `design.json` carries a `screenshot`
  beside each artboard, the design skill makes it part of done, and `odachi design list` and the
  canvas name the artboards still without one (0176).

## v0.12.1 - 2026-11-02

The installer verifies, and then installs.

- **`v0.12.0`'s PowerShell installer aborted when the signature verified.** cosign prints
  `Verified OK` on stderr, and PowerShell 5.1 with `$ErrorActionPreference 'Stop'` turns a native
  program's stderr into a terminating error: the check failed exactly when it succeeded, so nobody
  with cosign could install the release. The preference is relaxed for the call and restored, and the
  test's shim is a native program that writes to stderr, which is what the fixture was missing (0175).

## v0.12.0 - 2026-11-02

The installer verifies the signature before it installs.

- **Both installers require cosign.** They download `checksums.txt.sigstore.json` and verify the
  checksum file against this repository's release workflow and the tag before they read a hash, and
  they refuse when cosign is missing instead of degrading: a check that runs only sometimes is no
  check (0174).
- **Why it matters:** the checksum and the archive came from the same host, so a compromised release
  passed its own hash. The chain of trust now ends in the signature, not in GitHub.
- **Breaking for installers:** cosign must be on the PATH. That is a documented workflow step, so the
  next release is a `minor` and not a patch.

## v0.11.102 - 2026-11-02

A craft skill for the screen an agent builds.

- **`ui-craft` ships with the curriculum.** One skill holds the decision order a designer
  follows (content, hierarchy, layout, type, colour, detail, states, remove), nine tells of a
  generic interface with the move that replaces each, the numbers that make a screen look
  deliberate (scale, measure, contrast, targets, motion durations), and the read-back before
  delivering. It stands alone: it depends on no skill this project does not ship (0173).

## v0.11.101 - 2026-11-02

A one-line feature is reviewed against its own line, and a proof speaks while it runs.

- **The review of a light feature carries that feature's requirement as the vision.** The prompt
  used to read `brief.md`, which a one-line change never writes, so the reviewer judged the
  candidate against the previous feature's done-when. It now reads the task document whose base is
  the candidate's, quotes the requirement it carries, and names the file it came from (0168).
- **The lenses, the verdict and the store are unchanged.** Only the context the reviewer is handed
  moves: the domain never learns about lanes, and a feature with a brief reads it exactly as
  before. A task document that cannot be read is refused by name instead of being taken for
  "there is no light feature here".
- **`odachi check --run "<command>"` shows the command's output while it runs.** The proof is
  written when the process ends, so a run blocked on a lock printed nothing for seventeen minutes
  and read exactly like a slow suite. The bytes now reach the terminal as they arrive, and the
  proof keeps the same copy it always carried (0169).
- **Where it goes is the rule:** the stream is stderr, so `--json` keeps stdout as one document a
  caller can parse. The reviewer's own run stays quiet: the stream is a window, not a second
  answer.

## v0.11.99 - 2026-11-01

A line is enough for a change that fits one.

- **`odachi start --light "<requirement>"` opens a one-line feature.** The task doc carries the
  requirement quoted, one step with the same words and `Light: yes`, and nothing else is written:
  no brief, and the project's fence and voice stay exactly as they were. `check` no longer asks a
  light feature for a brief -the one line is the vision- and asks it for the fence, the open
  questions and the proof exactly as before.
- **The lane closes like any feature.** The one step is what the gate counts, so the closing tree
  owes its proof and its review, and `plan` refuses a second step: a change that needs one is a
  change that needs a brief.
- Why: the rule "without a written vision nothing is built" is the product, and the lane keeps it
  by making the written vision one line. `Light: yes` lives in the preamble, where the base and
  the waiver already live, so the artifact carries the shape and no command line has to remember
  it (0166).

## v0.11.98 - 2026-11-01

A name says what it is.

- **A step tail with two commit parts stays text.** A hand-written `R1 · tree:abc · note` used to
  keep the last part as the commit and lose the first on the next canonical write. Two parts in
  the same slot are ambiguous, so the line is kept as text, unchanged, instead of being rewritten
  behind its reader.
- **`--artifact` reads a bare name.** The name `--list` prints now answers which kind it is, so
  `--file brief --artifact 0002-x.md` is refused as the contradiction it is, instead of saying
  there is no backup of a document the run itself hid. A name outside the four is refused with
  the four paths it does keep.
- Why: the two rows 0163 and 0164 left were the same defect twice -a name read as something it is
  not- and the lesson is the same both ways: a parser that tolerates must also refuse what it
  cannot read (0165).

## v0.11.97 - 2026-11-01

A record can be repaired by naming it.

- **A step that closed with no tree takes the mark afterwards.** `odachi plan --done T2 --commit
  <sha>` on a step that is already done and carries no mark records it and says so; a step that
  already carries a mark is still a no-op, and with no `--commit` the message names the flag.
  Before, every already-done step was refused, so a mark lost to a dirty worktree could not be
  written by any command.
- **`undo` names the document it restores.** `--artifact odd/tasks/0163-....md`, or the bare file
  name `--list` prints, restores that one and backs up what it replaces. `--file task` in a
  project with several task documents used to demand `--file`, the flag the run already had.
- Why: a close can leave a mark nobody can write and a backup nobody can name, and it is the same
  defect twice -the record exists, and the verb cannot be pointed at it (0164).

## v0.11.96 - 2026-11-01

A step declares who runs it.

- **The step line carries its route.** `odachi plan "<text>" --route inline` says the parent runs
  the step; `--route delegated --agent odachi-worker` names the agent that does. The route lives
  in the task doc beside the requirement and the tree, so the document says who did the work
  instead of a chat nobody keeps.
- **A contradiction is refused when the step is written.** A delegated step with no agent, an
  inline step with one and an agent with no route are refused before the line exists, and the
  message names the flag to change. A step written before this change reads as not declared,
  never as inline.
- Why: delegation used to be a craft note in the history -a sentence about the work, not the
  step's own line- so the record of who ran a step lived where nothing reads it. Identity is
  declared, never inferred (ADR 0038), and the tool records the declaration without checking it:
  the gate and the review do not learn to read a route.

## v0.11.95 - 2026-11-01

The lenses name four defects they were missing.

- **reliability** names a focused or skipped test -`.only`, a disabled sibling- that makes CI
  pass by shrinking the suite, and a test that proves one function through a whole flow, which
  costs more than the behaviour it proves.
- **resilience** asks for a threshold the project chose -without a number, "it is fine" is a
  guess- and for something in production that would show an expected failure.
- **readability** names behaviour the diff changes and the brief never asked for.
- Why: the comparison with gentle-ai's lenses found five defects in four prompts; each landed
  with a test observed red first, no number was borrowed from another project, and the review's
  own engine stayed fenced out of the change. The fence that used to protect the four prompts
  now protects the engine, and it landed alone before the feature (policy PR #248) because CI
  reads the fence from the base ref.

## v0.11.94 - 2026-11-01

A second reading is not a second opinion.

- **The identity is the model, not its setting.** The gate compared the whole model string, so
  `provider/name#high` and `provider/name` read as two judges and a review by the same model at
  another effort passed as a second opinion it never was. Case and `#variant` are now ignored,
  and the sentence a human reads names the effort each side ran with.
- **A self-review is named by default.** It already was -every review refusal but a causal
  finding is softened (ADR 0043)- and nothing pinned it; there is a test now, and the sentence
  says what it is. `--require-review` still refuses, so a project that needs the strong
  guarantee asks for it.
- **A waiver is a decision on the document.** `odachi review waive --because "<why>"` writes
  `Review: waived` into the current feature's task doc and the reason into its history. No
  reason, no waiver; the proof and the fence stay, and a finding the review already made still
  stops the ship. It is the lane for a project that owns one model (ADR 0048).
- Why: a gate that only works for people who can afford two models is a gate that teaches
  people to turn it off.

## v0.11.93 - 2026-11-01

A block falls by evidence, and the quorum counts accounts.

- **A finding declares its evidence.** `cited` when the reviewer quoted the line of the diff
  that shows the defect, `argued` when it reasoned without one. The field is optional and
  absence reads as argued, so every verdict written before this release keeps the strict
  meaning, and an invented class is refused by name.
- **The class decides the quorum.** A cited claim falls to one cited refutation; an argued one
  needs two, and a `stands` or a `refuted` without a citation is never a vote to remove. Three
  witnesses is the ceiling, because a block decided by a crowd is a block nobody can read.
  **The quorum counts accounts, not files**: the store refuses a second answer from a refuter
  that already answered this tree.
- **Each answer keeps its own file.** The first is still `<tree>.refuted.json` and the next two
  are numbered, so no stored answer is rewritten and a tree refuted before this release is read
  as round 1. The gate names the votes a surviving block still needs, and the verdict schema
  moved to 4.
- Why: one refuter's citation lifted any block, and a claim nobody can point at was removed by
  being confident. The review of this feature itself found a CRITICAL -the quorum counted files
  instead of accounts- and three warnings, all paid before the close (ADR 0047).

## v0.11.92 - 2026-11-01

A failed review says which model it asked and where that model came from.

- **The failure names the source.** The three failure paths of `review --exec` -could not run,
  timed out, non-zero exit- name the lens, the model and the line that chose it: `--model`,
  `odd/review.toml` or the session. An offline tool cannot know what a provider still serves; it
  can make a retirement one edit instead of an investigation.
- **An archtest says this repository names a judge for every lens**, and the comment
  `odd/review.toml` carries tells the next human what to check when a model goes.
- Why: the run that cannot start is the one a person debugs, and a message that names only the
  command sends them to read three files. The measurement behind this feature was wrong once
  -a filtered list hid the active, cheapest model- and the review of the branch caught what the
  error left behind, which is why the row is in `docs/review-debt.md` and not gone.

## v0.11.91 - 2026-11-01

The gate runs before the push, and the craft has a trust boundary.

- **`check --ci` judges the tree the way CI does.** It resolves the base from the clone's
  remote -`origin/HEAD`, then `origin/main`, then `origin/master`- and reads both the fence
  and the compared paths from that ref, so a branch that loosened its own copy of the policy
  no longer passes locally and fails on the pull request. It takes no `--base` or
  `--fence-from` beside it, and a clone with no origin blocks naming the fetch that creates
  the ref. It never fetches: the run and the proof name the ref they read -`fence: <ref>` and
  a `# fence <ref>` header- so a stale base is visible instead of guessed around (ADR 0045).
- **The curriculum gains `security`.** One skill, loaded by situation -input from outside, a
  permission, a secret, a dependency, a log, a failure- with thirteen rules and four
  prohibitions in 39 lines, and every claim about an outside system carries its source and
  version: OWASP Top 10:2025, ASVS v5.0.0 (2025-05-30), NIST SSDF 1.1 and `govulncheck`.
- **ADR 0046 draws the line where enforcement stops.** The fence over paths, the proof over a
  command the project chooses and the risk lens over the diff are the enforcement; the trust
  boundary, the owner of an object and the life of a secret are judgement, and the skill says
  them once. No scanner: a pattern list measures the pattern, not the decision.
- **The impact map stopped counting skills by hand.** Its skills row had already lost `design`
  and spelled the payload as a word; `internal/archtest/impactmap_test.go` now pins the row
  against `skills.Names()`, so a new skill costs one line of the map.
- Why: a local green that CI contradicts is the worst kind of green, and a checklist nobody
  loads protects nobody.

## v0.11.90 - 2026-11-01

The design surface: any agent can design, and the person clicks to comment.

- **A `design` skill ships in the binary and travels to every client.** It carries the folder
  layout (`design/<slug>/`), the manifest fields, the comments protocol as a versioned file,
  the read-only design-system capture and the handoff, so a client with no tool of its own
  follows the same protocol as one with a canvas open.
- **Six commands, and no tool a single client owns.** `odachi design list`, `comments`, `open`,
  `stop`, `handoff` and `sync` drive the folder from a shell: list the directions with their
  tradeoff, read and resolve what a person asked, serve the artboards on loopback, print the
  chosen one as an implementation instruction, and write `design/system.md` from the project's
  own stylesheets without editing one.
- **The canvas is a command, not an extension.** `odachi design open` serves the artboards at
  their size, with a comment mode that computes a selector, a panel that resolves, and live
  reload by polling the folder. It writes `comments.json` atomically and refuses a file it
  cannot parse. It compares the request origin with the address it listens on -never the Host
  header- and refuses when it does not know its own address.
- **The fence learned to let a skill in.** It now names the sixteen skills that existed, so a
  new one is a feature and editing a shipped one is still a correction. The policy landed
  alone, before the feature: CI reads the fence from the base ref, and a candidate cannot
  loosen the policy it must obey.
- Why: gentle-design showed that a design direction is worth keeping as files a reviewer can
  read, and that the capability belongs in a protocol every client can follow, not in one
  editor's extension.

## v0.11.89 - 2026-10-31

The README stops naming a version that ages.

- **No version in prose.** The release paragraph, the footer and the cosign example carried `v0.11.40`, forty-eight releases behind: they point at the releases page and at the tag you downloaded, and say why a number here is a claim nobody measures.
- **The README declares the proof lifetime.** One file per feature, rewritten by the next run of that feature, never removed by `odachi clean`: the store keeps the claim and the folder keeps what the tree said.

## v0.11.88 - 2026-10-31

A proof says what it claims to cover, and the close reads the claim as a claim.

- **The proof header may carry `# covers R1 R2`.** `check --run` writes the requirements
  the closing feature declares, and a proof written before the field existed keeps proving
  its tree: `Parse` consumes the line only when the terminator follows, so an old proof's
  quotation survives.
- **The close names a requirement the proof does not declare**, and stays ready. The line is
  written by the agent that ran the command -the suite does not know which requirement a test
  belongs to- so blocking on it would give a claim the weight of a test. The README says so
  in those words: a promise a reviewer can challenge, never a certificate.
- Why: the gate knows a requirement has a step (0148); this is the half that says what the
  evidence claims to prove. Making it real evidence needs a command per requirement, and
  that is its own feature.

The fence guards the evidence, not the writer.

- **The rule named two promises and now names one.** *"No change to the proof format, and the
  stored proof is never rewritten"* protected `internal/domain/proof/**` and `proof/**`: the
  bytes a tree was certified with, and the code that writes them. It now protects `proof/**`
  and one thing -a stored proof is never rewritten- so the format may gain a versioned line
  while the evidence stays untouchable (ADR 0044).
- Why: 0148 made the gate read the requirements, and recording in the proof which requirement
  it covers needed a line the fence forbade. The correction is not a patch on the rule; it is
  the decision the rule was standing in for.

## v0.11.87 - 2026-10-31

A requirement nobody planned stops the close.

- **The gate reads the requirements.** `check` never looked at them: a feature could declare `R2`,
  plan a step for `R1` only, finish every step it had, and close green. At the close the
  requirements no step names are now a `NeedsWork` that names them and prints
  `odachi plan "<text>" --req R2`. Mid-feature the list is empty on purpose -a requirement a
  change added waits for its step, and the steps are planned before the work.
- Why: the gate joined "what you asked" (brief) to "what was proved" (proof) and skipped the middle
  artifact, so the plan could promise less than the vision and nobody heard it.

## v0.11.86 - 2026-10-31

The block closes the session, and `doctor` says when the note went stale.

- **The block carries the close.** Its last paragraph names the trigger a human uses, the file a project keeps for
  the next session, and what goes in it: the newest feature, what was measured and the decision that waits. The block
  is the only text that arrives before the agent acts, so it is where a habit that must not be forgotten lives.
- **`doctor` has a fourteenth check.** A `docs/next-session.md` that does not name the newest feature is a warning
  that names the number missing, and it never refuses: a stale note is a note to rewrite. A project that keeps no
  such file stays clean, because the convention belongs to the project.
- **The check asks one question**: does the note name the newest feature? It does not read the highest number the
  note mentions, because a bare year is four digits too and would warn forever on a note that is right.
- Why: `docs/next-session.md` is not one of the four artifacts, and nothing wrote it or read it, so it went stale
  silently for two releases. A rule nobody looks at ages by itself.

The CI gate stops asking to be excused from the review.

- **`gate.yml` runs `check` without `--no-review`.** The flag existed because the gate is built from the base ref and
  the base still refused a close without a verdict; since v0.11.85 that base names the review and stays ready, so a
  pull request that closes a feature carries the sentence instead of a flag that hid it.

## v0.11.85 - 2026-10-31

The review never authorises delivery, and what it found still stops a ship.

- **A close without a verdict is ready and names it.** With `mode = on`, `odachi check` reports the review the feature
  did not get and lets the commit through; the hook prints that reminder on a quiet run. Only `--require-review`
  refuses, and it refuses for that run.
- **What the review found still blocks.** A causal `BLOCKER` in the tree's verdict stops the ship until it is
  refuted, with or without the flag: the review saying no is not the review being absent.
- **CI stops passing `--no-review`**, so a pull request that closes a feature carries the sentence that the review was not verified
  there -the verdict lives in the clone- instead of a flag that hid it.
- Why: ADR 0042 proved the review cannot be verified from a clean checkout, and `mode = on` went on refusing commits
  locally, which was a promise larger than its mechanism. ADR 0043 decides the honest shape: the review is a
  discipline, and delivery is governed by the fence and the proof.

## v0.11.84 - 2026-10-31

The voice styles the chat and never the artifacts, and senior says what it does.

- **Both voices name what an artifact is** -code, comments, documents, commit messages and pull requests are
  English, with no regional slang and none of the voice's emphasis- and say that **only the human's message changes
  the conversation language**: a tool result, a subagent's report or a pasted fragment is input, never a reason to
  switch.
- **`senior` is a voice of its own**: lead with the decision, its cost in one line when there is a fork worth naming,
  and no recap of what the code already says -instead of being `novato` with the teaching taken out.
- **This repository's block carries the new voice**, regenerated by `init`, and its pinned test moved with it.
- Why: the two shared rules are the best of gentle-ai's persona, which separates how the agent talks from what it
  builds; the rest of that corpus is a character -years, titles, CAPS- that costs tokens every session and decides
  nothing.

## v0.11.83 - 2026-10-31

The clone's store has a door, and the bench measures the paths it opened.

- **`odachi review --all` lists every verdict the clone holds**, newest first, with the tree, the verdict, the
  lenses, the findings, the date, the base and the reviewer, and marks the tree that stands. The store is keyed by
  tree and nothing else listed it: a person who needs a hash -to correct a verdict, or to read what one judged- was
  reading `.git/odachi/reviews/` by hand.
- **`review --show` names the door** when this tree has no verdict and the store holds others, and **`check` names
  `odachi review --since <tree>`** when the feature has a predecessor for the same base, so the correction is one line
  away instead of a hash to find.
- **A `--since` naming a tree the clone does not hold names `odachi review --all`** as its way out, which the bench
  measures like every other block.
- **The bench grows three journeys**: the door on an empty store, a correction naming a missing verdict, and a close
  that names its debt; the corpus is eleven journeys and 29 steps, with no dead ends.

## v0.11.82 - 2026-10-31

A correction cannot drop a blocker in silence, and the base it completes is the base it keeps.

- **A blocker about bytes the fix moved becomes a follow-up**: it stays in the verdict, where a reader finds it, and it
  stops blocking. A blocker about bytes nobody can read any more cannot keep blocking -the correction exists to clear
  it- and dropping it would let a fix launder a reason not to ship.
- **The correction keeps the predecessor's base**, and the test that says so now fails when it does not: its fixture
  used the default base, so the assertion could not tell where the value came from.
- Why: the review of 0140 left two rows about what its tests did not prove, and the first one hid a rule rather than a
  test -keeping the blocker blocking would make a correction unable to clear the case it exists for.

## v0.11.81 - 2026-10-31

A fix after a verdict is reviewed as the delta, and a close names the debt it leaves.

- **`odachi review --since <tree>` reviews only what moved** since the tree that verdict judged: the prompt is the
  delta, and the stored verdict carries the predecessor's rows and base, so every path stays answered while the run
  pays for the fix alone. A correction with nothing to judge is refused, naming the read that already holds the answer.
- **The prompt says it is a correction**, so a reviewer handed a delta does not report the rest of the candidate as
  missing.
- **`odachi check` names the open rows of `docs/review-debt.md`** when a feature closes: the list a person reads before
  choosing the next feature is read at the moment it matters, and a row still never blocks.
- Why: 0139 needed five review rounds and each one re-sent a 64 KB candidate for a handful of lines, while the review
  gate is local by construction -`proof/` is not versioned, and evidence about a tree cannot live inside it-. ADR 0042
  records the cost and the boundary.

## v0.11.80 - 2026-10-31

The block names the copy of the craft the project actually keeps.

- **`odachi doctor` refuses a block that names the other folder**: it asks the same scope `init` asks -the store's
  choice, then the adoption of a repository that already lives- and the warning names the folder the block names, the
  one the project keeps, and `odachi init`. A question that could not be read keeps both variants current, because a
  failed read is not an absence.
- **This repository stops carrying its own copy**: `.agents/skills` (15 tracked files) and `.opencode/opencode.json`
  leave the tree, and `AGENTS.md` names `~/.agents/skills/`, the copy `setup` refreshes once per machine.
- Why: the block is what tells every agent where the craft lives, and one naming a copy nothing refreshes sends the
  agent to a folder that goes stale in silence. Both variants used to be current because the doctor could not know
  which one `init` had chosen; the catalogue adoption already answers that, so it asks it.

## v0.11.79 - 2026-10-31

The guard that decides whether an answer may merge into a stored verdict is one predicate, and a test pins it.

- **`collision` is the one place that decides**: it returns the first lens an answer shares with the stored verdict that
  this run did not land itself, and `storeIn` refuses exactly that case. The refusal names the lens already paid for.
- **The wrong question now fails a test.** A table pins the rule, and the split test plus the table both go red when the
  guard ignores what the run landed - the gap the review of 0137 found, where narrowing the guard left every test green.
- Why: a rule no test can contradict is a comment, and the guard this feature pins is the one that stops a second run
  from paying twice for the same lens.

## v0.11.78 - 2026-10-31

The review's repeat bypass asks what the run has landed, not what shape the candidate has.

- **`storeIn` takes the lenses this run already stored and judges the collision per lens**: a part of a
  split candidate may add the rows of the paths it carries to a lens an earlier part landed, and nothing
  else may touch an existing verdict.
- **A second run over the same oversized candidate is refused and starts no reviewer**, pinned by
  `TestASecondRunOverASplitCandidateStartsNoReviewer`. It was already refused one layer up, where `exec`
  counts the lenses still owed, so the guard now asks the question the run asks.
- Why: the bypass was keyed on `unit.of > 1`, a property of the candidate, and the shape of a candidate
  says nothing about what this run has stored. Found by the review of 0136 as a warning; 0137 pays it.

## v0.11.77 - 2026-10-31

A review whose prompt does not fit the project's limit is split, not stalled.

- **`odachi review --exec` sends a candidate over `max_prompt_kb` in parts**: each part carries its own
  changed paths and their hunks, the group's lenses and model, and its number in the header, and the store
  merges the answers into the tree's one verdict. A project that sets no limit is never split.
- **Every changed path is answered by exactly one part**, and each part's inspection declaration is judged
  against the paths it read -a part cannot vouch for bytes another part judged-.
- **The one refusal left is a file that cannot fit by itself**: it names the file, the size it costs and the
  limit. A reviewer handed half a file is not reviewing it.
- Why: 0132's prompt measured 114.4 KB and the attempt before it stalled `mimo-v2.6-flash` for 600s, so the
  biggest features were the ones the project's own reviewer could not take.

## v0.11.76 - 2026-10-31

The two rows the review of 0134 left are paid, and this repository declares its retention.

- **`odachi clean` says "one tree"**: the summary line uses `plural2` like the class rows above it, so
  `keep_verdicts = 1` no longer reads "the newest 1 trees" and `keep_prompts = 1` no longer drops the noun.
- **The doc comment in `cmd/odachi/review_test.go` names its own test again**, after 0134 inserted a test
  between the comment and the function it described.
- **This repository's `odd/review.toml` declares `schema = 3`, `keep_verdicts = 10` and
  `keep_prompts = 0`**, the numbers it already obeyed, now that the release that reads them is installed.

## v0.11.75 - 2026-10-31

The project decides how much review evidence survives.

- **`odd/review.toml` gains `keep_verdicts` and `keep_prompts`**, the lifetimes `odachi clean --apply`
  obeys. They are counted in trees -a prompt belongs to the review it served- and zero means the default
  odachi shipped: ten records, no prompt. More than a thousand, or a negative number, is refused by name.
- **Prompts are their own class in the report.** A prompt is the question, not the evidence: 597 KB in
  eleven files in this repository, the largest 114 KB, and nothing the gate reads touches it. A project
  that wants the question on file writes how many trees keep theirs.
- **The report says where the numbers came from**, `odd/review.toml` or odachi's defaults, so a lifetime
  nobody remembers choosing is never a mystery. A file the tool cannot parse is refused by name instead of
  falling back to a number nobody chose.
- **The schema moves to 3**, so a binary older than this one says upgrade instead of naming `keep_verdicts`
  as an unknown key. This repository's own `odd/review.toml` carries the keys once the release that reads
  them is on the machine: CI builds the gate from the base ref, and a base binary that reads schema 2
  refuses a candidate that says 3 before the suite runs.

## v0.11.74 - 2026-10-31

A verdict carries the rows of the paths it judged, so it stops growing with the store.

- **The store stopped copying itself forward.** A verdict inherited every potent row in the store: 126 KB
  and 769 clean rows, where 15 named a file the candidate changed. A review now inherits the rows of the
  paths its diff touches -before hashing them, so it no longer pays a git call per untouched row- and the
  prompt's own cut goes away: one rule, one place. The same verdict, bounded, is about 4.8 KB.
- **A standing blocker still travels.** A causal `BLOCKER` or `CRITICAL` about a file the candidate never
  touches is not a row about bytes somebody read; it is the reason not to ship, and it keeps blocking until
  it is fixed or refuted. `Finding.Blocks` is now the one definition of that rule.
- **Nothing old is rewritten.** A store from an earlier release keeps its copies; the bound applies to what
  is written from now on, and `odachi clean`'s retention retires the old ones.

## v0.11.73 - 2026-10-31

`odachi clean` names what a clone accumulates, and removes only what odachi wrote.

- **A clone grows four kinds of file and nothing said whose they were.** `odachi clean` prints a map:
  the class -review records, review scaffolding, undo backups, the agent's scratch- with its owner,
  its files, its bytes and what the run would do to it. It also names what it will never touch:
  `proof/`, `odd/tasks`, the client choice, and the rest of `.git/`, which belongs to git and to
  other tools.
- **`--apply` gives each class odachi wrote a lifetime**: the newest ten review trees (a verdict, its
  prompt and the records beside them are one record) and the newest five versions of each file an
  undo backup saves. The scratch folder is emptied. Nothing else enters the plan.
- **A class it cannot read stops the run before the first file**, because deleting on a partial view
  is how a tool loses evidence; a name in the backups folder that is not a version odachi wrote never
  joins a group, so somebody else's file is never counted as old.
- **`--apply --legacy` removes the notes an older agent left loose in `.git/`** -`PR*.md`,
  `PLAN*.txt`, `CLOSE*.txt`, `REL*.txt`, `*MSG.txt`, `T<n>_*.txt`, `ROWBASE.txt`, `HOLE.txt` and
  `fence.generated`- after naming them. Without `--legacy` they are reported and left.
- **The block says where an ad-hoc note goes**: `.git/odachi/scratch/`, never the tree and never
  loose in the git directory, and `odachi clean --apply` empties that folder. The rule is in the
  block because the block is the door every agent reads in every project.

The doctor tells a note from a plan.

- **A document in `odd/tasks` that is not a plan no longer makes the run unclean.** It carried no plan
  shape -no base, no requirement, no step, no history- so nothing was lost by not reading it, and the
  line now names it as what it is. Before, any document with a name odachi cannot read was a warning,
  which is how this repository's own bootstrap history screamed on every run.
- **A document that carries a plan's shape and a name odachi cannot read is still a warning**: that one
  is a feature that would be silently ignored, and it is the only case the line has to shout about.
- This repository's bootstrap history moved to `docs/odachi-bootstrap.md`, where its own first lines say
  it belongs: it is read to understand why odachi is the way it is, and no task is added to it.

`odachi start` stops writing a fence rule for a phrase.

- **A rule needs paths, and a phrase has none.** `start` turned every non-goal into a
  rule, so a non-goal in free text became a glob of its own words
  (`paths = ["datos-de-pacientes*"]`): the one thing the fence's own header forbids,
  because it matches files named after the phrase and protects nothing. The rule is
  written only for a non-goal that carries paths -the presets- and a phrase is
  carried as a commented template in `odd/fence.toml` with the paths left to fill.
- **The run says what it did**, naming the template and the file, instead of
  claiming the fence wrote a glob.
- The rule is ADR 0037, and it is 0120's rule with an enforcer: the verb that writes
  fences can no longer produce a rule without paths.

## v0.11.72 - 2026-10-30

The adoption is one inference, and `odachi doctor` stops calling it unfinished.

- **A project that adopted odachi is reported as what it is**, not as a gap. Check 11 saw no agent file
  and no `project_agents` key -the shape an adoption leaves, by design, because the inference is never
  stored- and advised `odachi init --global`, which records a choice nobody made. The note now says the
  rules and the hook live in the project and the craft and the agent live in the machine, and names
  `odachi setup` when the door is missing.
- **`init`, `role` and `doctor` ask one package** (`internal/app/adoption`) instead of keeping three
  copies of the same rule, so they cannot answer differently about the same project. Each verb keeps its
  own ports and its own refusal words; the answer is shared.
- **A note that cannot tell does not claim**: a repository git cannot read is check 2's business, and the
  adoption note says only what it sees.
- The rule is ADR 0036.

A failure at the destination the project lives in fails the run.

- **`init --skills home`, an adopted project and `odachi setup` now exit 1 when the
  home refuses the write**, instead of printing `pending:` and exiting 0. The block
  naming `~/.agents/skills` is written before the skills half runs, so the old
  behaviour left a green run that named a folder nothing had created.
- **Fatality follows the scope the run asked for, not the folder that failed**: a
  project-scope run that wrote its own copy keeps exit 0 when the home half fails,
  and still says what is pending.
- **The refusal names the verb that ran.** A failed `setup` used to say "run odachi
  init again", and its closing line claimed the bootstrap skill was installed while
  the write had failed.
- The rule is ADR 0035.

`odachi role` writes and reads where the project keeps its craft.

- **A role's skill lands where the craft lives.** In a project whose store says `skills = home`,
  or one that adopted odachi and keeps the crafts in the machine, `odachi role <slug>` writes
  `~/.agents/skills/<slug>/SKILL.md` and nothing into the project; a project that carries its own
  copy keeps getting `./.agents/skills/<slug>/SKILL.md`. Before, the skill went into the project
  whatever the scope, and that one folder made the next plain `init` install the whole curriculum
  in a project that keeps it in the machine.
- **The listing reads the same folder**, so `odachi role` says `up to date` about the copy that
  actually exists, and the scope is decided by one function that `init` and `role` share.

A check odachi could not make no longer counts as a "no".

- **`odachi init` stops instead of guessing the footprint.** The adoption inference reads two facts
  off the disk -whether the project carries the craft, whether it carries the native agent- and a
  failed read used to answer both of them "no": a project that keeps the curriculum in the machine
  could get all fifteen skills written into it, with a green exit. It now names the path it could
  not read and writes nothing.
- **A repository git cannot read is not a repository with no commit.** `.git` that git refuses used
  to choose the full install, as if the project were starting now. It is refused by name; a folder
  with no history, or a repository with nothing committed yet, keeps the full install.
- **`odachi role` refuses an `AGENTS.md` it cannot read** instead of writing the default block over
  a file nobody saw.
- The rule is written once in ADR 0034: an error is never an absence.

## v0.11.71 - 2026-10-30

Adopting odachi in a project that already lives stops scattering folders in it.

- **A repository with a commit and no copy of the craft at `.agents/skills` gets the minimal
  footprint.** The project keeps the rules, the `.gitignore` block and the hook; the craft is
  refreshed in the machine (`~/.agents/skills`, a no-op when `odachi setup` already put it there)
  and no project agent is written. Before, the same run wrote `.agents/skills`,
  `.opencode/agents` and `.claude/` into a project that only wanted the rules.
- **The inference is never stored.** Only `--skills project` and `--agents` write a choice, and
  the two are separate: naming a client does not decide where the craft goes. A project that
  already carries that copy keeps it (a bare `.agents/` is not the craft: that folder belongs to
  other tools too), and a folder with no history still gets the full install: there the project is
  starting now.

## v0.11.70 - 2026-10-30

The block names where the skills actually are.

- **`AGENTS.md` pointed every project at `.agents/skills`, even one that keeps the craft in the
  home.** With `skills = home` that folder is never written, so the block told every agent to load
  a file that does not exist -and kept telling it after every `init`. The text now names both
  folders and points at the clone's store as the authority, so it is true whatever the project
  chose and a project that never chose reads exactly what it read before.

## v0.11.69 - 2026-10-30

The rehearsal stops guessing about what it could not read.

- **A folder odachi could not read is not a client that is absent.** The run printed `could not
  check` for that client and then `no agent client found in the home: install OpenCode` -two
  lines that contradict each other, the second telling the human to install a client that may be
  right there. A failed read is now its own line with the error, and the advice only appears when
  nothing failed.
- **The three branches the review of 0121 left unpinned now have tests**, each checked backwards:
  the error arm of detection, a settled machine whose only drift is one changed skill (it must say
  `would replace`, never `nothing to write`), and a real `setup --json` that names
  `~/.config/opencode/opencode.json` with `dry_run` false. `docs/review-debt.md` loses the rows
  those tests paid.

## v0.11.68 - 2026-10-30

`odachi setup` can rehearse, and it names what it leaves alone.

- **`odachi setup --dry-run` writes no byte.** The machine verb only knew how to act, so after
  an upgrade the human ran it and hoped. The rehearsal reads every destination, compares the
  bytes it would write, and says `would write` - the same code that does the work builds the
  report, because the disk it writes through answers every question and swallows every write.
- **A settled machine says so.** Run it twice and the second run reports nothing to write; the
  real run twice changes nothing either.
- **A skill whose bytes differ is counted.** The join between `init` and `pack` dropped the
  count of replaced skills, so after an upgrade -the case this exists for- the rehearsal said
  "nothing to write" while every changed skill awaited a rewrite. The list of files a rehearsal
  would write now includes the ones it would replace, and the default agent in OpenCode's config
  is named as a file this run changes instead of a step that leans on the disk.
- **The report names the clients detection did not find**, with the folder it looked for, so an
  installer that covers two of three tools cannot look like it forgot one.

A fence rule protects a file, not a reason.

- **Two rules written from a rationale left main's policy.** 0119 closed with "no new verb", and
  `start` turned that reason into `internal/app/pack/**` and the persona writer. The feature this
  correction unblocks (F3, the machine refresh) needs both for its report, and CI refused it with
  a message about a verb it never added - the second time this session that a rule written for one
  feature blocked the next. The pack rule is gone, because it had no file behind it; the client
  rule is narrowed to `internal/domain/persona/selection.go`, the file that decides the choice,
  instead of deleted.
- **The lesson lands in `drive` and in ADR 0027:** a fence rule protects a file, never a reason.
  It is advice and not enforcement - `start` still asks for paths on every non-goal - so the
  changelog claims the lesson, not a rule that prevents the next one.

## v0.11.67 - 2026-10-30

The store remembers where the skills went, so an upgrade stops scattering them.

- **The skills scope is a stored choice.** `agents.toml` already kept which clients receive the
  agent (0081) and whether the project carries it (0106); the one choice it forgot was where the
  craft lives. `odachi init --skills home` installed the skills in the home, and the next plain
  `init` -the one every upgrade suggests- rebuilt `.agents/skills` in the project. The key is
  now `skills = project|home`, written when a run names it and honoured by every later run, and
  the output says where the answer came from: the flag, the store or the default.
- **A project that never chose keeps exactly what it had.** No key, no store, the project copy:
  the same bytes as before the key existed.
- **The store's schema moved to 2, so an older binary says upgrade instead of unknown key.**
  One counter serves the three strict files, and that cost is written where the counter lives: a
  verdict whose shape did not change also declares 2.

The fence of a feature names what it protects.

- **A rule that guarded the verdict listed the clone's store.** `internal/adapters/agentstoml`
  sat under "no change to the verdict shape", so the gate refused the store's next key with
  `out of scope` - and the only way out was to correct the policy in a change that does not
  touch the store. The rule now names every path it lists - the verdict's shape and its schema,
  and the review's settings - and no longer the store.
- **The deadlock is written down where the next reader looks:** a feature that needs a path the
  previous fence protected lands the correction first, in its own change (ADR 0027).

## v0.11.66 - 2026-10-30

The human is in the loop, and the block says what that means.

- **A question is answered, not implemented.** The block's rules gated decisions -the brief, the
  fence- and said nothing about acting: an agent could take a question as a request and start a
  feature, which is what happened. The block now says that an answer is not a proposal and that
  work starts when the human asks for it or approves one.
- **An irreversible action names its target and waits.** Push, pull request, merge, tag or
  release, an install, another folder: each one is proposed with the exact target and waits for
  a yes. It is the governance gentle-ai already had in its action gates, and this block had only
  for folders.
- **Ask at the gate, once**: a question that would not change what the agent does next is noise.
  The `drive` skill carries the same rule one line before the cycle.

## v0.11.65 - 2026-10-30

The suite checks that every Go file is what `gofmt` writes.

- **Formatting was the one hygiene nobody read.** `vet` was clean, `mod tidy` changed nothing,
  and `gofmt -l .` was red on a file this machine wrote without its final newline -the suite
  pins behaviour, and formatting is not behaviour-. `internal/archtest` now walks every `.go`
  file and compares it with `format.Source`, in process and in one pass, naming the path, the
  line and the two spellings when it fails.
- **The plan of this feature said the blob carried CRLF and it did not:** `.gitattributes`
  normalizes `*.go` to LF, so the blob was canonical and the CRLF lived in the worktree. The
  task doc records the correction.

## v0.11.64 - 2026-10-30

The bench stops lying about a check without a proof.

- **`odachi bench` reported one surprise and it was the journey.** `check-sin-prueba` expected
  `check` without a proof to exit 1; since 0093 the proof belongs to the tree that closes the
  feature, so an open feature is ready for a step and exits 0. The tool was right and the
  journey was stale.
- **The journey now walks the four moments** -start, a check ready for a step, a step, the
  close and the check that refuses until the proof runs- and a test pins its exit codes, which
  nothing did before: that is why the suite stayed green while the bench told the truth on its
  own.

## v0.11.63 - 2026-10-30

A checkout is not an edit.

- **A CRLF block reads like an LF block.** `check 11` called a block with CRLF line endings
  hand-edited - the parser looked for the markers by byte - so a Windows clone with
  `core.autocrlf=true` was sent to `odachi level --to` for a file that was fine. `BodyIn`,
  `VoiceIn` and `HasBlock` fold the endings once, where the block is parsed; the writer never
  folds, and a block edited by hand is still an edit.
- **`doctor` drops the fold it kept of its own**, which was the symptom that said the parser
  should have done it.

## v0.11.62 - 2026-10-30

The block comparison runs without a profile file too.

- **0111's check skipped the case it was most needed in.** A project that ran `init` and never
  ran `start` has no `odd/profile.toml`; the comparison treated an inferred profile as "no
  voice to compare" and stayed silent, so an older block passed -exactly where a project has
  just adopted odachi. The inferred name is the voice `init` writes the block with, so the
  comparison now runs there as well.

## v0.11.61 - 2026-10-30

Doctor says when the rules block is older than the binary.

- **Presence is not currency.** `doctor` looked for the block's markers and never at its
  text, so a project that upgraded odachi kept the rules it adopted with: a rule that
  arrived later -the one about other folders, for instance- never reached its agent, and
  nothing said so. Check 4 now compares the block with the one this binary would write for
  the project's voice and warns, naming `odachi init`, when they differ.
- **A warning, never a blocker**, and a checkout is not an older rule: CRLF and a trailing
  space fold away before the comparison, the same rule the brief has had since 0031.

## v0.11.60 - 2026-10-30

Another folder is another project.

- **The block says it.** After the four hard rules: this project is this repository, another
  folder is another project with its own brief, fence, store and memory, and odachi is never
  read or run there unless the human asks for it. The `drive` skill's *Never* list refuses the
  same thing, so an agent that reads either learns it before acting.
- **It is a rule of conduct, not a gate**, and the changelog says so on purpose: nothing can
  tell that an agent browsed a sibling folder, and a check that fired on a sibling directory
  would be noise. What the tool can do is put the rule where the agent reads it.

## v0.11.59 - 2026-10-29

The prompt says what it carries.

- **The carried section names its own contents.** After 0107 cut the clean rows of paths
  the candidate never touched, the heading still said "What the store already carries for
  these exact bytes" - the store carries more than the prompt shows. It now says what it
  is: the findings of earlier reviews and the clean rows of the paths this candidate
  changed, with one line admitting that the store keeps more.
- **Nothing else moves:** the rows, their order and the instruction that the verdict
  inherits them are the same.

## v0.11.58 - 2026-10-29

The review prompt carries what the review can use.

- **The memory stops being the whole repository.** The prompt carried 23,498 rows, and
  99.7 % of the clean rows named a path the candidate never touched. It now carries every
  finding -a known defect must not be re-raised as new- and the clean rows of the paths
  that changed, which are the ones a reviewer can act on.
- **The store is untouched.** Inheritance and coverage still read every row; the cut is
  the prompt's alone. On tree 56641e43 it takes the prompt from 94,678 B to 56,915 B
  (40 % less), and the same three factual probes answered correctly before and after.

## v0.11.57 - 2026-10-29

A global install leaves the project alone.

- **`odachi init --global` no longer writes the project's agent files.** The door in the
  home is the one that runs in every folder, and the project keeps only the rules and the
  plan; `doctor` calls that the clean mode instead of reporting missing files. A plain
  `odachi init` is the run that writes a project's own copies, and the choice stays in
  `agents.toml`, so a plain init after a global one keeps the agent in the home.
- **`--no-project-agents` is gone.** Its only legal use was with `--global`, where it is
  now the default; a flag that cannot change the outcome is a lie about what ran. The
  refusal for the old spelling names the run that does the same thing, and `--agents`
  brings the project's copies back.

## v0.11.56 - 2026-10-29

A local check no longer lies about committed code.

- **The gate anchors on the verdict.** After a close, "did anything non-document arrive"
  was answered from the candidate - the worktree against HEAD - which is empty once a
  change is committed. CI caught such a tree; a local `check` said ready. The gate now
  reads the newest verdict of the feature from the store and asks what changed between
  that tree and the one that stands: a committed code change owes the verdict, a release
  note stays paperwork.
- **A close with no verdict in the store keeps the answer it had**, and
  `--require-review` keeps its meaning.

## v0.11.55 - 2026-10-29

A documents-only close owes no verdict.

- **The review gets the proof's twin.** 0096 stopped asking a suite for a paragraph
  written after a close; the review kept asking for a verdict on a feature whose whole
  diff is documents, and `review` answers "none is due", so the only way out was
  `--lens all` - a four-lens review of a paragraph. `check` now classifies the
  feature's own diff, base to the closing tree, with the same function the review uses:
  a diff that asks for no lens has nothing to judge and the close is ready.
- **What still owes:** code after the close, an explicit `--require-review`, and a
  feature that carries code at its close. The closing tree is never measured alone,
  because it only marks the last step.

## v0.11.54 - 2026-10-29

A widened scope needs its words.

- **`--lens` that adds a lens now needs `--because`.** The scope record exists to say
  who asked, for what and in whose words; a run that would write it without words is
  refused before any reviewer starts, and the refusal names the flag, the lenses it
  would add and the words it needs. The comparison with gentle-ai agrees: its
  escalation takes the item and the reason together.
- **The words already on file count.** A tree whose record exists keeps working with
  `--lens` and no new words, and a run that adds no lens is never asked - the record
  is the memory, not the ceremony.
- **The help text is unchanged:** the fence protects `cmd/odachi/main.go`, so the rule
  arrives with the refusal rather than in the flag list.

## v0.11.53 - 2026-10-29

A refused answer writes nothing.

- **The scope record follows the verdict.** v0.11.52 wrote `reviews/<tree>.scope.json`
  before the declaration and coverage checks, so an `--attach` answer the tool refused
  still left a decision beside a verdict that does not exist. The record is now
  written after every refusal and immediately before the verdict it explains.
- **Why that order:** gentle-ai validates before it reads the repository and a refusal
  creates no authority and writes nothing; its escalation is frozen with the authority,
  never beside nothing. In odachi the verdict is the authority, so the record has no
  life of its own.

## v0.11.52 - 2026-10-29

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
  commit that renames `## v0.11.77 - 2026-10-31` to `## v0.11.50` is a new tree, and `check`
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



