# 0011 · Backups in the project store, not beside the artifact

Status: accepted · Date: 2026-10 · Session: v0.11.1

## Context

Every verb that rewrites a human-owned file saves the previous version first, so
`undo` can put it back. The backup was written beside the artifact:
`brief.md.bak-<hash>`, `0009-....md.bak-<hash>`, `SKILL.md.bak-<hash>`. The bytes
are safe and git ignores them, but the project becomes visually noisy: eight
`brief.md.bak-*` files in the root look like unfinished work, even though they
are the undo history.

The name already carries a content digest. The location was the only problem.

## Decision

- **New backups live in `.odachi/backups/`**, preserving the artifact's path
  relative to the project: `brief.md`, `odd/tasks/0009-....md`,
  `.odachi/profile.toml`. One store, no second naming scheme.
- **`fsx.BackupIn(root, path)` owns the store**, and `osFiles.Backup` finds the
  project by the nearest ancestor with a `.odachi` directory. A file outside a
  project has no store to join and keeps the backup beside it; refusing to save
  a version would be worse than one noisy file.
- **`undo` reads both locations.** The store is the new home; the legacy
  backups beside the artifact remain visible and restorable. A project migrates
  at its own pace.
- **`undo --collect` moves the legacy backups into the store.** It uses a
  rename, not a rewrite, so the bytes and the mtime survive; a second run is a
  no-op. If the target exists with different content, it refuses to overwrite:
  a content-addressed name that disagrees with its bytes is not a file to trust.
- **Backups remain ignored by `*.bak-*`.** They are tool state, not generated
  evidence and not project content.
- **No retention yet.** The store still accumulates versions; the difference is
  that they are out of sight and their location has one owner. A `--prune` can
  be added later without changing the read path.

## Consequences

- The root and `odd/tasks/` show only the artifacts a human reads. The backup
  history stays in `.odachi/backups/`, where `undo --list` still names it.
- `undo` reads two places, so a project that upgrades mid-feature does not lose
  the versions it already had. `--collect` is the explicit migration; it is not
  automatic because a read-only `--list` must not write.
- The project boundary is the nearest `.odachi`, not the current working
  directory. A backup made by `init` in a folder without `.odachi` may still
  land beside the file; once `start` creates `.odachi`, the store takes over.
- An outside file (a home skill, for example) keeps the old shape. It is not
  part of a project, so there is no project store to own it.
- Tests cover the low-level writer, the adapter's project detection, the undo
  read path, the migration, the no-op second run and the refusal to overwrite.
