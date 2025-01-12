package currency

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math"
)

// Currency represents a currency with a value and a symbol.
type Currency struct {
	// Amount is the amount of the currency.
	Amount int32 `json:"value"`

	// Symbol is the symbol of the currency.
	Symbol string `json:"symbol"`
}

func Zero() *Currency {
	// TODO(oxisto): Somehow make it possible to change default currency
	return &Currency{Symbol: "EUR"}
}

func Value(v int32) *Currency {
	// TODO(oxisto): Somehow make it possible to change default currency
	return &Currency{Symbol: "EUR", Amount: v}
}

func (c *Currency) PlusAssign(o *Currency) {
	if o != nil {
		c.Amount += o.Amount
	}
}

func (c *Currency) MinusAssign(o *Currency) {
	if o != nil {
		c.Amount -= o.Amount
	}
}

func Plus(a *Currency, b *Currency) *Currency {
	return &Currency{
		Amount: a.Amount + b.Amount,
		Symbol: a.Symbol,
	}
}

func (a *Currency) Plus(b *Currency) *Currency {
	if b == nil {
		return &Currency{
			Amount: a.Amount,
			Symbol: a.Symbol,
		}
	}

	return &Currency{
		Amount: a.Amount + b.Amount,
		Symbol: a.Symbol,
	}
}

func Minus(a *Currency, b *Currency) *Currency {
	return &Currency{
		Amount: a.Amount - b.Amount,
		Symbol: a.Symbol,
	}
}

func Divide(a *Currency, b float64) *Currency {
	return &Currency{
		Amount: int32(math.Round((float64(a.Amount) / b))),
		Symbol: a.Symbol,
	}
}

func Times(a *Currency, b float64) *Currency {
	return &Currency{
		Amount: int32(math.Round((float64(a.Amount) * b))),
		Symbol: a.Symbol,
	}
}

func (c *Currency) Pretty() string {
	return fmt.Sprintf("%.0f %s", float32(c.Amount)/100, c.Symbol)
}

func (c *Currency) IsZero() bool {
	return c == nil || c.Amount == 0
}

// Value implements the driver.Valuer interface.
func (c *Currency) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}

	return json.Marshal(c)
}

// Scan implements the sql.Scanner interface.
func (c *Currency) Scan(src interface{}) error {
	if src == nil {
		return nil
	}

	switch v := src.(type) {
	case string:
		return json.Unmarshal([]byte(v), c)
	case []byte:
		return json.Unmarshal(v, c)
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
}
