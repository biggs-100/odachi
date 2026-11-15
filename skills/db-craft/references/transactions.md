# The transaction is the boundary

A transaction is not "the code between BEGIN and COMMIT": it is the unit the database
can undo, the unit that holds locks, and the unit a retry has to replay whole. Three
questions decide it: what isolation, how long, and what happens when it loses a race.

## Isolation, by what it lets through

| level | a dirty read | a non-repeatable read | a phantom | a write conflict |
| --- | --- | --- | --- | --- |
| Read Committed (the default) | no | **yes** | **yes** | the second writer waits, then overwrites |
| Repeatable Read | no | no | no | `could not serialize access` |
| Serializable | no | no | no | `could not serialize access` |

- Read Committed is right for most requests: each statement sees a fresh snapshot, so
  two reads in one transaction can disagree. If the code reads, decides and writes, it
  is deciding on data that may already have moved.
- Repeatable Read is for a transaction that reads the same rows twice and needs the
  same answer. It does not prevent every anomaly; the write conflict it raises is the
  database protecting you.
- Serializable is for invariants that span rows no single lock covers -"the sum of the
  pending transfers never exceeds the balance"-. It is the only level where the
  database promises the result of some serial order, and it pays with retries.
- A read that decides nothing is read-only: `SET TRANSACTION READ ONLY` stops an
  accidental write and lets the planner pick a cheaper plan.

## The retry is the whole transaction

`could not serialize access due to concurrent update` (Repeatable Read, Serializable)
and `deadlock detected` are not bugs: they are the database refusing a history it
cannot make serial. The rule is the same for both:

```
for attempt := 0; attempt < 3; attempt++ {
    err := inTx(ctx, func(tx Tx) error { return doTheWork(ctx, tx) })
    if isRetryable(err) { backoff(attempt); continue }
    return err
}
```

- Retry **the whole function**, including the logic that decided which SQL to run: a
  statement retried alone runs against a different world.
- It must be safe to run twice, which means the body has no side effect that cannot
  repeat: see the outbox below.
- Bound the attempts and say what happens when they run out; an infinite retry is an
  outage with extra steps.
- A retry needs jitter, and it belongs where the transaction is opened, not inside the
  domain.

## Short, and with a fixed order

- No network, no model call, no email, no file inside a transaction. The lock waits for
  the slowest thing in the body, and that thing is not the database.
- Take locks in one order everywhere: two code paths that lock `orders` then
  `payments`, and `payments` then `orders`, produce a deadlock by design.
- `FOR UPDATE` when the row is about to change and the read decided it; `NOWAIT` when
  waiting is worse than failing; plain `SELECT` otherwise, because a lock you did not
  need is throughput you gave away.
- `idle_in_transaction_session_timeout` and `statement_timeout` on the role stop the
  transaction that forgot to close.

## A queue with `SKIP LOCKED`, and where it stops

```
SELECT id FROM jobs WHERE status = 'new'
ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1;
```

This is the right primitive for "any worker takes the next job": two workers never
claim the same row because the second skips what the first locked. It is **not** enough
when the invariant is global -"no more than three jobs running"-: two workers can each
see two running and both start a third. That count needs a lock on the counter row or
`pg_advisory_xact_lock`, released by the same commit.

## Effects that must not be lost, and not happen twice

Exactly-once delivery does not exist across a network; at-least-once plus an idempotent
handler is what does. Two patterns carry it:

- **The idempotency key**: the caller sends one, a table has `UNIQUE (key)`, and a
  repeated request finds the row and returns the stored result instead of paying twice.
  `INSERT … ON CONFLICT DO NOTHING` is the whole race handling.
- **The outbox**: the event is written to an `outbox` table in the same transaction as
  the state it describes, and a worker publishes after the commit, marking each row
  sent. Dual writes -the database, then the broker- have a window where one succeeded
  and the other did not, and no retry closes it.

## During a deploy

Two versions of the code run at once, so a transaction must tolerate both shapes: it
is the same reason the migration is expand/contract. A transaction that reads a column
the previous version still writes as null is a bug that only appears for ninety
seconds, on the deploy, at the worst time.

## What changes on another engine

MySQL's default is Repeatable Read and its `SKIP LOCKED` exists since 8.0; its
deadlock detection and retry surface is different but the rule is not. SQLite has one
writer at a time and serializes everything, which is simpler and does not scale past
one process: a long transaction there blocks every reader. The rule that does not
change: a transaction is short, retried whole, and never waits on the network.
