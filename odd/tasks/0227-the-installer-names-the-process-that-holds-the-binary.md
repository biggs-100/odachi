# the-installer-names-the-process-that-holds-the-binary
Base: aa5066c09dae07f00c126d35ffa390b4b142c47c

## Requirements (verbatim, from the user)
R1 · "the Windows installer never answers a running binary with a Copy-Item IOException: when `odachi.exe` cannot be replaced because a process holds it, it names the process, says what to do -close that terminal or the other project's gate, or install with `ODACHI_BIN_DIR`- and installs nothing, so the human sees the cause instead of a sharing violation"

## Done when
- `install.ps1` catches the failed replacement, finds the processes whose executable is the target, and exits 1 naming their PID
- the message names the cause (a terminal or another project's pre-commit hook running the binary) and the way out: close it, or install into another folder with `ODACHI_BIN_DIR`
- when no process holds the file, the message names the failing path and the action, and installs nothing
- a Windows test parks a process on `binDir\odachi.exe`, runs the real installer against the local release, and asserts exit 1, the PID in the output and the way out; the target stays as it was
- the installer's other paths stay as they are: a bad signature, a missing cosign and the user PATH still behave the same
- `go test ./...` is green, the Windows test included
- `odachi check --run "go test ./..."` is green
- the changelog's Unreleased section names the message

## Steps
- [x] T1 · the installer names the process that holds the binary   R1 · route:inline · tree:5b3f5d3b5266c0cc9d32030bde32a4a85fef8b64
- [x] T2 · the changelog names the message   R1 · route:inline · tree:edc9f5eeb638ce301749f912176bb82f9da25458

## History
L1 · "the Windows installer never answers a running binary with a Copy-Item IOException: when `odachi.exe` cannot be replaced because a process holds it, it names the process, says what to do -close that terminal or the other project's gate, or install with `ODACHI_BIN_DIR`- and installs nothing, so the human sees the cause instead of a sharing violation"
