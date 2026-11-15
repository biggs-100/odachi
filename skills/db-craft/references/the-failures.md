# The failures, and the decision each one replaced

Every row is something that works in development and costs an incident in production.
Read it as a question about the change in front of you, not as a ban list: when a row
appears, the fix is the decision it was hiding, and the last column is where that
decision lives.

| the mistake | what you see | the fix | the page |
| --- | --- | --- | --- |
| DDL without `lock_timeout` | a long report blocks the table for everyone, then the app | set it; a failed migration retries, a convoy does not | `migrations.md` |
| `SET NOT NULL` on a large table | a scan under an exclusive lock | `CHECK … NOT VALID`, `VALIDATE`, then `SET NOT NULL` | `migrations.md` |
| Backfill of every row in one `UPDATE` | bloat, replication lag, a lock nobody sees | bounded batches, keyset, commit between them | `migrations.md` |
| Rename in one deploy | every old process fails for ninety seconds | add, dual write, backfill, cut over, drop | `migrations.md` |
| Adding an index without `CONCURRENTLY` | writes blocked while it builds | `CREATE INDEX CONCURRENTLY` | `migrations.md` |
| Money in a float | a total that is off by cents, then by more | `numeric`, or integer cents | `schema.md` |
| `timestamp` without timezone | an hour of records in the wrong day | `timestamptz`, stored in UTC | `schema.md` |
| `serial` for an id | an overflow at 2.1 billion rows | `bigint generated` or `uuid` | `schema.md` |
| A Postgres `enum`, or `jsonb` for a field you filter | a type you cannot change; a table you cannot join | `CHECK` or a lookup table; a column | `schema.md` |
| Uniqueness checked in the application | a duplicate under concurrent requests | a `UNIQUE` constraint | `schema.md` |
| Soft delete without a partial unique | the deleted row blocks the person coming back | `UNIQUE … WHERE deleted_at IS NULL` | `schema.md` |
| A foreign key with no index on the child | every parent delete scans the child | index the referencing column | `indexes.md` |
| A composite index in the wrong order | a `Seq Scan` on a table that had an index | equality columns first, then range and sort | `indexes.md` |
| An index nobody uses | slower writes and a bigger cache for nothing | `idx_scan`, then `DROP INDEX CONCURRENTLY` | `indexes.md` |
| `OFFSET` past the first pages | slow deep pages, a row repeated between them | keyset with the id as tiebreaker | `queries.md` |
| A query per row (the N+1) | latency that grows with the list | one join, or one batched `= ANY` | `queries.md` |
| `NOT IN (subquery)` with a null | an empty result that should have rows | `NOT EXISTS` | `queries.md` |
| `SELECT *` in a hot path | transfer, coupling and no covering index | name the columns | `queries.md` |
| A network call inside a transaction | locks held for the slowest third party | move the call after the commit | `transactions.md` |
| Retrying the failed statement | the same error, or a wrong write | retry the whole transaction | `transactions.md` |
| `SKIP LOCKED` with a global limit | more jobs running than the limit allows | a lock on the counter, or `pg_advisory_xact_lock` | `transactions.md` |
| Writing the row and the event separately | an event lost or sent twice | the outbox in the same transaction | `transactions.md` |
| RLS without `FORCE`, or a policy whose `USING` is broader than the tenant | the owner reads every tenant; a write the filter never intended passes both halves | `FORCE`, the tenant predicate in `USING` and `WITH CHECK`, and `SET LOCAL` in the transaction | `access.md` |
| The application runs as the owner | DDL from a request, RLS bypassed | three roles, grants per table | `access.md` |
| A plain `SET` on a pooled connection | the next request inherits the tenant | `SET LOCAL` inside the transaction | `access.md` |
| A backup nobody restored | the RTO is a guess on the worst day | the drill: a mid-day timestamp, a real restore, a number | `operations.md` |
| A pool sized by the number of workers | a connection storm at peak | size to the database's cores | `operations.md` |
| "It worked in staging" | five rows prove nothing about locks | the lock table, before the statement | `migrations.md` |

## What the rows have in common

Each one trades a decision for a default: the missing `lock_timeout` decides by waiting,
the float decides by rounding, the `OFFSET` decides by scanning. The database does not
refuse a decision nobody took -it executes it, quietly, and the bill arrives with the
data. The last column is the point of this page: every mistake here has a page that
already decided it.
