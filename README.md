# money-gopher

<img src="img/gopher.png" width="200" alt="The Money Gopher" />

Money Gopher is a self-hosted investment tracker: portfolios, bank accounts,
securities, transactions and performance over time — for the whole household,
served from a single binary on your own machine.

> [!NOTE]
> This is version 2, a ground-up rewrite. The previous ConnectRPC/Next.js
> implementation lives in the git history and the `v0.x` releases. The
> redesign is documented in [docs/DESIGN.md](docs/DESIGN.md).

## Tech stack

- **Backend**: Go, SQLite ([sqlc](https://sqlc.dev) +
  [goose](https://github.com/pressly/goose))
- **API**: GraphQL via
  [graph-gophers/graphql-go](https://github.com/graph-gophers/graphql-go) —
  schema-first, see [api/schema.graphql](api/schema.graphql)
- **Frontend**: SvelteKit SPA with Tailwind CSS 4 and
  [Houdini](https://houdinigraph.org), embedded into the server binary
- **Imports**: bank PDFs via poppler's `pdftotext` and per-bank parsers

## Getting started

```sh
go run ./cmd/moneyd
```

`moneyd` listens on `:8080` and creates a `money.db` SQLite database in the
working directory. The GraphQL API is served at `/graphql`.

`mgo`, the command-line client, is not functional yet in v2.

## Development

```sh
go generate ./...   # regenerate sqlc query code (needs sqlc installed)
go test ./...
```

## Status

Version 2 is under heavy construction; see the milestones in
[docs/DESIGN.md](docs/DESIGN.md). Expect breaking changes, including to the
database schema, until the first v2 release.
