# The query is part of the design

A screen that loads in one query and a screen that loads in four hundred are the same
product until the data grows. Query cost is a design decision taken at the call site:
what is fetched, how much, and how many round trips.

## The N+1, and how it hides

An ORM that loads a list and then touches a relation per row issues one query for the
list and one per item. It is invisible in development, where the list has three rows,
and it is the first thing that breaks in production.

- The symptom is a request whose time grows with the number of rows on the screen, not
  with the size of the rows. A page of 50 items taking 400 ms while one item takes 8 ms
  is an N+1.
- Detection is counting, not guessing: `pg_stat_statements` ordered by `calls`, or a
  query log with a request id. The count of identical statements per request is the
  number.
- The fix is one join, one `WHERE id = ANY($1)` in a second query, or a batched loader.
  Eager loading is the ORM's version of the same fix, and it is a decision per
  relation: eager everything is how a page loads a table nobody rendered.

## Pagination that does not lie

- `OFFSET` gets slower with depth because the database scans and discards every skipped
  row. Worse, it is not stable: a row inserted while the person reads page one shifts
  everything and page two repeats or skips an item.
- Keyset resumes from the last row: `WHERE (created_at, id) < ($1, $2) ORDER BY
  created_at DESC, id DESC LIMIT 20`. Constant time at any depth, no duplicates, no
  gaps.
- The tiebreaker is not optional: `created_at` alone is not unique, and two rows with
  the same timestamp are the classic skipped row. The cursor is the ordered tuple.
- Keep `OFFSET` for the case it is good at: a small, bounded set where "jump to page 7"
  is a requirement. Then say the depth bound in the code.

## Fetching what the screen needs

- `SELECT *` couples the query to the schema, transfers columns nobody renders, and
  defeats a covering index. Name the columns.
- `COUNT(*)` on a large table is a scan. If the number is a UI hint, `reltuples` from
  `pg_class` is instant and approximate; if it is a fact, it is a materialized counter
  updated in the same transaction as the rows.
- `NOT IN (subquery)` with a `NULL` in the result returns nothing at all: use `NOT
  EXISTS`. It is the oldest silent wrong answer in SQL.
- An aggregate over a filtered join belongs in one query, not in the application. The
  network is slower than the database that already has the rows.

## Writing in volume

- One `INSERT` with many rows, or `COPY`, not a loop of single inserts: the round trip
  is the cost.
- `ON CONFLICT` for the upsert; `RETURNING` to get the generated values back instead of
  a second `SELECT`.
- A batch that can fail halfway is a transaction that should be bounded by design: pick
  the batch size so a failure retries a second of work, not ten minutes.

## The two numbers to watch

- **Calls per request**: the N+1. Read it from `pg_stat_statements` or a log.
- **p95, not the average**: a query that takes 2 ms on average and 4 s at the tail is an
  incident every fifth request. `pg_stat_statements` gives `mean_exec_time`, and the
  slow log gives the tail; the tail is the one a person feels.

## What changes on another engine

MySQL's `LIMIT`/`OFFSET` behaves the same and its `ON DUPLICATE KEY UPDATE` is the
upsert; SQLite has `INSERT OR REPLACE` with different semantics (it deletes and
reinserts, so triggers and foreign keys notice). The rule that does not change: one
round trip per screen is the budget, and a cursor is the only pagination that stays
correct while rows are written.
