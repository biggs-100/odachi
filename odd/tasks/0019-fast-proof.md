# fast-proof

## Requirements (verbatim, from the user)
R1 · "the suite and the proof loop stop being the slowest part of a commit"

## Done when
- go test ./cmd/odachi finishes under 25 seconds on this machine (it was ~70)
- the full suite stays green, race-clean on ./internal/..., and gofmt/vet clean
- the workflow gets the proof by running the suite once per tree, not twice
- go test ./... and odachi check are green

## Steps
- [ ] T1 · the cmd e2e tests run in parallel where their fixtures are isolated   R1
- [ ] T2 · docs: how the proof runs the suite once per tree, and the measured result   R1

## History
L1 · "the suite and the proof loop stop being the slowest part of a commit"
L2 · OFICIO · "The e2e tests own their fixtures, so they run in parallel; the ones that redirect the home or PATH stay serial" (porque t.Setenv forbids t.Parallel and a shared home is real state; isolation is the condition for concurrency, not a detail)
L3 · OFICIO · "TestMain prepares one empty git repository and every test copies its .git directory" (porque On Windows a git process costs ~80ms and the suite creates a repository per test; copying a template is the same repository without the process)
L4 · OFICIO · "Measured: cmd/odachi 69.5s -> 44s, full suite ~74s -> 60s; the serial doctor tests dominate the tail" (porque Doctor redirects PATH with t.Setenv, so it cannot run in parallel; the honest result is recorded instead of moving the target silently)
