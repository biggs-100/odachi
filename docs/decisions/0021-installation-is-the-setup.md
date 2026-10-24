# 0021 — Installation is the setup

Status: accepted · Date: 2026-10 · Session: v0.11.8

## Context

ADR 0012 made the home door opt-in: a primary agent in every folder hijacked
the selector of projects that never asked for it, so `init` stopped writing the
human's home and `--global` became the explicit door. ADR 0020 completed that
path: `--global` also makes the door OpenCode's default agent.

The human then asked the question the flow did not answer: if installing odachi
is the only thing they did, why are they the one typing `odachi init` and
`odachi init --global`? Two gaps caused it.

1. The installer only copied the binary. On a fresh machine there was no
   bootstrap skill and no door, so no agent knew odachi existed.
2. The door told the agent to *guide* the human: "first `odachi init`, then
   `odachi start`". The agent had the knowledge and handed the work back.

The product's promise is a binary the agent operates. The human decides the
vision; the agent runs the verbs.

## Decision

- **`odachi setup` installs the global half and nothing else.** The bootstrap
  skill in `~/.agents/skills`, the home door for the agent clients this machine
  already has, and OpenCode's `default_agent`. It writes nothing in the folder it
  runs in, so it can run from anywhere.
- **The installer runs it.** `install.sh` and `install.ps1` call the binary they
  just installed. The human installs odachi and opens the agent; the first
  conversation is the setup.
- **Setup detects the clients present.** A client whose config directory is not
  in the home gets no door: an installer does not create another tool's folder.
  `--agents` is the explicit override.
- **The door and the bootstrap skill instruct the agent to run the verbs.** The
  door says "set it up yourself: run `odachi init`, then `odachi start` and ask
  the human its questions. Never ask the human to type these commands." The
  bootstrap skill says "You run them; the human never types them."
- **Permission stays where it was.** In a repository that is not the human's,
  the agent asks in one sentence and waits for a yes. The human answers in
  words; the permission is about writing, not about typing.
- **Setup is idempotent.** A machine already set up is left as it is: the door
  and the config are compared, not rewritten, and the line says so.
- **This is the reversal of ADR 0012, for the installer path only.** The door
  stays opt-in for `init` in a project. What changed is the moment of consent:
  running the odachi installer is the human saying "put odachi on this machine".
  A plain `init` in a repository still writes nothing in the home.

## Alternatives rejected

- **The installer running `init --global` in a temporary folder.** It writes the
  project half too, prints project lines, and depends on a temp directory being
  disposable. A verb that says what it does is cheaper than a trick.
- **The installer copying the skill and the door itself.** The payload is
  embedded in the binary. A shell script with a second copy is a second source
  that drifts on the next release.
- **Keeping "guide the human".** It was the defect: the agent had the skill and
  still asked the human to type. The knowledge and the action belong together.
- **Installing the door in `init` again.** Without `--global`, initializing a
  project is not the moment the human chose an agent for every folder.
- **A silent global default with no door.** OpenCode would start with an agent
  that is not installed. Setup writes the door first, then the default.

## Consequences

- A new machine: install, open the agent, describe the project. The agent runs
  `init` (permission in a foreign repo), then `start`, and the human answers the
  interview in natural language.
- A machine without OpenCode or Claude gets the bootstrap skill and one honest
  line; no folder is created for a client that is not there.
- The installer's success is no longer "the binary exists": it reports the
  setup's lines, and a failed setup says so without uninstalling the binary.
- `init` without `--global` still leaves the home untouched.
- `odachi doctor` reads the result (door, config) and names the drift.

## Sources

- ADR 0012 (the door is opt-in), ADR 0020 (the door is the default agent).
- The human's requirement: "el humano no debe escribir esos comandos; el agente
  ya lo debe saber usar, para que el humano solo instale odachi y lo empiece a
  usar".
