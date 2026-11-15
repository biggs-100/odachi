---
name: bootstrap
description: Trigger: when someone wants to start a project from scratch, asks what odachi is, or does not know where to begin. Do not load it inside a project that is already odachi.
---

## When

When someone wants to start a project from scratch, asks what odachi is, or does
not know where to begin. Also when the folder is empty and the first step has to
be decided.

## Rules

- **You run them; the human never types them.** A project is not set up until the first proof exists.
- `odachi init` leaves the project ready: the rules block the agent reads, the hook that stops a commit without proof, the `.gitignore`, and the skills.
- It creates the git repository when the folder has none.
- The skills live in `.agents/skills` (this project's copy, the default) or `~/.agents/skills` (`--skills home`); `odachi doctor` says which, and the block names it.
- `odachi start` is the interview that writes `brief.md` (the vision and what is NOT included), the fence and the first task doc.
- The answers are the human's: ask them in chat, then pass their own words with `--outcome`, `--audience`, `--done`, `--exclude`, `--sensitive`, `--deadline` and `--learn`, or a draft with `--brief`.
- A tool shell has no terminal: `start` exits 2 naming the missing flags, and that is the prompt to ask, not a reason to invent the vision.
- Load `drive` next: the day-to-day manual, one step planned, test-first, one commit at a time.
- Close with `odachi check --run "<command>"`: the gate runs the proof and writes the real output in `proof/`, and a green proof belongs to the tree it ran on.
- A project is already odachi if there is `brief.md`, an odachi block in `AGENTS.md`, or `odd/tasks/`: read that block and follow its rules.
- `odachi doctor` says the state without writing anything; in a folder that is not a project yet, it names this cycle.
- The human decides the vision; you execute and explain the why in one line.
- The rules that keep a shape alive already have an owner: `drive` owns one owner per rule and the fence, `tests` the guard test and the mutation, `commits` the risk line. Read them there; this door says what to build.

## Decision gates

| Situation | What to read |
| --- | --- |
| The first file is next, whatever the stack | Load `arch-project` after this one: the kind of project, the twelve decisions, the guard and the UI contract are its door |
| The stack is Tauri v2 | Load `arch-tauri` after this one: that stack's boundary is its own door |

## Output

The next command and why, in this order: `init`, `start`, `drive`, the proof. A folder that is not a project yet returns the cycle it needs and the state it is in, never a file nobody asked for.

## Never

- Never invent the interview's answers, the exclusions or the done criteria: an assumption is not the human's vision. Ask in chat and wait.
- Never call the setup done before the first proof exists: a project that has never run `check --run` has not proved anything.
- Run `odachi check`, `odachi review` or the hook in a folder that is not an odachi project: gates on somebody else's work say nothing.
- Run `odachi init` in a repository that is not the human's without asking: it writes in the project and in their home. Ask in one sentence and wait for a yes.
- Start over a project that already exists, or rewrite somebody else's `brief.md`.
- Never carry a reference of its own: the shape of the code belongs to `arch-project`, and a file here would be a second owner of the same material.
