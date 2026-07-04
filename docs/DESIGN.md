# Money Gopher v2 — Redesign

> Status: draft, decided 2026-07-04. This document captures the ground-up redesign of
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
| GraphQL client | [Houdini](https://houdinigraph.org) | generates typed Svelte stores from queries + the introspected schema |
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
Houdini (frontend) and genqlient (CLI) both generate from it.

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
ui/                    SvelteKit SPA (Tailwind 4, Houdini)
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
3. 🚧 **M2 — Frontend scaffold.** SvelteKit SPA + Tailwind 4 + Houdini against the running
   backend: portfolio list, portfolio detail, transaction entry.
4. **M3 — Finance engine.** Port/redo snapshot & performance calculations (positions,
   market value, gains, time-weighted return). `snapshot(time:)` resolver + UI dashboard.
5. **M4 — Quotes.** QuoteProvider interface, first provider, background refresh,
   quote history table for charts.
6. **M5 — PDF import.** `/upload`, document storage, pdftotext extraction, bank
   detection, first real bank parser (whichever bank's PDFs we have fixtures for),
   staging + review UI, `confirmImport`. CSV import rides the same staging pipeline.
7. **M6 — Real auth.** OIDC against a real provider, user provisioning, ownership
   enforcement tests.
8. **M7 — Ship it.** `go:embed` the built SPA, single-binary release (goreleaser),
   Dockerfile, docs. `mgo` CLI rebuilt on genqlient for the endpoints that matter.
9. **M8 — LLM fallback extraction** (opt-in), more bank parsers as needed.

## Status / handoff notes (updated 2026-07-04)

Work happens on the `v2` branch. M0 and M1 are committed; M2 is mid-flight.

**Where M2 stands:**

- `ui/` scaffolded with `sv create` (SvelteKit 2, Svelte 5 runes mode, TS).
  Note: SvelteKit config now lives inline in `ui/vite.config.ts`, not in a
  `svelte.config.js`.
- Tailwind 4 installed manually (`tailwindcss` + `@tailwindcss/vite` plugin +
  `@import 'tailwindcss'` in `src/app.css`) — `sv add tailwindcss` failed at
  an interactive prompt, don't bother with it.
- SPA mode done: `adapter-static` with `fallback: 'index.html'`, `ssr = false`
  in `src/routes/+layout.ts`, vite dev proxy `/graphql` → `localhost:8080`.
- `npm run build` is green.

**Next steps for M2 (not started):**

1. Houdini setup — do it manually, not via `npx houdini init` (interactive):
   `npm i houdini houdini-svelte`, then `houdini.config.js` with
   `schemaPath: '../api/schema.graphql'`, plugin `houdini-svelte` with
   `client: './src/client'`, scalar `Time` mapped to string/Date;
   `src/client.ts` with `new HoudiniClient({ url: '/graphql' })`; add
   `houdini/vite` plugin to `vite.config.ts` (before sveltekit); add
   `$houdini` to `ui/.gitignore`; run `npx houdini generate`.
2. Pages: layout with nav + gopher logo (copy `img/gopher.png` to
   `ui/static/`), dashboard (portfolios + cash accounts + create forms),
   portfolio detail (transactions table + entry form), securities
   (list/create with ISIN + listing).
3. CI: add a `ui` job to `.github/workflows/build.yml` (npm ci, npm run
   build in `ui/`).

**Gotchas learned so far:**

- sqlc's SQLite engine silently drops `@name` params it cannot parse (e.g.
  inside `IN (...)`) — always eyeball the generated SQL when using named
  params. GetTransaction ownership is checked in Go for this reason.
- SQLite `:memory:` + `database/sql` pooling: without
  `conn.SetMaxOpenConns(1)` every pooled connection gets its own empty
  database. Set in `persistence.OpenDB`.
- Backend smoke test: `go run ./cmd/moneyd` then POST to
  `localhost:8080/graphql`; GraphiQL at `/graphiql`.

## Open questions

- Which banks first? Need sample PDFs as test fixtures (anonymized) to pick the initial
  parser targets.
- Quote provider(s): what did we like/dislike about v1's providers? Yahoo-style scraping
  vs. an API with a key.
- Portfolio sharing between users (read-only household view): schema should not preclude
  it, but is it a v2.0 feature or later?
