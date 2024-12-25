module github.com/oxisto/money-gopher

go 1.22.1

require (
	connectrpc.com/connect v1.17.0
	connectrpc.com/vanguard v0.3.0
	github.com/MicahParks/keyfunc/v3 v3.3.5
	github.com/alecthomas/kong v1.6.0
	github.com/fatih/color v1.17.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/jotaen/kong-completion v0.0.6
	github.com/lmittmann/tint v1.0.6
	github.com/mattn/go-colorable v0.1.13
	github.com/mattn/go-isatty v0.0.20
	github.com/mattn/go-sqlite3 v1.14.24
	github.com/oxisto/assert v0.0.6
	github.com/oxisto/oauth2go v0.14.0
	github.com/posener/complete v1.2.3
	golang.org/x/net v0.33.0
	golang.org/x/text v0.21.0
	google.golang.org/protobuf v1.36.0
)

require (
	github.com/MicahParks/jwkset v0.5.19 // indirect
	github.com/riywo/loginshell v0.0.0-20200815045211-7d26008be1ab // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/oauth2 v0.20.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/genproto v0.0.0-20230807174057-1744710a1577 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241223144023-3abc09e42ca8
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241219192143-6b3ec007d9bb // indirect
)

replace github.com/posener/complete v1.2.3 => github.com/oxisto/complete v0.0.0-20231209194436-0b605e2b5bff
