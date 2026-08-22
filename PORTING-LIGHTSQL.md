# money-gopher on lightsql

**The port is done.** `internal/persistence` opens a lightsql directory, goose
runs on the `postgres` dialect, and both migrations are written for it. What
follows is the record of what was changed and why, kept because the *why* is not
recoverable from the diff.

Written 2026-08-21 while the port was still ahead of us; rewritten 2026-08-22,
after it landed in `d250a48`, to describe what is rather than what to do.

## Outstanding

Two things, unrelated to each other.

### Multi-write operations without a transaction

Reviewed separately in [TRANSACTIONS.md](TRANSACTIONS.md). The short version:
`db.Tx` exists and is used twice, and there are three places that need it and do
not have it. `ConfirmImport` is the one worth fixing first, because a partial
failure there duplicates financial records on the next attempt without saying so.

### The lightsql version

`go.mod` pins **`v0.4.0`**. Move it to `v0.6.0`:

```sh
go get github.com/oxisto/lightsql@v0.6.0
```

Two of the releases in between matter here:

- **v0.5.1** fixes `BYTEA` literals. `documents.data` is the one `BYTEA` column,
  and `CreateDocument` writes it as a **bound parameter**, which has stored real
  bytes correctly on every version — so nothing on disk is wrong. The fix only
  changes literals written inside SQL text, which this codebase does not do.
  Worth taking anyway, along with `length(bytea)` and `octet_length`.
- **v0.6.0** makes `NUMERIC` exact, and changes one thing: a numeric column
  scanned into an `any` is now a string rather than a `float64`. Scanning into a
  concrete `float64` or `string` is unaffected, and this schema has no `NUMERIC`
  column today — see below for why it might want one.
- **v0.5.0** adds a shell, which is the reason to bump beyond curiosity:

  ```sh
  go install github.com/oxisto/lightsql/cmd/lightsql@v0.5.1
  lightsql ./money.db
  lightsql> .schema documents
  ```

  Being able to open the directory `moneyd` wrote and ask what the migrations
  actually did is worth more than any amount of reasoning about them.

Also note `CURRENT_TIMESTAMP` works as of v0.3.0. `sessions.sql` was changed to
`now()` because it had to be at the time; both are correct now and there is no
reason to change it back.

## Worth considering: units is a REAL

`transactions.units` and `staged_transactions.units` are `REAL`, which is
float64. `units × price` is the core computation in this application, and a
float cannot represent most decimal fractions exactly — accumulating those
products is where a portfolio total drifts by a cent and nobody can say from
where.

Money is already safe: `price`, `fees`, `taxes` and `cash_delta` are `INTEGER`
cents, which is the other correct answer. It is only `units` that is inexact.

As of lightsql v0.6.0 the alternative exists:

```sql
ALTER TABLE transactions ADD COLUMN units_exact NUMERIC(18,8);
UPDATE transactions SET units_exact = CAST(units AS NUMERIC(18,8));
-- and later, once nothing reads the old column
```

Two things to weigh before doing it. lightsql has no `DROP COLUMN`, so the old
`units` would linger the way `user_id` does. And a `NUMERIC` column arrives in
Go as a string, so the sqlc-generated field changes type and every caller of it
has to be looked at — which is a real change, not a mechanical one.

Not urgent, and not something the port created. Flagged because the schema is
being read carefully right now and this is the kind of thing that is easier to
change before there is a lot of data than after.

## What was changed, and why

| Change | Reason |
|---|---|
| `sql.Open("lightsql", "file:"+path)` | The `file:` prefix is load-bearing. Without it the name selects an in-memory instance, which works until the process restarts and the data is gone. |
| `lightsql.Drop(dsn)` on close | Checkpoints the log, so the next start replays the rows that survived rather than every change ever made. Skipping it is safe, only slower. |
| `_pragma=foreign_keys(1)` dropped | lightsql always enforces foreign keys and has no switch. This is what forced migration 2 to be rewritten. |
| `_time_format=sqlite` dropped | lightsql stores a `time.Time` natively, so there is no format to agree on. |
| `goose.SetDialect("postgres")` | Every statement goose's stock postgres dialect issues is supported as of v0.4.0, and pinned upstream by a regression test. |
| `DATETIME`→`TIMESTAMP`, `BLOB`→`BYTEA`, `STRFTIME(...)`→`now()` | Dialect. Mechanical. |
| Migration 2 rewritten | SQLite's `PRAGMA foreign_keys = OFF` table-rebuild idiom cannot work where foreign keys cannot be switched off. It uses `ADD COLUMN`, backfill, then `SET NOT NULL` — which is what a PostgreSQL migration would have looked like from the start. |
| `sqlc.yaml` left on `engine: "sqlite"` | The queries are dialect-clean and all 150 placeholders are `?`, which lightsql accepts. Generated code is untouched. |

