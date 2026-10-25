# the-release-proves

## Requirements (verbatim, from the user)
R1 · "the release proves where it came from: the checksums are signed without keys, an SBOM shows what went into each archive, and the docs say how to verify"

## Done when
- the release publishes checksums.txt.sigstore.json beside the archives, the cosign v3 bundle that holds the signature and its certificate
- the release publishes one SBOM per archive
- the documented cosign verify-blob --bundle command succeeds against the published checksums and identity
- the installer still verifies sha256 only: cosign is not a requirement to install
- the goreleaser comments no longer claim there are no keys or no LICENSE

## Steps
- [x] T1 · release.yml signs checksums.txt keyless and publishes the signature and certificate   R1 · a3a5ec8
- [x] T2 · the release publishes one SBOM per archive   R1 · d191a02
- [ ] T3 · README and SECURITY show how to verify, and the goreleaser comments tell the truth about keys and LICENSE   R1

## History
L1 · "the release proves where it came from: the checksums are signed without keys, an SBOM shows what went into each archive, and the docs say how to verify"
L2 · PREGUNTA P1 · "cosign v3 replaced the .sig + .pem pair with one bundle (.sigstore.json): may the release publish the bundle as the signature asset instead of the pair the brief names?"
L3 · VOS P1 · "yes: publish the cosign v3 bundle (checksums.txt.sigstore.json) as the signature asset; do not pin cosign v2 for the pair"
L4 · OFICIO · "T1 keeps the words 'signature and certificate' although the asset is one bundle" (porque the cosign v3 bundle carries both, and plan is the only writer of a step's text)
L5 · OFICIO · "The signing contract is pinned by a test that reads the two config files" (porque no Go code reads them: a later edit that drops the signature leaves no red test until a release publishes unsigned checksums)
L6 · OFICIO · "The review runs once on the feature's final tree, not before each delivery commit" (porque a verdict binds to a tree and the next commit invalidates it (0027); every commit still names Risk: delivery so the review cannot be skipped)
L7 · OFICIO · "The .goreleaser.yaml header comment lands whole in T3" (porque it is one surface and one commit: half-editing it in T1 and T2 would make the file contradict itself twice)
L8 · OFICIO · "The SBOM catalogues the archive, not the binary inside it" (porque the archive is what a person downloads and what the signature covers; an SBOM of a file nobody downloads describes the wrong thing)
L9 · OFICIO · "The release config's header is rewritten in English and lists what the file does not publish" (porque artifacts are English since 0015, and a header written before signing describes a file that no longer exists)
L10 · OFICIO · "The verify command names the tag and the OIDC issuer, not just the bundle" (porque the bundle alone proves that somebody signed; the certificate identity proves this workflow, from this tag, signed)
L11 · OFICIO · "The docs name the first signed release (v0.11.11)" (porque a verification command with a placeholder cannot be run, and an unrunnable command is a decoration)
