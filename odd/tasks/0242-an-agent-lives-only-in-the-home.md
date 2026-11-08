# an-agent-lives-only-in-the-home
Base: d79d4226061a9e62fd1ba19c5b9b0ce37a68ba9d

## Requirements (verbatim, from the user)
R1 · "an agent lives only in the home: init writes no project agent and refreshes the door the machine keeps, doctor names a project agent an older release left behind, the bundled craft defaults to the home, and `odachi skill new` asks where a user skill lives"

## Done when
- `odachi init` writes no agent file inside the project and installs or refreshes the home door for the clients `--agents` names
- `odachi init --skills` defaults to home; `--skills project` is the explicit pin, and the block names the folder it uses
- `odachi skill new <slug>` without `--scope` fails asking for the decision; with `--scope project|home` it writes the skill under `.agents/skills` or `~/.agents/skills` and updates that root registry
- `doctor` names a project agent an older release left as a warning, and keeps its 15 checks
- nothing init wrote is deleted in silence: the removal is named and stays the human call
- go test ./... is green and odachi check --run "go test ./..." is green
- Unreleased names the change: a contract change, minor

## Steps
- [x] T1 · init writes no project agent and the craft defaults to the home   R1 · route:inline · tree:249a4570551aecc6cb8ff6209473a2ee0c7b1a81
- [x] T2 · skill new asks the scope and writes the skill where it belongs   R1 · route:inline · tree:d8e4fc7134d2b8493f569e71c35e2e7447748d6e
- [ ] T3 · doctor names a project agent an older release left   R1 · route:inline

## History
L1 · "an agent lives only in the home: init writes no project agent and refreshes the door the machine keeps, doctor names a project agent an older release left behind, the bundled craft defaults to the home, and `odachi skill new` asks where a user skill lives"
L2 · OFICIO · "The default craft scope is the machine, unless the project already carries .agents/skills" (porque a copy the human already has is a fact: the new default may not move a project that adopted odachi the old way, and a bare .agents/ belongs to other tools)
L3 · OFICIO · "init keeps --global accepted but out of the help" (porque an older script must not break, and offering a flag that changes nothing would be a lie, the rule 0106 used for --no-project-agents)
L4 · OFICIO · "The app keeps its Global and NoProjectAgents defaults; the CLI forces both" (porque library callers keep the meaning they had, and the product still has one behavior: an agent in the home)
L5 · OFICIO · "init names the folders an older release left and deletes nothing" (porque odachi never deletes a file it did not write in this version; the removal stays the human's call)
L6 · OFICIO · "doctor's advice stops offering --agents as the way back to a project mirror" (porque the flag no longer writes a project agent, and an advice must name a command that works)
L7 · OFICIO · "The leftover line names the files odachi generated, never their folders" (porque .opencode/agents and .claude/output-styles also hold a role's agent and the human's own styles; a folder is the wrong unit to offer for deletion)
L8 · OFICIO · "skill new writes .agents/skills, never the curriculum" (porque skills/ is what odachi ships and pack installs; a user's skill lives where the agent reads it)
L9 · OFICIO · "A user skill's registry lives inside its skills folder" (porque the folder a human opens is the whole thing; the curriculum keeps SKILLS.md at the project root for its own layout)
L10 · OFICIO · "skill new refuses a run without --scope" (porque a default would hide a decision the human has to make, and the old default was the wrong place)
L11 · OFICIO · "The CLI splits the slug from the flags itself" (porque Go's flag package stops at the first positional, and the documented form has the slug first)
