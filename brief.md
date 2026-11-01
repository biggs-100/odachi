# the-evidence-is-immutable

## What must exist
- the fence guards the bytes of a stored proof, not the code that writes them: the format may gain a versioned line, and no stored proof is ever rewritten

## Done when
- `odd/fence.toml` protects `proof/**` and stops naming `internal/domain/proof/**`
- `brief.md` says the same rule the fence carries, and it is a new non-goal, not a comment
- `check --fence-from <base>` does not call the change a loosening, and the ADR says why the wallet and the hand may differ

## Not included (first version)
- a stored proof is never rewritten, whatever the writer learns
- no change to the bundled skills

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: por que una cerca debe proteger la evidencia y no al que la escribe.
