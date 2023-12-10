module github.com/oxisto/money-gopher

go 1.21

require (
	connectrpc.com/connect v1.13.0
	github.com/alecthomas/kong v0.8.1
	github.com/fatih/color v1.16.0
	github.com/jotaen/kong-completion v0.0.6
	github.com/mattn/go-colorable v0.1.13
	github.com/mattn/go-sqlite3 v1.14.18
	github.com/oxisto/assert v0.0.6
	github.com/posener/complete v1.2.3
	golang.org/x/net v0.19.0
	golang.org/x/text v0.14.0
	google.golang.org/protobuf v1.31.0
)

require (
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/riywo/loginshell v0.0.0-20200815045211-7d26008be1ab // indirect
	golang.org/x/sys v0.15.0 // indirect
)

replace github.com/posener/complete v1.2.3 => github.com/oxisto/complete v0.0.0-20231209194436-0b605e2b5bff
