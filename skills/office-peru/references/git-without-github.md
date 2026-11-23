# Git as a local notebook of versions

A public servant does not need a hosting platform to know which draft was
reviewed. A repository on the machine records that, with no account, no remote
and no publish step.

## The four commands that matter

- `git init` — start a notebook in the folder that holds the document.
- `git add <file>` — put the current version of one file on the list to record.
- `git commit -m "reviewed draft with sources 2026-11-23"` — record that version
  with a sentence in ordinary language. The sentence is the label a person reads.
- `git log --oneline` — see the labels in order, to find the version that was
  reviewed.

Everything else is optional. Do not introduce branches, pull requests, rebases
or continuous integration unless the person asks for them by name.

## What to record

- The intent paragraph and the requirement in the requester's words.
- Each draft that was sent for review, with the date.
- The source list and the checked excerpts for that draft.
- The reviewer's note and the version it refers to.

A useful label names the state and the date, not a code: "draft with three
sources checked", "reviewed by the officer", "final sent 2026-11-23".

## Boundaries

- Never publish the repository, add a remote or push, unless the person asks and
  the entity allows it. Local history is enough for the evidence habit.
- If the entity forbids storing certain data on disk, do not put that data in the
  notebook. Record where the approved source lives instead.
- Personal data about citizens or colleagues does not belong in a commit label.
- The notebook is evidence of a process, not proof that a claim is true. The
  checked primary source is the proof.

## If the person asks why not just use a shared folder

A shared folder keeps the latest file and loses the earlier ones. The notebook
keeps the sequence: which draft had which sources, and which one the reviewer
saw. That sequence is the answer when someone asks how the document was made.
