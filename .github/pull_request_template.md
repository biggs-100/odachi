<!-- One thing per pull request. If the title needs "and", it is two. -->

## What changed

<!-- For the person using odachi, in one line, in the imperative. -->

## Why

<!-- The issue, or the brief for a feature. A feature without an issue is a talk
     that has not happened yet. -->

## The proof

<!-- The CI run that passed, or the command you ran locally and its real output.
     `proof/` is generated and ignored by git: do not commit it, do not paste a
     proof from another tree, and never write one by hand. If the run failed,
     the gate kept the whole output in proof/<feature>.failed.txt: read it and
     say what it showed. -->

## Risk

<!-- none | data | security | contract | concurrency | delivery | no-test -->
Risk:

## Rollback

<!-- How to undo it if it breaks: revert this pull request, or the migration, the
     flag or the release step that follows. "Revert" is a valid answer when the
     change is one commit and has no data. -->

## Checklist

- [ ] The tests can fail: I saw them red for the right reason.
- [ ] Documentation travels with the change (README, CHANGELOG, ADR, impact map).
- [ ] No secrets, tokens or private project content in the diff, the proof or the
      screenshots.
- [ ] AI assistance: none / disclosed here, and a person read the result.
