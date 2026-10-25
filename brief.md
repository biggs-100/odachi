# the-release-proves

## What must exist
- the release proves where it came from: the checksums are signed without keys, an SBOM shows what went into each archive, and the docs say how to verify

## Done when
- the release publishes checksums.txt.sigstore.json beside the archives, the cosign v3 bundle that holds the signature and its certificate
- the release publishes one SBOM per archive
- the documented cosign verify-blob --bundle command succeeds against the published checksums and identity
- the installer still verifies sha256 only: cosign is not a requirement to install
- the goreleaser comments no longer claim there are no keys or no LICENSE

## Not included (first version)
- own keys or a KMS: keyless only, with the workflow as the identity
- Apple notarization
- a Homebrew tap
- changing what the installer verifies
- an SBOM of the source tree: only the archives

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: what a signature proves that a checksum does not.
