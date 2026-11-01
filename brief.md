# check-ci-runs

## What must exist
- check --ci runs the gate exactly as CI runs it: the fence read from the remote base and the base resolved from the remote, so the policy lands before the feature is verifiable before a push instead of after it

## Done when
- a candidate the base fence forbids passes --fence-from HEAD and is blocked by --ci, and both were observed
- check --ci beside --base or --fence-from is refused, naming that --ci is what they are for
- a clone with no remote base blocks and names the command that creates it

## Not included (first version)
- no fetch and no network: the gate stays offline and names the base it read
- no change to .github/workflows/gate.yml

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: por que el gate debe leer su politica de una referencia que el candidato no puede editar.
