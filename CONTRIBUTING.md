# Contribution Guidelines

In general, I am looking forward for all kinds of contributions, especially in
this early stage. However, contributions should promote and not conflict the
core motivations about this project described
[here](https://github.com/oxisto/money-gopher#why).

## Technology

So far, the only fixed piece is the use of Go in the backend and the use of some
kind of RPC mechanism to provide a server. Code should follow the usual
guidelines, such as [Effective Go](https://go.dev/doc/effective_go). All code is
automatically built and and tested using GitHub Actions.

## Dependencies

We want to keep the amount of dependencies minimal. `golang/x` packages are fair
game, we need `github.com/bufbuild/connect-go` and `google.golang.org/protobuf`.
We also probably need some kind of assertion library for testing. For everything
else (logging, database) we stick to the Go standard library for now, however we
need to import an SQL driver, like `modernc.org/sqlite`.
