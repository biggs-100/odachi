# A tag moves and a digest cannot

What you pin, what you scan and what you can prove about the bytes you deploy. Every
rule here is about the registry side of an artifact whose build lives in
`images-and-build.md`; the decision is always "which of these two claims do I need".

## A tag is a pointer, a digest is content

- **Pin by digest where the code is deployed**: a retagged image leaves "a mix of Pods running the old and new code" (kubernetes.io/docs/concepts/containers/images).
- **A moved tag may not pull at all**: `imagePullPolicy` defaults to `IfNotPresent` for any non-`latest` tag, so a node that already cached `:1.4` keeps the old bytes after `:1.4` moves (same source).
- **The policy names `non-latest` on purpose**: `latest` is the exception the default is written around, so keeping it for a deployed service is a policy you chose, not a default you inherited.
- **Decide the pull policy where you keep tags**: `Always` pays a registry round trip on every container start; `IfNotPresent` pays correctness when someone retags.
- **Record which of the two you chose**, because a cached moved tag and a digest can look identical in a `docker ps`.

## Immutability is the registry's feature, not the Dockerfile's

- **ECR `IMMUTABLE` applies to all tags**, with exceptions added Jul 2025 (docs.aws.amazon.com/AmazonECR/latest/userguide/image-tag-mutability.html).
- **So it is an all-or-nothing switch per repository**: it stops the retag that overwrote a release, and it also forbids a moving `:staging` beside a fixed `:1.4` in that same repository.
- **Immutability answers "can this tag change", never "is this the right build"**: two repositories with the same immutable tag still carry different bytes.

## A scanner reports inventory, not risk

- **Read the OS coverage matrix before you trust a zero**: Trivy's matrix is per distribution *and version*, and RHEL 10 is SBOM only, no vulnerability verdict (trivy.dev/latest/docs/coverage/os).
- **An unsupported base yields an SBOM with no CVE opinion**: the scan succeeds, the report comes back empty, and empty reads as clean.
- **A static scan sees package-manager inventory and nothing else**: it cannot see `dlopen`, a vendored binary, runtime config, or whether the CVE is reachable from your code.
- **Choose the base for coverage as well as size**: a base outside the matrix trades a smaller image for no verdict, and that trade belongs in writing.
- **Re-read the report as a list**, of packages and versions, because that is what it is; the sentence "no vulnerabilities" is an inference the tool did not make.

## A signature binds an identity to a digest

- **Keyless signing goes through Fulcio, and Rekor records it** (docs.sigstore.dev/cosign/verifying/verify).
- **`cosign verify` checks the digest claim by default against `--certificate-identity` and `--certificate-oidc-issuer`** (same source), so passing neither proves a signature existed, not who made it.
- **Pass both arguments or you verified nothing**: one alone accepts whatever the other dimension allows.
- **A signature is per-digest, so it does not survive a rebuild**: retagging keeps it, and a new build needs a new signature.
- **Notation/Notary Project v1.1.0 (2024)** is the vendor-neutral OCI signature spec, built on a trust store and a trust policy (notaryproject.dev/blog/2024/announcing-notation-v1-1).
- **Choose the format your gate already consumes**: a vendor-neutral signature that no admission controller reads is a file in a registry.

## Admission is where a signature becomes a gate

- **Verification has to happen at admission**, in OPA/Gatekeeper, Kyverno or sigstore policy-controller (kubernetes.io/docs/reference/access-authn-authz/admission-controllers).
- **Without a policy engine, signing is a dashboard**: it records intent and blocks nothing that deploys.
- **A policy you have not tested against an unsigned image is unproven**, so the gate earns its trust by rejecting one deliberately.

## Provenance says which builder produced it

- **SLSA v1.2 is current (2026); the L1-L3 ladder is from v1.0** (slsa.dev/spec/v1.0/levels).
- **L1 provenance exists and is forgeable**: it lists the steps, and anything inside the build can write it.
- **L2 is signed provenance from a hosted platform**, and it deters tampering *after* the build, not during it.
- **L3 is hardened isolation with a signing key the build steps cannot reach**, which is what deters tampering *during* the build, and it is the level that costs a platform.
- **Provenance is the builder's attestation**: "an attestation that a particular build platform produced a set of software artifacts through execution of the buildDefinition" (slsa.dev/spec/v1.0/provenance).
- **A self-hosted runner attesting to itself proves little**, because the attestation and the thing attested share one trust boundary.

## An SBOM asserts presence

- **CycloneDX 1.7 (2025-10-21, Ecma-424) is security-native**: VEX, vulnerabilities and formulation live in the document (cyclonedx.org).
- **SPDX 3.0 (2025-08) is license and graph first**, so it is the format when the consumer is legal review rather than a scanner (spdx.dev).
- **ISO/IEC 5962:2021 standardizes SPDX 2.2.1**, not 3.0: "we have an ISO SBOM" and "we emit SPDX 3.0" are two different claims.
- **An SBOM asserts presence, not exploitability**: it lists what shipped, and says nothing about whether a listed CVE can be reached.
- **Completeness is self-declared**: `compositions` can be `incomplete`, so a short SBOM is an assertion someone made, not a scan that missed a line.

## The registry is state you operate

- **Registry garbage collection is stop-the-world**: run it read-only or stopped, or in-flight pushes lose layers (distribution.github.io/distribution/about/garbage-collection).
- **`--delete-untagged` is opt-in**, and without it every untagged manifest keeps its blobs, so storage grows straight through a deletion policy.
- **A pull-through cache has one upstream per mirror**, and tag pulls revalidate against it (distribution docs, `recipes/mirror`).
- **Private credentials on the mirror expose that content**: the cache serves whatever it can authenticate to, to everyone who can reach the cache.
- **`delete` must be enabled for a mirror to shrink**, otherwise it is a second full copy of the upstream that only grows.
- **The daemon's `registry-mirrors` supports Docker Hub only**, so a general mirror is a proxy in front, not a daemon setting.

## Contested, and both sides are true

- **Scanner as a merge gate**: gate on newly introduced and fixable findings, because a gate on absolute counts fires on a base image nobody in this PR touched and gets disabled within a week.
- **SBOM as a gate**: requiring the file proves your build emits one; requiring a clean one claims a completeness no scanner can establish.
- **Keyless against long-lived keys**: keyless moves custody to the OIDC identity and needs the network to verify; a long-lived key verifies offline and puts the key on you to hold. Decide by who holds the secret and whether the verifier has a network.
