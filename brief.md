# the-tree-is

## What must exist
- the tree is gofmt-clean and a test keeps it so: every Go file in the repository is the bytes gofmt would write, checked in process so a CRLF or a missing final newline cannot land again

## Done when
- gofmt -l . prints nothing, and the file that was red is canonical
- a test walks every .go file and fails when format.Source would rewrite it, with the path and the first difference
- the test is fast: it parses the tree in one pass, without spawning a process per file
- the changelog says one line

## Not included (first version)
- no change to the verdict shape
- no change to the proof format, and the stored proof is never rewritten

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: cómo una higiene que nadie mira se queda roja catorce releases.