The old `user_id` columns survive because lightsql has no `DROP COLUMN`. They are
still `NOT NULL`, so the application has to keep writing them. That is the one
piece of debt the port took on.

## Historical record

Everything below describes the port as it was planned, before it was carried
out. It is kept for the reasoning, not as instructions.

### What this port was, and was not

It is a **dialect** port, not just a driver swap: lightsql speaks the PostgreSQL
dialect and the schema is written for SQLite.

The good news is that the split is lopsided. The **queries are already
dialect-clean** — no `DATETIME`, `BLOB`, `STRFTIME` or `PRAGMA` in any of them,
and all 150 placeholders are `?`, which lightsql accepts alongside `$N`. So:

- `sqlc.yaml` stays as it is (`engine: "sqlite"`).
- The generated code in `internal/persistence/*.sql.go` stays as it is.
- Only the **migrations** need real work, plus two lines in `sessions.sql`.

## The work, in order

### 1. Swap the driver — `internal/persistence/persistence.go`

One call site, line 35.

```go
_ "modernc.org/sqlite"                  →  _ "github.com/oxisto/lightsql/driver"

sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&...")
                                        →  sql.Open("lightsql", "file:"+path)
```

**The `file:` prefix is not decoration.** Without it the name selects an
in-memory instance, and everything works right up until the process restarts and
the data is gone. `path` becomes a *directory* rather than a file — lightsql
creates it, and puts a single `wal` file inside.

Every committed transaction is appended to that log and flushed before the commit
returns, so a crash loses at most transactions that had not been acknowledged.
`?fsync=off` trades that guarantee for speed and is worth it for the test suite,
not for anything else.

Notes on what falls away:

- The `_pragma=foreign_keys(1)` argument is unnecessary: lightsql always
  enforces foreign keys and has no way to turn them off. **This matters for
  migration 2 — see step 3.**
- `_time_format=sqlite` is unnecessary. lightsql stores a `time.Time` natively
  rather than as formatted text, so there is no format to agree on and no
  lexicographic-sorting trick needed.
- `conn.SetMaxOpenConns(1)` can go. It is there because SQLite allows one writer
  and because `:memory:` databases are per-connection. lightsql is MVCC with a
  shared instance behind a DSN, so pooled connections are the expected case.
  Leaving it set is harmless if you would rather change one thing at a time.
- For tests, drop the `file:` prefix: `sql.Open("lightsql", "moneyd")` is a named
  in-memory instance, and `lightsqldriver.Drop("moneyd")` releases it between
  tests. Passing `?fsync=off` on a directory-backed test database is the other
  option if a test needs to exercise a restart.
- Call `lightsqldriver.Drop("file:"+path)` on shutdown. It closes the database
  and compacts the log, so the next start replays the rows that survived rather
  than every change ever made. Skipping it is safe — recovery is identical
  either way — it is just slower to open.

### 2. `sessions.sql` — nothing to change

This step used to say the two `CURRENT_TIMESTAMP` occurrences had to become
`now()`. **lightsql implements `CURRENT_TIMESTAMP` now**, along with
`CURRENT_DATE`, `CURRENT_TIME`, `LOCALTIMESTAMP` and `LOCALTIME`, all reporting
the transaction's start time as PostgreSQL does. The queries are unchanged, and
`go generate ./...` does not need re-running.

### 3. Migrations — the actual work

**`0001_schema.sql`** needs only mechanical substitutions:

| SQLite | lightsql | occurrences in `0001` |
|---|---|---|
| `DATETIME` | `TIMESTAMP` | 12 |
| `BLOB` | `BYTEA` | 1 |
| `DEFAULT (STRFTIME('%Y-%m-%dT%H:%M:%fZ'))` | `DEFAULT now()` | 6 |
| `DEFAULT CURRENT_TIMESTAMP` | *(leave it)* | 1 |

A `sed` pass handles the first two; the `STRFTIME` default is a fixed string, so
it substitutes cleanly too. `CURRENT_TIMESTAMP` needs no change at all now.

