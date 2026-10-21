# Impact map

Who depends on each package, and which tests prove it. Update this table in the
same work unit as the code change.

| Package | Who depends on it | Tests that cover it |
| --- | --- | --- |
| `cmd/odachi` | Nobody imports it; it is the binary | `cmd/odachi/main_test.go`, `cmd/odachi/ask_test.go`, `cmd/odachi/level_test.go`, `cmd/odachi/reuse_test.go`, `cmd/odachi/version_test.go`, `cmd/odachi/install_test.go`, `cmd/odachi/skills_test.go`, `cmd/odachi/pack_test.go` |
| `internal/app/start` | `cmd/odachi`: `start` and `odachi` with no arguments | `internal/app/start/start_test.go` |
| `internal/app/tasks` | `internal/app/{ask,level,summary}`: find the current or every feature document | `internal/app/tasks/tasks_test.go` |
| `internal/app/ask` | `cmd/odachi`: `ask` in its three forms | `internal/app/ask/ask_test.go` |
| `internal/app/level` | `cmd/odachi`: `level` and `level --to` | `internal/app/level/level_test.go` |
| `internal/app/summary` | `cmd/odachi`: `summary` | `internal/app/summary/summary_test.go` |
| `internal/app/check` | `cmd/odachi`: `check` in all its forms | `internal/app/check/check_test.go` |
| `internal/app/init` | `cmd/odachi`: `init` (the AGENTS.md block and the pre-commit hook) | `internal/app/init/init_test.go` |
| `internal/app/skills` | `cmd/odachi`: `skills` (list, `--check`, `--write`) and `skill new` | `internal/app/skills/skills_test.go` |
| `internal/app/pack` | `cmd/odachi`: `pack` (the detected agents, `--agents`, `--dry-run`) | `internal/app/pack/pack_test.go` |
| `internal/domain/brief` | `internal/app/start` (render and read `brief.md`), `internal/app/check` (is it complete?) | `internal/domain/brief/brief_test.go`, `internal/domain/brief/render_test.go` |
| `internal/domain/taskdoc` | `internal/app/start` (write the numbered task doc), `internal/app/{ask,level,summary}` (history, questions, counts), `internal/app/check` (open questions) | `internal/domain/taskdoc/taskdoc_test.go`, `internal/domain/taskdoc/history_test.go` |
| `internal/domain/fence` | `internal/app/check` (the fence gate); `start` when it writes `fence.toml` (T5) | `internal/domain/fence/fence_test.go` |
| `internal/domain/gate` | `internal/app/check` (the four gates and the exit code) | `internal/domain/gate/gate_test.go` |
| `internal/domain/proof` | `internal/app/check` (write and read `proof/NNNN-<slug>.txt`) | `internal/domain/proof/proof_test.go` |
| `internal/domain/profile` | `internal/app/start`, `internal/app/level`, the profile decoder adapter | `internal/domain/profile/profile_test.go`, `internal/domain/profile/next_test.go` |
| `internal/domain/skill` | `internal/app/skills` (parse each `SKILL.md`, validate the set, generate `SKILLS.md`) | `internal/domain/skill/skill_test.go` |
| `internal/domain/agent` | `internal/app/pack`: the verified agent → skills directory table and the refused names | `internal/domain/agent/agent_test.go` |
| `skills` (repository root) | `cmd/odachi` (`pack` reads the embedded payload); the Go package lives in the same directory as the `SKILL.md` files it embeds, so there is no second copy | `skills/embed_test.go` |
| `internal/adapters/{fsx,gitx,profiletoml}` | `cmd/odachi` only; the domain never imports them | the matching `_test.go` in each adapter |
| `internal/adapters/shellcmd` | `cmd/odachi` only (the `check --run` shell); one file per platform | `shellcmd_test.go` plus the per-platform `_test.go` |
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
| Bundled five | `skills/{acceptance,data,ship,slices,tests}/SKILL.md`, each with the three sections and at most 60 lines | `cmd/odachi/skills_test.go` runs `skills --check` over this repository |

`odachi init` does not own the skills: its block only gains the line that points
at `SKILLS.md`. Copying the bundled skills into a project is `pack`, a later
unit.

