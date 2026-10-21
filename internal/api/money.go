package api

// Money resolves the Money GraphQL type: an amount in minor units together
// with its ISO 4217 currency code. GraphQL Int is 32 bits, which caps a
// single monetary value at about 21 million major units — plenty for a
// personal finance tool.
type Money struct {
	Amount   int32
	Currency string
}

func money(amount int64, currency string) Money {
	return Money{Amount: int32(amount), Currency: currency}
}
