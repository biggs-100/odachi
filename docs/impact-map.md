# Impact map

Who depends on each package, and which tests prove it. Update this table in the
same work unit as the code change.

| Package | Who depends on it | Tests that cover it |
| --- | --- | --- |
| `cmd/odachi` | Nobody imports it; it is the binary | `cmd/odachi/main_test.go`, `cmd/odachi/ask_test.go`, `cmd/odachi/level_test.go`, `cmd/odachi/reuse_test.go` |
| `internal/app/start` | `cmd/odachi`: `start` and `odachi` with no arguments | `internal/app/start/start_test.go` |
| `internal/app/tasks` | `internal/app/{ask,level,summary}`: find the current or every feature document | `internal/app/tasks/tasks_test.go` |
| `internal/app/ask` | `cmd/odachi`: `ask` in its three forms | `internal/app/ask/ask_test.go` |
| `internal/app/level` | `cmd/odachi`: `level` and `level --to` | `internal/app/level/level_test.go` |
| `internal/app/summary` | `cmd/odachi`: `summary` | `internal/app/summary/summary_test.go` |
| `internal/app/check` | `cmd/odachi`: `check` in all its forms | `internal/app/check/check_test.go` |
| `internal/app/init` | `cmd/odachi`: `init` (the AGENTS.md block and the pre-commit hook) | `internal/app/init/init_test.go` |
| `internal/domain/brief` | `internal/app/start` (render and read `brief.md`), `internal/app/check` (is it complete?) | `internal/domain/brief/brief_test.go`, `internal/domain/brief/render_test.go` |
| `internal/domain/taskdoc` | `internal/app/start` (write the numbered task doc), `internal/app/{ask,level,summary}` (history, questions, counts), `internal/app/check` (open questions) | `internal/domain/taskdoc/taskdoc_test.go`, `internal/domain/taskdoc/history_test.go` |
| `internal/domain/fence` | `internal/app/check` (the fence gate); `start` when it writes `fence.toml` (T5) | `internal/domain/fence/fence_test.go` |
| `internal/domain/gate` | `internal/app/check` (the four gates and the exit code) | `internal/domain/gate/gate_test.go` |
| `internal/domain/proof` | `internal/app/check` (write and read `proof/NNNN-<slug>.txt`) | `internal/domain/proof/proof_test.go` |
| `internal/domain/profile` | `internal/app/start`, `internal/app/level`, the profile decoder adapter | `internal/domain/profile/profile_test.go`, `internal/domain/profile/next_test.go` |
| `internal/adapters/{fsx,gitx,profiletoml}` | `cmd/odachi` only; the domain never imports them | the matching `_test.go` in each adapter |
| `internal/adapters/shellcmd` | `cmd/odachi` only (the `check --run` shell); one file per platform | `shellcmd_test.go` plus the per-platform `_test.go` |
| `internal/archtest` | Nobody imports it; it guards every domain package | `internal/archtest/boundary_test.go` |

## `cmd/odachi` behaviour

- Commands: `start` (default), `ask`, `level`, `summary`, `check`, `init`. `odachi`
  with no arguments is `odachi start`; a first argument that starts with `-` is
  treated as `start` flags.
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
