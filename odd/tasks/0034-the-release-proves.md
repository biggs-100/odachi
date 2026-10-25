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
- [ ] T1 · release.yml signs checksums.txt keyless and publishes the signature and certificate   R1
- [ ] T2 · the release publishes one SBOM per archive   R1
- [ ] T3 · README and SECURITY show how to verify, and the goreleaser comments tell the truth about keys and LICENSE   R1

## History
L1 · "the release proves where it came from: the checksums are signed without keys, an SBOM shows what went into each archive, and the docs say how to verify"
L2 · PREGUNTA P1 · "cosign v3 replaced the .sig + .pem pair with one bundle (.sigstore.json): may the release publish the bundle as the signature asset instead of the pair the brief names?"
L3 · VOS P1 · "yes: publish the cosign v3 bundle (checksums.txt.sigstore.json) as the signature asset; do not pin cosign v2 for the pair"
L4 · OFICIO · "T1 keeps the words 'signature and certificate' although the asset is one bundle" (porque the cosign v3 bundle carries both, and plan is the only writer of a step's text)
