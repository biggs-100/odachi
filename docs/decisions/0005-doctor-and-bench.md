# 0005 · Two instruments: a health report and a friction bench

Status: accepted · Date: 2026-10 · Session: T22, T23 (the pre-flight: T29)

## Context

By T21 the product had eleven verbs and no way to answer two questions that the
verbs themselves cannot answer:

1. **"¿Está todo bien instalado?"** A person who installs odachi has a binary, a
   project, an `AGENTS.md`, a hook, a profile, a brief and a skills catalogue, and
   nothing tells them whether those seven things are actually wired together. A
   real OpenCode session answered part of it the hard way: with `odachi` outside
   the PATH, the agent made 41 shell calls — 7 of them looking for the binary,
   including a PowerShell recursion over `C:\` — and executed **zero** rules. The
   content worked; the installation did not. That failure is invisible from
   inside the tool and obvious from a health check.

2. **"¿Cada mensaje de bloqueo tiene salida?"** Contract C promises that every
   blocked result carries a `Next` that starts with `odachi `. A unit test proves
   the string; nothing proves the command *works*. The difference matters: a
   `Next` of `odachi ask` looks right in a test and is a usage error when run.

Neither is a feature a user asked for. Both are instruments: one reads, one
measures, and both are useful precisely because they are not part of the happy
path.

## Decision

### `doctor` is ten fixed checks and cannot write

The ten checks, their order and their numbering live in `internal/domain/doctor`
and are pinned by a test. The ordering is not cosmetic: the number is the line's
number, and inserting a check in the middle would silently rename every check
after it. There is exactly one folder where these ten lines are the wrong
answer —one where odachi was never started— and it has its own section below.

Read-only is structural, not a promise: `internal/app/doctor.Files` has no write
method, so "doctor never writes" is a property of the type. The CLI test also
compares the whole project tree before and after, because a comment is not a
guarantee.

Each problem carries the phrase that repairs it. Nine of the ten name an `odachi`
command. One does not, and it says so honestly:

| Check | When it fails, the fix is |
| --- | --- |
| `odachi` on the PATH | the exact `export PATH="<folder>:$PATH"` line, built from the running binary's folder |
| no `.git` | `git init` |
| no rules block | `odachi init` |
| hook missing or foreign | `odachi init` (and paste its block) |
| hook names a binary that is gone | `odachi init` |
| profile does not resolve | **fix or delete the file** and re-run `odachi doctor` |
| no `brief.md` | `odachi start` |
| `brief.md` incomplete | `odachi start --force` |
| `SKILLS.md` drifted | `odachi skills --write` |
| review store / `proof/` not ignored | `odachi review` / `odachi check --run "<comando>"` |

The profile line is the exception because **no verb can repair a malformed
`profile.toml`**: `start` and `level` both resolve the profile before they do
anything, so a broken file blocks the two commands that would rewrite it. Naming
`odachi start` there would be a dead end, and a dead end in the check that exists
to find dead ends would be exactly the wrong lesson. A missing profile is not a
problem at all: the profile is inferred from what the tool can see, which always
resolves, and `check` already reports an unconfigured project through the brief.

### A warning is said and never fails

The exit code is 0 as soon as there is no problem, warnings included. The one
warning today is a `proof/` or `.odachi/reviews/` directory that exists, is empty,
and has no `.gitignore`: nothing is broken yet, and the next write is the one that
would invalidate the evidence already there.

### A folder that is not a project yet gets one step, not ten checks

Measured, in an empty folder: the ten checks answered `4 problemas de 10
chequeos` and named three different commands —`git init`, `odachi init` twice,
`odachi start`— and none of the four problem lines said the only thing that
person needed: this folder is not a project yet, start here. A beginner cannot
order three commands out of a health report, and the two that come first were
not even distinguishable in it.

So `doctor` has two shapes. The pre-flight comes first and asks four questions:
is there a `.git` (a directory, or the file a linked worktree has), a
`brief.md`, an `AGENTS.md`, a `.odachi/`? When the answer is no to all four, the
report is one situation and one step —`odachi init` first, because it leaves the
rules the agent reads, and `odachi start` after, because it is the interview—
with exit `1` and no check lines at all. The two commands appear in the order a
person runs them, and each says what it leaves behind.

Everything else keeps the ten checks. A git repository that never ran `init`, a
project with an `AGENTS.md` and nothing else, an empty folder with a `.odachi/`
from a half-finished `start`: all of them get their ten lines with their
per-check fixes, because they are projects and their problems are per check. The
pre-flight is not a shortcut for the healthy case; it is the answer to the only
case the ten checks cannot phrase.

A read error is not a verdict. Saying "esta carpeta todavía no es un proyecto"
about a folder odachi could not read would be a lie, so an unreadable `.git`
falls through to the checks, where the unreadable path becomes its own line.

The pre-flight is read-only for the same structural reason as the rest:
`internal/app/doctor.Files` has no write method, and the test asserts no file and
no directory was created. `--json` carries the same two lines in `lines` and an
empty `checks` array, so a machine never gets a different story than a person.

### `bench` runs journeys, not code

The corpus is data in `internal/bench`: eight recorridos, each a name, a reason,
the setup that creates its state, and the invocations that are measured. Adding a
journey is adding a literal, not a branch.

Each journey runs in its own temporary directory, with `HOME`, `USERPROFILE` and
the git config pointed inside it. `stdin` is an empty pipe and never a terminal,
so a journey cannot depend on someone typing.

A classification is only as good as its extraction. `bench.Command` finds the
`odachi <verb> …` invocation in a Spanish sentence, using the tool's own verb and
flag list, and **returns nothing else**: the bench never runs a command that is
not `odachi`, because a message could name anything and the bench is not a shell.
The named command is run as the binary under test, not as a literal `odachi`, so a
machine with nothing on its PATH is still measurable.

| Kind | Meaning | Fails the bench |
| --- | --- | --- |
| `in_band` | named a command, ran it in the same sandbox, exit 0 | no |
| `out_of_band` | named no odachi command | no |
| `templated` | the command carries a `<placeholder>` | no, counted apart |
| `dead_end` | named a command that, run, did not resolve the block | **yes** |

A journey that could not be measured at all — no sandbox, a setup step that
failed, a binary that will not execute — also fails the run. A bench that measured
nothing may not report success.

## The first reading

`odachi bench` against the v0.5.0 binary, 8 journeys and 17 invocations:

```text
bench: 8 recorridos, 17 pasos, 5 bloqueos
in_band 1 · out_of_band 0 · templated 1 · dead_end 3
```

Three dead ends, all of them exit 2:

- `start-dos-veces` → `odachi start --force`
- `archivo-prohibido` → `odachi ask`
- `repo-sin-commits` → `odachi start`

They are reported here and **not fixed in this unit**, because fixing them is
choosing product behaviour, and the instrument's job was to make them visible.
Two of them are interactive handoffs (`start` and `start --force` ask the seven
questions when a human is there and name the missing flags when one is not), and
one is a bare usage error (`ask` with no question). None of them is a message that
sends someone to a place that does not exist; all three are messages that name a
command which cannot be finished without more input. Whether that counts as a dead
end is a product decision, and it is now a visible one.

## What this unit deliberately did not do

- No new verb behaviour, no new flag on an existing verb.
- No new dependency; both new pure packages are stdlib-only and the boundary
  guard stays green.
- No fix to any message the bench flagged.
- No `bench` in CI: a ratchet that starts red would be turned off within a week.
  The instrument is run by hand, and its reading is in the T23 report.

## Consequences

- `doctor` duplicates two strings it cannot import: the hook marker and the
  AGENTS.md markers are unexported in `internal/app/init`. The CLI test runs
  `init` and then `doctor` and asserts check 5 is green, so the two cannot drift
  apart silently.
- `doctor`'s check 10 reads the ignore files rather than asking git whether the
  paths are ignored. A project that ignores them through `.git/info/exclude` (the
  repair L48 recommends) would still be reported. It is the tool's own mechanism
  that is being checked, and asking git would add a git port to a verb whose
  contract has none.
- `bench`'s boolean-flag table is a copy of the CLI's surface. A new boolean flag
  would have to be added there or its value could be mistaken for a command
  argument; a new verb likewise, since `bench.Command` only recognises verbs it
  knows.
- The bench's journeys drive a real binary and a real `git`; they will notice a
  change in the CLI's exit codes as a "surprise", which is a signal and not a
  failure.
- Check 9 reads the **project's own** registry: `skills/` and `SKILLS.md`, the
  ones `odachi skills --write` generates. It does not read the installed
  `./.agents/skills` that `init` writes, because a project that never wrote its
  own skills has not drifted — it has not started. A person who just ran `init`
  reads `ninguna: no hay skills/` on that line and it is not an error: it is a
  different question, and the tool has no line that asks the other one.
