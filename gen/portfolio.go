package portfoliov1

import "time"

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
