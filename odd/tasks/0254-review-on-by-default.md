# review-on-by-default
Base: d2187207e1c85114b0f36d76fe1e76d54df14c5d

## Requirements (verbatim, from the user)
R1 · "a new project is born with the review on: `init` writes `odd/review.toml` with `mode = "on"` and never replaces one"

## Done when
- `odachi init` writes `odd/review.toml` with `mode = "on"` when the project has none, through the writer `review mode on` uses
- a project that already has the file keeps it byte for byte, and the run says it kept it
- `init --json` reports the file it wrote or kept
- `odachi check` on a project with that file asks for the verdict the mode owes
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- Unreleased names the change: a contract change, minor (0254)

## Steps
- [x] T1 · the port and the write: init leaves odd/review.toml with mode on   R1 · route:inline · tree:89119879ffea64766ac123d7d7f9bd9dbd426576
- [x] T2 · the tests: written, kept byte for byte, and named in --json   R1 · route:inline · tree:1aac406a6b691e0b382d912f6f6e3014a12d7143
- [x] T3 · Unreleased and the impact map   R1 · route:inline · tree:39d4ad7c29f754e8dd5777376c2bd7fb8b41b0e6

## History
L1 · "a new project is born with the review on: `init` writes `odd/review.toml` with `mode = "on"` and never replaces one"
