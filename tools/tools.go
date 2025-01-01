//go:build tools
// +build tools

package tools

import (
	_ "github.com/99designs/gqlgen"
	_ "github.com/mfridman/tparse"
	_ "github.com/mpyw/sqlc-restruct/cmd/sqlc-restruct"
	_ "github.com/sqlc-dev/sqlc/cmd/sqlc"
)
