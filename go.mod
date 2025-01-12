module github.com/oxisto/money-gopher

go 1.23.4

require (
	github.com/99designs/gqlgen v0.17.61
	github.com/fatih/color v1.18.0
	github.com/google/uuid v1.6.0
	github.com/lmittmann/tint v1.0.6
	github.com/mattn/go-colorable v0.1.13
	github.com/mattn/go-isatty v0.0.20
	github.com/mattn/go-sqlite3 v1.14.24
	github.com/oxisto/assert v0.1.2
	github.com/oxisto/oauth2go v0.14.0
	github.com/pressly/goose/v3 v3.24.0
	github.com/shurcooL/graphql v0.0.0-20230722043721-ed46e5a46466
	github.com/urfave/cli/v3 v3.0.0-beta1
	github.com/vektah/gqlparser/v2 v2.5.20
	golang.org/x/net v0.33.0
	golang.org/x/sync v0.10.0
)

require (
	github.com/agnivade/levenshtein v1.2.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/oauth2 v0.22.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)

replace github.com/99designs/gqlgen v0.17.61 => github.com/oxisto/gqlgen v0.17.62-0.20241227140449-4bf1c5c27bad
