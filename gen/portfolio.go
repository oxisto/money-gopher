package portfoliov1

import (
	"hash/fnv"
	"log/slog"
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
		slog.Float64("price", float64(tx.Price)),
		slog.Float64("fees", float64(tx.Fees)),
		slog.Float64("taxes", float64(tx.Taxes)),
	)
}

// LogValue implements slog.LogValuer.
func (ls *ListedSecurity) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("name", ls.SecurityName),
		slog.String("ticker", ls.Ticker),
	)
}
