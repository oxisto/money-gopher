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

Carried over from v1, extended for multi-user and imports:

- **User** — OIDC subject, display name. Owns portfolios and bank accounts.
- **Portfolio** — belongs to a user; may be shared (read) with other users later.
- **BankAccount** — cash account, linked to portfolios for settlement.
- **Security** — global (shared across users), identified by ISIN-style ID.
- **ListedSecurity** — a security on a specific exchange: ticker, currency, latest quote.
- **Transaction** (v1: `PortfolioEvent`) — buy/sell/dividend/deposit/withdrawal/fees/taxes,
  with time, amounts, linked security. Provenance: manual, CSV, or a source **Document**.
- **Document** — an uploaded file (PDF/CSV): raw bytes, extraction state, detected bank.
- **Snapshot / Position** — computed, not stored: portfolio value, positions, performance
  at a point in time (from `internal/finance`).

Money is stored as integer minor units (cents) + ISO currency code — never floats.
GraphQL exposes custom scalars `Money` and `Time`.

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

1. **M0 — Teardown & scaffold.** Remove v1 code (keep `img/`, LICENSE, community files),
   new directory layout, Go module tidy, CI (build + test + sqlc/vet), README rewrite.
2. **M1 — Core backend.** SQLite schema + migrations + sqlc (users, portfolios, bank
   accounts, securities, transactions). GraphQL schema v1 with portfolio/security/
   transaction CRUD. Dev auth. `moneyd` serves `/graphql` with GraphiQL in dev.
3. **M2 — Frontend scaffold.** SvelteKit SPA + Tailwind 4 + Houdini against the running
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

## Open questions

- Which banks first? Need sample PDFs as test fixtures (anonymized) to pick the initial
  parser targets.
- Quote provider(s): what did we like/dislike about v1's providers? Yahoo-style scraping
  vs. an API with a key.
- Portfolio sharing between users (read-only household view): schema should not preclude
  it, but is it a v2.0 feature or later?
