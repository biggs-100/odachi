# the-release-says

## Requirements (verbatim, from the user)
R1 · "the release says which source built it: each archive carries an attestation that binds it to this workflow, this commit and this tag"

## Done when
- gh attestation verify against a released archive succeeds for this repository
- the release workflow attests every archive listed in checksums.txt, with the attestation permission and the action pinned by sha
- the guard test fails when the attestation step or its permission disappears
- SECURITY and the README name the third link and what it adds over the signature

## Steps
- [ ] T1 · the release job attests every archive with actions/attest, and the guard pins the step and its permission   R1
- [ ] T2 · SECURITY, the README and the release config name the third link and what it adds   R1

## History
L1 · "the release says which source built it: each archive carries an attestation that binds it to this workflow, this commit and this tag"
L2 · OFICIO · "The attestation runs after goreleaser and reads dist/checksums.txt, not a glob of archives" (porque the checksums are the signed list: an archive the list does not name has no business carrying an attestation from this workflow)
