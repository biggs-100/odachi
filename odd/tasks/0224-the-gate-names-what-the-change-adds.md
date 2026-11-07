# the-gate-names-what-the-change-adds
Base: 9a211be952e250bfc7adc0acdbe496c88143ff1f

## Requirements (verbatim, from the user)
R1 · "when the feature closes, `check` names what the diff adds —the new files and the dependency files— in one line of the ready output, so the reviewer sees the scaffolding before reading the code; the exit code and the verdict do not change"

## Done when
- the ready output names the new files (capped) and the dependency files the diff adds, and stays silent when there are none
- the line never changes the exit code, the reason or the proof
- a test seeds a diff with an added file and a `go.mod` change and asserts the line; another asserts silence on a diff with neither
- `odachi check --run "go test ./..."` is green
- the changelog's Unreleased section names the line
- `go test ./...` is green

## Steps
- [x] T1 · check names the new files and the dependency files   R1 · route:inline · tree:106b0086bb4bcf7183bc9963d5aab99bde81a163
- [x] T2 · the changelog names the line   R1 · route:inline · tree:0a569d98ebe5f8d35a9b6ca3b5d6a68d11d817bf

## History
L1 · "when the feature closes, `check` names what the diff adds —the new files and the dependency files— in one line of the ready output, so the reviewer sees the scaffolding before reading the code; the exit code and the verdict do not change"
