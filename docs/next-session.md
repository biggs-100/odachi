# Next session: adopting odachi without scattering, and one upgrade per machine

Written at the close of the session that shipped v0.11.66, at the human's request: the plan
and the method live here, because the chat is gone tomorrow.

## Read this first, before any feature

Open gentle-ai and read how it solved the same problem. It is not in this repo: it lives at
`C:\Users\USER\Desktop\herramientas\gentle-ai` on the machine where this was written, and the
human authorised reading it for this work.

Read, in this order:

1. `docs/usage.md`, the sections `### sync` and `### uninstall`.
2. `~/.gentle-ai/state.json` (or the `state.json` its docs name): what it remembers and why.
3. `docs/components.md#gga-behavior` only if the sync scope is unclear.

What to extract, in one line each:

- managed assets live with the agent in the home, never per project;
- a state file records **which** agents are managed, so `sync` touches only those;
- `gentle-ai sync` refreshes every managed asset to the current binary, is idempotent, has
  `--dry-run`, and names the changed files;
- version-bound assets **fail closed** until sync succeeds.

## The problem, measured

In a scratch project with one commit (`git init`, one file, one commit), with `HOME` redirected:

```
odachi init                       -> .agents/ (15 skills), .opencode/, .claude/, AGENTS.md, .gitignore
odachi init --global --skills home -> AGENTS.md, .gitignore, hook; 16 skills and 4 agents in the home
odachi init  (plain, afterwards)  -> .agents/skills: 15 installed   <- scatters again
store after all of it             -> agents = "both", project_agents = "none"   (no skills scope)
```

So the flags exist, and the money quote is the third line: **the store remembers the agent
choice (0081) but not the skills scope**, so any later plain `init` rebuilds the project copy.
That is the bad experience the human named: review and update every project after an upgrade.

## The plan, in the order the human chose

Each one is a feature with its own brief, fence, proof and close. Nothing here changes a
default without the human's words.

### F1 - the store remembers the skills scope

**Done, 0117.** `skills = "project" | "home"` lives in `agents.toml`, a plain init honours it,
`--skills` changes it, and a project that never chose keeps the project copy. The store's schema
moved to 2. What the next session still owes: nothing here.

The original note, kept because it is the reasoning: `odd/profile.toml` is the profile; the
store is `.git/odachi/agents.toml`. It already carried `agents` and `project_agents`. The skills
scope joined them with the same rule 0081 gave the agents. Schema care: a new key in a file the
binary reads is a release-bound change (see the schema note in the debt history - an old binary
must say "upgrade", never "unknown key").

Acceptance, now pinned as tests: after `init --skills home`, a plain `init` writes no
`.agents/skills`; after `init --skills project`, it does; a project with no stored choice keeps
today's answer.

### F3 - one command refreshes the machine, not each project

The gentle-ai shape: a verb that rewrites every managed asset in the home (skills, door,
subagents, default agent) to this binary's version, idempotent, `--dry-run`, naming what changed
and what it skipped. It is not `init`: `init` is per project, and this is per machine.

Acceptance: run it twice, the second is a no-op; `--dry-run` changes nothing and lists the
files that would change; a home written by an older binary becomes current in one command.

### F2 - a project that already lives gets the minimal footprint

Adopting odachi in an advanced repository should propose "the project keeps the rules, the craft
lives in the machine": no `.agents/`, no `.opencode/`, no `.claude/`, only `AGENTS.md`,
`.gitignore` and the hook. `--skills project` and `--agents` ask for the copies back.

This one changes a default, so it needs the human's words before anything is written. Do not
infer them from this file.

## What is already true, so it is not re-decided

- 0106: `init --global` leaves the project's agent files alone (the store records `none`).
- 0081: the home-only agent choice survives a plain init.
- 0117: the skills scope is the third stored choice (`skills = project|home`), so a plain init
  after `--skills home` writes no `.agents/skills`. The store's schema is 2; one counter serves
  the three strict files, and the accepted cost is written where the counter lives.
- 0111 + 0112: `doctor` warns when `AGENTS.md` carries a block older than the binary, naming
  `odachi init`; the comparison runs even without `odd/profile.toml`.
- The block (0116) says the human is in the loop: a question is answered, not implemented, and
  an irreversible action names its target and waits for a yes. **Push, PR, merge, release and
  any install wait for the human.**

## How this session ended, so the next one does not repeat it

`docs/review-debt.md` has nothing open but the accepted rows. The last features (0110-0116)
came from the human's own report or from running the tool's harness (`odachi bench`) and the
hygiene trio (`go vet`, `go mod tidy`, `gofmt -l .`). That is still the cheapest way to find
real work when no row is open.
