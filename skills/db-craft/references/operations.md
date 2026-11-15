# The database is the part that keeps running

Everything here is about the day after the feature ships: the backup nobody restored, the
connection pool that ran out at 9 a.m., the table that grew a hundred million rows
overnight. The list is short because most of it is one habit: measure the thing before
you need it in an incident.

## A backup is the drill, not the copy

- **Physical, for recovery**: a base backup plus continuous WAL gives point-in-time
  recovery (`pgBackRest`, `wal-g`, or the provider's PITR). A `pg_dump` per night gives
  you yesterday, not 14:03, and not a 200 GB restore in a useful time.
- **The drill is the only proof**: pick a timestamp from yesterday, in the middle of the
  business day, restore it into a separate environment, and check a table the business
  cares about. Measure the wall time: that number is the RTO, and assuming it is faster
  is how an incident doubles.
- Say the two numbers where the team can read them: how much data can be lost (RPO) and
  how long the restore takes (RTO). A backup with no drill has neither.
- Store the backups somewhere the database cannot delete them, and test the restore with
  the credentials a responder actually has at 3 a.m.

## What to watch, in order of usefulness

1. **`pg_stat_statements`**: calls, mean and total time. The query that costs the most
   total time is the one to fix first, and it is rarely the slowest one.
2. **The slow log**: `log_min_duration_statement` above the p95 you accept, so the tail
   is visible instead of averaged away.
3. **Locks**: `pg_locks`, `pg_blocking_pids()`, and a query that has waited longer than a
   few seconds. A lock wait is an outage starting.
4. **Bloat**: `n_dead_tup` per table. A table that only grows its dead tuples has an
   autovacuum that cannot keep up, and every scan pays for it.
5. **Connections**: used against `max_connections`. Full in the afternoon means the
   pool, not the database, is the bottleneck.
6. **Replication lag** where there is a replica: a read that must be current does not go
   to a lagging one, and the feature that forgot it reads yesterday.

## The pool is part of the design

- A connection is a process in Postgres: more connections is not more throughput, it is
  more context switching. Size the pool to the database's cores, not to the number of
  workers.
- A pooler in transaction mode rewrites the assumptions: `SET` (without `LOCAL`), session
  advisory locks, prepared statements and `LISTEN` all outlive the statement the way they
  should not. Use `SET LOCAL`, `pg_advisory_xact_lock`, and named prepared statements
  only where the pooler supports them.
- `idle_in_transaction_session_timeout` and `statement_timeout` on the role are cheaper
  than the incident a forgotten transaction causes.

## Vacuum, and the table that grows

- Autovacuum exists and is often too gentle for the hot tables: `autovacuum_vacuum_scale_factor`
  per table, set for the tables that take the writes.
- A partition by time is right for an append-only table that is read by range and deleted
  by age (`DROP PARTITION` is instant where `DELETE` is a night). It is wrong for a table
  that is read by anything other than the partition key.
- Free space is not returned to the filesystem by `DELETE`; it is reused. Disk planning
  is about the peak, not the average.

## In CI, and in the deploy

- Migrations run in CI against a fresh database of the same major version, and the
  down-migration or the backup path is exercised where it exists.
- The expand phase and the contract phase are separate deploys, and the release notes say
  so: a rollback that stops at the expand phase is a code rollback, not a data rollback.
- A major version upgrade is a project: logical replication moves a live database with
  seconds of switchover, `pg_upgrade` is a window. Extensions are the compatibility list
  to check first.

## What changes on another engine

MySQL's PITR is binlogs rather than WAL, and its `performance_schema` replaces
`pg_stat_statements`. SQLite's backup is a file copy done correctly (`VACUUM INTO` or the
backup API), and its "restore drill" is opening a copy and running integrity_check. The
rule that does not change: the backup is proven by the restore, and the numbers are
measured before the incident.
