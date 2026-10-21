// Package api holds the GraphQL schema, which is the single source of truth
// for the moneyd API. The Go resolvers live in internal/api; the frontend and
// CLI clients are generated from this schema.
package api

import _ "embed"

// Schema is the raw GraphQL schema definition.
//
//go:embed schema.graphql
var Schema string
