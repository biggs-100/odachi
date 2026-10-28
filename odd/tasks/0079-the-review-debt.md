# the-review-debt
Base: 5bd880699e35d9483b94feb73c87a1d0cab19f12

## Requirements (verbatim, from the user)
R1 · "the review debt of 0078 is paid, and three rows leave the debt file"

## Done when
- the client-selection adapter tells no repository from a git failure, and init names the store it could not find
- review refuses a store it could not resolve instead of building a path relative to the process
- TreeOf, Has and CommitWithTree have adapter tests against a real repository
- the three open rows leave docs/review-debt.md, and the 0073 divergence stays written down

## Steps
- [x] T1 · the selection adapter tells no repository from a git failure, and init names the store it could not find   R1 · tree:67613cfe402629a1585a4bb4a71820fd0726bcc9
- [x] T2 · review refuses a store it could not resolve instead of building a relative path   R1 · tree:dd36158be36ba9dcd4858d4184c5e63e6e9a2fff
- [x] T3 · TreeOf, Has and CommitWithTree are pinned by adapter tests, and the debt file loses the three rows   R1 · tree:d5bfd1ff41b51b66c2e9380223c12561e2ce0ed2

## History
L1 · "the review debt of 0078 is paid, and three rows leave the debt file"
