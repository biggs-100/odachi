# the-release-says

## Requirements (verbatim, from the user)
R1 · "the release says which source built it: each archive carries an attestation that binds it to this workflow, this commit and this tag"

## Done when
- gh attestation verify against a released archive succeeds for this repository
- the release workflow attests every archive listed in checksums.txt, with the attestation permission and the action pinned by sha
- the guard test fails when the attestation step or its permission disappears
- SECURITY and the README name the third link and what it adds over the signature

## Steps
- [x] T1 · the release job attests every archive with actions/attest, and the guard pins the step and its permission   R1 · 65aabba
- [x] T2 · SECURITY, the README and the release config name the third link and what it adds   R1 · 0ca24c9

## History
L1 · "the release says which source built it: each archive carries an attestation that binds it to this workflow, this commit and this tag"
L2 · OFICIO · "The attestation runs after goreleaser and reads dist/checksums.txt, not a glob of archives" (porque the checksums are the signed list: an archive the list does not name has no business carrying an attestation from this workflow)
L3 · OFICIO · "The docs name v0.11.12 as the first release with provenance" (porque a reader on v0.11.11 would run gh attestation verify against a file that has no attestation yet, and conclude the command is broken)
L4 · OFICIO · "The attestation keeps its place after publishing, and gains a retry job instead of a reorder" (porque goreleaser has no publish-only command, so build-attest-publish would duplicate its publishing for nothing; a dispatch that attests the published bytes fixes the failure without touching the release)
L5 · OFICIO · "The publishing job is push-only and the guard reads each job on its own" (porque a dispatch reached goreleaser too, so the recovery would have published a second time while the retry attested a checksums file the other job was rewriting)
