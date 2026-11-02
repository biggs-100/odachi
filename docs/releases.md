# Releases

A release is a promise someone can install: the tag exists, the binaries are signed, and the notes
say what changed for them. Three things that look like a release and are not:

- a heading in the changelog with no tag -it describes no bytes anyone can download;
- a build with a version injected by hand -it was never published, so nobody else can reproduce it;
- a branch, a pull request or a merge -the code is in `main`; the promise starts at the tag.

The rules below come from SemVer 2.0.0, Keep a Changelog 1.1.0, the Go modules reference,
trunk-based development, and the two automation doctrines that disagree with each other: release-please
keeps a release pull request a human merges, semantic-release publishes on every merge. This project
follows the first, because the decision belongs to a person.

## When a release is owed

When what a person can do, or can trust, changes. That is the whole rule; the table says what it means
for the change classes this project has.

| the change | release? | version | why |
| --- | --- | --- | --- |
| tests, CI, docs, an internal refactor | **no** | - | nothing observable moved; it rides the next release |
| a defect someone suffered | **yes**, once green | patch | waiting to batch a defect that hurts is not a favour |
| a new capability | yes, when it is usable and proven | patch | cap at two or three before publishing |
| a contract change: a flag that means something else, the shape of a stored record, a documented workflow step | yes | **minor** | it is the only visible "you may have to adapt" signal |
| a security fix | **yes, now** | patch | not batched, not held for a feature |
| a pre-release `-rc.N` | only if someone outside this machine must try it first | - | it sorts before the release it precedes |

A release is never scheduled for its own sake. There is no calendar here: the trigger is value, and a
quiet week is not a reason to publish. The safety valve is the other way around -an `Unreleased` that
keeps growing for weeks is a release waiting to happen.

## When no release is owed

- The work was tests, CI, documentation or an internal refactor: `Unreleased` gained nothing a person
  can observe, and a version that says "the same thing with another number" costs every reader.
- The tag already exists. A version is never moved, edited or reused (SemVer 3): a mistake is fixed by
  publishing the next one, and a version that must not be used is marked `[YANKED]`.
- The tree is red, or the close has no proof. The tag names tested bytes, or it names a guess.
- A decision is open. A tag resolves nothing: publish when the decision is taken, never to unlock it.

## Versioning while 0.x

SemVer says `0.y.z` is initial development and anything may change. That is a licence to move fast, not
a licence to be arbitrary: the number still tells a reader how much attention to pay.

- **minor** when the contract changes: a CLI flag that means something else, the shape of a stored
  record (task doc, fence, verdict, proof, store), or a step of the documented workflow.
- **patch** for every other observable change: a fix, or an addition that breaks nothing.
- **1.0.0** the day someone outside this machine depends on the artifacts. Not before: 1.0.0 is a
  promise of compatibility, and a project that cannot keep it should not make it.

## The tag

Annotated, `vX.Y.Z`, on the merge commit of the release pull request. The release workflow runs the
suite before goreleaser builds, so the tag names a tree that passed. Never a branch name, never a moved
tag, never a deleted one: the checksum and the attestation are bound to those bytes.

## The notes

The body of the release is the changelog section, verbatim. One source: a second text written for
GitHub is a second truth. Order: what changes for the person, then how to install (the exact command),
and any break or migration before everything else. Generated notes are a `git log` in another font.

## The flow here

1. `## Unreleased` grows while features close.
2. A release pull request renames it to `## vX.Y.Z - <date>`, a one-line change and nothing else.
3. Merge, then push the tag. **The tag is what publishes**: the workflow builds, signs and attests.
4. Install from the release, not from a local build: those are the bytes that were signed.
5. `## Unreleased` is empty again, and the next feature writes under it.

## What odachi should say, and never block on

When a feature closes, `check` and `summary` may say in one line that there is something to publish,
and how many entries `Unreleased` carries. It is a suggestion, not a gate: the publish is the human's
call, like every other irreversible step.

It should never suggest a release when the change was tests, CI, documentation or internal only, and a
release pull request that renames an empty `Unreleased`, or whose notes name no person's change, is a
refusal worth having.

## Decided here

- A version heading with no tag is not a release, so it may stop existing: the changelog names only
  versions a person could install.
- While `0.x`, a contract change bumps the minor and everything else observable bumps the patch.
