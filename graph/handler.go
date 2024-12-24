package graph

import (
	"net/http"

	"github.com/99designs/gqlgen/handler"
	"github.com/oxisto/money-gopher/db"
)

// NewHandler returns a new graphql endpoint handler.
func NewHandler(q *db.Queries) http.Handler {
	return handler.GraphQL(NewExecutableSchema(Config{
		Resolvers: &Resolver{
			Queries: q,
		},
	}))
}

// NewPlaygroundHandler returns a new GraphQL Playground handler.
func NewPlaygroundHandler(endpoint string) http.Handler {
	return handler.Playground("GraphQL Playground", endpoint)
}
