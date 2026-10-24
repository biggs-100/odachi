# Odachi

**Two hands: you decide, the agent builds. And you learn the craft.**

## What it is

- A CLI that makes the agent do the dirty work with senior discipline, while you
  make the decisions that matter: the vision, the scope, what is non-negotiable.
- A tutor: every question explains why it exists, so the person stops being a
  beginner and does not depend on the agent forever.
- Four plain-text artifacts that act as memory: `brief.md`,
  `odd/tasks/NNNN-<slug>.md`, `proof/NNNN-<slug>.txt` and `.odachi/profile.toml`.

## Installation

If you do not have Go, the installer downloads the binary from a release, verifies
its sha256 and leaves it ready. The three commands:

```sh
# linux and macOS
curl -fsSL https://raw.githubusercontent.com/biggs-100/odachi/main/scripts/install.sh | bash

# windows (PowerShell)
irm https://raw.githubusercontent.com/biggs-100/odachi/main/scripts/install.ps1 | iex

# and to confirm which version you got
odachi --version
```

Those three lines work. The latest release is `v0.11.9`: six binaries (Linux,
macOS and Windows, amd64 and arm64) and a `checksums.txt`. The installer resolves
the latest version through the API, verifies the sha256 and only then writes; a
checksum that does not match exits with an error and leaves **nothing** at the
destination. The whole chain is tested against the real release, not a fixture: a
test builds the `tar.gz` and the checksums the way goreleaser publishes them,
runs `install.sh` offline through `file://`, and the installed binary answers its
version.

The installer also runs `odachi setup`: the bootstrap skill in
`~/.agents/skills`, the home door for the agent clients it finds, and OpenCode's
`default_agent`. You open your agent and describe the project; the agent runs
`odachi init` and `odachi start` itself, and asks permission before writing in a
repository that is not yours. You never type those verbs.

If you prefer to build: `go build ./cmd/odachi` (you need Go 1.25 or newer), and
the binary says `odachi dev`, because the release injects the version, not the
code.

## What it is NOT

- It is not an agent installer or a per-harness adapter.
- It is not a TUI or an application with screens.
- It is not a proprietary memory or knowledge base: the artifacts are files you
  can read, edit and version.
- It sends no telemetry, opens no network, runs in no cloud.

## The four artifacts

| File | Owner | For |
| --- | --- | --- |
| `brief.md` | human | vision, non-goals, durable decisions |
| `odd/tasks/NNNN-<slug>.md` | agent | requirements R#, steps and history |
| `proof/NNNN-<slug>.txt` | agent | real output of the proof command |
| `.odachi/profile.toml` | human | 8 lines of ceremony, never of rigor |

Complete examples in `testdata/examples/`.

`proof/` and `.odachi/reviews/` are **generated evidence**, not versioned: the
proof carries the hash of the tree that produced it, and the verdict is named
after the tree it judges. That is why both must stay out of git: otherwise the
first proof would invalidate the previous review and the first review would
invalidate the previous proof.

`init` takes care of that for you, and of two things that are not artifacts: it
leaves a marked block in the project's `.gitignore` (`/proof/` and
`/.odachi/reviews/`, versioned, so it survives a clone) and a rules block in
`AGENTS.md` —what the agent reads before touching anything— plus the pre-commit
hook that runs `odachi check --quiet` on this tree's proof. In a project without
`init`, `check --run` and `review` protect themselves with their own internal
`.gitignore`, at the price of moving the tree once.

Before rewriting one of your files, `odachi` keeps the previous version so `undo`
can go back. Those backups live in `.odachi/backups/`, preserving the artifact's
path, not beside `brief.md` or the task docs. The old ones left beside the
artifact are still read, and `odachi undo --collect` moves them into the store
without rewriting them: they keep their bytes and date, and a second run does
nothing.

## The skills, in your agent

There are fourteen skills, and `odachi init` already leaves thirteen in the
project and one in your home. `odachi pack` copies the fourteen to where your
agent looks for them, for everything `init` does not do: another client, another
folder, another scope. Each one is triggered by a situation, not by a command:

| Skill | When it triggers |
| --- | --- |
| `acceptance` | deciding whether something is done |
| `bootstrap` | starting a project from scratch, or explaining what odachi is |
| `commits` | what goes into a commit, how to split it, what the message says |
| `data` | saving, migrating, reading or deleting data: owner, lifetime and backup |
| `drive` | working inside an odachi project: which verb at each moment and how to close a feature |
| `go-tests` | Go tests, golden files or Bubbletea flows |
| `pull-request` | a change goes to a PR or the diff passes 400 lines |
| `ship` | something is ready for another person to use |
| `skill-craft` | writing a new skill or fixing one that is not used |
| `skill-improve` | a skill is not used, grew too long or the agent does not load it |
| `slices` | splitting a large request, or when the code no longer fits in one read |
| `tests` | what a change tests: the failing test first, the code after |
| `triage` | several bugs, a backlog or a report that repeats |
| `writing` | a message, a PR, a comment or documentation someone will read |

