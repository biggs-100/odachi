# 0020 — The global door is the default agent, and the human's file is not rewritten

Status: accepted · Date: 2026-10 · Session: v0.11.7

## Context

ADR 0012 made the home door opt-in: a plain `init` does not put an agent in the
human's home, because a primary agent in every folder hijacked the selector of
projects that never asked for it. `--global` installs the door in
`~/.config/opencode/agents/odachi.md`.

The door was installed but not chosen. OpenCode starts a session with
`default_agent`, and the generated agent only appears in the Tab selector until
someone picks it. The human asked for the missing half: installing the door
globally should make it the agent a session starts with.

The file is the human's: it carries their MCP servers, their comments, their key
order, their line endings. Writing it is not like writing a generated persona.

## Decision

- **`init --global` writes the default.** When the opencode client is selected
  and the door is installed, the global config gets `default_agent: "odachi"`.
  It is the same opt-in decision: the flag that puts odachi in every folder also
  chooses it.
- **Another default is never overwritten.** The outcome names the agent and the
  file: installing a door is not hijacking a selector. There is no force flag in
  the first version; the human edits the file if they want to change their own
  choice.
- **The edit inserts one member; it never re-encodes the document.** Comments,
  key order and CRLF survive. The insertion goes right after the root brace, with
  the file's own line ending, and a comma only when the object already has
  members.
- **JSONC is a first-class citizen.** OpenCode documents `opencode.json(c)`.
  Comments are stripped for validation only; the written bytes keep them. A
  `.jsonc` file is edited when it is the one that exists; a project with neither
  gets a `.json` carrying the published `$schema` and the agent.
- **A file that cannot be read safely is refused, not repaired.** A document
  that is not a JSON object, a duplicate `default_agent`, a non-string value or
  a strict-invalid document (a trailing comma, a missing brace) leaves the file
  byte-for-byte and is reported as pending, with the exact member to add by
  hand. The project half of `init` still succeeds: the door is installed, one
  convenience is missing.
- **Doctor reads, never writes.** When the door is installed and the config does
  not choose it, check 11 says so and names `odachi init --global`; the human's
  file is not touched by a question.
- **The project-level default is out of scope.** This feature is the global
  install. A project's `.opencode/opencode.json` is the project's decision.

## Alternatives rejected

- **Re-encode with `encoding/json`.** It loses comments and order, and the human
  reads that file. An edit of one key may not rewrite the document around it.
- **Refuse JSONC.** The format OpenCode publishes is JSONC; refusing it would
  fail exactly the configs the docs invite.
- **Overwrite another default.** The human picked it. The door is an offer, and
  the line names what stands.
- **Do it from the installer.** The installer is machine-wide and has no client
  selection, no project, and no opt-in. `--global` is the moment the human says
  "every folder".
- **A `--force` flag now.** No human has needed it yet; adding it before the
  first case is ceremony. The file is one edit away.

## Consequences

- After `odachi init --global`, new OpenCode sessions start as odachi without
  pressing Tab. The current session keeps the agent it was created with.
- On a machine where another `default_agent` is set, nothing changes and the
  line says which agent is in charge and where the file is.
- `init` without `--global` writes nothing in the home, config included.
- A malformed global config becomes a pending line, not a failed install: the
  door is there, the default is one hand edit away.
- `doctor` can report the drift, and the fix is the same command that installed
  it.

## Sources

- OpenCode's configuration guide: `default_agent` "chooses the primary agent
  used when a session does not select one explicitly"; project and global
  `opencode.json(c)` are both discovered.
- ADR 0012 (the door is opt-in), feature 0011 (client selection).
