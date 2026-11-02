# the-installer-refuses
Base: d20f49bf5d02773ec290409c1c286b6d5aba1e5c

## Requirements (verbatim, from the user)
R1 · "the installer refuses bytes it cannot verify against the release signature"

## Done when
- both installers download checksums.txt with its sigstore bundle and refuse to install when the signature does not verify
- the refusal names cosign and how to install it, and nothing is written into the bin directory
- the installer test proves the refusal with a failing cosign and the happy path with a passing one
- the docs say the installer now requires cosign

## Steps
- [ ] T1 · both installers verify the sigstore bundle before they install   R1 · route:inline
- [ ] T2 · the docs say the installer needs cosign   R1 · route:inline

## History
L1 · "the installer refuses bytes it cannot verify against the release signature"
L2 · OFICIO · "the installer requires cosign instead of verifying only when it is present" (porque a check that runs sometimes is a silent degradation, and this repository already refuses instead of degrading)
