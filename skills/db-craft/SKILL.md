---
name: db-craft
description: Trigger: a schema, a migration, a query or a transaction on a SQL database. Postgres first -locks, expand and contract, indexes with evidence-, and what changes on another engine.
---

## When

A table, a column, a type, a constraint, an index, a migration, a query or a
transaction is being written or reviewed. `data` decides what is kept and for how long;
this door decides how it is written. A file on disk is `data`'s, not this one's.

## Rules

- **A migration is an online operation**: `SET LOCAL lock_timeout` and `statement_timeout` before any DDL, so a lock it cannot take fails fast instead of queuing every query behind it.
- **A change that rewrites the table is a deployment, not a statement**: expand the new shape, write both, backfill in batches, cut the readers over, contract in a later deploy.
- **Integrity lives in the database**: `NOT NULL`, `CHECK`, `UNIQUE` and `FOREIGN KEY` catch what the code forgets, and a rule that lives only in the application will be broken by the next endpoint.
- **A constraint on a large table lands in two steps**: `NOT VALID` is instant and enforces new writes, `VALIDATE CONSTRAINT` scans under a lock that blocks nobody.
- **Money is `numeric`, time is `timestamptz` in UTC, and an enum is a `CHECK` or a lookup table**: three choices that are cheap now and a migration under lock later.
- **An index exists because a query needs it**: the columns in the order the query filters and sorts, and an index nothing uses is write amplification with a memory bill.
- **A transaction holds locks**: no network call, no model, no email inside it; the work that leaves the database happens after the commit.
- **A serialization failure retries the whole transaction**, including the logic that chose the SQL, and never the statement alone.

## Decision gates

| Situation | Load |
|---|---|
| a table, a column, a type, a key or a constraint | `references/schema.md` |
| a migration, a rename, a backfill or a drop | `references/migrations.md` |

## Output

The schema or the migration, the lock it takes and what it blocks, the way back, and the query or the constraint that proves the change did what it claims. A rewrite on a large table returns the phases of an expand and contract, not a window.

## Never

- Run a DDL without `lock_timeout`, or a backfill of every row in one transaction.
- Add a constraint that scans a large table under an exclusive lock when `NOT VALID` and `VALIDATE` exist.
- Store money in a float, time without a timezone, or a mutable value in a Postgres enum.
- Decide what data is kept, or how long it lives: that is `data`'s door.

## References

- `references/schema.md` - types, keys, constraints, soft delete and names: the table is the contract.
- `references/migrations.md` - the lock each operation takes, expand and contract, and what no staging catches.
