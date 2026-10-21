package persistence

import (
	"database/sql/driver"
	"fmt"
	"math/big"
	"strconv"
)

// Decimal is an exact decimal number backed by math/big.Rat. Unlike the rest
// of this schema's INTEGER-minor-units convention, price columns
// (DECIMAL(19,8)) need more than 2 decimal places — real ING confirmations
// quote some prices and FX rates to 6 decimals (a subscription-rights Kurs,
// an FX rate), so the column (and this type's String()/Value()) use 8 for
// headroom — and lightsql's DECIMAL type is exact as of v0.6.0
// (github.com/oxisto/lightsql), so this reads and writes it as text through
// database/sql's driver.Value rather than through an intermediate float64,
// which would reintroduce the precision loss this type exists to avoid.
//
// The zero Decimal represents SQL NULL; Valid reports whether it holds a
// real value, matching the database/sql sql.Null* convention.
type Decimal struct {
	Valid bool
	r     *big.Rat
}

// NewDecimal parses s (e.g. "8.2656") as an exact decimal.
func NewDecimal(s string) (Decimal, error) {
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return Decimal{}, fmt.Errorf("invalid decimal %q", s)
	}
	return Decimal{Valid: true, r: r}, nil
}

// DecimalFromFloat converts f to an exact Decimal via its precise rational
// value. Meant for values that already came from a float64 (e.g. a market
// data provider's JSON response, or Units, which the schema stores as
// REAL) — it cannot recover precision the float64 conversion already lost,
// but avoids losing any more on top of that.
//
// This is the only place a float64 should enter a Decimal computation.
// Scaling by a whole number (e.g. converting major units to minor units by
// a factor of 100) should use MulInt/DivInt instead of folding the factor
// into a float64 expression before conversion — the latter performs real
// float64 arithmetic that this type exists to avoid, even though it's
// often exact in practice for small whole-number factors.
func DecimalFromFloat(f float64) Decimal {
	return Decimal{Valid: true, r: new(big.Rat).SetFloat64(f)}
}

// NewDecimalFromInt64 converts an exact integer (e.g. a minor-units Money
// amount) to a Decimal, with no float64 involved.
func NewDecimalFromInt64(n int64) Decimal {
	return Decimal{Valid: true, r: new(big.Rat).SetInt64(n)}
}

func (d Decimal) rat() *big.Rat {
	if d.r == nil {
		return new(big.Rat)
	}
	return d.r
}

// IsZero reports whether d is exactly zero. A NULL Decimal is not zero.
func (d Decimal) IsZero() bool { return d.Valid && d.r.Sign() == 0 }

// Sign returns -1, 0 or +1 depending on the sign of d. A NULL Decimal has
// sign 0.
func (d Decimal) Sign() int { return d.rat().Sign() }

// Add, Sub and Mul return a new Decimal; d and other are never mutated.
func (d Decimal) Add(other Decimal) Decimal {
	return Decimal{Valid: true, r: new(big.Rat).Add(d.rat(), other.rat())}
}

func (d Decimal) Sub(other Decimal) Decimal {
	return Decimal{Valid: true, r: new(big.Rat).Sub(d.rat(), other.rat())}
}

func (d Decimal) Mul(other Decimal) Decimal {
	return Decimal{Valid: true, r: new(big.Rat).Mul(d.rat(), other.rat())}
}

// MulFloat multiplies by a float64 (e.g. Units, which the schema stores as
// REAL) via its precise rational value.
func (d Decimal) MulFloat(f float64) Decimal {
	return d.Mul(DecimalFromFloat(f))
}

// MulInt multiplies by an exact integer, with no float64 involved — the
// right way to apply a whole-number scale factor (e.g. 100, to convert
// major units to minor units) without folding it into a float64 expression.
func (d Decimal) MulInt(n int64) Decimal {
	return d.Mul(NewDecimalFromInt64(n))
}

// Div returns a new Decimal; d and other are never mutated. Dividing by zero
// panics, matching big.Rat.
func (d Decimal) Div(other Decimal) Decimal {
	return Decimal{Valid: true, r: new(big.Rat).Quo(d.rat(), other.rat())}
}

// DivFloat divides by a float64 via its precise rational value.
func (d Decimal) DivFloat(f float64) Decimal {
	return d.Div(DecimalFromFloat(f))
}

// DivInt divides by an exact integer, with no float64 involved.
func (d Decimal) DivInt(n int64) Decimal {
	return d.Div(NewDecimalFromInt64(n))
}

// RoundToInt64 rounds d to the nearest integer (half away from zero) —
// meant for converting an exact per-unit Decimal price into an int64
// minor-units Money amount, e.g. price.MulFloat(units).RoundToInt64() for a
// position's market value in cents.
func (d Decimal) RoundToInt64() int64 {
	n, err := strconv.ParseInt(d.StringPrec(0), 10, 64)
	if err != nil {
		// Only possible for a value wider than int64, which nothing in this
		// app's domain produces.
		panic("persistence.Decimal: RoundToInt64 overflow: " + err.Error())
	}
	return n
}

// StringPrec formats d with exactly prec decimal digits, e.g. StringPrec(8)
// for a DECIMAL(19,8) column.
func (d Decimal) StringPrec(prec int) string {
	return d.rat().FloatString(prec)
}

// String formats d at 8 decimal digits, matching the DECIMAL(19,8) columns
// this type is used for today.
func (d Decimal) String() string {
	if !d.Valid {
		return ""
	}
	return d.StringPrec(8)
}

// Scan implements sql.Scanner. lightsql hands back DECIMAL columns as a
// plain decimal string (verified directly against v0.6.0), so this never
// round-trips through a float64.
func (d *Decimal) Scan(src any) error {
	if src == nil {
		*d = Decimal{}
		return nil
	}
	var s string
	switch v := src.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("persistence.Decimal: cannot scan %T", src)
	}
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return fmt.Errorf("persistence.Decimal: invalid decimal %q", s)
	}
	*d = Decimal{Valid: true, r: r}
	return nil
}

// Value implements driver.Valuer.
func (d Decimal) Value() (driver.Value, error) {
	if !d.Valid {
		return nil, nil
	}
	return d.StringPrec(8), nil
}