`bootstrap` is the entry point, and it is the only one that lives in
`~/.agents/skills` and not in the project: it teaches how to start in an empty
folder, and in an empty folder there is no project to put it in yet. The other
thirteen are the craft a project carries, and `init` copies them to
`./.agents/skills`. That folder is the convention the Agent Skills standard asks
for, so clients odachi does not know can find it; it also travels with the
repository, and that is why the content of somebody else's repo is untrusted
input. `bootstrap` only explains the flow: it runs nothing.

`odachi init --skills home` leaves the fourteen in `~/.agents/skills` and writes
no copy in the project; it is for one copy per machine. The default is still
`project`, because the project versions the craft and the catalog stays smaller.

Supported agents:

| Agent | `--agents` | Destination |
| --- | --- | --- |
| Claude Code | `claude` | `~/.claude/skills` |
| Codex | `codex` | `~/.codex/skills` |
| Cursor | `cursor` | `~/.cursor/skills` |
| Gemini | `gemini` | `~/.gemini/skills` |
| Copilot | `copilot` | `~/.copilot/skills` |
| Qwen | `qwen` | `~/.qwen/skills` |
| Kiro | `kiro` | `~/.kiro/skills` |
| **OpenCode** | `opencode` | `./.opencode/skills` (in the project) |

Without `--agents` it installs into the agents it detects: those with a config
folder in your home, and OpenCode if the project has `.opencode/`. It never
deletes anything, backs up what differs and does not repeat what is already
identical.

### Subagents: the worker and the verifier

`init` also writes two native subagents per client: `odachi-worker` (implements
one bounded step) and `odachi-verify` (verifies without editing). They are not
picked with a key: the main agent invokes them when `slices` triggers. The parent
passes them the task doc, the step, the edit surfaces, the commands and the exact
skill paths; they return evidence and ≤2k tokens. The parent keeps git, the gate
and the vision; there is no separate orchestrator.

### Three more ways to say where

- `odachi pack --agents standard`: not a client, the folder every client reads;
  `./.agents/skills` in the project, `~/.agents/skills` in the home. It is the
  one `init` already leaves in place.
- `odachi pack --agents opencode --scope home`: the same client, in the other
  root (`~/.config/opencode/skills`), to have one copy for all your projects
  instead of one per project.
- `odachi pack --agents claude --dest ./skills-x`: the folder you name, for a
  client that is not in the table. `--dest` and `--scope` do not go together:
  both answer "where", and two answers are no answer.

`odachi pack --agents opencode` still exists and writes to `./.opencode/skills`,
OpenCode's native folder, for when you want it there and not in `.agents/skills`.
`init` does not delete or move it: if it finds it, it names it and leaves the
project one as the `.agents/skills` one.

### Testing odachi with OpenCode

```sh
mkdir my-project && cd my-project
odachi init    # leaves the project ready in one step
odachi start   # the interview: brief.md, fence and first task doc
odachi check   # tells you what is missing and how to unblock it
```

`odachi init` leaves the project ready in one step: the git history if it was
missing, the rules block the agent reads, the project agent (picked with Tab),
the `.gitignore` block, the pre-commit hook, the thirteen skills in
`./.agents/skills`, the entry skill in `~/.agents/skills` and the profile voice
in the `AGENTS.md` block. The home door -the agent that appears in any folder- is
opt-in: `odachi init --global` installs it, and an init without the flag writes
nothing in your home. With the door, OpenCode also gets `default_agent` in
`~/.config/opencode/opencode.json` (or the `.jsonc` you already have), so a new
session starts as odachi; a default you picked before is named, never replaced.

If you do not use Claude Code, `odachi init --agents opencode` writes only the
OpenCode agent and stores the choice in `.odachi/agents.toml`; `init`, `start`,
`level` and `doctor` respect it, and `--agents claude` does the same the other
way. The files of the client you do not choose are not deleted: they simply stop
being written. Without a file, the default is still both.

Open the project with OpenCode and that is it: it reads `.agents/skills` as an
external directory in addition to its own, so it sees the thirteen without
`pack`. `pack` is for taking the skills to another client's native folder, or to
wherever you say.

## Ceremony and voice

Two different things, with different owners.

The **profile** is the ceremony the binary enforces: how much it asks, how much
it prints, what counts as proof. It lives in `.odachi/profile.toml`, `odachi
start` writes it with your answers and the binary reads it. The agent does not
read it: it cannot loosen or change a gate.

Without a file, the default profile is `novato`. `start` does not raise it to
`senior` just because there is a `.git`: `init` is the one that creates that
repo. `senior` is earned with `odachi level --to`, or inferred from a brief the
person already brought.

The **voice** is how the agent talks: whether it explains the why of every
decision or goes to the point. It lives in the last section of the `AGENTS.md`
block, `## How you speak`, and in the native agent that mirrors that block.
`odachi init` writes the `novato` voice if there is no profile yet, and
`odachi level --to <profile>` changes the profile, that section and the mirror
together —and nothing else in the block—.

