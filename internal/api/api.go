// Package api wires up the GraphQL API of moneyd. The schema in
// api/schema.graphql is the source of truth; resolver structs in this package
// are verified against it when the schema is parsed at startup.
package api

import (
	"net/http"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/oxisto/money-gopher/api"
)

// Version is the moneyd version reported by the API. Overridden at build time
// via -ldflags "-X github.com/oxisto/money-gopher/internal/api.Version=...".
var Version = "dev"

// RootResolver is the root GraphQL resolver.
type RootResolver struct{}

// Version resolves Query.version.
func (r *RootResolver) Version() string {
	return Version
}

// NewSchema parses the embedded schema and binds it to the root resolver. It
// panics if the resolvers do not match the schema, which is exercised by
// tests, so a mismatch cannot reach a release.
func NewSchema() *graphql.Schema {
	return graphql.MustParseSchema(api.Schema, &RootResolver{}, graphql.UseFieldResolvers())
}

// NewHandler returns the HTTP handler serving the GraphQL API.
func NewHandler(schema *graphql.Schema) http.Handler {
	return &relay.Handler{Schema: schema}
}
