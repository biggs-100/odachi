---
name: ship
description: Trigger: when something is ready for another person to use. Verify the installer's path, not the developer's.
---

## When

When you are about to say "done", publish a version, hand something to someone
who does not have your environment, or close a feature.

## Rules

- "Done" is `odachi check` green on the current tree, with the real output saved
  in `proof/`. Everything else is "I think it is done".
- Test the installer's path, not the developer's: the downloaded binary, without
  Go, without your PATH, on another machine if possible.
- Verify against the real artifact, not fixtures. An installer that only worked
  with a fake file is not verified.
- Say what was NOT verified and why. That paragraph is part of the delivery, not
  a confession.
- The changelog talks about what changed for the human, not about the files you
  touched.
- If the behavior of something that already existed changes, say it first and
  plainly.

## Never

- Publish with red tests, without a version or without a checksum.
- Say "done" when it only compiles.
- Change an existing verb without announcing it in the same release.
- Sign as verified something that was tested by hand once.
- Tag a version from a tree that is not the one you verified.