## Pack (T13)

`odachi pack` is what makes the skills reachable by the agent: they live in this
repository, and an agent only loads what is in its own skills directory.

| What | Where | How it is verified |
| --- | --- | --- |
| Source | The five `skills/<name>/SKILL.md` files compiled into the binary with `//go:embed */SKILL.md`, in the `skills` package. No network, no repository at run time, no duplicated copy | `skills/embed_test.go` compares the embedded bytes with the files on disk |
| Destination table | `internal/domain/agent`: `claude .claude/skills`, `codex .codex/skills`, `cursor .cursor/skills`, `gemini .gemini/skills`, `copilot .copilot/skills`, `qwen .qwen/skills`, `kiro .kiro/skills`, all home-relative. It is data, not branches of code, because every path was verified against the agent's adapter | `internal/domain/agent/agent_test.go` pins every row, `Display()` and `Marker()` |
| Refusals | `pi` is refused with `Pi no usa un directorio de skills` (its adapter reports none); `opencode` is refused because its path depends on the platform and is not guessed; any other name is refused naming the supported ones | `agent_test.go`, `internal/app/pack/pack_test.go`, `cmd/odachi/pack_test.go` |
| Write rule | `<home>/<dest>/<name>/SKILL.md` through `fsx.WriteAtomic`; an existing file that differs is backed up first; an identical one is reported as `sin cambios` and not rewritten | `pack_test.go` (second run: no write, no backup; edited file: backup before write) and the end-to-end run |
| Never deletes | Nothing is deleted or moved; a `SKILL.md` odachi did not write is reported with its path and left alone | `TestForeignFileIsReportedAndNotTouched` |
| Failure isolation | One agent's failure does not abort the others, and the message names the agent and the path | `TestOneFailureDoesNotAbortTheOthers` |
| `--dry-run` | Zero writes and zero backups, asserted on the double, never on the output; the report still lists the plan and says `por instalar` | `TestDryRunTouchesNothing` |
| Project untouched | `pack` never reads or writes the project's `skills/` or `SKILLS.md`: those belong to `odachi skills --write` | `TestPackDoesNotTouchTheProject` |
| Home | The new `Home` port resolves the home; `internal/app/pack` imports no `os`, so every write goes through the `Files` double in tests and through `osFiles` in the binary | `internal/app/pack/pack_test.go` (in-memory `Files` + `Home` doubles), `osHome` in `cmd/odachi/files.go` |

## `cmd/odachi` behaviour

- Commands: `start` (default), `ask`, `level`, `summary`, `check`, `init`,
  `skills`, `skill new`, `pack`, `version`. `odachi`
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
  same profile name.
- `--json` is machine output and implies non-interactive. `start`'s JSON shape
  is unchanged; other verbs add `lines` only when they report a block.
- `check` is the only verb whose exit code is the gate's, not a fixed one: `0`
  ready · `1` the agent can act · `2` only the human can decide. `--run` returns
  the gate's code, never the command's. Its third adapter is `shellcmd`; the
  other two are `osFiles` and `gitRepo` (through `gitx`).
- `check --run` keeps `proof/` out of the git tree with `proof/.gitignore`:
  without it the proof would change the tree it certifies and every run would
  invalidate the previous proof. `--quiet` prints nothing only when ready (the
  pre-commit hook depends on that) and still prints the block when not.
- `init` writes the rules block between `<!-- odachi:begin -->` and
  `<!-- odachi:end -->` and the `.git/hooks/pre-commit` hook. It never replaces a
  hook it did not write: it prints the three lines to paste and exits `1`.
- `pack` installs the bundled skills into the agents' own directories and takes
  no project directory at all. Without `--agents` it serves the agents whose
  config directory (`~/.claude`, `~/.codex`, ...) exists; when none does, it exits
  `0` with one line naming `--agents`. A refused agent (`pi`, `opencode`, an
  invented name) exits `1` and writes nothing, so a typo never leaves half a
  payload installed. The summary is one line per agent, and it is the only
  surface that prints a home-relative path.
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
  invisible to them, so their canonical render can never rewrite it.
