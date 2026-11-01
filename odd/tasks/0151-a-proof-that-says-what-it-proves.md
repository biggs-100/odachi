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
- [x] T1 · the boundary 0150 opened is pinned by a test   R1 · tree:f16a2e42ae95a4da768fc4b017bc047484fdb229
- [x] T2 · a proof may carry a coverage line, and an old proof still proves its tree   R1 · tree:984e0200accb67230030e2a2915350209b4e9eff
- [x] T3 · check --run writes the coverage of the feature it proves   R1 · tree:0cdd66c5261e4e6d85ffa4de75596e628ccbc028
- [x] T4 · the close names a requirement the proof does not claim   R1 · tree:0cdd66c5261e4e6d85ffa4de75596e628ccbc028
- [x] T5 · the docs say the line is a claim, not evidence   R1 · tree:028772bfabf916fd2a8765e678f07b8bbf043fb7

## History
L1 · "a proof may carry the requirements it claims to cover, a proof without that line stays valid, and the close names a requirement the proof does not cover"
L2 · OFICIO · "T3 and T4 land in one commit: the line check --run writes and the close that reads it" (porque they are one behaviour seen from two sides, and splitting edits of the same file would leave the writer without its reader in the first commit)
