# the-release-says

## What must exist
- the release says which source built it: each archive carries an attestation that binds it to this workflow, this commit and this tag

## Done when
- gh attestation verify against a released archive succeeds for this repository
- the release workflow attests every archive listed in checksums.txt, with the attestation permission and the action pinned by sha
- the guard test fails when the attestation step or its permission disappears
- SECURITY and the README name the third link and what it adds over the signature

## Not included (first version)
- own keys or a KMS: the workflow's identity again
- Apple notarization
- a Homebrew tap
- changing what the installer verifies
- attesting anything outside dist/checksums.txt

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: what provenance adds to a signature.
