package portfoliov1

import (
	"fmt"
	"hash/fnv"
	"log/slog"
	"math"
	"strconv"
	"time"
)

func (p *Portfolio) EventMap() (m map[string][]*PortfolioEvent) {
	m = make(map[string][]*PortfolioEvent)

	for _, tx := range p.Events {
		name := tx.GetSecurityName()
		if name != "" {
			m[name] = append(m[name], tx)
		}
	}

	return
}

func EventsBefore(txs []*PortfolioEvent, t time.Time) (out []*PortfolioEvent) {
	out = make([]*PortfolioEvent, 0, len(txs))

	for _, tx := range txs {
		if tx.GetTime().AsTime().After(t) {
			continue
		}

		out = append(out, tx)
	}

	return
}

func (tx *PortfolioEvent) MakeUniqueName() {
	// Create a unique name based on a hash containing:
	//  - security name
	//  - portfolio name
	//  - date
	//  - amount
	h := fnv.New64a()
	h.Write([]byte(tx.SecurityName))
	h.Write([]byte(tx.PortfolioName))
	h.Write([]byte(tx.Time.AsTime().Local().Format(time.DateTime)))
	h.Write([]byte(strconv.FormatInt(int64(tx.Type), 10)))
	h.Write([]byte(strconv.FormatInt(int64(tx.Amount), 10)))

	tx.Name = strconv.FormatUint(h.Sum64(), 16)
}

// LogValue implements slog.LogValuer.
func (tx *PortfolioEvent) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("name", tx.Name),
		slog.String("security.name", tx.SecurityName),
		slog.Float64("amount", float64(tx.Amount)),
		slog.String("price", tx.Price.Pretty()),
		slog.String("fees", tx.Fees.Pretty()),
		slog.String("taxes", tx.Taxes.Pretty()),
	)
}

// LogValue implements slog.LogValuer.
func (ls *ListedSecurity) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("name", ls.SecurityName),
		slog.String("ticker", ls.Ticker),
	)
}

func Zero() *Currency {
	// TODO(oxisto): Somehow make it possible to change default currency
	return &Currency{Symbol: "EUR"}
}

func Value(v int32) *Currency {
	// TODO(oxisto): Somehow make it possible to change default currency
	return &Currency{Symbol: "EUR", Value: v}
}

func (c *Currency) PlusAssign(o *Currency) {
	if o != nil {
		c.Value += o.Value
	}
}

func (c *Currency) MinusAssign(o *Currency) {
	if o != nil {
		c.Value -= o.Value
	}
}

func Plus(a *Currency, b *Currency) *Currency {
	return &Currency{
		Value:  a.Value + b.Value,
		Symbol: a.Symbol,
	}
}

func Minus(a *Currency, b *Currency) *Currency {
	return &Currency{
		Value:  a.Value - b.Value,
		Symbol: a.Symbol,
	}
}

func Divide(a *Currency, b float64) *Currency {
	return &Currency{
		Value:  int32(math.Round((float64(a.Value) / b))),
		Symbol: a.Symbol,
	}
}

func Times(a *Currency, b float64) *Currency {
	return &Currency{
		Value:  int32(math.Round((float64(a.Value) * b))),
		Symbol: a.Symbol,
	}
}

func (c *Currency) Pretty() string {
	return fmt.Sprintf("%.0f %s", float32(c.Value)/100, c.Symbol)
}
