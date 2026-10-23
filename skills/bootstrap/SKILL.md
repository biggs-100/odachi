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

Three commands, in this order. Each one does one thing and leaves it done:

1. `odachi init` - leaves the project ready: the rules block the agent reads, the
   hook that stops a commit without proof, the `.gitignore` that keeps the tree
   clean and the skills in `.agents/skills`. This skill, the entry point, is
   already in `~/.agents/skills`.
2. `odachi start` - the interview that writes `brief.md` (the vision and what is
   NOT included), the fence and the first task doc. It is what gives the work a
   direction.
3. `odachi check` - says whether it is ready: brief, fence, open questions and
   proof. With `--run "<command>"` it runs the proof and saves it.

- A project is already odachi if there is `brief.md`, an odachi block in
  `AGENTS.md`, or `odd/tasks/`. If there is one, do not start over: read that
  block and follow its rules.
- If you do not know the state, `odachi doctor` says it without writing anything.
- Once the project is set up, the day-to-day manual is the `drive` skill.
- The human decides the vision. You execute and explain the why in one line.

## Never

- Run `odachi check`, `odachi review` or the hook in a folder that is not an
  odachi project: gates on somebody else's work say nothing.
- Run `odachi init` in a repository that is not the human's: it writes in the
  project and in their home.
- Start over a project that already exists, or rewrite somebody else's
  `brief.md`.
