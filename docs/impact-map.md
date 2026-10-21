# Impact map

Who depends on each package, and which tests prove it. Update this table in the
same work unit as the code change.

| Package | Who depends on it | Tests that cover it |
| --- | --- | --- |
| `cmd/odachi` | Nobody imports it; it is the binary (T5, and later T6/T7 commands) | `cmd/odachi/main_test.go` |
| `internal/app/start` | `cmd/odachi` (`start`, and `odachi` with no arguments) | `internal/app/start/start_test.go` |
| `internal/domain/brief` | `internal/app/start` (render and read `brief.md`); CLI `check` later | `internal/domain/brief/brief_test.go`, `internal/domain/brief/render_test.go` |
| `internal/domain/taskdoc` | `internal/app/start` (write the numbered task doc); CLI `check` later | `internal/domain/taskdoc/taskdoc_test.go` |
| `internal/domain/fence` | CLI `check` (T7); `start` when it writes `fence.toml` (T5) | `internal/domain/fence/fence_test.go` |
| `internal/domain/gate` | CLI `check` (T7), CLI `summary` (T6) | `internal/domain/gate/gate_test.go` |
| `internal/domain/profile` | CLI `start` (T5), CLI `level` (T6), profile decoder adapter (T4) | `internal/domain/profile/profile_test.go` |
| `internal/archtest` | Nobody imports it; it guards every domain package | `internal/archtest/boundary_test.go` |

## `cmd/odachi` behaviour

- `odachi` with no arguments is `odachi start`. A first argument that starts
  with `-` is also treated as `start` flags, so `odachi --outcome ...` works.
- `odachi start` is the only available command in this session. Any other first
  argument exits `1` and prints `comandos disponibles: start`.
- Exit codes: `0` artifacts written, `1` blocked (invalid flag, unreadable or
  nameless profile, existing `brief.md` without `--force`), `2` a human answer
  is missing and this run cannot ask for it.
- Two adapters are wired here and nowhere else: `osFiles` (through
  `internal/adapters/fsx`, so an existing file is backed up before the atomic
  rewrite) and `terminalInterview` (stdin/stdout). `internal/app/start` declares
  the ports and never imports `os`.
- `.odachi/profile.toml` is read when it exists and wins over
  `profile.Infer`: the human owns that file. A field left out keeps the
  built-in default of the same profile name.
- `--json` is machine output and implies non-interactive: the run uses the flags
  or exits `2` naming them.
- The profile is resolved by `internal/domain/profile`; `cmd/odachi` only loads,
  validates and merges the human's file.

## Notes

- Package paths are stable and are the review unit: a change to
  `internal/domain/gate` is reviewable by reading `gate.go` and `gate_test.go`
  only.
- `internal/archtest` has no implementation file on purpose — it is a guard, and
  the guard must exist before the code it guards.
- `internal/app/start` has no filesystem tests of its own: `Files` is an
  in-memory double in `fakes_test.go`, and `cmd/odachi/main_test.go` exercises
  the real adapter end to end (backup before rewrite, atomic write).
