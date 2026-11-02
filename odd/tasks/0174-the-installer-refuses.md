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
- [x] T1 · both installers verify the sigstore bundle before they install   R1 · route:inline · tree:98c25c27f41102bf9c2050b038db1e2b040f9e0e
- [x] T2 · the docs say the installer needs cosign   R1 · route:inline · tree:cce285225156bef848904e9483f99751319cebf6
- [x] T3 · the windows installer verifies the bundle, with a test that runs it   R1 · route:inline · tree:cd5577c4e6c9be284f81ffbab219ccac1e340cc4

## History
L1 · "the installer refuses bytes it cannot verify against the release signature"
L2 · OFICIO · "the installer requires cosign instead of verifying only when it is present" (porque a check that runs sometimes is a silent degradation, and this repository already refuses instead of degrading)
