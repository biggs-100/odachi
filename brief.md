# the-release-doctrine

## What must exist
- the release doctrine is written, drive points at it, and the changelog names only versions that were published

## Done when
- docs/releases.md says what a release is, when to publish and when not, how to version while 0.x, how the tag is named and placed, and what the notes carry
- the drive skill names the doctrine in one line and stays inside its sixty-line format
- CHANGELOG.md no longer carries a version heading with no tag: the one that was never published stops existing, so every heading is a version a person could install
- the next release is one that names the observable changes and nothing else

## Not included (first version)
- no change to how a release is built or published: the workflow, the signing and the installers stay as they are
- a stored proof is never rewritten, whatever the writer learns

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why a release is a promise and not a milestone.
