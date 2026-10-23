# Impact map

Who depends on each package, and which tests prove it. Update this table in the
same work unit as the code change.

| Package | Who depends on it | Tests that cover it |
| --- | --- | --- |
| `cmd/odachi` | Nobody imports it; it is the binary | `cmd/odachi/main_test.go`, `cmd/odachi/ask_test.go`, `cmd/odachi/bench_test.go`, `cmd/odachi/check_test.go`, `cmd/odachi/doctor_test.go`, `cmd/odachi/doctor_debts_test.go`, `cmd/odachi/evidence_test.go`, `cmd/odachi/init_test.go`, `cmd/odachi/init_personas_test.go`, `cmd/odachi/init_skills_test.go`, `cmd/odachi/install_test.go`, `cmd/odachi/level_test.go`, `cmd/odachi/lock_test.go`, `cmd/odachi/pack_test.go`, `cmd/odachi/plan_test.go`, `cmd/odachi/reuse_test.go`, `cmd/odachi/review_test.go`, `cmd/odachi/reviewexec_unix_test.go`, `cmd/odachi/reviewexec_windows_test.go`, `cmd/odachi/reviewlens_test.go`, `cmd/odachi/skills_test.go`, `cmd/odachi/undo_test.go`, `cmd/odachi/undo_ambiguous_test.go`, `cmd/odachi/version_test.go` |
| `internal/app/start` | `cmd/odachi`: `start` and `odachi` with no arguments; after deciding the profile it refreshes the block's voice and the native agent when they exist | `internal/app/start/start_test.go` |
| `internal/app/tasks` | `internal/app/{ask,plan,level,summary}`: find the current or every feature document, and the single write guard every task-document rewrite goes through | `internal/app/tasks/tasks_test.go`, `internal/app/tasks/save_test.go` |
| `internal/app/ask` | `cmd/odachi`: `ask` in its three forms | `internal/app/ask/ask_test.go` |
| `internal/app/plan` | `cmd/odachi`: `plan` (add a step, close a step) | `internal/app/plan/plan_test.go` |
| `internal/app/level` | `cmd/odachi`: `level` and `level --to`; the accepted profile rewrites the profile, the block's voice and the native agent that already exists | `internal/app/level/level_test.go` |
| `internal/app/summary` | `cmd/odachi`: `summary` | `internal/app/summary/summary_test.go` |
| `internal/app/check` | `cmd/odachi`: `check` in all its forms, plus the optional `--require-review` gate: verdict per tree, verdict per required lens, and no unrefuted blocker | `internal/app/check/check_test.go`, `internal/app/check/review_test.go`, `internal/app/check/lens_test.go` |
| `internal/app/review` | `cmd/odachi`: `review` in its five forms. Reads the brief, the changed paths and the diff through ports, loads the selected lens prompts through the `Lenses` port, and hands the reviewer the same prompt the verb prints (one builder, two verbs); takes the lock on nothing: it writes one verdict or one refutation, atomically, inside `.odachi/reviews/` | `internal/app/review/review_test.go`, `internal/app/review/lenses_test.go` |
| `internal/app/init` | `cmd/odachi`: `init` — the rules block in `AGENTS.md`, the native agent that mirrors the block in the project, the worker and verifier subagents, the door in the home that makes odachi available in any folder, the marked block in `.gitignore`, the pre-commit hook, and the `Skills` port the CLI joins to `pack`: the project's curriculum and the home's bootstrap skill | `internal/app/init/init_test.go`, `internal/app/init/personas_test.go`, `internal/app/init/agents_block_internal_test.go`, `internal/app/init/gitignore_internal_test.go`, `internal/app/init/path_warning_test.go`, `cmd/odachi/init_test.go`, `cmd/odachi/init_skills_test.go`, `cmd/odachi/init_personas_test.go`, `cmd/odachi/evidence_test.go` |
| `internal/app/skills` | `cmd/odachi`: `skills` (list, `--check`, `--write`) and `skill new` | `internal/app/skills/skills_test.go` |
| `internal/app/pack` | `cmd/odachi`: `pack` (the detected agents, `--agents`, `--scope`, `--dest`, `--dry-run`), and `init` through the CLI's `initSkills` join — two destinations, one report. Resolves every name to a target directory, groups agents that share one, and refuses `--dest` with `--scope` before reading anything | `internal/app/pack/pack_test.go`, `cmd/odachi/pack_test.go`, `cmd/odachi/init_skills_test.go` |
| `internal/domain/doctor` | `internal/app/doctor` (the eleven checks' numbering, titles, hook and AGENTS.md reading, the installed skills and the generated personas, the PATH fix and the one-line rendering, plus the two-line pre-flight for a folder that never was a project), `cmd/odachi` (`--json` prints the same lines) | `internal/domain/doctor/doctor_test.go` |
| `internal/app/doctor` | `cmd/odachi`: `doctor` — the pre-flight for a folder that is not a project yet, and otherwise the eleven read-only checks over three ports; check 11 also compares the native agent and its subagents with what the binary generates, and the global door in the home | `internal/app/doctor/doctor_test.go`, `internal/app/doctor/preflight_test.go`, `internal/app/doctor/ignore_test.go`, `cmd/odachi/doctor_test.go` |
| `internal/bench` | `internal/app/bench` (the corpus, the command a blocked message names, the four-way classification, the sandbox environment) | `internal/bench/bench_test.go` |
| `internal/app/bench` | `cmd/odachi`: `bench` (runs each journey in its own sandbox and removes it) | `internal/app/bench/bench_test.go`, `cmd/odachi/bench_test.go` |
| `internal/app/lock` | `cmd/odachi` (the `Lock` adapter, wired into `start` and `undo`), `internal/app/start` and `internal/app/undo` (they take the lock before their first deciding read and release it with `defer`) | `internal/app/lock/lock_test.go`, `internal/app/start/lock_test.go`, `internal/app/undo/lock_test.go`, `cmd/odachi/lock_test.go` |
| `internal/domain/brief` | `internal/app/start` (render and read `brief.md`), `internal/app/check` (is it complete?) | `internal/domain/brief/brief_test.go`, `internal/domain/brief/render_test.go` |
| `internal/domain/taskdoc` | `internal/app/start` (write the numbered task doc), `internal/app/{ask,plan,level,summary}` (history, questions, steps, counts), `internal/app/check` (open questions), `internal/app/tasks` (which lines a canonical rewrite would drop) | `internal/domain/taskdoc/taskdoc_test.go`, `internal/domain/taskdoc/history_test.go`, `internal/domain/taskdoc/step_test.go`, `internal/domain/taskdoc/unaccounted_test.go` |
| `internal/domain/fence` | `internal/app/check` (the fence gate); `start` when it writes `fence.toml` (T5) | `internal/domain/fence/fence_test.go` |
| `internal/domain/gate` | `internal/app/check` (the four gates and the exit code) | `internal/domain/gate/gate_test.go` |
| `internal/domain/proof` | `internal/app/check` (write and read `proof/NNNN-<slug>.txt`) | `internal/domain/proof/proof_test.go` |
| `internal/domain/review` | `internal/app/review` (parse, seal and render a verdict or a refutation; the tier and the lens selection; build both prompts), `internal/app/check` (does this tree have a verdict, does it cover its lenses, and does anything still block?), `cmd/odachi` (the JSON payload it prints back) | `internal/domain/review/review_test.go`, `internal/domain/review/prompt_test.go`, `internal/domain/review/tier_test.go`, `internal/domain/review/refute_test.go`, `internal/domain/review/lens_test.go` |
| `internal/assets/review` | `cmd/odachi` (the `Lenses` adapter): the four lens prompts, embedded with `//go:embed` | `internal/assets/review/lens_test.go` |
| `internal/domain/profile` | `internal/app/start`, `internal/app/level`, the profile decoder adapter | `internal/domain/profile/profile_test.go`, `internal/domain/profile/next_test.go` |
| `internal/domain/skill` | `internal/app/skills` (parse each `SKILL.md`, validate the set, generate `SKILLS.md`) | `internal/domain/skill/skill_test.go` |
| `internal/domain/agent` | `internal/app/pack`: the verified agent → skills directory table, its two scopes, the `standard` pseudo-agent (`.agents/skills`, the cross-client location, never detected), the one scope override the table cannot predict (`opencode` under home is `~/.config/opencode/skills`), `--dest` labels that skip the table, and the refused names | `internal/domain/agent/agent_test.go` |
| `internal/domain/rules` | `internal/app/init` (escribe el bloque de `AGENTS.md` con la voz del perfil), `internal/app/level` y `internal/app/start` (reescriben **solo** la sección `## Cómo hablás`), `internal/domain/doctor` (una sola definición de «hay bloque») | `internal/domain/rules/rules_test.go` |
| `internal/domain/persona` | `internal/domain/rules` (el texto de cada voz), `internal/app/init` (el agente nativo, la puerta del home y los subagentes), `internal/app/level` y `internal/app/start` (refrescan el espejo cuando existe), `internal/app/doctor` (compara el espejo, la puerta y los subagentes con lo generado) | `internal/domain/persona/persona_test.go` |
| `skills` (repository root) | `cmd/odachi` (`pack` reads the embedded payload; `init` reads `Curriculum()` and `Bootstrap()` from it); the Go package lives in the same directory as the `SKILL.md` files it embeds, so there is no second copy | `skills/embed_test.go` |
| `internal/adapters/{fsx,gitx,profiletoml}` | `cmd/odachi` only; the domain never imports them | the matching `_test.go` in each adapter |
| `internal/adapters/shellcmd` | `cmd/odachi` only (the `check --run` and `review --exec` shell); one file per platform. `RunInput` delivers a prompt as the command's stdin from a file, and `{prompt}` as that file's quoted path; Windows builds the `cmd` command line with `/s /c "…"` because `cmd.exe` does not understand the backslash escaping Go applies to a quoted argument | `shellcmd_test.go` plus the per-platform `_test.go` |
| `internal/archtest` | Nobody imports it; it guards every domain package | `internal/archtest/boundary_test.go` |

## Distribution (T9)

| File | What it is | How it is verified |
| --- | --- | --- |
| `cmd/odachi` (`var version`, `--version`, `version`) | The two surfaces that report the version; the variable is the ldflags seam `-X main.version=<tag>` | `cmd/odachi/version_test.go`: `--version` and `version` print `odachi dev` and exit `0`; the injected value is what gets printed; an unknown command still exits `1` |
| `.goreleaser.yaml` | v2 config: `CGO_ENABLED=0`, `-trimpath`, `-s -w -X main.version={{ .Version }}`, linux/darwin/windows × amd64/arm64, `tar.gz` (zip on windows), `checksums.txt` (sha256), snapshot named `dev`. No `brews`, no signing, no provenance — the comment at the top says why | `goreleaser check` (locally or in the `release-check` CI job). It cannot complete in this repository yet: there is no git remote, and the scm step needs one |
| `.github/workflows/ci.yml` | `test` matrix on ubuntu/macos/windows (+ `go vet` and `gofmt -l .` on ubuntu), `race` on `./internal/...`, `build`, `release-check`; `permissions: contents: read`, no secrets | The workflow itself is not executed by any Go test; the commands inside it are exactly the ones run in the verification section of this work unit |
| `scripts/install.sh` | POSIX shell installer: platform and arch detection, version from `ODACHI_VERSION` or the latest release, download base from `ODACHI_BASE_URL`, sha256 verification before touching the destination, install to `${ODACHI_BIN_DIR:-$HOME/.local/bin}` (falling back to `./odachi`), version line and `PATH` hint | `cmd/odachi/install_test.go`: builds a real binary with `-ldflags -X main.version=1.2.3`, packs a real tarball and a real `checksums.txt`, runs the script offline over `file://`, then runs the installed binary and reads `odachi 1.2.3`. Second case: a tampered `checksums.txt` exits non-zero and leaves nothing installed |
| `scripts/install.ps1` | Windows equivalent: `AMD64`/`ARM64`, zip download, `Get-FileHash -Algorithm SHA256`, install to `%LOCALAPPDATA%\odachi\bin`, user `PATH` entry when missing | Not verified: it needs a Windows PowerShell run that this work unit did not perform. See the T9 report |

## Skills (T12)

| What | Where | How it is verified |
| --- | --- | --- |
| Format | `skills/<name>/SKILL.md`: frontmatter with `name` and a `description` that starts with `Trigger:`, then `## Cuándo`, `## Reglas`, `## Nunca` | `internal/domain/skill/skill_test.go` (valid file, CRLF, no frontmatter) |
| Rules | `name` equals the directory and matches `^[a-z][a-z0-9-]{2,40}$`; description starts with `Trigger:` and is at most 200 characters; body at most 60 lines; no empty section; no `TODO`; unique names; the registry is exactly what `--write` generates | `skill_test.go` plus `internal/app/skills/skills_test.go` and the `cmd/odachi/skills_test.go` table, one violation per case |
| Registry | `SKILLS.md`, first line `<!-- generado por odachi skills --write; no lo edites a mano -->`, then the `Skill \| Trigger \| Path` table sorted by name | `skill_test.go` pins the exact bytes; `cmd/odachi/skills_test.go` proves two `--write` runs are identical and a hand edit makes `--check` fail naming `--write` |
| `skill new` | validates the slug before any write, refuses an existing one, writes the template, then regenerates the registry | `internal/app/skills/skills_test.go` and `cmd/odachi/skills_test.go` |
| Bundled ten | `skills/{acceptance,bootstrap,commits,data,drive,ship,skill-craft,slices,tests,writing}/SKILL.md`, each with the three sections and at most 60 lines | `cmd/odachi/skills_test.go` runs `skills --check` over this repository |
| Delegation policy | `slices` owns it: precedence over global skills, writer only by parallelism or the context backstop, handoff by task-doc path, bounded return, verify against the brief and the diff. `drive` forces loading it when the request does not fit. The decision and what was not copied from gentle-ai live in `docs/decisions/0007-delegation-policy.md` | `skills/embed_test.go` `TestDelegationPolicyLivesInSlicesAndDrive`, `skills --check` |
| Native subagents | `init` writes `odachi-worker` (bounded writer) and `odachi-verify` (read-only) per client. `slices` passes the task doc, step, allowed surfaces, verification commands and exact `SKILL.md` paths; the worker returns evidence and ≤2k tokens. OpenCode uses `mode: subagent` and a permission block; Claude Code uses `.claude/agents/` and, for the verifier, a tools allowlist without Edit/Write. The parent keeps git, the gate and vision | `internal/domain/persona/persona_test.go` `TestSubagentFilesAreSubagentsForEveryClient`, `TestTheVerifierIsReadOnly`; `internal/app/init/personas_test.go`; `internal/app/doctor/doctor_test.go` `TestAMissingSubagentIsSaid`, `TestAStaleSubagentIsSaid`; `opencode debug agents` live check |
| ODD-lite | `drive` owns the workflow discipline (requirements verbatim, observable `Listo cuando`, RED before GREEN, delegated step recorded with `ask --craft`, risk line); `commits` owns the risk vocabulary (`none` or the six items). The task-doc grammar is untouched; the decision is `docs/decisions/0008-odd-lite.md` | `skills/embed_test.go` `TestDriveTeachesTheWorkflowDiscipline`, `skills --check` |

`odachi init` copies the skills now: the eight that are craft go to the project's
`.agents/skills` and `bootstrap` goes to `~/.agents/skills`. Its block still only
gains the line that points at `SKILLS.md`, and that registry is the project's
own — not the installed set.

## Pack (T13)

`odachi pack` is what makes the skills reachable by the agent: they live in this
repository, and an agent only loads what is in its own skills directory.

| What | Where | How it is verified |
| --- | --- | --- |
| Source | The nine `skills/<name>/SKILL.md` files compiled into the binary with `//go:embed */SKILL.md`, in the `skills` package. No network, no repository at run time, no duplicated copy. `All()` is the payload `pack` installs; `Curriculum()` is the eight that are not the entry point; `Bootstrap()` is that one | `skills/embed_test.go` compares the embedded bytes with the files on disk |
| Destination table | `internal/domain/agent`: `claude .claude/skills`, `codex .codex/skills`, `cursor .cursor/skills`, `gemini .gemini/skills`, `copilot .copilot/skills`, `qwen .qwen/skills`, `kiro .kiro/skills`, all home-scoped, plus `opencode .opencode/skills`, the only project-scoped row. It is data, not branches of code, because every path was verified against the agent's documentation or source | `internal/domain/agent/agent_test.go` pins every row, its scope, `Display()` and `Marker()` |
| Scopes | A `home` agent resolves its destination against the human's home; a `project` agent (`opencode`) against the project directory the verb runs in. OpenCode reads `{skill,skills}/**/SKILL.md` relative to the project, so this path has no platform-dependent shape; the global `~/.config/opencode/skills` exists but is not the default — `--scope home` is what installs there | `TestOpenCodeIsProjectScoped`, `TestOnlyOpenCodeIsProjectScoped`, `TestOpenCodeInstallsIntoTheProject`, `TestOnlyTheProjectAgentWritesIntoTheProject`, `TestOpenCodeHomeLandsInItsGlobalConfigDir` |
| Refusals | `pi` is refused with `Pi no usa un directorio de skills` (its adapter reports none); any other name is refused naming the supported ones, opencode included | `agent_test.go`, `internal/app/pack/pack_test.go`, `cmd/odachi/pack_test.go` |
| Write rule | `<root>/<dest>/<name>/SKILL.md` through `fsx.WriteAtomic`; an existing file that differs is backed up first; an identical one is reported as `sin cambios` and not rewritten | `pack_test.go` (second run: no write, no backup; edited file: backup before write; opencode: same rule in the project), `TestPackOpenCodeSecondRunIsIdentical` |
| Never deletes | Nothing is deleted or moved; a `SKILL.md` odachi did not write is reported with its path and left alone | `TestForeignFileIsReportedAndNotTouched` |
| Failure isolation | One agent's failure does not abort the others, and the message names the agent and the path | `TestOneFailureDoesNotAbortTheOthers` |
| `--dry-run` | Zero writes and zero backups, asserted on the double, never on the output; the report still lists the plan and says `por instalar` | `TestDryRunTouchesNothing` |
| Detection | With no `--agents`: a home agent is here when its config directory exists in the home, a project agent when `.opencode/` exists in the project. Nothing detected exits `0` naming `--agents` | `TestDetectsInstalledAgents`, `TestDetectsOpenCodeInTheProject`, `TestPackDetectsOpenCodeInTheProject` |
| Display | `~/.claude/skills` for a home agent, `./.opencode/skills` for a project one; an absolute path is never printed, and the JSON carries `dest` and `scope` separately | `TestOpenCodeInstallsIntoTheProject`, `TestPackOpenCodeWritesIntoTheProject` |
| Project untouched | `pack` never reads or writes the project's `skills/` or `SKILLS.md`: those belong to `odachi skills --write`. A project-scoped agent writes inside `.opencode/skills` | `TestPackDoesNotTouchTheProject`, `TestOnlyTheProjectAgentWritesIntoTheProject` |
| Home | The `Home` port resolves the home, and it is asked for only when a home-scoped agent is being served (or when detecting). `internal/app/pack` imports no `os`, so every write goes through the `Files` double in tests and through `osFiles` in the binary | `internal/app/pack/pack_test.go` (in-memory `Files` + `Home` doubles), `TestProjectAgentDoesNotNeedAHome`, `osHome` in `cmd/odachi/files.go` |

## Plan (T15)

`odachi plan` is how the agent adds and closes steps without editing Markdown
by hand. The step is appended to the parsed document and the file is written
back through the canonical renderer, so the document cannot drift out of shape.

| What | Where | How it is verified |
| --- | --- | --- |
| Add a step | `odachi plan "<texto>" [--req R1]`; the id is `max+1` (never a reused hole); `--req` must be a requirement the document declares | `internal/domain/taskdoc/step_test.go` (numbering, link, bad text), `internal/app/plan/plan_test.go`, `cmd/odachi/plan_test.go` |
| Text rule | one line, no leading dash, at most `taskdoc.MaxStepText` (120) characters; runs of spaces collapse to one so the rendered step parses back to the same text | `step_test.go`, `plan_test.go` |
| Close a step | `odachi plan --done T2 [--commit abc1234]`; unknown id lists the open steps; an already-done step is a no-op with exit `0` and `ya estaba hecho`, writing nothing | `internal/app/plan/plan_test.go`, `cmd/odachi/plan_test.go` |
| Flag contradictions | `--commit` without `--done`, both forms at once, `--req` without a new step: blocked before the document is read | `internal/app/plan/plan_test.go` |
| The one write guard | `internal/app/tasks.Save` / `SaveWithBackup`: canonical render, then `taskdoc.Unaccounted` decides whether any human line would be lost; `SaveWithBackup` backs up only after the guard allows the write | `internal/app/tasks/save_test.go` (`ask` and `plan` share it), `internal/domain/taskdoc/unaccounted_test.go` |
| Refusal is real | a document with an extra section is refused naming `## Historial`, and the filesystem bytes are unchanged with no backup left behind | `internal/app/plan/plan_test.go`, `cmd/odachi/plan_test.go` |
| Never touches other artifacts | `brief.md`, `.odachi/profile.toml` and `.odachi/fence.toml` are byte-identical after a plan | `internal/app/plan/plan_test.go`, `cmd/odachi/plan_test.go` |

The guard distinguishes canonical lines from unaccounted ones: a step marked
done legitimately rewrites its own line, while a `## Notas` section cannot be
represented by `Render` at all. `taskdoc.Unaccounted` is the grammar's answer to
"what would be lost", and it is the only place that answer is computed.

## `cmd/odachi` behaviour

- Commands: `start` (default), `ask`, `plan`, `level`, `summary`, `undo`,
  `check`, `review`, `doctor`, `bench`, `init`, `skills`, `skill new`, `pack`,
  `version`. `odachi`
  with no arguments is `odachi start`; a first argument that starts with `-` is
  treated as `start` flags, except `--version`, which is answered before the
  dispatch.
- `odachi --version` and `odachi version` print exactly `odachi <version>` and
  exit `0`. `version` is a package variable, `"dev"` unless goreleaser injects
  the tag with `-X main.version=...`; both surfaces print the same variable so
  there is one source of truth. `-v` and `-version` are deliberately not
  accepted: two spellings are enough, and a third would be a silent alias nobody
  documents.
- Unknown command: exit `1`, message lists every available command.
- Exit codes: `0` ok · `1` blocked (invalid flag, unreadable or nameless
  profile, existing `brief.md` without `--force`, no feature document, a
  contradiction between flags, unknown profile) · `2` a human answer is missing
  and this run cannot ask for it (non-interactive without flags, bare `odachi ask`).
- Every blocked message is one line of Spanish and names the command or flag
  that unblocks it.
- Two adapters are wired here and nowhere else: `osFiles` (through
  `internal/adapters/fsx`, so an existing file is backed up before the atomic
  rewrite) and `terminalInterview` (stdin/stdout). The `internal/app/*` packages
  declare the ports and never import `os`.
- `.odachi/profile.toml` is read when it exists and wins over `profile.Infer`:
  the human owns that file. A field left out keeps the built-in default of the
  same profile name. With no file, `Infer` starts on `novato` unless a brief
  already exists: `.git` is not a signal because `init` creates it.
- `--json` is machine output and implies non-interactive. `start`'s JSON shape
  is unchanged; other verbs add `lines` whenever they print a human block —
  `doctor` included, in both of its shapes.
- `check` is the only verb whose exit code is the gate's, not a fixed one: `0`
  ready · `1` the agent can act · `2` only the human can decide. `--run` returns
  the gate's code, never the command's. Its third adapter is `shellcmd`; the
  other two are `osFiles` and `gitRepo` (through `gitx`).
- `check --run` keeps `proof/` out of the git tree, and which mechanism it uses
  is decided by the repository: it asks the Git port (`gitx.Ignored`) whether the
  proof it is about to write is already ignored, and only when nothing ignores it
  does it write `proof/.gitignore` —before hashing— as a fallback. A probe that
  fails writes the fallback too: evidence must never move the tree because git
  could not answer. `--quiet` prints nothing only when ready (the pre-commit hook
  depends on that) and still prints the block when not.
- The diff a review reads is `gitx.Diff`, and `cmd/odachi` has ONE Git adapter
  (`gitRepo`) for both `check` and `review`. gitx decides that "no base" means
  the empty tree, so the app never spells that hash. The diff builds a throwaway
  `GIT_INDEX_FILE` exactly like `TreeHash`: untracked files enter the candidate,
  git-ignored files do not, and the human's real index is never staged.
- `init` leaves the project ready in one command. It runs `git init` when there
  is no `.git` (a proof is only a proof against a tree, and a tree needs
  history), then owns five artifacts: the rules block in `AGENTS.md` between
  `<!-- odachi:begin -->` and `<!-- odachi:end -->`, the marked block in the
  project's `.gitignore` between `# odachi:begin` and `# odachi:end` (`/proof/`
  and `/.odachi/reviews/`, refreshed in place, every other byte preserved, never
  backed up because the edit is additive), the `.git/hooks/pre-commit` hook, the
  native agent per client (the block without markers, see the T29–T32 section
  below), and the skills: the curriculum to `./.agents/skills` and `bootstrap` to
  `~/.agents/skills`, through the `Skills` port that the CLI wires to `pack`. It
  never replaces a hook it did not write: it prints the three lines to paste and
  exits `1`. Order matters: AGENTS.md, `.gitignore`, the hook, the native agent,
  then the skills, so a later refusal still reports what was already written. A
  failure inside the project is exit `1`; a home that cannot be written is
  `pendiente:` and exit `0`, because the project is done.
