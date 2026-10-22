# 0006 · Profile and voice: two axes, two owners

Status: accepted · Date: 2026-10 · Session: T32 (T29–T31 built the ground it
stands on)

## Context

The word *perfil* was carrying two different things. `.odachi/profile.toml`
holds the ceremony the binary enforces —`questions`, `output`, `proof`,
`assumptions`, `brief_source`— and next to them `explain = why|none`, which is
not ceremony at all: it is how the agent talks. The binary never shows that file
to the model, and the model cannot change what the binary does. So the human set
one thing in a file and watched a different thing happen: one word, two sources
of truth, neither of them in charge.

Meanwhile every client already has a selector for the other axis. OpenCode cycles
its primary agents with Tab (`switch_agent`) and reads them from
`.opencode/agents/<name>.md`. Claude Code has output styles in
`.claude/output-styles/<name>.md`, chosen from `/config`. Both are per session
and per repository, which is exactly what a voice needs and what a config file
cannot do.

Two facts fixed by earlier units and respected here:

- The profile schema is frozen at eight fields, and a reflection test fails the
  build if a ninth appears (Contract B). The voice cannot grow there even if
  somebody wanted it back.
- The artifact that teaches the agent —`AGENTS.md`, the skills— is written by
  odachi. So whatever teaches the persona has to be written by the same command
  that writes the rules (the L71/T29 lesson: in a folder where nothing exists,
  nobody knows which of three commands comes first).

## Decision

### Two axes, two owners

| Axis | Where it lives | Who writes it | Who reads it | What it decides |
| --- | --- | --- | --- | --- |
| profile | `.odachi/profile.toml` | `start`, `level --to`, the human | the binary | how much the tool asks and prints, what counts as proof |
| persona | `.opencode/agents/*.md`, `.claude/output-styles/*.md` | `init` | the client, at session start | how the agent talks |

The binary never reads a persona to decide anything: it compares the bytes only
to know whether to rewrite them. The client never reads the profile. If one file
could be both, one of the two owners would be lying, and the tool would claim an
authority it does not have.

### Both personas, always

`init` cannot know who the human is: it can look at a folder and know it is
empty; it cannot look at a person. So it writes `odachi-novato` and
`odachi-senior`, in every client that has a native mechanism, and the human picks
with the switch they already use.

Deriving the persona from the profile was the obvious alternative and it is the
one that reproduces the defect: it would push the decision back into the TOML
this change exists to get `explain` out of, create a second thing that has to
agree with the first, and —because `init` runs before `start`— it would have to
guess a profile that does not exist yet. The client's switch is reversible in a
keystroke; a generated file is not.

The cost is two files per client instead of one. On a full install that is four
files of a few hundred bytes each, and a rerun rewrites none of them.

The separation is pinned, not promised: `TestInitPersonasDoNotDependOnTheProfile`
asserts identical bytes whichever profile is in force, and
`TestLevelToDoesNotRewriteThePersonas` asserts that the command that changes the
ceremony never touches the voice.

### Every persona file says which axis it owns

Both personas carry the same sentence as the second line of the body —the first
one says the file is generated— in both clients:

> La ceremonia la decide .odachi/profile.toml; este archivo decide solo cómo hablás. Si los dos no coinciden, decilo y pedí odachi level --to <perfil>.

In the file and not only here, because the line is read by the agent at session
start — the exact moment the two axes could disagree. One line, because an agent
reads a paragraph as a section and takes the first rule out of it; one sentence
is the whole rule. It names `odachi level --to <perfil>` because a disagreement
needs an exit that exists, and that is the command which changes the ceremony
without editing the voice.

The first line of the body says the file is generated and is rewritten whole,
because a hand edit is lost work and the human deserves to know before making
it. A differing file is backed up before the rewrite; an identical one is not
rewritten.

### The block avoided writing a global `AGENTS.md`

`AGENTS.md` is a project standard; there is no standard path for global rules,
and each client keeps its own (`~/.config/opencode/AGENTS.md`, `~/.codex/AGENTS.md`,
`~/.claude/CLAUDE.md`, `~/.gemini/GEMINI.md`,
`~/.copilot/copilot-instructions.md`). Writing those would be five copies of the
block and five things to drift, and the next client would get none.

Worse, a global gate rule lies. `odachi check` in a folder that is not an odachi
project has nothing to say, so a global block that says "brief antes de código"
would apply a rule nobody can satisfy to every repository the person opens. A
rule that is false half the time teaches the agent to ignore rules.

What replaced it is a skill: `bootstrap`, at `~/.agents/skills`. A skill loads by
trigger —a catalogue line at rest, the body only when a project is being
started— so it costs nothing in folders that are not odachi projects, and it can
say out loud that a gate in someone else's repository says nothing. It explains
the three commands in order and it never executes anything.

### `.agents/skills`, with the standard as the source

The Agent Skills standard asks clients to scan, besides their own directory, a
cross-client one per scope. That convention is the standard's, not a client's,
which is why it is the destination for a client odachi has never heard of:

- `./.agents/skills` in the project: the eight craft skills, written by `init`.
- `~/.agents/skills` in the home: `bootstrap` only, because before a project
  exists the home is the only place a skill can live.

`odachi pack` names this destination `standard`: a pseudo-agent, offered by name,
refused by nobody, and never detected — it is a destination, not a program.
`pack` also learned the two escape hatches the table cannot predict: `--scope`
installs a known client under the other root, and `--dest` installs every named
client into the folder the human typed, where the name is only a label.
`--dest` and `--scope` together are refused, because two answers to "where" is no
answer.