The native agent has one copy per chosen client (`both` is the default). The
**project** one is the same block in the folder where your client already picks
agents: `.opencode/agents/odachi.md` (Tab in OpenCode) and
`.claude/output-styles/odachi.md` (`/config` in Claude Code). It is generated: if
you edit it by hand, `odachi init` backs it up before rewriting it and `doctor`
tells you it stopped matching. Measured on OpenCode 2.1.3: a project with `.git`
lists the file as a primary agent; a folder without `.git` lists none.

The **home** copy is the door: `~/.config/opencode/agents/odachi.md` and
`~/.claude/output-styles/odachi.md`, one per chosen client. It is opt-in:
`odachi init --global` installs it, because a primary agent that appears in
**any** folder and can replace the one you chose is your decision, not an effect
of initializing a project. Once installed it appears in any folder, not only the
one that ran `init`. If the folder has the `odachi` block, it reads it and adopts
its voice; if not, it loads the entry skill and guides you to start a project. It
is short on purpose: it does not repeat the rules, it only says where they are.
`doctor` tells you if it is missing or stopped matching, and for both the command
is `odachi init --global`. The two per-profile personas from v0.10.0 remain as
leftovers: `init` names them and does not delete them, they are yours.

## Review: four lenses, no ceremony

`odachi review` prints the review prompt for the candidate —brief, changed paths
and diff— and chooses the lenses from the evidence of the change: none if only
documents were touched, `reliability` if there is logic, all four if the diff
touches security, identity, money, data shape, a public contract, concurrency or
the machinery that publishes it. You do not choose the ceremony; the change asks
for it.

```sh
odachi review                            # the prompt, to paste into your agent
odachi review --exec "agent {prompt}"    # or have it run it: prompt on stdin and in {prompt}
odachi review --exec "cli --model {model} {prompt}"   # and with the session's model
odachi review --attach verdict.json      # or save what it answered
odachi review --show                     # which verdict exists for this tree
```

`{model}` is the model this session is using, so `--exec` can judge with the
same one you are talking to. The order is `--model`, then `ODACHI_MODEL` —any
host can set it—, then the OpenCode session. When the placeholder is there and
no model resolves, the review stops before starting the reviewer and names
`--model`; it never picks a fallback. A command without `{model}` runs exactly as
before, and the verdict records the command that really ran.

The verdict lives in `.odachi/reviews/<tree-hash>.json`, so approval cannot
float: the file name *is* the state of the code that was approved. A causal
`BLOCKER` blocks, and to unblock it there is `odachi review --refute`: the
refuter has to cite the diff line that disproves the finding, because what cannot
be cited cannot be refuted and what is not refuted stands.

Review is a human decision, not the agent's. By default it is off; the project
switch lives in `.odachi/review.toml` and is written with
`odachi review mode status|on|off`. With `on`, `odachi check` asks for this
tree's verdict; with `off`, it leaves it optional. `--require-review` and
`--no-review` force one run without changing the file. The file is versioned;
changing it moves the tree, so the proof is retaken with
`odachi check --run "<command>"`. A passive change —documents only— asks for no
verdict even with the mode on: there is no lens to answer, so there is no review
to fabricate.

## CI: the gate from the base ref

A local hook can be skipped, and a change can edit the fence it must obey. In CI,
`check --fence-from origin/main` reads the fence from that ref, not the worktree,
and the proof names it. Build the gate from the base ref, not the candidate:

```sh
git worktree add ../gate origin/main
(cd ../gate && go build -o /tmp/odachi ./cmd/odachi)
/tmp/odachi check --fence-from origin/main --run "go test ./..."
```

`.github/workflows/gate.yml` is that recipe; it runs on pull requests and on
`main`, builds the gate from the base ref, and uploads `proof/` even when the
gate fails. The workflow file itself is not trusted: `.github/CODEOWNERS` asks
for an owner review of `.github/` and `.odachi/`, and the repository ruleset can
require that review. Mark the gate as a required check when the project is ready
to stop pushing to `main` directly.

`ODACHI_NO_SETUP=1` makes the installers leave only the binary — no home door, no
default agent — which is what a shared runner wants. A missing file at a valid
ref is "no fence"; a ref that does not exist blocks, because a typo must not turn
the gate off.

## Status

Pure domain in `internal/domain/...` with an import guard, adapters in
`internal/adapters/...`, and sixteen surfaces: `start`, `ask`, `plan`, `level`,
`summary`, `undo`, `check`, `review`, `doctor`, `bench`, `init`, `setup`,
`skills`, `skill new`, `pack` and `--version`. 41 packages, green on Linux, macOS and
Windows —ubuntu repeats the internal suite with the race detector—.

Published: `v0.11.9`. What changed in each version is in
[CHANGELOG.md](CHANGELOG.md). The one-line installation is tested against the
real release, not a fixture. `install.ps1` was verified by hand on Windows with
the published release: there is no automated test because it would need an HTTP
listener in CI.

The release config **has no `brews`, signing or provenance**: there is no tap and
no keys, and pretending they exist would be worse than not having them.

## License

MIT. See [LICENSE](LICENSE).
