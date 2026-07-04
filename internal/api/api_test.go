package api

import (
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
)

// TestNewSchema ensures the resolvers match the schema; MustParseSchema panics
// on any mismatch.
func TestNewSchema(t *testing.T) {
	schema := NewSchema()

	gqltesting.RunTest(t, &gqltesting.Test{
		Schema:         schema,
		Query:          `{ version }`,
		ExpectedResult: `{"version": "dev"}`,
	})
}