The report is grouped per destination, not per agent: two clients that share a
folder share a line, and `--json` has one row per destination with `agents: []`
inside, so the human block and the machine payload cannot tell different stories.

## Known limits

1. `doctor`'s check 9 reads the project's own `skills/` and `SKILLS.md`, not the
   installed `./.agents/skills`. A person who just ran `init` reads `ninguna: no
   hay skills/` on that line. It is not an error: it is a different question, and
   no line asks the other one.
2. A home that cannot be written is not a failed project. `init` reports
   `pendiente:` for `bootstrap`, names the command that retries, and exits `0`,
   because the project half is done and the home belongs to the machine.
3. The persona key can be remapped in OpenCode (`switch_agent`). The generated
   line names Tab because that is the default; a line that named a binding odachi
   cannot read would be worse than a line that names the default.
4. `.agents/skills` travels with the repository, so in someone else's repo its
   content is untrusted input. The standard asks clients to gate project skills
   on trust before loading them, and that gate is the client's to enforce.
   `bootstrap` only explains the flow and never executes anything, which is the
   part odachi can guarantee; it cannot approve a file it did not write.

## Alternatives rejected

- **Derive the persona from the profile.** One word, two sources of truth again,
  and `init` would have to guess a profile that `start` has not written yet.
- **Keep `explain` as the only switch.** It lives in a file the model never
  reads and it cannot be changed mid-session; it is the defect, not the fix.
- **Write a global `AGENTS.md`.** No standard path, one copy per client, and a
  gate rule that is false in every folder that is not an odachi project.
- **`init --persona novato|senior` at install time.** It asks a question about
  the person instead of the project, and the answer is wrong the moment they
  change their mind. A file per persona plus the client's switch has no such
  moment.
- **One persona, and let the agent detect the level.** The block already says
  "no lo adivines"; a detected level is a guess written to disk.
- **Hand-written personas per client.** Nine copies of one voice, nine chances
  to drift, and the next client gets none. One template, one wrapper per client,
  and a table that cannot name a client the skills table does not know.
- **Put the persona body in `AGENTS.md` only.** It would apply whether or not the
  human wants it, and it cannot be switched inside a session.

## Consequences

- `internal/domain/persona` is pure and renders; its client table hangs off the
  names in `internal/domain/agent` and a test fails if it names a client the
  skills table does not have. Destination and frontmatter are per client because
  the mechanism is.
- `AGENTS.md` keeps its novice section, and `persona.Body(novato)` reuses those
  sentences: the tool may not say one thing in the block and another in the
  persona, and `internal/app/init/agents_block_internal_test.go` pins the match.
- The profile still carries `explain = why|none`: the schema is frozen at eight
  fields, so the field is written, validated, rendered and compared by
  `level --to`, and nothing consumes it to change what the tool does or how the
  agent talks. It is residue, not a second owner. Removing it is a schema change
  with its own unit.
- `init` grew from three artifacts to five plus two generated halves, and keeps
  its idempotence: a second run writes nothing —personas, skills and the
  `AGENTS.md` block included— and the summary says `sin cambios` instead of
  claiming work it did not do.
- One more thing the tool cannot see: whether a project's `.agents/skills` copy
  is the one its clients actually load, or a second copy an earlier explicit
  `pack` left somewhere else. `init` names the other copy and leaves both in
  place; it never deletes a directory it did not write.

## Amendment · the voice lives in the block (v0.11.0)

The decision above shipped in v0.10.0 with generated per-client files:
`novato` and `senior` as OpenCode primary agents and as Claude Code output
styles, so the human could pick one with the client's own switch.

Measured, it does not work, and the reference implementation never did it
that way:

- OpenCode 2.1.2 listed the files from `.opencode/agents/` in one run and
  not in the next, with the same files and a fresh service; `opencode debug
  agents` is not a reliable witness either.
- gentle-ai does not install agent files for OpenCode at all. Its capability
  manifest for that client declares `SlashCommands`, `Skills`, `SystemPrompt`
  and `MCP` — no `FileSubAgents` — its `SubAgentsDir()` and
  `EmbeddedSubAgentsDir()` return `""`, and its persona is delivered by
  replacing the instructions file (`~/.config/opencode/AGENTS.md`,
  `SystemPromptStrategy = StrategyFileReplace`).

So the channel changed, not the two axes. The profile still owns the
ceremony and the voice still owns how the agent talks. What changed:

- **One owner for the block.** `internal/domain/rules` renders the whole
  `AGENTS.md` block, with `## Cómo hablás` as its last section. `init` writes
  it, `start` and `level --to` rewrite only that section. A test pins the
  `novato` block byte for byte and asserts the two voices differ **only**
  inside the section.
- **No per-client files.** The client table, the wrappers and the generated
  headers are gone. A project that still has `.opencode/agents/odachi-*.md` or
  `.claude/output-styles/odachi-*.md` from v0.10.0 is told they are leftovers
  and nothing is deleted: they are the human's files.
- **The binary decides, the human does not guess.** `init` uses
  `.odachi/profile.toml` when it exists and `novato` otherwise — never
  inferred, because inference turned every new repository into `senior`.
  `level --to <perfil>` writes the profile and the section together.
- **Doctor compares instead of guessing.** Check 11 asks whether the block's
  voice matches the profile, and names `level --to` or `init` as the fix.
  Leftover persona files are reported as `ok`, not as a problem.

What this costs: the switch is no longer a keystroke. The voice is a decision
of the project, written by the binary, and it changes with `odachi level --to`.
That is the price of one channel every client reads instead of one file per
client that half of them ignore.
