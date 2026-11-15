# The table is the contract

A schema outlives the code that wrote it, the framework that generated it and usually
the person who chose it. So it is written for the reader in five years: names that say
the thing, constraints that refuse the impossible row, and a type for each value that
cannot be the wrong shape.

## Types, and the four that are always chosen wrong

| value | the type | why not the usual one |
| --- | --- | --- |
| money | `numeric(precision, scale)` | `float` rounds cents away and the error compounds in a total |
| time | `timestamptz`, stored in UTC | `timestamp` loses the offset; two servers disagree about "now" |
| identifiers | `bigint` generated, or `uuid` when the id leaves the system | `serial` is `int4`: 2.1 billion rows is a Tuesday |
| free text | `text` with a `CHECK (length(…))` if a bound exists | `varchar(n)` is not faster and the limit is painful to raise |

- An enum is a `CHECK` on a text column or a lookup table, not a Postgres `enum` type:
  the type cannot drop a value, and the migration to undo a mistake is a rewrite.
- `jsonb` is for attributes that are genuinely open -metadata, a provider's payload, a
  saved filter-. A field you query, index and constrain is a column; a table hiding in
  JSON is a table you cannot join.
- Booleans are not tri-state: if "unknown" is real, it is a nullable column with a name
  that says so, or a small text status.
- Arrays are a list you cannot query well; a child table is a list you can.

## Keys, and the policy of the foreign key

- Every table has a primary key. A surrogate `bigint` is the default; a natural key is
  for values that are truly immutable and truly unique (a country code, not an email).
- A `UNIQUE` constraint is the only unique that survives concurrency: "check then
  insert" in the application is a race with a duplicate at the end of it.
- A foreign key says what happens when the parent goes: `ON DELETE RESTRICT` by
  default, `CASCADE` only where the child has no meaning alone, `SET NULL` where the
  relationship is optional. The choice is documented next to the column.
- Postgres does not index the referencing side: an unindexed foreign key is a lock and
  a scan on every parent delete.

## Constraints are named, and they are the documentation

- Name them (`orders_total_positive`), because the error a person gets names the
  constraint and that name is the first line of the bug report.
- `NOT NULL` is the default for a new column; nullable is a decision taken on purpose,
  and the application treats null as a case, not as an absence of one.
- `CHECK` carries the rule the application would otherwise repeat in three places: a
  positive total, a status in a known set, an end after a start.
- A default is a statement about the normal case, not a way to avoid choosing: a
  `DEFAULT now()` on a `created_at` is honest; a `DEFAULT ''` on a name is a lie.

## Soft delete, and why it usually backfires

`deleted_at` looks cheap and costs in three places: every query has to remember the
filter (and one will not), every unique index has to be partial (`WHERE deleted_at IS
NULL`), and the table grows with rows nobody may ever read. It is right when the person
can undo the delete or when the record must be kept for audit; it is wrong as a
substitute for deciding what happens to a row. When it is right, say the retention and
the purge job in the same migration.

## Names

Tables plural and `snake_case` (`reservations`), columns singular (`reserved_at`),
booleans read as a question (`is_active`, `has_paid`), timestamps `*_at`, money
`*_cents` or `amount` with its currency beside it, and no prefix that repeats the table
(`reservations.reservation_status` is `reservations.status`). A join table is
`a_b` in alphabetical order, so it is found by either side.

## What changes on another engine

SQLite has no real types (affinity), no `timestamptz` and no `numeric` decimal: money
becomes an integer of cents and time becomes UTC text or an integer. MySQL is closer,
but its default collation, its `utf8mb3` history and its silent truncation make the
explicit types worth spelling out. The rules that do not change: the constraint lives
in the database, and the type says what the value is.
