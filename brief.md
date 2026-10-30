# doctor-compares-the

## What must exist
- doctor compares the block even when the profile is inferred: a project that ran init and never ran start still learns that its block is older than the binary

## Done when
- a project with no odd/profile.toml and an older block gets the warning and names odachi init
- a project with no profile and the current block stays ok
- the tests that already exist keep their answer, and one of them fails without the change

## Not included (first version)
- no change to the verdict shape
- no change to the proof format, and the stored proof is never rewritten

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: por qué un skip silencioso se descubre solo al probar.
