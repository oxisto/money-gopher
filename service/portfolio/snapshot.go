// Copyright 2023 Christian Banse
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This file is part of The Money Gopher.

package portfolio

import (
	"context"

	moneygopher "github.com/oxisto/money-gopher"
	"github.com/oxisto/money-gopher/finance"
	portfoliov1 "github.com/oxisto/money-gopher/gen"

	"github.com/bufbuild/connect-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (svc *service) GetPortfolioSnapshot(ctx context.Context, req *connect.Request[portfoliov1.GetPortfolioSnapshotRequest]) (res *connect.Response[portfoliov1.PortfolioSnapshot], err error) {
	var (
		snap   *portfoliov1.PortfolioSnapshot
		p      portfoliov1.Portfolio
		m      map[string][]*portfoliov1.PortfolioEvent
		names  []string
		secres *connect.Response[portfoliov1.ListSecuritiesResponse]
		secmap map[string]*portfoliov1.Security
	)

	// Retrieve transactions
	p.Events, err = svc.events.List(req.Msg.PortfolioName)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// If no time is specified, we assume it to be now
	if req.Msg.Time == nil {
		req.Msg.Time = timestamppb.Now()
	}

	// Set up the snapshot
	snap = &portfoliov1.PortfolioSnapshot{
		Time:      req.Msg.Time,
		Positions: make(map[string]*portfoliov1.PortfolioPosition),
	}

	// Record the first transaction time
	if len(p.Events) > 0 {
		snap.FirstTransactionTime = p.Events[0].Time
	}

	// Retrieve the event map; a map of events indexed by their security name
	m = p.EventMap()
	names = keys(m)

	// Retrieve market value of filtered securities
	secres, err = svc.securities.ListSecurities(
		context.Background(),
		connect.NewRequest(&portfoliov1.ListSecuritiesRequest{
			Filter: &portfoliov1.ListSecuritiesRequest_Filter{
				SecurityNames: names,
			},
		}),
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Make a map out of the securities list so we can access it easier
	secmap = moneygopher.Map(secres.Msg.Securities, func(s *portfoliov1.Security) string {
		return s.Name
	})

	// We need to look at the portfolio events up to the time of the snapshot
	// and calculate the current positions.
	for name, txs := range m {
		txs = portfoliov1.EventsBefore(txs, snap.Time.AsTime())

		c := finance.NewCalculation(txs)

		if c.Amount == 0 {
			continue
		}

		snap.Positions[name] = &portfoliov1.PortfolioPosition{
			Security:      secmap[name],
			Amount:        c.Amount,
			PurchaseValue: c.NetValue(),
			PurchasePrice: c.NetPrice(),
			MarketValue:   marketPrice(secmap, name, c.NetPrice()) * float32(c.Amount),
			MarketPrice:   marketPrice(secmap, name, c.NetPrice()),
		}

		// Add to total value(s)
		snap.TotalPurchaseValue += snap.Positions[name].PurchaseValue
		snap.TotalMarketValue += snap.Positions[name].MarketValue
	}

	return connect.NewResponse(snap), nil
}

func marketPrice(secmap map[string]*portfoliov1.Security, name string, netPrice float32) float32 {
	ls := secmap[name].ListedOn

	if ls == nil || ls[0].LatestQuote == nil {
		return netPrice
	} else {
		return *ls[0].LatestQuote
	}
}

func keys[M ~map[K]V, K comparable, V any](m M) (keys []K) {
	keys = make([]K, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	return keys
}
