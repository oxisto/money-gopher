# Contribution Guidelines

In general, I am looking forward for all kinds of contributions, especially in
this early stage. However, contributions should promote and not conflict the
core motivations about this project described
[here](https://github.com/oxisto/money-gopher#why).

## Technology

The backend is Go, serving a GraphQL API (schema-first, see
[api/schema.graphql](api/schema.graphql)); the frontend is a SvelteKit SPA in
`ui/`. Go code should follow the usual guidelines, such as
[Effective Go](https://go.dev/doc/effective_go). All code is automatically
built and tested using GitHub Actions.

## Dependencies

We want to keep the amount of dependencies minimal. `golang/x` packages are
fair game; beyond that, the backend uses
[graph-gophers/graphql-go](https://github.com/graph-gophers/graphql-go),
[lightsql](https://github.com/oxisto/lightsql) with sqlc and goose, and a few
small libraries. For everything else (logging, HTTP) we stick to the Go
standard library.

## Pull requests and commits

Pull requests are squash-merged, so a pull request's title becomes the commit
message on `main`. Titles follow the Go project's style:

```
<area>: <what it does>
```

- `<area>` is the part of the code it touches, usually a package path without
  the `internal/` prefix: `api`, `importer`, `finance`, `persistence`,
  `cmd/moneyd`, `ui`, `ui/holdings`. Several areas are separated by a comma
  (`api, ui: ...`); use `all` for sweeping changes, and `docs`, `ci` or
  `deps` for those.
- The rest is a short phrase in the imperative, lowercase unless it starts
  with an identifier, and without a trailing period:
  `importer: skip Storno statements` rather than
  `importer: Skipped Storno statements.`
- Keep the whole title under about 72 characters.

The description says *why* the change is needed and anything a reviewer
should know; the diff already shows what changed. Commits inside a pull
request can be as informal as you like.

A check on every pull request enforces the title format.
