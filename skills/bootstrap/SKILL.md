---
name: bootstrap
description: Trigger: when someone wants to start a project from scratch, asks what odachi is, or does not know where to begin. Do not load it inside a project that is already odachi.
---

odachi is a binary and four plain-text artifacts -brief, task doc, proof and
profile-. The binary enforces the gate; the artifacts are the memory. The human
decides the vision and what is non-negotiable; the agent executes.

## When

When someone wants to start a project from scratch, asks what odachi is, or does
not know where to begin. Also when the folder is empty and the first step has to
be decided.

## Rules

One cycle, in this order. **You run them; the human never types them.**
The project is not set up until the first proof exists:

1. `odachi init` - leaves the project ready: the rules block the agent reads, the
   hook that stops a commit without proof, the `.gitignore`, and the skills. It
   creates the git repository when the folder has none. The skills live in
   `.agents/skills` (this project's copy, the default) or `~/.agents/skills` (one
   copy for the machine, the `--skills home` choice); `odachi doctor` says which,
   and the block names it.
2. `odachi start` - the interview that writes `brief.md` (the vision and what is
   NOT included), the fence and the first task doc. The answers are the human's:
   ask them in chat, then pass their own words with `--outcome`, `--audience`,
   `--done`, `--exclude`, `--sensitive`, `--deadline` and `--learn` (or a draft
   with `--brief`). A tool shell has no terminal, so `start` exits 2 naming the
   missing flags; that is the prompt to ask, not a reason to invent the vision.
3. Load `drive` - the day-to-day manual: plan one step, work test-first, commit
   one step at a time.
4. Close with `odachi check --run "<command>"` - the gate runs the proof and
   writes the real output in `proof/`. `odachi check` is the gate; a green proof
   belongs to the tree it ran on.

- A project is already odachi if there is `brief.md`, an odachi block in
  `AGENTS.md`, or `odd/tasks/`. If there is one, do not start over: read that
  block and follow its rules.
- If you do not know the state, `odachi doctor` says it without writing anything;
  in a folder that is not a project yet, it names this cycle.
- The human decides the vision. You execute and explain the why in one line.

## Output

The next command and why, in this order: `init`, `start`, `drive`, the proof. A folder that is not a project yet returns the cycle it needs and the state it is in, never a file nobody asked for.

## Never

- Never invent the interview's answers, the exclusions or the done criteria: an
  assumption is not the human's vision. Ask in chat and wait.
- Never call the setup done before the first proof exists: a project that has
  never run `check --run` has not proved anything.
- Run `odachi check`, `odachi review` or the hook in a folder that is not an
  odachi project: gates on somebody else's work say nothing.
- Run `odachi init` in a repository that is not the human's without asking: it
  writes in the project and in their home. Ask in one sentence and wait for a
  yes; the human answers in words, never by typing the command.
- Start over a project that already exists, or rewrite somebody else's
  `brief.md`.
