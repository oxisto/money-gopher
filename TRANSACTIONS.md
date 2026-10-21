# Multi-write operations that are not in a transaction

A review of transaction use in `moneyd`, done from the lightsql side on
2026-08-22 against `d250a48`, as a list of findings with a suggested order.

**All four have since been fixed** (2026-10-21, `4aeb8e1`..). The findings are
kept below because they explain why the code looks the way it does; line
numbers in them refer to `d250a48`.

`internal/persistence.(*DB).Tx` already exists and does the right thing —
rollback if `fn` returns an error, commit otherwise, with sqlc's `WithTx`
rebinding the queries onto it:

```go
func (db *DB) Tx(ctx context.Context, fn func(q *Queries) error) error
```

It is called from exactly **two** places, both in `internal/api/security.go`:
`CreateSecurity` and the security update. Both are correct — a security plus its
identifiers and listings genuinely belong together.

Everything else writes on the bare pool. Most of that is fine: a single
statement gets an implicit transaction of its own, so `DeleteTransaction` and
friends are already atomic. The problem is the places where **several** writes
have to succeed or fail together and do not.

## 1. `ConfirmImport` — the one worth fixing first

> **Done.** Wrapped in `db.Tx` as described below; covered by
> `TestConfirmImport_RollsBackOnFailure` in `internal/api/documents_test.go`.

`internal/api/documents.go:93`, loop at `:128`, state update at `:207`.

```
for _, stx := range staged {      // line 128
    ...
    r.db.CreateTransaction(...)   // line 185 — commits immediately
}
r.db.UpdateDocumentState(...)     // line 207 — "IMPORTED"
```

Two ways this goes wrong, and neither reports anything useful:

- The loop fails partway. Some transactions are committed; the document is still
  `STAGED`. Re-running the import creates the earlier ones **again**.
- The loop succeeds and the state update fails. Every transaction is imported
  and the document does not know it. Same duplication on the next attempt.

This is the only finding here whose failure mode is silent duplicate financial
records, which is why it is first.

The fix is the existing helper. `created` is built inside the closure and read
after it, so it has to be declared outside:

```go
var created []*TransactionResolver
err := r.db.Tx(ctx, func(q *persistence.Queries) error {
    for _, stx := range staged {
        // ... unchanged, but calling q.CreateTransaction rather than r.db.*
    }
    _, err := q.UpdateDocumentState(ctx, ...)
    return err
})
if err != nil {
    return nil, err
}
```

Note the `TransactionResolver` values keep `db: r.db`, not `q` — the resolver
outlives the transaction and must read through the pool.

## 2. First-login provisioning leaves an account permanently broken

> **Done.** All three sites now call `db.CreatePersonWithAccess`, which does
> both writes in one `db.Tx`; covered by
> `TestCreatePersonWithAccess_RollsBackOnFailedGrant`. Accounts already left
> without a grant are not repaired by this.

`EnsureDevUser` at `internal/auth/auth.go:81` (writes at `:87`, `:100`, `:107`),
`EnsurePersonForUser` at `:121` (writes at `:130`, `:137`), and
`RootResolver.CreatePerson` at `internal/api/person.go:50` (writes at `:56`,
`:64`).

```
db.CreateUser(...)          // three separate commits
db.CreatePerson(...)
db.GrantPersonAccess(...)
```

If the grant fails, there is a person with no access to it. The next login takes
the `GetPerson` path, finds the person, and **skips the grant entirely** — so the
account does not heal on retry, it stays broken. That is worse than the import
case in kind, if rarer in practice.

All three sites are `CreatePerson` + `GrantPersonAccess` and want the same
treatment.

## 3. Two things to know before wrapping anything

**Isolation.** `db.Tx` passes `nil` options, which is `READ COMMITTED`: a fresh
snapshot per statement. lightsql honours `sql.TxOptions` rather than ignoring it,
so anything that reads a row and then writes based on what it read wants:

```go
db.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
```

Worth considering for the security update at `security.go:135`, which does
`GetSecurity` and then `UpdateSecurity`.

**Write conflicts are an error, not a wait.** lightsql is MVCC, so two
transactions writing the same row do not block — the second fails with SQLSTATE
`40001` (`serialization_failure`). `db.Tx` returns that straight to the caller.
`SetMaxOpenConns` is not set, so the pool is unbounded and concurrent GraphQL
mutations can genuinely reach it.

A retry is the standard answer, and belongs in `db.Tx` so no caller has to
remember:

```go
// A write conflict means another transaction got there first. Retrying is the
// documented response, and the whole closure is safe to re-run because the
// first attempt left nothing behind.
for attempt := range 3 {
    err := db.tx(ctx, opts, fn)
    var e interface{ SQLState() string }
    if errors.As(err, &e) && e.SQLState() == "40001" && attempt < 2 {
        continue
    }
    return err
}
```

`SQLState() string` is the interface `pgx` and `lib/pq` also satisfy, so this
code is not lightsql-specific.

## Suggested order

> All done. Retry on `40001`: `db.Tx` re-runs `fn` up to three times, covered
> by `TestTx_RetriesWriteConflict`. Isolation: `db.TxWith` takes
> `sql.TxOptions`, and the security update runs at `LevelRepeatableRead`,
> covered by `TestTxWith_RepeatableReadRetriesStaleRead`.

1. `ConfirmImport` (#1) — silent duplicates, real money.
2. Retry on `40001` in `db.Tx` (#3) — cheap, and wanted before more callers use it.
3. The three provisioning sites (#2).
4. Isolation level on the security update (#3) — only if the read-then-write
   there matters.