- `pack` installs the bundled skills into the agents' own directories: the home
  for the seven home-scoped agents, the project (`.opencode/skills`) for
  opencode, and `./.agents/skills` or `~/.agents/skills` for the `standard`
  pseudo-agent. Without `--agents` it serves the agents detected here — their
  config directory in the home, or `.opencode/` in the project; when none is
  detected it exits `0` with one line naming `--agents`. A refused agent (`pi`,
  an invented name) exits `1` and writes nothing, so a typo never leaves half a
  payload installed. `--scope` installs the named agents under the other root,
  `--dest` installs them all into one directory the human typed (the names stop
  being lookups and become labels), and the two flags are refused together
  before anything is read. The report is grouped per destination — two agents
  that share a folder share a line — and `--json` has one row per destination
  with `agents: []` inside. The summary is the only surface that prints a
  relative path: `~/...` for the home, `./...` for the project, the typed path
  for `--dest`.
- `skills` is the only verb that writes a generated file: `--write` regenerates
  `SKILLS.md` from the frontmatter and `skill new` runs the same writer, so the
  registry and the skills cannot disagree. `--check` exits `1` with one Spanish
  line per violation, each naming the file; a repository without `skills/` exits
  `0` and says `odachi skill new <slug>`. No skill means no error.

## Notes

