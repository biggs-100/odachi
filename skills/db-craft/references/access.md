# Who the database thinks you are

Every query runs as a role, and that role is the last line of defence: a bug that builds
the wrong `WHERE` still returns nothing if the database itself decided the rows are not
the caller's. The design work is choosing which role runs what, and how the tenant gets
into the transaction.

## Roles: three, not one

| role | can | cannot |
| --- | --- | --- |
| the owner (migrations) | DDL, grant, `BYPASSRLS` | serve application traffic |
| the application | `SELECT`, `INSERT`, `UPDATE`, `DELETE` on its tables | DDL, `BYPASSRLS`, read another schema |
| the reader (reports, analytics) | `SELECT`, often on a replica | write anything |

- The application never runs as the owner. `ALTER TABLE` from a request path is a
  migration in disguise, and the owner is the role that ignores row-level security.
- Grants are per schema and per table, not `GRANT ALL ON DATABASE`: a new table starts
  with no access and the migration that creates it grants what the feature needs.
- `GRANT` to a role, `GRANT` a role to a person; never share a login.

## Row-level security, and the four ways it does nothing

```
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices FORCE ROW LEVEL SECURITY;          -- the owner obeys too
CREATE POLICY tenant_isolation ON invoices
  USING (tenant_id = current_setting('app.tenant_id')::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);
```

**1. No `FORCE`.** Without it the table's owner -often the role that created it- reads
every tenant, and the policy on the screen proves nothing.

**2. The session variable on a pooled connection.** A pooled connection is reused, so:
set the variable with `SET LOCAL` **inside the transaction** that reads, never with a
plain `SET`. And define the setting: `SET LOCAL` on a custom variable that was never
defined reverts to an empty string, and an empty string cast to `uuid` is an error, not
a wildcard -treat unset as "no rows", never as "all rows".

**3. A filter narrower than the tenant.** A write is checked by `WITH CHECK`, and when a
policy omits it for `FOR ALL` or `FOR UPDATE` Postgres fills it from `USING`. That
default is why the defect is not the omission but a `USING` broader than the tenant
filter -or an explicit `WITH CHECK (true)` added to make a test pass-. A table whose only
policy is `FOR SELECT` refuses writes outright. Write the tenant predicate in both
halves on purpose: the default is a copy, and a copy of a wrong filter is still wrong.

**4. The doors that walk around every policy.** `SECURITY DEFINER` functions,
`BYPASSRLS` roles and a `row_security = off` setting ignore the policy entirely. Keep
them out of the request path and name who owns them.

## Choosing the multi-tenant shape

| shape | when | the cost |
| --- | --- | --- |
| one table, `tenant_id`, RLS | many small tenants, one schema to migrate | one policy mistake leaks everything |
| a schema per tenant | hundreds of tenants, per-tenant backup or restore | migrations run N times, the catalog grows |
| a database per tenant | compliance, residency, a loud tenant | connection count and operations multiply |

RLS is the right default for a SaaS with ordinary tenants. It is not a substitute for
the `WHERE`: it is the constraint that makes forgetting the `WHERE` harmless.

## Credentials and the wire

- The password lives in a secret manager or the environment, never in the repository,
  a fixture or a log. A connection string printed in an error is a credential leak with
  a timestamp.
- `sslmode=verify-full` in every environment that leaves the machine; `require` encrypts
  and trusts any certificate, which is not the same thing.
- Rotate on a schedule and after any leak; a rotated credential is cheap, an unrotated
  one is a decision to keep the exposure.
- Parameters, always: `WHERE id = $1`. String concatenation is the injection, and an
  identifier cannot be a parameter -that one is an allowlist of names, never a quote.

## What changes on another engine

MySQL has no row-level security: tenant isolation is a view, a `WHERE` in every query or
a schema per tenant, and the design must say which. SQLite has one file and no roles: its
isolation is the filesystem. The rule that does not change: the database decides who may
read what, and the application is not trusted to remember.
