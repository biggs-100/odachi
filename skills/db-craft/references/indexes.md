# An index is a bet on one query

An index makes reads cheaper and every write more expensive, so it is not a decoration
added "for performance": it is a bet that this shape of query will keep coming. The bet
is written in the column order, and the order is why two indexes on the same columns
are not the same index.

## The order is the index

- Postgres can use a leading part of the key: `(tenant_id, created_at)` serves
  `WHERE tenant_id = ?` and `WHERE tenant_id = ? ORDER BY created_at`, and it does not
  serve `WHERE created_at > ?` alone. Put the equality columns first, the range and the
  sort after.
- Equality before range, always: `(status, created_at)` for "pending, oldest first" is
  the index; `(created_at, status)` makes the database scan every recent row and filter.
- A sort that matches the index order is free; a sort on the other side of a range is
  not. `EXPLAIN` names it: `Sort Method: external merge` is the one to remove.
- `INCLUDE (column)` makes a covering index: the query is answered from the index alone
  when the visibility map is fresh. It is the cheapest fix for "it only needs two more
  columns".
- A partial index (`WHERE deleted_at IS NULL`) is smaller, faster and honest when the
  query always has that filter, which is exactly the soft delete case: the unique
  constraint stays `WHERE deleted_at IS NULL` and the dead rows leave the index.

## The ones that are not B-tree

- `GIN` for `jsonb`, arrays and full text: it indexes the contents, and it is the reason
  a `jsonb` column can be queried without becoming a table.
- `BRIN` for a huge table whose column correlates with physical order -a log by time-:
  tiny, and it skips blocks instead of rows.
- A hash index is for equality only and rarely beats a B-tree. An expression index
  (`lower(email)`) is how a case-insensitive lookup stops being a scan.

## When not to have one

- A column with three values and a query that returns most rows: the planner is right to
  scan, and the index is only a write cost.
- A table that takes more writes than reads: every index is a tax on the insert path.
- An index nothing uses. Read `pg_stat_user_indexes.idx_scan`; a zero after a week of
  production traffic is a decision nobody took. Drop it with `DROP INDEX CONCURRENTLY`.
- Duplicate indexes: `(a)` and `(a, b)` are not duplicates, `(a)` and `(a)` are, and the
  second one is pure cost.

## What Postgres does not index for you

- The referencing side of a foreign key. Every `DELETE` of a parent scans the child
  table, and the lock it takes waits for that scan. Index the `child.parent_id`.
- `NULL` in a unique index: `UNIQUE (email)` allows many nulls, which is usually what
  you want and is sometimes the bug.
- A `LIKE '%term%'`: a B-tree cannot help; that is full text or a trigram index.

## Reading the plan, in four lines

`EXPLAIN (ANALYZE, BUFFERS)` is the only honest reading: it runs the query.

1. **The row estimate against the actual** (`rows=100` vs `actual rows=100000`) means the
   statistics are stale or the predicate is not sargable: `ANALYZE`, or rewrite it.
2. **`Seq Scan` on a small table is correct.** On a large one filtered to a hundred rows
   it is a missing index or a cast that defeated it.
3. **`Nested Loop` with a big outer side** is the N+1 in SQL form; `Hash Join` is
   usually the fix the planner already found.
4. **`Buffers: shared hit` vs `read`** is cache: a fast query reading from disk is a
   fast query today and a slow one when the cache turns over.

## What changes on another engine

MySQL builds a clustered index on the primary key and every secondary index carries it,
which makes the primary key choice a storage decision. SQLite has indexes too, with no
`CONCURRENTLY` and no planner hints. The rule that does not change: the index follows
the query, the order follows the filter, and an unused index is a cost with no benefit.
