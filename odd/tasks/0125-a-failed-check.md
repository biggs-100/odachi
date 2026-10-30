# a-failed-check
Base: e71f8992ce75c36cf0faff717d96db95ec77091e

## Requirements (verbatim, from the user)
R1 · "the footprint reads a failed check as not there: a stat that fails, an AGENTS.md that cannot be read, or a git that refuses the repository flips a project that does carry its craft into the full install, or the other way"

## Done when
- a plain init in a repository that already lives whose .agents/skills cannot be checked exits 1, names the path, and writes nothing
- the same for the client's agent file: the agent half is never decided from a failed check
- a .git that git cannot read makes init exit 1 and say so, instead of choosing the full install from no commit
- a repository with no commit keeps today's answer: the full install
- odachi role with an AGENTS.md it cannot read is refused by name instead of writing the default block

## Steps
- [ ] T1 · init refuses a footprint decided from a failed check, naming the path, and the rule is said once   R1
- [ ] T2 · the CLI tells no commit apart from a repository git cannot read   R1
- [ ] T3 · role refuses an AGENTS.md it cannot read instead of writing the default block   R1

## History
L1 · "the footprint reads a failed check as not there: a stat that fails, an AGENTS.md that cannot be read, or a git that refuses the repository flips a project that does carry its craft into the full install, or the other way"
