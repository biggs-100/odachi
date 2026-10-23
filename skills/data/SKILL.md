---
name: data
description: Trigger: when data has to be saved, migrated, read or deleted. Decide owner, lifetime and backup before writing the first row.
---

## When

When a file has to be persisted, a schema changes, a migration appears, a delete
is requested, or the question "where do we store this?" comes up.

## Rules

- Before writing, answer three things: who owns the data, how long it lives and
  who reads it. If there is no answer, it is not time to store it yet.
- Write atomically: a temporary file in the same directory and a rename at the
  end. A crash halfway leaves the old file whole, not a truncated one.
- Every migration is reversible or has a backup you have already restored. A
  backup that was never restored is not a backup.
- Read verifying: if the file carries a checksum or a length, check it when
  opening. Assuming it is whole is how data gets lost.
- Sensitive data (names, health, money) does not enter the model or the log if it
  is not needed. What is not stored cannot leak.
- Test data lives in small, readable fixtures; real data never lives in the
  repository.

## Never

- Run a destructive migration without a verified backup and a way back.
- Store secrets, tokens or personal data in the repo, fixtures or logs.
- Change the schema and the code that reads it in a deploy that cannot be
  reverted.
- Delete on the spot: mark it, wait, and delete with a date and an owner.
- Invent data to make a test pass.
