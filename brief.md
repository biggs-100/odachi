# the-installer-refuses

## What must exist
- the installer refuses bytes it cannot verify against the release signature

## Done when
- both installers download checksums.txt with its sigstore bundle and refuse to install when the signature does not verify
- the refusal names cosign and how to install it, and nothing is written into the bin directory
- the installer test proves the refusal with a failing cosign and the happy path with a passing one
- the docs say the installer now requires cosign

## Not included (first version)
- the release workflow and the signing stay as they are: nobody re-signs by hand
- a stored proof is never rewritten, whatever the writer learns

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why a chain of trust must end in a signature and not in a host.
