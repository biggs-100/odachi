# the-release-body

## What must exist
- the release body comes from the changelog, not from memory: a step of the workflow, so nobody writes it twice by hand

## Done when
- scripts/release-notes.sh prints the tag's changelog section verbatim and the install block for that tag, and refuses a tag with no section
- the release workflow calls it after goreleaser and edits the release with that body
- a test pins the section, the install block and the refusal

## Not included (first version)
- the gate and the signing stay: the step only edits the body of a release that already exists
- a stored proof is never rewritten, whatever the writer learns

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why a release note has one source and not a second text written for GitHub.
