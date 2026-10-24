# review-answers-from-stdout

## Requirements (verbatim, from the user)
R1 · "review --exec reads the reviewer's answer from stdout, and a failing reviewer shows both streams; the proof still keeps both"

## Done when
- a reviewer that writes a banner and ANSI to stderr still stores a valid verdict from its stdout
- a reviewer that fails shows stdout and stderr in the failure message
- check --run keeps both streams in the proof
- go test ./... and odachi check are green

## Steps
- [x] T1 · Shell ports return stdout and stderr separately; the adapter stops merging them   R1 · d4e61d3
- [ ] T2 · review parses stdout and shows both on failure; check keeps both in the proof   R1
- [ ] T3 · docs: CHANGELOG, impact map and the real-LLM dogfood rerun   R1

## History
L1 · "review --exec reads the reviewer's answer from stdout, and a failing reviewer shows both streams; the proof still keeps both"
L2 · OFICIO · "The Shell port returns stdout and stderr separately; the proof composes both and the reviewer's answer is stdout only" (porque A stream is a contract: decoration belongs to stderr and evidence to the proof, while a JSON answer cannot survive either being mixed in)
L3 · OFICIO · "review parses stdout only; a failing reviewer shows stdout and stderr together" (porque The answer is a contract on stdout; mixing decoration into it turns a good verdict into a parse error, while failures still need the whole picture)
