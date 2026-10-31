# the-destination-that
Base: f2ff5ec4cb50636aabd7b78f66e5961cc12b4f5b

## Requirements (verbatim, from the user)
R1 · "initSkills treats a failed home destination as pending: and exit 0, because before the adoption the project still had its copy. Under the minimal footprint the home is the only destination and the block is written before the skills half runs, so a home that refuses the write leaves a green run with a block naming a folder nothing created -the state 0123 removed. A failure there has to fail the run, which changes the rule for every verb that installs in the home"

## Done when
- a run whose chosen scope is the machine exits 1 when the home refuses the write, and says the run failed instead of calling it pending
- a project-scope run that wrote its own copy keeps exit 0 when the home half fails, and still names what is pending
- odachi setup, whose only destination is the home, exits 1 when that write fails, and names odachi setup as the command that retries it
- the refusal names the file, the reason, and the verb that ran
- a home odachi cannot resolve at all is a failure of the chosen destination too, not a destination that quietly vanished

## Steps
- [x] T1 · the scope the run chose is the destination that must be written, and a failure there fails the run naming the verb   R1 · tree:5816b99cc1b9d89714eb333c846864d978779b03
- [ ] T2 · the ADR and the changelog say which failure fails the run   R1

## History
L1 · "initSkills treats a failed home destination as pending: and exit 0, because before the adoption the project still had its copy. Under the minimal footprint the home is the only destination and the block is written before the skills half runs, so a home that refuses the write leaves a green run with a block naming a folder nothing created -the state 0123 removed. A failure there has to fail the run, which changes the rule for every verb that installs in the home"
L2 · OFICIO · "fatality follows the scope the run asked for, not the folder that failed" (porque under the minimal footprint the machine is the only destination, so a home that refuses the write used to leave a green run naming a folder nothing created)
