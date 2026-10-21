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
  [gql.tada](https://gql-tada.0no.co) for typed GraphQL documents, embedded
  into the server binary
- **Imports**: bank PDFs via poppler's `pdftotext` and per-bank parsers

## Getting started

```sh
(cd ui && npm ci && npm run build)   # the SPA, into ui/build/site
go build ./cmd/moneyd                # one binary with the SPA embedded
./moneyd                             # http://localhost:8080
```

`moneyd` stores its data in the `moneyd-data` directory (`-db` to change it),
serves the UI at `/` and the GraphQL API at `/graphql`. Without `-auth`, it
signs everyone in as a development user. PDF import needs poppler's
`pdftotext` on the `PATH`.

A `go build` without the npm step still works; the server then says so at `/`
instead of serving the UI.

### Docker

Releases are published as `ghcr.io/oxisto/money-gopher` (linux/amd64 and
arm64), with `pdftotext` included. Data lives in `/data`:

```sh
docker run -d --name money-gopher -p 8080:8080 -v money-gopher:/data \
  ghcr.io/oxisto/money-gopher:0.3 \
  -auth builtin -auth-password change-me -public-url http://localhost:8080
```

Without `-auth`, the server runs in development mode and signs everyone in
as the same user — fine on your own machine, not for anything reachable by
others. `-auth builtin` runs a small login server inside `moneyd`;
`-auth oidc` with `-oidc-issuer`, `-oidc-client-id` and `-oidc-client-secret`
uses an external identity provider. `moneyd -help` lists everything.

`mgo`, the command-line client, is not functional yet.

## Development

```sh
go run ./cmd/moneyd                  # API on :8080 (serves the last-built UI, if any)
cd ui && npm run dev                 # UI on :5173 with hot reload, proxies to :8080

go generate ./...                    # regenerate sqlc query code (needs sqlc installed)
go test ./...
cd ui && npm run generate            # refresh GraphQL types after changing api/schema.graphql
```

## Status

Under heavy construction; see the milestones in
[docs/DESIGN.md](docs/DESIGN.md). Expect breaking changes, including to the
database schema, until v0.3.0 ships as a single binary.
