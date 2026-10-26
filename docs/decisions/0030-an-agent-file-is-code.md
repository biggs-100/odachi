# 0030 - An agent file is code

Status: accepted · Date: 2026-10 · Session: v0.11.27

## Context

The review decided which files are the tool's own text by a prefix: any `.md`
under `.opencode/agents` or `.claude/agents` whose name started with `odachi` was
treated as generated plumbing and left out of the candidate. The comment above the
function already said why that directory must stay: agent privileges and secrets
live there. The code did not keep the promise. `odachi-helper.md` -a file the human
wrote, granting tools or carrying a token- disappeared from the candidate and
never reached a lens.

Reading the classifier while planning the fix found the second half: every `.md`
under those directories was also classified as documentation, so a human agent
file with no `odachi` prefix (`helper.md`) stayed passive for a different reason.
Two rules were hiding the same class of file.

## Decision

- **The names odachi writes are a closed set.** `persona.AgentNames` derives it
  from the writers: the primary mirror, the three subagents and the two leftovers
  of v0.10.0. The review asks for exactly those names; it no longer knows a
  prefix.
- **Everything else under an agent directory is the human's.** `.opencode/agents`,
  `.claude/agents` and `.claude/output-styles` hold instructions an agent obeys, so
  their Markdown is code and reaches a lens. The extension decides for every other
  directory.
- **The default is human.** A name the tool does not write is reviewed, even when
  it starts with `odachi`; a name it writes is not, even if the file was edited by
  hand -the note inside says a hand edit is lost work, and reviewing it would
  charge every project for the tool's own prompt text.

## Alternatives rejected

- **Prefix matching.** It is the defect: `odachi-helper.md` is not odachi's.
- **Exempting the whole directory.** The opposite hole, and the comment said so:
  tool grants and secrets live exactly there.
- **A second literal list in the review.** Two lists drift; `persona` writes the
  files today and will write the next subagent tomorrow, so it owns the names.
- **Reading the generated note inside the file.** The tier is decided by paths
  alone: a content rule would need the diff of every file, would change with a
  comment, and would let a hand-edited copy pass as generated.
- **Treating every Markdown as code.** It would ask for a lens on every README in
  the repository, which is the noise this classifier exists to avoid.

## Consequences

- A human agent file asks for at least the reliability lens; a name that carries a
  signal of its own (`worker`, `permission`) raises the tier as any path does.
- Adding a subagent to `persona` adds its name to the set the day it exists, and a
  test pins the list so a rename is a decision, not a silent hole.
- The v0.10.0 leftovers under `.claude/output-styles` count as plumbing like the
  rest of odachi's own text; before, only `odachi.md` did.
- Documents outside the agent directories are unchanged: the existing test still
  pins a nested `AGENTS.md` as documentation.

## Sources

- The carried review finding on `odachiAgentFile`, left open by feature 0055.
- `isDocumentation` and the promise in the comment it contradicted.