- Package paths are stable and are the review unit: a change to
  `internal/domain/gate` is reviewable by reading `gate.go` and `gate_test.go`
  only.
- `internal/archtest` has no implementation file on purpose — it is a guard, and
  the guard must exist before the code it guards.
- `internal/app/*` has no filesystem tests of its own: the `Files` port is an
  in-memory double, and `cmd/odachi` exercises the real adapter end to end.
- `ask` and `summary` only recognise `NNNN-<slug>.md`. A hand-written document
  such as this repository's own `odd/tasks/odachi-bootstrap.md` is deliberately
  invisible to them, so their canonical render can never rewrite it. `plan`
  resolves the current document the same way, so it inherits the same limit.

## Review lenses and the refuter (T21)

The four lens prompts are the product. They live as content with a registry
header, next to the code that selects them, and a test counts their rules.

| What | Where | How it is verified |
| --- | --- | --- |
| Four prompts | `internal/assets/review/{risk,readability,reliability,resilience}.md`, each with frontmatter (`name`, `description: Trigger: …`), `## Qué mirar` with 12–20 concrete rules and `## Qué NO reportar`; all four inherit the five rules and state that `WARNING` and `SUGGESTION` never block | `internal/assets/review/lens_test.go` (frontmatter, rule count, rule length, both sections, severity floor, style ban) |
| Selection | `internal/domain/review/tier.go`: `Classify(changedPaths)` is pure and volume-blind; `Lenses(tier)` returns none, `[reliability]`, or the four in fixed order; `Add` never removes; `ParseLenses` reads `--lens` | `internal/domain/review/tier_test.go` (one test per signal, docs-only, 1000 docs, signal beats docs, `Add`, `ParseLenses`) |
| Stamping | The verdict records `lenses`: the tool stamps what the candidate required, never the reviewer | `internal/app/review/lenses_test.go`, `internal/domain/review/lens_test.go` (`Missing`) |
| Refuter prompt | `--refute <veredicto>` prints three questions per causal `BLOCKER`/`CRITICAL`, with ids and the diff; warnings are never refutable | `internal/domain/review/refute_test.go`, `cmd/odachi/reviewlens_test.go` |
| Default stands | A `refuted` with no citation becomes `stands`; an unanswered blocking finding is recorded as `stands`; an off-list id is refused | `internal/domain/review/refute_test.go` |
| One refutation per tree | `.odachi/reviews/<treehash>.refuted.json`, written once, listed by `--show` | `internal/app/review/lenses_test.go`, `cmd/odachi/reviewlens_test.go` |
| Stricter gate | `check --require-review` requires the verdict, every required lens, and no unrefuted blocker; a refuted blocker stops blocking | `internal/app/check/lens_test.go`, `cmd/odachi/reviewlens_test.go` |
| No new dependency | the domain stays stdlib-only; the assets package embeds content | `internal/archtest/boundary_test.go` stays green |

