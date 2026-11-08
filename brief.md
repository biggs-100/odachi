# an-agent-lives-only-in-the-home

## What must exist
- an agent lives only in the home: init writes no project agent and refreshes the door the machine keeps, doctor names a project agent an older release left behind, the bundled craft defaults to the home, and `odachi skill new` asks where a user skill lives

## Done when
- `odachi init` writes no agent file inside the project and installs or refreshes the home door for the clients `--agents` names
- `odachi init --skills` defaults to home; `--skills project` is the explicit pin, and the block names the folder it uses
- `odachi skill new <slug>` without `--scope` fails asking for the decision; with `--scope project|home` it writes the skill under `.agents/skills` or `~/.agents/skills` and updates that root registry
- `doctor` names a project agent an older release left as a warning, and keeps its 15 checks
- nothing init wrote is deleted in silence: the removal is named and stays the human call
- go test ./... is green and odachi check --run "go test ./..." is green
- Unreleased names the change: a contract change, minor

## Not included (first version)
- no change to the bundled skill texts: the 19 keep their words
- no change to the review or the classifier
- the doctor keeps its 15 checks: it gains a warning that names a project agent an older release left, and loses none
- no deletion of an older project agent files: init stops writing and doctor names them
- no change to the worker refusal or the door text
- no change to odachi bench and no new verb
- no change to the brief format or to start
- no change to the stored proofs: never touched
- no change to the design canvas, its token or the installers
- no change to the release workflow

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: this week.
- Wants to learn: whether a single global agent removes the per-project refresh, read from what a fresh init writes.
