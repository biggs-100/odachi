# A migration is an online operation

The SQL is the easy part. The risk is the lock: `ALTER TABLE` takes `ACCESS
EXCLUSIVE` unless the subcommand says otherwise, and that lock conflicts with
`SELECT` too. Postgres grants locks fairly, so a migration waiting three minutes for
a long report also makes every query that arrives after it wait. That queue, not the
lock, is what takes an application down. Every DDL starts here:

```
BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
-- the change
COMMIT;
```

A migration that cannot take its lock in five seconds fails; a failed migration that
runs again later costs nothing, and a five-minute convoy costs an incident.

## What each operation costs

| operation | lock | blocks |
| --- | --- | --- |
| `SELECT`, `INSERT`, `UPDATE`, `DELETE` | `ROW EXCLUSIVE` (shared by writers) | other DDL, not each other |
| `CREATE INDEX` | `SHARE` | writes, not reads |
| `CREATE INDEX CONCURRENTLY` | `SHARE UPDATE EXCLUSIVE` | nothing you care about; cannot run inside a transaction |
| `ALTER TABLE` | `ACCESS EXCLUSIVE` | everything, including reads |
| `ADD CONSTRAINT … NOT VALID` | `ACCESS EXCLUSIVE`, briefly | the catalog only, no scan |
| `VALIDATE CONSTRAINT` | `SHARE UPDATE EXCLUSIVE` | neither reads nor writes |
| `ALTER TABLE ADD COLUMN` with a constant default | `ACCESS EXCLUSIVE`, instant since Postgres 11 | nothing measurable: the default is read, not written |
| `ALTER COLUMN TYPE` (not binary-coercible) | `ACCESS EXCLUSIVE` | a full rewrite of the table and its indexes |

## Catalog-only, and the ones that rewrite

Instant: adding a column with a constant default or none, dropping a column (invisible
at once, the space returns later), `ADD CONSTRAINT … NOT VALID`, and
`CREATE INDEX CONCURRENTLY`.

A rewrite: a volatile default such as `clock_timestamp()`, a stored generated column,
an identity column, and most type changes. A rewrite holds the table for its whole
duration and needs up to twice the disk. On a large table that is a maintenance window
or an expand/contract, and `varchar(50)` to `varchar(100)` is not the same risk as
`int` to `text`.

## `SET NOT NULL` is a hidden scan

Making a column `NOT NULL` verifies every row under an exclusive lock. Since Postgres
12 the planner accepts a validated check as proof, so the safe order is:

```
ALTER TABLE orders ADD CONSTRAINT orders_user_id_not_null
  CHECK (user_id IS NOT NULL) NOT VALID;
ALTER TABLE orders VALIDATE CONSTRAINT orders_user_id_not_null;
ALTER TABLE orders ALTER COLUMN user_id SET NOT NULL;   -- metadata only
```

## Expand, backfill, contract

Anything that is not instant and not catalogue-only is a deployment with phases, and
the application owns the transition:

1. **Expand**: add the new column, table or index, nullable or unconstrained. Old code
   keeps working.
2. **Dual write**: the code writes both shapes; a reader may still read the old one.
3. **Backfill** in bounded batches, keyset-paginated, committing between them, with a
   short pause so `VACUUM` and replication keep up. One `UPDATE` of ten million rows is
   a lock, a bloat and a lag in a single transaction.
4. **Cut the readers** over, and only then
5. **Contract**: drop the old shape in a later deploy, once nothing reads it.

Each phase is reversible on its own. A rename is the canonical case: `RENAME COLUMN`
is instant and breaks every running copy of the old code, so the safe rename is add,
dual write, backfill, cut over, drop.

## Indexes and keys without the rebuild

`CREATE UNIQUE INDEX CONCURRENTLY` and then adopt it, instead of `ADD PRIMARY KEY`
which rebuilds:

```
CREATE UNIQUE INDEX CONCURRENTLY orders_id_temp ON orders (id);
ALTER TABLE orders DROP CONSTRAINT orders_pkey,
  ADD CONSTRAINT orders_pkey PRIMARY KEY USING INDEX orders_id_temp;
```

## After it lands

- Watch `pg_stat_activity` and `pg_blocking_pids()` while the migration runs: if the
  lock is held longer than expected, cancel the migration, do not wait it out.
- Expect a short recompile of cached plans after a schema change; a pooler in
  transaction mode makes that herd larger, not smaller.
- A migration is reviewed like code, and the expand phase is deployed apart from the
  contract phase. Staging forgives what production does not: it has five rows.

## What changes on another engine

MySQL 8 has instant `ADD COLUMN` too, but the safe path is explicit (`ALGORITHM=INPLACE`,
`LOCK=NONE`) and it silently falls back to a copy; check the result. SQLite rewrites the
table for most `ALTER`, so its migrations are a new table, a copy and a swap. The rule
that does not change: know the lock before you write the statement.
