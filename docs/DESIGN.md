# Money Gopher — Redesign

> Status: decided 2026-07-04; M0–M6 done, see "Status" below. Developed as
> "v2" on a branch, merged into `main` in October 2026; releases continue
> as v0.x, the next being v0.3.0.
>
> This document captures the ground-up redesign of
> money-gopher. Everything from v1 is replaced except the gopher logo (`img/gopher.png`)
> and the domain knowledge encoded in the old code.

## Goals

- Self-hosted investment tracker: portfolios, bank accounts, securities, transactions,
  performance over time.
- Multi-user from day one (households: several people, shared securities/quotes).
- Import transactions from bank documents — primarily **PDF** (statements, buy/sell
  confirmations), CSV as a secondary path.
- Single deployable Go binary that eventually embeds the frontend.
- Clear, typed API schema with auto-generated clients on both ends.

## Non-goals

- Multi-tenant SaaS scale, horizontal scaling, non-SQLite databases.
- Real-time trading / brokerage integration. This is bookkeeping, not execution.

## Tech stack (decided)

| Area | Choice | Notes |
|---|---|---|
| Backend | Go (single module `github.com/oxisto/money-gopher`) | binaries: `moneyd` (server), `mgo` (CLI) |
| API | **GraphQL** via [`graph-gophers/graphql-go`](https://github.com/graph-gophers/graphql-go) | schema-first `.graphql` file, handwritten resolver structs, schema↔resolver match verified at startup. Chosen deliberately over REST/huma — nested parameterized reads fit the domain, and giving GraphQL a real try is an explicit goal. |
| Uploads | Plain `POST /upload` REST endpoint | GraphQL multipart is not worth it; documents go through a normal endpoint, results are queried via GraphQL. |
| Frontend | SvelteKit **SPA** (`adapter-static`, SSR off) + **Tailwind CSS 4** | lives in `ui/`; embedded into `moneyd` via `go:embed` in a later milestone |
| GraphQL client | [gql.tada](https://gql-tada.0no.co) + `fetch` | typed documents inferred from the schema, loaded in SvelteKit `load` functions (replaced Houdini in October 2026, see Status) |
| Go GraphQL client (CLI) | [`Khan/genqlient`](https://github.com/Khan/genqlient) | typed Go client for `mgo` from the same schema |
| Persistence | SQLite + [sqlc](https://sqlc.dev) | typed query code generated from SQL; migrations via goose (as in v1) |
| PDF extraction | poppler `pdftotext -layout` (external binary) + per-bank Go parsers | LLM-based extraction (Claude API) as optional fallback for unknown formats |
| Auth | Multi-user, **OIDC** (pluggable provider) | bundled dev/simple mode so self-hosters don't need an IdP to try it |

## Architecture

```
┌──────────────────────── moneyd (one binary) ────────────────────────┐
│                                                                     │
│  /graphql   GraphQL API (graph-gophers/graphql-go)                  │
│  /upload    document upload (multipart REST)                        │
│  /auth/*    OIDC callback / session handling                        │
│  /*         embedded SvelteKit SPA (go:embed, later milestone)      │
│                                                                     │
│  internal/finance    snapshot & performance calculations            │
│  internal/quotes     pluggable quote providers                      │
│  internal/importer   extract → detect bank → parse → stage → commit │
│  internal/persistence  sqlc-generated queries over SQLite           │
└─────────────────────────────────────────────────────────────────────┘
```

Everything is request-driven except quote updates (periodic background job) and
document processing (async job after upload, status queryable via GraphQL).

## Domain model

Redesigned from scratch (deliberately *not* v1's schema, which had known pain
points: decorative bank accounts, per-field currency pairs, no quote history,
ISIN as primary key). Core idea: **typed events with an explicit cash leg** —
every transaction that moves money names the cash account it settles against
and its signed effect on that account, so account balances are always
derivable as `SUM(cash_delta)`.

- **User** — OIDC identity (issuer + subject), provisioned on first login.
  Owns portfolios and cash accounts.
- **CashAccount** — a real bank/cash account with a denomination currency.
  Balance is derived, never stored.
- **Portfolio** — a collection of security positions. Cash is deliberately
  not part of a portfolio; transactions link the two.
- **Security** — global (shared across users), opaque surrogate ID. External
  identifiers (ISIN, WKN, tickers) live in **SecurityIdentifier** rows —
  imports match on whatever the bank prints. ISIN/WKN are globally unique.
- **Listing** — a security on an exchange: ticker, currency, quote provider.
  The thing that has a price.
- **Quote** — append-only price history per listing; "latest quote" is just
  the newest row. First-class from day one for charts and snapshots.
- **Transaction** — typed event (BUY, SELL, DELIVERY_IN/OUTBOUND, DIVIDEND,
  INTEREST, DEPOSIT_CASH, WITHDRAW_CASH, ACCOUNT_FEES, TAX_REFUND) with
  units, price, fees, taxes, and the cash leg (`cash_account_id` +
  `cash_delta`). Single currency per transaction. The server derives the
  cash delta for trades/dividends; pure cash events state it explicitly.
  `source` records provenance (manual vs. import).
- **Document** (M5) — an uploaded file (PDF/CSV): raw bytes, extraction
  state, detected bank; staged transactions reference it.
- **Snapshot / Position** — computed, not stored: portfolio value, positions,
  performance at a point in time (from `internal/finance`, M3).

Money is stored as integer minor units (cents) + ISO 4217 code — never
floats. Share quantities are floats (fractional units exist). In GraphQL,
money surfaces as a `Money { amount, currency }` type.

## GraphQL schema (sketch)

`api/schema.graphql`, embedded via `go:embed`, is the single source of truth.
gql.tada (frontend) and genqlient (CLI) both derive their types from it.

```graphql
scalar Time
scalar Money   # { amount: Int (minor units), currency: String }

type Query {
  me: User!
  portfolios: [Portfolio!]!
  portfolio(id: ID!): Portfolio
  securities(filter: String): [Security!]!
  documents(state: DocumentState): [Document!]!
}

type Portfolio {
  id: ID!
  displayName: String!
  transactions(after: Time, before: Time): [Transaction!]!
  snapshot(time: Time): PortfolioSnapshot!   # defaults to now
}

type PortfolioSnapshot {
  time: Time!
  totalValue: Money!
  positions: [Position!]!
  performance(period: Period!): Performance!
}

type Mutation {
  createPortfolio(input: CreatePortfolioInput!): Portfolio!
  createTransaction(input: CreateTransactionInput!): Transaction!
  # ... updates/deletes
  confirmImport(documentID: ID!, input: ConfirmImportInput!): [Transaction!]!
  triggerQuoteUpdate(securityIDs: [ID!]): QuoteUpdateResult!
}
```

N+1 note: graph-gophers exposes the selected field set to resolvers; list resolvers
prefetch child rows in one query where it matters (positions → securities → quotes).

## PDF import pipeline

```
upload → store Document → extract text (pdftotext -layout)
       → detect bank (fingerprint match on text)
       → parse (per-bank Go parser)  ──fail──→ optional LLM extraction (Claude API)
       → staged transactions (user reviews in UI, maps securities/accounts)
       → confirmImport mutation commits them
```

- `internal/importer/extract` — wraps the `pdftotext` binary behind an interface, so a
  pure-Go or LLM extractor can be swapped in. `pdftotext` presence is checked at startup
  and surfaced in the UI, not a hard crash.
- `internal/importer/banks/<bank>` — one package per supported bank format. Each parser
  declares a fingerprint (strings/regex that identify its documents) and returns staged
  transactions. Registry pattern; adding a bank is one package + tests against fixture PDFs.
- LLM fallback is **opt-in** (config: API key present): extracted text (or the PDF) is sent
  to the Claude API with a structured-output schema matching staged transactions. Never
  silently — documents state which path produced their data.
- Staged transactions are never auto-committed; the review step is where security/account
  mapping and duplicate detection happen.

## Auth

- OIDC authorization-code flow handled by `moneyd` (`/auth/*`); server-side session
  (cookie) so the SPA never touches tokens.
- Users are provisioned on first login (subject + issuer → `users` row).
- Dev mode (`--auth=dev`): a built-in fake issuer with a fixed user, so `moneyd` runs
  standalone. (Same spirit as v1's oauth2go setup, but simpler.)
- All GraphQL resolvers are scoped to the session user; securities/quotes are global.

## Repository layout (target)

```
cmd/moneyd/            server entrypoint
cmd/mgo/               CLI (genqlient-based)
api/schema.graphql     GraphQL schema (source of truth)
internal/api/          resolvers, GraphQL/HTTP wiring, upload handler
internal/auth/         OIDC + sessions + dev mode
internal/persistence/  sqlc output, migrations (goose), db bootstrap
internal/finance/      snapshot & performance calculations
internal/quotes/       QuoteProvider interface + implementations
internal/importer/     extract/, banks/, llm/, staging
ui/                    SvelteKit SPA (Tailwind 4, gql.tada)
img/gopher.png         the one survivor
docs/                  this file, ADRs as needed
```

Deleted from v1: all buf/proto/Connect artifacts (`buf*.yaml`, `mgo.proto`, `gen/`,
`openapi.yaml`), `server/`, `service/`, `cli/`, old `persistence/`, `finance/`,
`import/`, the Next.js `ui/`.

## Milestones

1. ✅ **M0 — Teardown & scaffold.** Remove v1 code (keep `img/`, LICENSE, community files),
   new directory layout, Go module tidy, CI (build + test + sqlc/vet), README rewrite.
2. ✅ **M1 — Core backend.** SQLite schema + migrations + sqlc (users, portfolios, cash
   accounts, securities, transactions). GraphQL schema v1 with portfolio/security/
   transaction CRUD. Dev auth. `moneyd` serves `/graphql` with GraphiQL at `/graphiql`.
3. ✅ **M2 — Frontend scaffold.** SvelteKit SPA + Tailwind 4 + Houdini against the running
   backend: portfolio list, portfolio detail, transaction entry.
4. ✅ **M3 — Finance engine.** Port/redo snapshot & performance calculations (positions,
   market value, gains, time-weighted return). `snapshot(time:)` resolver + UI dashboard.
5. ✅ **M4 — Quotes.** QuoteProvider interface, first provider, background refresh,
   quote history table for charts.
6. ✅ **M5 — PDF import.** `/upload`, document storage, pdftotext extraction, bank
   detection, first real bank parser (whichever bank's PDFs we have fixtures for),
   staging + review UI, `confirmImport`. CSV import is out of scope — PDF covers
   the real import path.
7. ✅ **M6 — Real auth.** OIDC against a real provider, user provisioning, ownership
   enforcement tests.
8. **M7 — Ship it.** `go:embed` the built SPA, single-binary release (goreleaser),
   Dockerfile, docs. `mgo` CLI rebuilt on genqlient for the endpoints that matter.
9. **M8 — LLM fallback extraction** (opt-in), more bank parsers as needed.
10. **M9 — Assets.** A security is one kind of asset. The office room rented
    out, the car, a private loan: an asset with a name and a kind, a price
    series (a listing with a provider for a security, manual valuations for
    the rest), ownership and cost basis as a delivery-in transaction, income
    against it like a dividend. The engine values it through the same quote
    lookup; nothing there changes. Done: kinds beyond securities, manual
    valuations with an entry form (#658); ownership — `securities.person_id`,
    visible to the owner alone, "shared or mine" in every lookup by ID or
    ISIN — the `bookAsset` mutation and the Assets page, with personal
    assets held in one "Other assets" portfolio per person
    (`portfolios.purpose = 'ASSETS'`, created on first use); the naming:
    the API type is `Asset` with `AssetKind`, `createAsset`/`updateAsset`/
    `deleteAsset`, `asset` on positions, transactions and staged rows,
    `assetID` in inputs; `securities` (shared, listed) and `assets` (own)
    stay the two list queries, and the Securities page keeps its name. The
    database keeps `securities`/`security_id` — renaming would churn every
    foreign key and query for no behaviour.
11. **M10 — Recurring transactions, accounts and cash flow.** Gets its own
    planning pass before code. Three things that are one mechanism:
    - *Schedules*: amount, interval, start, optional end, sign, optionally
      the asset or account they belong to — rent coming in from the room, a
      salary (Gehaltseingang), a savings plan going out, average living
      costs (the credit card) as a rough outflow. Expanded into expected
      cash flows and matched against imported statements, so expected and
      actual line up and a missed month shows.
    - *Accounts with kinds, which may go negative*: a loan is an account
      whose balance is a ledger (disbursement, interest charged,
      repayments); a Bausparvertrag is a savings account (Sparrate in,
      interest credited). Net worth sums accounts with their sign plus the
      holdings; equity in the house is its value minus the financing.
      Transfers link a Giro withdrawal and a loan or savings deposit as one
      event. The worked example is the user's Bausparvorfinanzierung: an
      endfälliges Darlehen (interest only, principal at maturity) plus a
      Bausparvertrag that at Zuteilung repays it, partly from its Guthaben
      and partly as an amortising Bauspardarlehen — all ordinary events.
    - *A rough cash-flow view*: expected in versus out per month. Rough by
      design; the app must not become a household expense tracker.
    Needed by the plan in M12.
12. **M11 — Agent and chat.** A backend agent against an OpenAI-compatible
    model (self-hosted first): model client with tool calling and streaming,
    a tool registry (internal: holdings, performance, cash, net worth history,
    quotes, assets; external: web search, page fetch, dividend data,
    instrument search, all cached with a timestamp), and a run executor with
    a persisted step log and limits. Runs stream over SSE; a chat page per
    person starts runs and renders their steps; a finished run is a stored,
    reopenable report. Every number in a report comes from a tool result.
    Candidates the agent researches land on a watchlist with their facts.
    The same tools serve an MCP endpoint for Claude Code.
13. **M12 — Plans and simulation.** A plan per person: starting amount,
    monthly contribution, allocation, horizon, cost assumptions. A
    deterministic path and a Monte Carlo fan from per-kind return and
    volatility assumptions (editable), plan versus actual, and the agent
    creating or editing plans through a tool. Then the German tax layer
    (Sparer-Pauschbetrag, Vorabpauschale, Freistellungsauftrag) and inflation.

**Where M5 landed:**

- `internal/importer/`: extraction via `pdftotext -layout` behind an `Extractor`
  interface (`PlainText` for text files, `PDFToText` wrapping the binary). Bank
  detection via `Matches()` fingerprint methods; registry in `DefaultParsers()`.
- ING-DiBa parser (`ing.go`): handles Kauf/Verkauf (BUY/SELL), Dividende/Ertrag
  (DIVIDEND), Zinsen (INTEREST), Rückzahlung (bond maturity → SELL),
  Wertpapier Eingang (spin-off → DELIVERY_INBOUND). Vorabpauschale and Storno
  return `ErrSkipped` → document state `SKIPPED` (not FAILED).
- `SKIPPED` added to `DocumentState` enum and `documents.state` CHECK constraint.
- `transaction_date` column on `documents`: extracted from first staged transaction
  during processing; used to sort the import list chronologically (`NULLS LAST`
  for FAILED/SKIPPED).
- `iban` on `cash_accounts`; `settlement_iban` on `documents`. After parsing,
  settlement IBAN is stored on the document; `suggestedCashAccount` resolver
  performs a live IBAN lookup so the review UI can auto-select the right account.
- `StagedTransaction.Security` resolver and `ConfirmImport` both do a live ISIN
  lookup when `security_id` is NULL — securities created after a document was
  processed are matched without reprocessing.
- Review UI (`DocumentReview.svelte`): split-pane PDF + form; portfolio and cash
  account selectors; inline "Create security" modal pre-filled with name + ISIN.
- Document list (`DocumentList.svelte`): cards show transaction date, type,
  security name, and cash delta; filename demoted to secondary line.
- `cmd/reprocess/`: CLI tool for bulk re-uploading all PDFs in a directory tree
  (used to populate a fresh DB from a folder of bank statements).
- CSV import is out of scope — PDF pipeline covers the real import path.
- Integration test `TestImportPipeline` in `internal/api/import_test.go` covers
  the full path: upload → process → IBAN match → confirm → balance.

**Where M6 landed:**

- `internal/auth/oidc.go`: `OIDCHandler` with `LoginHandler` (generates state+nonce
  cookies, redirects to IdP), `CallbackHandler` (verifies state/nonce, exchanges code,
  provisions user via `provisionUser`, creates session, redirects to `/`), and
  `LogoutHandler` (clears session, redirects to `/login`).
- `internal/auth/session.go`: `CreateSession`/`ClearSession` helpers and
  `SessionMiddleware` — resolves the `mg_session` cookie to a user on every request;
  returns 401 if missing or expired. Sessions are stored in the `sessions` table (goose
  migration `0004_sessions.sql`), TTL 30 days.
- `cmd/moneyd/main.go`: three auth modes via `--auth` flag:
  - `dev` (default): fixed dev user, no credentials needed.
  - `builtin`: embedded `oauth2go` authorization server, mounted on moneyd's own
    listener under `/auth/builtin/` with `--auth-user` / `--auth-password`, so one
    port and one `--public-url` cover the app and its login page — no external IdP
    required for self-hosting. The browser is sent to `<public-url>/auth/builtin/authorize`;
    the code exchange goes to the local listener directly. Users are provisioned under
    the fixed issuer `urn:money-gopher:builtin`, so changing the public URL keeps them.
  - `oidc`: external provider via `--oidc-issuer`, `--oidc-client-id`,
    `--oidc-client-secret`.
  All protected routes are wrapped with `SessionMiddleware`; OIDC routes
  (`/auth/login`, `/auth/callback`, `/auth/logout`) are registered for both
  `builtin` and `oidc`.
- `ui/src/client.ts`: monkey-patches `window.fetch` to redirect to `/login` on any
  401 — catches session expiry for all GraphQL and REST calls without per-call
  error handling.
- `ui/src/routes/login/+page.svelte`: standalone sign-in card (no AppNav), links to
  `/auth/login` to start the authorization-code flow.
- AppNav: "Sign out" link to `/auth/logout` always visible; in dev mode it just
  calls the logout handler which clears the dev session cookie (harmless).

## Status / handoff notes (updated 2026-10-21)

Work happens on `main`. M0–M6 are done. M7 (ship it): the SPA is embedded
(`ui/ui.go`, `go:embed` of `ui/build`), and a release publishes a Docker
image (`Dockerfile`, `.github/workflows/release.yml`; built but not pushed on
every PR by the `docker` job in `build.yml`). Decided against downloadable
binaries — the image is the distribution. Publishing v0.3.0 is a matter of
creating the GitHub release. The `mgo` CLI rebuild from the M7 plan is
deferred.

**Since M6 (August–October 2026):**

- Persistence moved from SQLite to lightsql — see `PORTING-LIGHTSQL.md`;
  multi-write operations run in `db.Tx`, which retries write conflicts —
  see `TRANSACTIONS.md`.
- ING Kontoauszug (bank statement) importer, FX for foreign-currency bonds,
  re-analysis of documents with reconciliation of already-imported rows.
- `Query.netWorthHistory` for the overview chart (net worth against
  invested capital over a `Period`).
- **UI redesign** ("calm & dense"): design tokens with light/dark theme in
  `ui/src/app.css`, a sidebar shell, and rebuilt pages — Overview,
  Holdings (+ a page per holding with trades on the price chart),
  Activity (one feed across accounts; also on cash account and portfolio
  pages), Import inbox, Income (`/gains` redirects), Securities, People.
  Shared pieces: `ui/StatStrip`, `Panel`, `Segmented`, `Dialog`, `Icon`,
  `describeTransaction`/`labelOf`, `formatSignedMoney`/`toneOf`.
  Open: ⌘K palette, a combined household view (needs backend support),
  closed positions on Holdings, a daily-change column (needs
  previous-day quotes).
- **Houdini replaced by gql.tada.** Houdini tied the SvelteKit version to
  its own (it blocked SvelteKit 3), ran a second codegen beside kit's, and
  its normalized cache was mostly worked around with forced refetches.
  Now: documents are `graphql()` calls typed from `api/schema.graphql`
  (`ui/src/lib/graphql/graphql-env.d.ts`, refreshed by `npm run generate`
  and checked in CI), `query()`/`mutate()` in `ui/src/lib/graphql/client.ts`,
  data in ordinary `+page.ts` load functions, and `refreshAll()` after a
  mutation. View state that should reload data lives in the URL
  (`/?range=1Y`, `/income?year=2025`).
- **Currencies.** A security has no currency; its listings do, and that is
  the currency of their quotes. The finance engine works in the currency the
  transactions settle in: a foreign-currency purchase (a US Treasury bought
  in USD with euros) is booked at `price / fx_rate`, the bank's own rate,
  and quotes are requested in the settlement currency. `internal/fx` keeps
  the ECB's euro reference rates in `fx_rates` (full history on first sight
  of a currency, the 90-day file afterwards; refreshed with the quotes) and
  converts via the euro. A quote with no usable rate falls back to the
  purchase price rather than mixing currencies. Without any quote, a
  position is valued at the price it last changed hands at (buy or sell,
  in the settlement currency), so a bond that has left its quote feed
  still carries a sensible value through the history; the average purchase
  price is the fallback only before any trade carried a price. The
  scheduled quote refresh skips securities that are sold out or redeemed
  (`ListListingsWithProvider`); an explicit `triggerQuoteUpdate` with IDs
  still fetches them. The import's "Create
  security" dialog makes a listing (ticker = ISIN, the document's price
  currency, ING quotes), so this works without a trip to the Securities page.
- **Duplicate uploads.** `documents.content_hash` (SHA-256) is checked per
  person at upload; the same bytes again return `importer.DuplicateError`,
  which `/upload` answers with 409 and a message naming the earlier file.
  Documents from before the column are hashed at moneyd startup
  (`Service.BackfillHashes`).
- **Accrued interest on bond trades.** ING folds Stückzinsen into the
  Zwischensumme. The parser keeps it in Fees so that Kurswert ± Fees ± Taxes
  reproduces the Endbetrag: paid on a purchase (positive), received on a
  sale (negative — Fees can go below zero). It is therefore part of the
  sale's realized gain rather than interest income on the Income page; a
  separate INTEREST transaction would classify it better but would have to
  split the document's taxes by guesswork.
- **Security kind.** `securities.asset_kind` (STOCK, ETF, FUND, BOND,
  PROPERTY, VEHICLE, LOAN, OTHER; null until set) and `SecurityKind` in the
  schema. It started as `kind` with a CHECK listing the values; lightsql can
  neither drop a constraint nor rebuild a table other tables reference, so
  the column was superseded (migration 0011) and `kind` is dead weight.
  Lesson: no CHECK on enum-like columns — validate in the GraphQL enum.
  Kinds beyond securities have no quote feed: their listing has no provider
  and is priced by `recordQuote` (the holding page's "Record valuation");
  the finance engine values them through the same quote lookup. Ownership
  and cost basis of such an asset is a DELIVERY_INBOUND with a price —
  `bookAsset` makes all of it in one transaction. `securities.person_id`
  (migration 0012) marks a personal asset: `ListSecurities` is the shared
  rows, `ListAssetsForPerson` the owned ones, and every lookup by ID or
  identifier goes through a `...VisibleTo` query (`person_id IS NULL OR
  person_id = ?`); positions, transactions and gains start from
  person-scoped rows and keep plain `GetSecurity`. Known gap:
  `NetWorthAt`'s "Invested" counts transfers at raw price × units, not at
  `purchasePrice` (FX-converted). The ING parser suggests one
  per staged transaction (`security_kind`: a face value and a percentage
  price make it a bond, a Zinsgutschrift a bond coupon, "ETF"/"ETC"/"ETN" in
  the name an ETF, "Fonds"/"Fund" a fund, anything else a stock); the import
  dialog preselects it. Nothing computes on it yet — an allocation by asset
  class is the obvious consumer.

**Where M4 landed:**

- `internal/quotes`: `Provider` interface (`LatestQuote(ctx, Instrument)`,
  where `Instrument` carries ticker/exchange/currency plus ISIN/WKN from the
  identifiers), a `Registry`, and an `Updater` that fetches quotes for all
  listings with a `quote_provider` and appends them to the history
  (`ON CONFLICT` on listing+time makes re-fetching idempotent). `moneyd`
  refreshes in the background (`-quote-interval`, default 1h, 0 disables).
- Providers ported from v1: `yf` (Yahoo chart API, by ticker — use
  Yahoo-style tickers like `EUNL.DE`; needs a User-Agent header) and `ing`
  (by ISIN). **The v1 ING endpoint is dead** (404, retired API) — the
  provider is kept for the interface's sake but needs a new endpoint or
  removal; `yf` is verified working against production.
- GraphQL: `triggerQuoteUpdate(securityIDs)` mutation returning
  `QuoteUpdateResult { updatedListings, errors }` — per-listing failures go
  into `errors` instead of failing the mutation. `Listing.quotes(from, to)`
  exposes the history for future charts.
- UI: securities page shows the latest quote per listing, the create form
  has a quote-provider select, and an "Update quotes" button triggers the
  mutation and refetches.
- **Time-format bug found & fixed:** the modernc SQLite driver stores
  `time.Time` as Go's `t.String()` by default (`2026-07-03 15:36:11 +0000
  UTC`), which is not lexicographically sortable across offsets and broke
  `ORDER BY time DESC` for quotes. `OpenDB` now sets `_time_format=sqlite`
  in the DSN and all write paths normalize to UTC first; a regression test
  (`TestQuoteTimeOrdering`) guards it. Databases written before this change
  have mixed formats and should be recreated (or `replace(time, ' +0000
  UTC', '+00:00')`).
- UI pattern: route pages gate on `fetching && !data`, not `fetching` alone —
  otherwise every refetch unmounts the page tree and wipes component state
  (this ate the quote-update status message until fixed).

**Where M3 landed:**

- `internal/finance`: FIFO lot books per security (ported from v1's
  `finance/calculation.go`, but on the new typed-transaction model),
  `SnapshotAt(txs, time, quote)` for positions/market value/gains, and
  `TimeWeightedReturn(txs, from, to, quote)` chaining sub-period returns at
  external flows. Prices come through a `QuoteFunc`; without a quote a
  position is valued at its purchase price (so returns are flat-but-honest
  until M4 fills the quotes table).
- Deliberate simplifications, revisit when they hurt: single currency per
  portfolio (no FX), quotes come from a security's *first* listing, TWR
  treats dividends as distributions (they add to return) and counts fees and
  taxes on trades as cost (flows use the cash leg).
- GraphQL: `Portfolio.snapshot(time:)` returning `PortfolioSnapshot` with
  `positions`, totals, and `performance(period:)` (Period enum, ALL_TIME
  uses the first transaction). New sqlc query `GetLatestQuoteBefore` for
  historical snapshots. The quote lookup is per-security (n+1); fine at
  personal-portfolio scale, batch it if it ever shows up in profiles.
- UI: dashboard portfolio list shows market value + gain badge; the detail
  page gained a snapshot summary (market value, P/L, all-time TWR) and a
  positions table. New components: SnapshotSummary, PositionsTable,
  ui/GainBadge; `formatPercent` in `src/lib/format.ts`.
- Quotes can only be inserted directly into the database so far (no
  mutation, no provider) — that is exactly M4.

**Where M2 landed:**

- `ui/` scaffolded with `sv create` (SvelteKit 2, Svelte 5 runes mode, TS).
  Note: SvelteKit config lives inline in `ui/vite.config.ts`, not in a
  `svelte.config.js`.
- Tailwind 4 installed manually (`tailwindcss` + `@tailwindcss/vite` plugin +
  `@import 'tailwindcss'` in `src/app.css`) — `sv add tailwindcss` failed at
  an interactive prompt, don't bother with it.
- SPA mode: `adapter-static` with `fallback: 'index.html'`, `ssr = false`
  in `src/routes/+layout.ts`, vite dev proxy `/graphql` → `localhost:8080`.
- Houdini 2 + houdini-svelte 3 set up (see gotchas — several things moved
  since Houdini 1.x). Query documents live in `.gql` files co-located with
  their route; each route has a `+page.ts` calling the generated
  `load_<Query>` helper. Mutations live inline (`graphql(...)`) in the
  component that owns the form and use `@list` insert fragments
  (`All_Portfolios`, `All_CashAccounts`, `All_Securities`) so lists update
  without refetching; the transaction form refetches via an `oncreated`
  callback instead.
- UI is composed from small components: `src/lib/components/ui/` holds
  primitives (Button, TextField, SelectField, Section, ErrorNote), feature
  components (AppNav, PortfolioList, CashAccountList, TransactionTable,
  TransactionForm, SecurityList, SecurityForm) sit next to them; pages only
  compose. Money/date helpers in `src/lib/format.ts`.
- Pages: dashboard (portfolios + cash accounts + create forms), portfolio
  detail (transactions table + entry form), securities (list/create with
  ISIN + one listing).
- CI: `ui` job in `.github/workflows/build.yml` (npm ci, houdini generate,
  svelte-check, build). `npm run build` and `npm run check` are green.
- Verified end to end with headless Chromium against `moneyd`: pages render,
  forms create data, list cache updates work, no console errors.
- `cmd/moneyd` had never actually been committed in M1 (only a stale compiled
  binary at the repo root) — recreated in M2: `/graphql` with dev auth,
  GraphiQL page at `/graphiql`, flags `-addr` and `-db`.

**Gotchas learned so far:**

- (Houdini gotchas removed with Houdini; see the October 2026 status note.)
- sqlc's SQLite engine silently drops `@name` params it cannot parse (e.g.
  inside `IN (...)`) — always eyeball the generated SQL when using named
  params. GetTransaction ownership is checked in Go for this reason.
- SQLite `:memory:` + `database/sql` pooling: without
  `conn.SetMaxOpenConns(1)` every pooled connection gets its own empty
  database. Set in `persistence.OpenDB`.
- Backend smoke test: `go run ./cmd/moneyd` then POST to
  `localhost:8080/graphql`; GraphiQL at `/graphiql`. Frontend:
  `npm run dev` in `ui/`, which proxies `/graphql` to moneyd.
- The `money.db` (+ `-shm`/`-wal`) and `moneyd`/`mgo` binaries at the repo
  root are leftovers from the lost pre-M2 build and predate the current
  migrations ("no such table: users") — delete them and start fresh.

## Open questions

- Which banks first? Need sample PDFs as test fixtures (anonymized) to pick the initial
  parser targets.
- Quote provider(s): what did we like/dislike about v1's providers? Yahoo-style scraping
  vs. an API with a key.
- Portfolio sharing between users (read-only household view): schema should not preclude
  it, but is it a v2.0 feature or later?
