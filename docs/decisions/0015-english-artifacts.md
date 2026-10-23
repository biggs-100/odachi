# 0015 — English artifacts, docs and CLI; the chat in the human's language

Status: accepted · Date: 2026-10 · Session: v0.11.2

## Context

Odachi grew with two languages mixed: the skills, the brief, the task doc, the
`AGENTS.md` block, the README, the ADRs and the CLI messages were Spanish, while
the code, the domain and most of the impact map were English. Gentle-ai's
standard is the opposite: everything the tool writes is English, and only the
agent chat follows the human's language.

The human asked for that standard. The migration is not a find-and-replace: the
skill format was `Cuándo / Reglas / Nunca`, the brief and task doc had Spanish
headers, the rules block and the voices were Spanish, and tests pinned those
strings.

## Decision

- **English for everything the tool writes**: skills, brief, task doc, `AGENTS.md`
  block, native agents, subagent prompts, review lens prompts, docs and CLI
  messages.
- **Spanish only in the chat**: the agent answers the human in the human's
  language. The voice says so explicitly.
- **Human quotes stay verbatim** in their language inside the English artifact. A
  requirement quoted from the human is not translated; the prose around it is
  English.
- **The skill format is `When / Rules / Never`.** The validator enforces the
  English sections; the 60-line cap and the `Trigger:` prefix stay.
- **Old artifacts stay readable.** The brief and task-doc parsers accept the
  Spanish headers as read-only aliases; `## Cómo hablás` is an alias for
  `## How you speak`; a task doc written before the migration is still visible to
  `summary`, `plan` and `level`.
- **History stays history.** Closed task docs, old CHANGELOG entries and ADRs
  0001–0012 keep their original language. Living docs —README, impact map, the
  current CHANGELOG section and new ADRs— are English.
- **The CLI is English too**, including help, errors and tests. This is a
  breaking change for scripts that pinned the Spanish messages.

## Consequences

- New projects are born in English; old ones keep reading their Spanish artifacts.
- Custom skills written with the Spanish sections stop validating: the format is
  a contract, and the migration changes it deliberately.
- The README, impact map and new ADRs describe the English standard; the
  historical record keeps the language it was written in.
- The agent still talks to the human in Spanish (or whatever language the human
  uses), because the voice is a rule and the chat is not an artifact.
- The migration is split into steps: skills and format, artifacts, `AGENTS.md` and
  voices, docs, domain/app messages, `cmd` messages, review lenses and dogfood.
