package persistence

// Currency represents a currency with a value and a symbol.
type Currency struct {
	// Value is the value of the currency.
	Value int32 `json:"value"`

	// Symbol is the symbol of the currency.
	Symbol string `json:"symbol"`
}

func Zero() *Currency {
	// TODO(oxisto): Somehow make it possible to change default currency
	return &Currency{Symbol: "EUR"}
}

func Value(v int32) *Currency {
	// TODO(oxisto): Somehow make it possible to change default currency
	return &Currency{Symbol: "EUR", Value: v}
}