The one `BLOB` column, on `documents`, is worth a second look. lightsql writes a
`BYTEA` as a string literal in PostgreSQL's format — `'\x0102'` for two bytes —
whereas SQLite takes `X'0102'`. Nothing in `queries/` writes one as a literal;
they all bind a parameter, which carries real bytes and needs no change. If you
add a literal later, write it PostgreSQL's way. **This only works on `v0.5.1`
and later**: earlier versions stored the characters that spell the bytes rather
than the bytes.

`TEXT`, `INTEGER` and `REAL` are common to both. Everything else in this file —
including the partial unique index
`ON security_identifiers(kind, value) WHERE kind IN ('ISIN','WKN')` — works
as written and is enforced.

**`0002_persons.sql` needs rewriting**, and this is the only non-mechanical part
of the port.

It currently uses SQLite's table-rebuild idiom: `PRAGMA foreign_keys = OFF`,
then for each of `portfolios`, `cash_accounts` and `documents`: create a `_new`
table, copy the rows across, drop the original, rename. That only works because
foreign key enforcement can be switched off — lightsql has no such switch, so
`DROP TABLE portfolios` is correctly refused while `transactions` still
references it.

Write it the way a PostgreSQL migration would be, using `ADD COLUMN`:

```sql
-- +goose Up
CREATE TABLE persons (
    id           TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    created_at   TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE user_person_access (
    user_id   TEXT NOT NULL REFERENCES users(id)   ON DELETE CASCADE,
    person_id TEXT NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, person_id)
);

INSERT INTO persons (id, display_name) SELECT id, display_name FROM users;
INSERT INTO user_person_access (user_id, person_id) SELECT id, id FROM users;

ALTER TABLE sessions      ADD COLUMN person_id TEXT REFERENCES persons(id) ON DELETE SET NULL;
ALTER TABLE portfolios    ADD COLUMN person_id TEXT REFERENCES persons(id) ON DELETE CASCADE;
ALTER TABLE cash_accounts ADD COLUMN person_id TEXT REFERENCES persons(id) ON DELETE CASCADE;
ALTER TABLE documents     ADD COLUMN person_id TEXT REFERENCES persons(id) ON DELETE CASCADE;

UPDATE portfolios    SET person_id = user_id;
UPDATE cash_accounts SET person_id = user_id;
UPDATE documents     SET person_id = user_id;

-- Now that every row has a value, tighten the columns that should be required.
ALTER TABLE portfolios    ALTER COLUMN person_id SET NOT NULL;
ALTER TABLE cash_accounts ALTER COLUMN person_id SET NOT NULL;
ALTER TABLE documents     ALTER COLUMN person_id SET NOT NULL;
```

The three `SET NOT NULL` statements are new: lightsql did not have them when
this document was first written, which meant `person_id` would have stayed
nullable at the schema level forever and sqlc would have generated
`sql.NullString` for a column the application always populates. They are checked
against the rows already there, so they will fail loudly if the backfill above
missed any.

This is shorter and clearer than the rebuild, and it is what the schema would
have looked like had it targeted PostgreSQL from the start.

Two consequences worth deciding on deliberately:

- The old `user_id` columns survive rather than being dropped, because lightsql
  does not implement `DROP COLUMN` (see limitations). Leaving them is harmless;
  if you want them gone, that is a reason to ask for the feature. Note they are
  still `NOT NULL`, so the application has to keep writing them.
- Rows written before the `ADD COLUMN` read `person_id` as NULL until the
  `UPDATE` statements run — which is why those statements are there.

### 4. goose — the one genuine unknown

`persistence.go` calls `goose.SetDialect("sqlite3")`. Change it to:

```go
goose.SetDialect("postgres")
```

Every query goose's stock postgres dialect issues is now supported, and each was
checked against lightsql directly:

| goose does | status |
|---|---|
| `CREATE TABLE goose_db_version (id integer PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY, ...)` | works verbatim |
| `SELECT EXISTS (SELECT 1 FROM pg_tables WHERE tablename = ...)` | works, unqualified |
| `INSERT INTO goose_db_version (version_id, is_applied) VALUES ($1, $2)` | works |
| `SELECT version_id, is_applied FROM goose_db_version ORDER BY id DESC` | works |
| `DELETE FROM goose_db_version WHERE id = $1` | works |

So the custom `dialect.Querier` that dropped the surrogate `id` column is no
longer needed, and neither is applying the migrations by hand.

Those statements are now a regression test in lightsql
(`driver/goose_test.go`), run in the order goose runs them, so the
combination stays working rather than each piece working separately.

