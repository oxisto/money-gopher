# money-gopher

<img src="img/gopher.png" width="200" alt="The Money Gopher" />

Money Gopher is a self-hosted investment tracker: portfolios, bank accounts,
securities, transactions and performance over time — for the whole household,
served from a single binary on your own machine.

> [!NOTE]
> This is a ground-up rewrite. The previous ConnectRPC/Next.js implementation
> lives in the git history and the `v0.2.x` releases. The design is documented
> in [docs/DESIGN.md](docs/DESIGN.md).

## Tech stack

- **Backend**: Go, [lightsql](https://github.com/oxisto/lightsql)
  ([sqlc](https://sqlc.dev) + [goose](https://github.com/pressly/goose))
- **API**: GraphQL via
  [graph-gophers/graphql-go](https://github.com/graph-gophers/graphql-go) —
  schema-first, see [api/schema.graphql](api/schema.graphql)
- **Frontend**: SvelteKit SPA with Tailwind CSS 4 and
  [Houdini](https://houdinigraph.org); embedding it into the server binary is
  still to come
- **Imports**: bank PDFs via poppler's `pdftotext` and per-bank parsers

## Getting started

```sh
go run ./cmd/moneyd        # API on :8080, data in ./moneyd-data
cd ui && npm install && npm run dev   # UI on :5173, proxies to moneyd
```

`moneyd` stores its data in the `moneyd-data` directory (`-db` to change it)
and serves the GraphQL API at `/graphql`. Without `-auth`, it signs everyone
in as a development user. PDF import needs poppler's `pdftotext` on the
`PATH`.

`mgo`, the command-line client, is not functional yet.

## Development

```sh
go generate ./...   # regenerate sqlc query code (needs sqlc installed)
go test ./...
```

## Status

Under heavy construction; see the milestones in
[docs/DESIGN.md](docs/DESIGN.md). Expect breaking changes, including to the
database schema, until v0.3.0 ships as a single binary.
