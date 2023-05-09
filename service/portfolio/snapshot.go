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

	"github.com/oxisto/money-gopher/finance"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/bufbuild/connect-go"
)

func (svc *service) GetPortfolioSnapshot(ctx context.Context, req *connect.Request[portfoliov1.GetPortfolioSnapshotRequest]) (res *connect.Response[portfoliov1.PortfolioSnapshot], err error) {
	var (
		snap   *portfoliov1.PortfolioSnapshot
		p      *portfoliov1.Portfolio
		secres *connect.Response[portfoliov1.ListSecuritiesResponse]
		secmap map[string]*portfoliov1.Security
	)

	// Retrieve portfolio
	p = svc.portfolio

	// If no time is specified, we assume it to be now
	if req.Msg.Time == nil {
		req.Msg.Time = timestamppb.Now()
	}

	// Set up the snapshot
	snap = &portfoliov1.PortfolioSnapshot{
		Time:      req.Msg.Time,
		Positions: make(map[string]*portfoliov1.PortfolioPosition),
	}

	m := p.EventMap()

	names := maps.Keys(m)

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

	// Make a map out of the securities list
	secmap = Map(secres.Msg.Securities, func(s *portfoliov1.Security) string {
		return s.Name
	})

	// We need to look at the portfolio events up to the time of the snapshot
	// and calculate the current positions.
	for name, txs := range m {
		txs = portfoliov1.EventsBefore(txs, snap.Time.AsTime())

		c := finance.NewCalculation(txs)
		snap.Positions[name] = &portfoliov1.PortfolioPosition{
			SecurityName:  name,
			Amount:        c.Amount,
			PurchaseValue: c.NetValue(),
			PurchasePrice: c.NetPrice(),
			MarketValue:   *secmap[name].ListedOn[0].LatestQuote * float32(c.Amount),
		}
	}

	return connect.NewResponse(snap), nil
}

func Map[K comparable, V any](slice []V, key func(V) K) (m map[K]V) {
	m = make(map[K]V)

	for _, v := range slice {
		m[key(v)] = v
	}

	return
}
