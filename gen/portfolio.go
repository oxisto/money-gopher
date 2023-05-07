package portfoliov1

// Apply applies the portfolio event to the snapshot
func (e *PortfolioEvent) Apply(snap *PortfolioSnapshot) {
	switch v := e.EventOneof.(type) {
	case *PortfolioEvent_Buy:
		pos, ok := snap.Positions[v.Buy.GetSecurityName()]
		if ok {
			// Position already exists, we can just add it
			// TODO(oxisto): re-calculate entry price
			pos.Amount++
		}
	}
}
