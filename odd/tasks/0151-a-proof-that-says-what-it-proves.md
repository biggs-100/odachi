# a-proof-that-says-what-it-proves
Base: f0452f17f9d0b88ffcd88cd4b845e9d751faefba

## Requirements (verbatim, from the user)
R1 · "a proof may carry the requirements it claims to cover, a proof without that line stays valid, and the close names a requirement the proof does not cover"

## Done when
- the boundary 0150 opened is pinned by a test: a change under the stored proofs is blocked and one under the package that writes them is not
- the proof header may carry a coverage line, and a proof written before the field existed still proves its tree
- check --run writes the requirements the closing feature declares, and the close names the ones the proof does not claim
- the documentation says the line is a claim by the agent, not evidence the suite produced

## Steps
- [ ] T1 · the boundary 0150 opened is pinned by a test   R1
- [ ] T2 · a proof may carry a coverage line, and an old proof still proves its tree   R1
- [ ] T3 · check --run writes the coverage of the feature it proves   R1
- [ ] T4 · the close names a requirement the proof does not claim   R1
- [ ] T5 · the docs say the line is a claim, not evidence   R1

## History
L1 · "a proof may carry the requirements it claims to cover, a proof without that line stays valid, and the close names a requirement the proof does not cover"