**This is still the one step I cannot verify from the lightsql side.** The test
pins the statements, not goose: if goose issues something I did not think to
look for, it will not catch it. That is a bug report worth filing rather than a
workaround worth writing. The fallback remains: apply the migrations with a few
lines of your own.

Do this step first if you want to know quickly whether the port is smooth or
fiddly.

One thing that now matters more than it did: **DDL is not transactional and is
written to the log as it runs.** A migration that fails halfway leaves the
statements it already applied both in memory and on disk, so re-running it hits
"relation already exists" rather than starting clean. That is also true of
PostgreSQL-minus-transactional-DDL engines like MySQL, and goose's usual answer —
one migration per file, each idempotent where it can be — applies unchanged.

## Verified against lightsql

Every one of these was executed against a real lightsql instance, not reasoned
about:

- All 65 generated queries bind.
- `GrantPersonAccess` — `ON CONFLICT DO NOTHING`, run twice, inserts once.
- `ListPersonsForUser` — the `persons`/`user_person_access` join, scanned into
  `string`/`time.Time`.
- `CreateCashAccount` — `INSERT ... RETURNING *`, including a nullable `iban`
  and the `ALTER`-added `person_id`.
- `GetCashAccountBalance` — `CAST(COALESCE(SUM(cash_delta), 0) AS INTEGER)`,
  returning 0 over an empty table and the correct sum over a populated one.
- `CreateQuote` — the upsert keyed on `(listing_id, time)` with a `time.Time`,
  replacing rather than duplicating.
- The partial unique index — refuses a duplicate ISIN across securities, allows
  duplicate tickers.
- `ON DELETE CASCADE` through a column added by `ALTER TABLE`, and onward two
  levels (`persons → cash_accounts → transactions`).
- A schema, its rows, its sequences, its `CHECK` and foreign-key constraints and
  a partial unique index all coming back after the process that wrote them is
  gone — including from a directory copied mid-write, with nothing checkpointed.

## lightsql limitations you will meet

Not bugs — deliberate, each with a stated reason in the compatibility matrix.

| Missing | Consequence here |
|---|---|
| `DROP COLUMN`, column type changes | The old `user_id` columns stay after migration 2. `SET`/`DROP NOT NULL` do work. |
| `OVERRIDING SYSTEM VALUE` | So a `GENERATED ALWAYS` identity column cannot be given an explicit value at all. Prefer `GENERATED BY DEFAULT` unless you mean it. |
| `DROP TABLE ... CASCADE` | Drop the referencing table first, or name both in one statement. |
| `PRAGMA` | No way to disable foreign keys; this is what forces the migration 2 rewrite. |
| `information_schema` / `pg_catalog` | Present but partial: `tables`, `columns`, `table_constraints`, `key_column_usage`, `pg_tables`, `pg_namespace`, `pg_class`, `pg_attribute`, with the columns tools actually read. Unqualified names find `pg_catalog` only, as in PostgreSQL. |
| `INTERVAL` | No interval arithmetic. `CURRENT_TIMESTAMP` and `now()` both work. |
| DDL is not transactional | A failed migration does not roll back its DDL. Matters for `goose.Up` on a half-applied migration. |
| Renaming a column named by a `CHECK`, `DEFAULT` or partial index predicate | Refused. Not hit by this schema. |
| Log compaction | Only at close. A long-running process that is killed rather than shut down accumulates a log until it next exits cleanly. |
| Multi-process access | Nothing stops a second process opening the same directory, and the result would be two engines overwriting each other. There is no lock file yet. |

Neither of those last two is a problem for a single `moneyd` process, but both
are worth knowing before pointing a second one at the same directory.

## Suggested order

0. `go get github.com/oxisto/lightsql@v0.5.1`. Earlier tags are not enough.
1. goose dialect (step 4) — one line, and no longer expected to fight back.
2. Driver swap (step 1) — and mind the `file:` prefix.
3. Migration 1 substitutions (step 3).
4. Migration 2 rewrite, including the three `SET NOT NULL` (step 3).
5. Run the test suite. Anything that fails here is new information — please
   report it upstream rather than working around it. Every issue raised from
   this port so far has been a real one: three became features (#39, #40, #42)
   and one became a documentation fix (#43). The only one that did not
   reproduce, #41, turned out to have been filed against `v0.1.0` rather than
   `v0.2.0` — worth checking which version is in `go.mod` before writing the
   next one up.
