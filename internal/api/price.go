package api

import "github.com/oxisto/money-gopher/internal/persistence"

// Price resolves the Price GraphQL type: an exact per-unit price together
// with its ISO 4217 currency code. Unlike Money, the amount is a Decimal —
// see that type for why per-unit prices need more precision than the
// 2-decimal minor-unit convention Money uses for amounts.
type Price struct {
	Amount   Decimal
	Currency string
}

func price(amount persistence.Decimal, currency string) Price {
	return Price{Amount: decimal(amount), Currency: currency}
}
