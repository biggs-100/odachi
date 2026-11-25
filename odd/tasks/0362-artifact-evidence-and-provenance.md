# artifact-evidence-and-provenance
Base: 431973abc13d4a06d7c86e9a5c5111c4bbe74932

## Requirements (verbatim, from the user)
R1 · "odachi's existing writing and ui-craft skills require evidence for delivered artifacts and provenance for content and assets, without adding a new skill or a second delivery authority"

## Done when
- `skills/writing/references/the-doc.md` defines a stack-neutral verification ladder from source correctness through the rendered artifact, real destination behavior and readback
- an unavailable check is reported as `not verified`, never silently passed
- `skills/ui-craft/references/task-and-content.md` requires creator or source, usage rights, version or date, and generated or demo status for external content and assets
- `skills/ui-craft/references/the-read.md` verifies the exported artifact at its destination, including link response and actual media dimensions, and leaves aesthetic judgment to a human decision
- `TestArtifactDeliveryRequiresEvidenceAndProvenance` fails before the content change for the missing contracts and passes after it
- `odachi skills --check`, the focused skill tests, `go test ./internal/archtest`, and `go test ./...` are green
- only the `skills:payload` pin changes deliberately, and `## Unreleased` names the feature

## Steps
- [x] T1 · add artifact evidence and provenance to the existing writing and ui-craft skills   R1 · route:delegated/odachi-worker · skills:tests,writing,skill-craft · tree:d8dd91b2b0fb2f41e0a9fa6a397c0a98f8e6072a

## History
L1 · "odachi's existing writing and ui-craft skills require evidence for delivered artifacts and provenance for content and assets, without adding a new skill or a second delivery authority"
