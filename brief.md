# the-installer-names-the-process-that-holds-the-binary

## What must exist
- the Windows installer never answers a running binary with a Copy-Item IOException: when `odachi.exe` cannot be replaced because a process holds it, it names the process, says what to do -close that terminal or the other project's gate, or install with `ODACHI_BIN_DIR`- and installs nothing, so the human sees the cause instead of a sharing violation

## Done when
- `install.ps1` catches the failed replacement, finds the processes whose executable is the target, and exits 1 naming their PID
- the message names the cause (a terminal or another project's pre-commit hook running the binary) and the way out: close it, or install into another folder with `ODACHI_BIN_DIR`
- when no process holds the file, the message names the failing path and the action, and installs nothing
- a Windows test parks a process on `binDir\odachi.exe`, runs the real installer against the local release, and asserts exit 1, the PID in the output and the way out; the target stays as it was
- the installer's other paths stay as they are: a bad signature, a missing cosign and the user PATH still behave the same
- `go test ./...` is green, the Windows test included
- `odachi check --run "go test ./..."` is green
- the changelog's Unreleased section names the message

## Not included (first version)
- no change to `install.sh`: on Linux and macOS a running binary is replaced by rename, and there is no lock to explain
- no new flag, no new environment variable and no new verb
- no change to the release workflow or to the assets the release publishes
- no change to the binary, its verbs or the gate: the installer is the whole change
- no change to the brief's format or to `start`
- no change to the review, the doctor or the classifier
- no change to the stored proofs: never touched
- no change to the skills or to the voice
- no change to the design canvas, its token or the push plugin
- no change to the artboard format

## Decisions taken
- the message is built where the copy fails, not before it: a pre-check has a race with the process that starts after it, and the failed copy is the fact worth explaining
- the search compares the executable path, not the process name: another `odachi.exe` in another folder is not the file this install replaces, and naming it would send the human to the wrong window
- a failure with no holder still gets a sentence: an antivirus or an indexer can hold the file without a process odachi can name, and "nothing was installed" must stay true in that case too
- the installer keeps its voice in Spanish, like the rest of `install.ps1`; the test asserts the PID and the way out, not the prose
