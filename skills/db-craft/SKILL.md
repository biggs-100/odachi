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
- **One round trip per screen**: a query per row is an N+1, and a page that gets slower as the list grows is the first thing production finds.
- **Page with a cursor, not an offset**: `OFFSET` scans what it skips and repeats or loses a row when something is written between two pages.
- **The application role cannot change the schema and does not bypass row-level security**: the database decides who may read a row, and the request path is not the owner.
- **A transaction holds locks**: no network call, no model, no email inside it; the work that leaves the database happens after the commit.
- **A serialization failure retries the whole transaction**, including the logic that chose the SQL, and never the statement alone.
- **An effect that leaves the database is at-least-once**: an idempotency key for the request, or an outbox row written in the same transaction as the state it announces.

## Decision gates

| Situation | Load |
|---|---|
| a table, a column, a type, a key or a constraint | `references/schema.md` |
| a migration, a rename, a backfill or a drop | `references/migrations.md` |
| a transaction, a retry, a lock or a background job | `references/transactions.md` |
| an index, a plan or a query that is slow | `references/indexes.md` |
| a list, a page, a count or a batch write | `references/queries.md` |
| a role, a grant, row-level security or a tenant | `references/access.md` |
| a backup, a restore, a pool, vacuum or an upgrade | `references/operations.md` |

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
- `references/transactions.md` - isolation, the retry that replays everything, queues and the outbox.
- `references/indexes.md` - the column order, the index that is not a B-tree, and reading a plan.
- `references/queries.md` - the N+1, the cursor, and the two numbers to watch.
- `references/access.md` - three roles, row-level security and the doors around it.
- `references/operations.md` - the restore drill, what to watch, the pool and vacuum.
