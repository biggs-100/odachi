# If you delegate

Inside an odachi project, this is the delegation policy: if another global skill
says otherwise, this one wins.

A subagent starts without cache: **it does not save tokens, it protects your context.**

- Budget: more than ~3 reads, ~10k tokens or a long session sends the next unit
  to a subagent; what is read a lot gets delegated, what is written little does
  not. Delegate for a reason, never for size.
- A **writer** is delegated for one reason only: two or more independent units
  with disjoint surfaces, each heavier than starting a subagent; or ~150k tokens
  of session. Never for size, files or price; without a reason, write it yourself.
- One writer per unit; in parallel they do not share files and you own git and the
  final check. If you delegate the writing, delegate the reading that prepares it.
- The request carries the **task doc path and the step**, not a summary; also the
  **edit surfaces**, the **verification commands** and the **exact skill paths**
  (`## Skills to load before work`).
- To write, use **`odachi-worker`**; to verify, **`odachi-verify`** (read-only).
  The worker does not commit or delegate; the parent keeps git and the gate.
- A change that names one of the six risks in `commits` gets a verifier even when
  you wrote it inline: `odachi-verify` reads the task doc and answers per
  requirement.
- The return is bounded: **=2k tokens**, `path:line` evidence and the list of files
  read; the rest stays inside.
- Ask it to record with `odachi plan` and not to say "done" without
  `odachi check --run`.
- When it returns, verify its output against `brief.md` and the diff, **not
  against its summary**: the subagent can lie like the main agent.
- **Never delegate the vision decision.** The subagent executes; the human decides.
