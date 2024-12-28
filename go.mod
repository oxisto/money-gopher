module github.com/oxisto/money-gopher

go 1.23.4

require (
	connectrpc.com/connect v1.16.2
	connectrpc.com/vanguard v0.3.0
	github.com/99designs/gqlgen v0.17.61
	github.com/MicahParks/keyfunc/v3 v3.3.5
	github.com/fatih/color v1.18.0
	github.com/golang-jwt/jwt/v5 v5.2.1
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
	golang.org/x/text v0.21.0
	google.golang.org/protobuf v1.36.1
)

require github.com/google/go-cmp v0.6.0 // indirect

require (
	github.com/MicahParks/jwkset v0.5.19 // indirect
	github.com/agnivade/levenshtein v1.2.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/oauth2 v0.22.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241230172942-26aa7a208def
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241223144023-3abc09e42ca8 // indirect
)

replace github.com/99designs/gqlgen v0.17.61 => github.com/oxisto/gqlgen v0.17.62-0.20241227140449-4bf1c5c27bad
