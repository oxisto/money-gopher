module github.com/oxisto/money-gopher

go 1.21.5

require (
	connectrpc.com/connect v1.14.0
	github.com/MicahParks/keyfunc/v3 v3.1.1
	github.com/alecthomas/kong v0.8.1
	github.com/fatih/color v1.16.0
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/jotaen/kong-completion v0.0.6
	github.com/lmittmann/tint v1.0.3
	github.com/mattn/go-colorable v0.1.13
	github.com/mattn/go-isatty v0.0.20
	github.com/mattn/go-sqlite3 v1.14.19
	github.com/oxisto/assert v0.0.6
	github.com/oxisto/oauth2go v0.12.0
	github.com/posener/complete v1.2.3
	golang.org/x/net v0.19.0
	golang.org/x/text v0.14.0
	google.golang.org/protobuf v1.32.0
)

require (
	github.com/MicahParks/jwkset v0.5.4 // indirect
	github.com/riywo/loginshell v0.0.0-20200815045211-7d26008be1ab // indirect
	golang.org/x/crypto v0.16.0 // indirect
	golang.org/x/oauth2 v0.15.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
)

// needed until https://github.com/golang/oauth2/issues/615 is resolved
require (
	github.com/golang/protobuf v1.5.3 // indirect
	google.golang.org/appengine v1.6.7 // indirect
)

replace github.com/posener/complete v1.2.3 => github.com/oxisto/complete v0.0.0-20231209194436-0b605e2b5bff
