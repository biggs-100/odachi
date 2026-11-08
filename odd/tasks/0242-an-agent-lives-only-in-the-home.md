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
- [ ] T1 · init writes no project agent and the craft defaults to the home   R1 · route:inline
- [ ] T2 · skill new asks the scope and writes the skill where it belongs   R1 · route:inline
- [ ] T3 · doctor names a project agent an older release left   R1 · route:inline

## History
L1 · "an agent lives only in the home: init writes no project agent and refreshes the door the machine keeps, doctor names a project agent an older release left behind, the bundled craft defaults to the home, and `odachi skill new` asks where a user skill lives"