## Doctor and bench (T22, T23)

Two instruments, not two features: one reads the project and writes nothing, the
other runs the binary in throwaway directories and measures how a person gets
stuck. Neither changes any existing verb.

| What | Where | How it is verified |
| --- | --- | --- |
| Eleven checks, fixed order | `internal/domain/doctor`: `CheckVersion` … `CheckIgnored`, `CheckPersonas` (appended: the number is the check's identity), `CheckCount = 11`, a `title(number)` for each, and a test that asserts the numbering and the count | `internal/domain/doctor/doctor_test.go` |
| Two facts in check 9 | the project's own `skills/` registry AND the skills installed in `.agents/skills`, each labelled by path so one line cannot lie about the other | `internal/app/doctor/doctor_test.go`, `cmd/odachi/doctor_debts_test.go` |
| La voz coincide con el perfil | check 11 lee la sección `## Cómo hablás` del bloque y la compara con `.odachi/profile.toml`: si no coinciden nombra las dos y el arreglo (`odachi level --to`), si el bloque no tiene sección manda a `init`, y si no hay perfil no hay nada que comparar. Los archivos de persona de la v0.10.0 se reportan como `ok`, nunca como problema | `internal/app/doctor/doctor_test.go`, `cmd/odachi/doctor_debts_test.go` |
| One line each, Spanish | `doctor.Line`: `ok/mal/aviso · <n> · <título>: <detalle>[; <arreglo>]`; a problem always carries the phrase that repairs it | `doctor_test.go`, `cmd/odachi/doctor_test.go` |
| Exit code | `doctor.Code`: 1 as soon as one check found a problem. A warning is said and never fails | `TestTallyAndCode`, `TestEmptyProtectedDirectoryIsAWarning` |
| Read-only, structurally | `internal/app/doctor.Files` has no write method: a write is not expressible. The CLI test compares the whole project tree before and after | `TestDoctorNeverWrites`, `cmd/odachi/doctor_test.go` `TestDoctorWritesNothing` |
| The check that matters most | `odachi` on the PATH, with the exact `export PATH=…` line built from the running binary's folder (`selfExe`) | `TestEachProblemNamesItsCommand` (PATH case), `TestHookNamingOdachiByNameUsesThePath`, `TestPathFix` |
| Hook and AGENTS.md markers | `doctor.IsOurHook` reads the marker `init` writes, `doctor.HookExe` reads the `ODACHI=` line, `doctor.HasBlock` needs both markers. The CLI test runs `init` and then `doctor`, so the writer and the reader cannot drift apart | `doctor_test.go`, `cmd/odachi/doctor_test.go` `TestDoctorSaysAConfiguredProjectIsFine` |
| Corpora as data | `internal/bench.Journeys()`: eight recorridos with setup actions and measured steps, no logic | `TestJourneysAreTheCorpusContractXVAsksFor` |
| Four-way classification | `bench.Command` finds only `odachi <verb> …` in the message; `bench.IsTemplated`; `bench.KindOf` → `in_band` (ran, exit 0), `out_of_band` (named nothing), `dead_end` (ran, did not resolve), `templated` (a `<placeholder>`) | `TestCommandFindsTheContinuation`, `TestIsTemplated`, `TestKindOf` |
| Never a shell | `bench.Command` returns only commands starting with `odachi`, and `internal/app/bench` runs that command as the binary under test. Setup commands (`git init`) are the only other thing it executes | `TestTheBinaryUnderTestRunsTheContinuation`, `TestTemplatedIsNeverRun` |
| Controlled HOME | `bench.SandboxEnv` replaces `HOME`, `USERPROFILE` and the git config with paths inside the sandbox, in a fixed order | `TestSandboxEnvIsAControlledHome`, `TestSandboxGetsAControlledHome` |
| Removed sandboxes | `internal/app/bench` creates one directory per journey and always removes it; a directory that cannot be removed is said out loud | `TestSandboxesAreRemoved`, `cmd/odachi/bench_test.go` `TestBenchRemovesItsSandboxes` (real temp directory) |
| Ratchet | exit 1 when a `dead_end` exists, and also when a journey could not be measured at all | `TestTheThreeCounts`, `TestNoDeadEndIsGreen`, `TestSandboxFailureIsAReport` |
| Report | journeys, steps, blocks, the four counts, one line per journey, and the dead ends with the command, its exit and the message that named it; `--json` carries the same | `TestHumanLinesCarryTheCounts`, `cmd/odachi/bench_test.go` `TestBenchJSONSaysTheSameThing` |
| No new dependency | `internal/domain/doctor` and `internal/bench` are stdlib-only; the app layers use the existing ports | `internal/archtest/boundary_test.go` stays green |

## Init, personas and portability (T29–T32)

The block that turned three steps into one: `init` leaves the project ready,
`pack` says where, and the voice stops living inside the file the model never
reads. v0.11.0 added the native agent as a generated mirror of the block.

| What | Where | How it is verified |
| --- | --- | --- |
| One command | `init` runs `git init` when `.git` is missing, then writes the AGENTS.md block, the `.gitignore` block, the hook, the project agent, the home door and the skills, in that order, each step idempotent | `cmd/odachi/init_skills_test.go` `TestInitInstallsTheSkillsInOneCommand`, `TestInitSecondRunWritesNothing`; `internal/app/init/init_test.go` |
| The block says what odachi is | `internal/app/init/agents.go`: the block opens with what odachi is, who decides what, and where the context lives (`brief.md`, `odd/tasks/`, `SKILLS.md`), then the four hard rules and the six verbs | `internal/app/init/agents_block_internal_test.go` (byte for byte, and the novice sentences must still match `persona.Voice(novato)`) |
| Curriculum vs entry point | eight skills to `./.agents/skills`, `bootstrap` to `~/.agents/skills`; `internal/app/init` decides the two destinations and the CLI's `initSkills` join lets `pack` own how a skill is written | `TestInitInstallsTheSkillsInOneCommand`, `TestInitWritesOnlyTheBootstrapSkillIntoAnIsolatedHome` |
| The suite never reaches the real home | `cmd/odachi/init_skills_test.go` `TestMain` points the home adapter at a temp directory before any test runs, and removes it after | `TestMain` in `cmd/odachi/init_skills_test.go` |
| A second project copy is named, not touched | somebody who ran `pack --agents opencode` before `init` keeps that copy; `init` says so and leaves `.agents/skills` as the project's | `TestInitAcknowledgesTheOpenCodeCopy`, `internal/app/init/skills.go` `Notice` |
| A home that cannot be written is not a failed project | the project half is written, exit `0`, and the line starts with `pendiente:` plus the command that retries | `TestInitKeepsTheProjectWhenTheHomeIsNotWritable` |
| The native agent is a mirror | `.opencode/agents/odachi.md` (`description`, `mode: primary`) and `.claude/output-styles/odachi.md` (`name`, `description`, `keep-coding-instructions: true`); both wrap the same `rules.Body(voice)`, so the agent a human picks reads the block the project publishes | `internal/domain/persona/persona_test.go` `TestFilesMirrorTheBlockForEveryClient`, `TestOpenCodeFileIsAPrimaryAgent`, `TestClaudeFileKeepsTheCodingInstructions`; `cmd/odachi/init_personas_test.go` `TestInitWritesTheVoiceIntoTheBlock` |
| The door in the home | `~/.config/opencode/agents/odachi.md` and `~/.claude/output-styles/odachi.md`: the same agent name in the client's global directory, so it appears in any folder. The body is a dispatcher, not a copy of the block: it reads the project's block when there is one and hands over to `bootstrap` when there is not | `internal/domain/persona/persona_test.go` `TestGlobalFilesAreTheDoorInTheHome`, `TestProjectAndGlobalShareTheName`; `internal/app/init/personas_test.go` `TestInitInstallsTheGlobalDoor`, `TestInitKeepsTheProjectWhenTheHomeIsNotWritable`; `cmd/odachi/init_skills_test.go` `TestInitWritesOnlyTheExpectedThingsIntoAnIsolatedHome` |
| The mirror follows the profile | `init` installs it, `level --to` and `start` refresh it when it exists; a missing file is init's to install, not level's to invent | `internal/app/level/level_test.go` `TestLevelToRefreshesTheNativeAgentWhenItExists`, `TestLevelToDoesNotInstallAMissingNativeAgent`; `internal/app/start/start_test.go` `TestStartRefreshesTheNativeAgentItFinds` |
| The mirror is compared, not assumed | doctor's check 11 appends the mirror's state: it names a missing file or a hand edit and points at `odachi init`; the block still works without a selector, so it is a note, not a failed check | `internal/app/doctor/doctor_test.go` `TestAStaleNativeAgentIsSaidAndNotAFailure`, `TestAMissingNativeAgentIsSaidAndNotAFailure` |
| Generated, therefore backed up | a hand edit is lost work: the first line says the file is generated, an identical file is not rewritten, a differing one is backed up first | `TestInitBacksUpAHandEditedAgent`, `TestInitRerunStillWritesNothing` |
| v0.10.0 leftovers are named, never deleted | the old `odachi-{novato,senior}.md` files still work as a voice but do not mirror the block; `init` and `doctor` name them in one line and leave them alone | `TestInitReportsTheV0100PersonaFilesAndTouchesNothing`, `TestLeftoverPersonaFilesAreNotAProblem` |
| `standard` | a pseudo-agent for the cross-client location: `.agents/skills` in the project, `~/.agents/skills` in the home. Offered by name, refused by nobody, never detected | `internal/domain/agent/agent_test.go` `TestStandardIsTheCrossClientLocation`, `pack_test.go` `TestStandardInstallsTheCrossClientLocation`, `TestDetectionIsUnchangedAndNeverInventsStandard` |
| `--scope` | the same agent under the other root; the one exception the table cannot predict is `opencode` under home, which is `~/.config/opencode/skills` because that is the config dir its own docs name | `TestScopeOverridesTheRootAndKeepsTheAgent`, `TestOpenCodeHomeLandsInItsGlobalConfigDir`, `TestEveryAgentHasADestinationInBothRoots`, `pack_test.go` `TestScopePutsTheSameAgentInTheOtherRoot` |
| `--dest` | one directory for every named agent, absolute as typed or relative to the project; the names stop being lookups and become labels, so a client odachi has never heard of is served | `TestLabelsServeAClientTheTableDoesNotKnow`, `pack_test.go` `TestDestServesAClientTheTableDoesNotKnow`, `TestDestAbsoluteIsUsedAsGiven` |
| Two answers are no answer | `--dest` with `--scope` is refused before anything is read, and each flag needs named agents | `TestDestAndScopeAreRefusedTogether`, `TestDestAndScopeNeedNamedAgents` |
| Grouped report | one target per directory, so agents that share one share a line; `--json` has one row per destination with `agents: []` inside | `internal/app/pack/pack.go` `targetsFor`, `cmd/odachi/pack.go` `packDest`, `cmd/odachi/pack_test.go` |
| The failure names the file | a destination that fails keeps the first `Path` and `Reason` as fields, so `init` can say which half is pending without parsing a line | `TestTheFirstFailedFileKeepsItsPathAndReason`, `cmd/odachi/init_skills_test.go` |
| Pre-flight | a folder with no `.git`, no `brief.md`, no `AGENTS.md` and no `.odachi/` gets one situation and one step (`init`, then `start`), exit `1`, nothing written | `internal/app/doctor/preflight_test.go`, `internal/domain/doctor/doctor_test.go` `TestNotAProjectNamesBothCommandsInOrder`, `cmd/odachi/doctor_test.go` `TestDoctorOnAFolderThatIsNotAProject` |
| The second shape does not leak | a folder with any one of those four keeps the eleven checks with their per-check fixes, and a read error falls through to the checks instead of claiming there is no project | `TestThePreflightOnlyFiresWhenThereIsNothing`, `TestAnUnreadableFolderFallsThroughToTheChecks` |
| `--json` says the same two lines | the machine payload gains `lines`, and `checks` is empty in the pre-flight shape | `cmd/odachi/doctor_test.go` `TestDoctorJSONSaysTheSameThing` |
| No new dependency | `internal/domain/persona` and `internal/domain/agent` are stdlib-only, and the persona table hangs off the agent names with a test that fails on a client the skills table does not have | `internal/archtest/boundary_test.go`, `internal/domain/persona/persona_test.go` `TestClientsAreTheAgentsTheSkillsTableKnows` |
