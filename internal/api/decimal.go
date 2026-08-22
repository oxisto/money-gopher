package api

import (
	"encoding/json"
	"fmt"

	"github.com/oxisto/money-gopher/internal/persistence"
)

// Decimal is the GraphQL-facing wrapper around persistence.Decimal,
// marshaled as an exact decimal string (e.g. "8.2656") rather than a
// number — GraphQL's Int is 32 bits and Float is IEEE 754 binary, neither
// of which can hold an arbitrary decimal exactly. Implements the same
// interface graph-gophers/graphql-go's own built-in Time scalar does.
type Decimal persistence.Decimal

// ImplementsGraphQLType maps this custom Go type to the graphql scalar
// type in the schema.
func (Decimal) ImplementsGraphQLType(name string) bool {
	return name == "Decimal"
}

// UnmarshalGraphQL is called whenever the Decimal scalar is used as input.
func (d *Decimal) UnmarshalGraphQL(input any) error {
	s, ok := input.(string)
	if !ok {
		return fmt.Errorf("Decimal must be a string, got %T", input)
	}
	dec, err := persistence.NewDecimal(s)
	if err != nil {
		return err
	}
	*d = Decimal(dec)
	return nil
}

// MarshalJSON is called whenever a field using the Decimal scalar is queried.
func (d Decimal) MarshalJSON() ([]byte, error) {
	return json.Marshal(persistence.Decimal(d).StringPrec(8))
}

func decimal(d persistence.Decimal) Decimal { return Decimal(d) }
