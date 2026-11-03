# the-release-body
Base: 41d82ddbb9807479d4138fb319518e277fa4c728

## Requirements (verbatim, from the user)
R1 · "the release body comes from the changelog, not from memory: a step of the workflow, so nobody writes it twice by hand"

## Done when
- scripts/release-notes.sh prints the tag's changelog section verbatim and the install block for that tag, and refuses a tag with no section
- the release workflow calls it after goreleaser and edits the release with that body
- a test pins the section, the install block and the refusal

## Steps
- [x] T1 · a script prints the tag's section from the changelog and the install block, and refuses a tag with no section   R1 · route:inline · tree:28febecd2f725bc22d11ab3076b1e60562daa2a8
- [x] T2 · the release workflow edits the release with that body after goreleaser   R1 · route:inline · tree:01377c30030dfa2ebda126308973b30061e17b96

## History
L1 · "the release body comes from the changelog, not from memory: a step of the workflow, so nobody writes it twice by hand"
