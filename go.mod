module github.com/oxisto/money-gopher

go 1.22.1

require (
	connectrpc.com/connect v1.16.0
	github.com/MicahParks/keyfunc/v3 v3.3.2
	github.com/alecthomas/kong v0.9.0
	github.com/fatih/color v1.16.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/jotaen/kong-completion v0.0.6
	github.com/lmittmann/tint v1.0.4
	github.com/mattn/go-colorable v0.1.13
	github.com/mattn/go-isatty v0.0.20
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/oxisto/assert v0.0.6
	github.com/oxisto/oauth2go v0.13.0
	github.com/posener/complete v1.2.3
	golang.org/x/net v0.22.0
	golang.org/x/text v0.14.0
	google.golang.org/protobuf v1.33.0
)

require (
	github.com/MicahParks/jwkset v0.5.17 // indirect
	github.com/riywo/loginshell v0.0.0-20200815045211-7d26008be1ab // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/oauth2 v0.15.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/time v0.5.0 // indirect
)

// needed until https://github.com/golang/oauth2/issues/615 is resolved
require (
	github.com/golang/protobuf v1.5.3 // indirect
	google.golang.org/appengine v1.6.7 // indirect
)

replace github.com/posener/complete v1.2.3 => github.com/oxisto/complete v0.0.0-20231209194436-0b605e2b5bff
