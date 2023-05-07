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
		if tx.GetTime().After(t) {
			continue
		}

		out = append(out, tx)
	}

	return
}

func (ev *PortfolioEvent) GetSecurityName() (name string) {
	if buy := ev.GetBuy(); buy != nil {
		name = buy.SecurityName
	} else if sell := ev.GetSell(); sell != nil {
		name = sell.SecurityName
	} else if div := ev.GetDividend(); div != nil {
		name = div.SecurityName
	}

	return
}

func (ev *PortfolioEvent) GetTime() (t time.Time) {
	if buy := ev.GetBuy(); buy != nil {
		t = buy.Time.AsTime()
	} else if sell := ev.GetSell(); sell != nil {
		t = sell.Time.AsTime()
	} else if div := ev.GetDividend(); div != nil {
		t = div.Time.AsTime()
	}

	return
}
