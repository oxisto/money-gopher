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

	"github.com/bufbuild/connect-go"
)

func (svc *service) GetPortfolioSnapshot(ctx context.Context, req *connect.Request[portfoliov1.GetPortfolioSnapshotRequest]) (res *connect.Response[portfoliov1.PortfolioSnapshot], err error) {
	var (
		snap *portfoliov1.PortfolioSnapshot
		p    *portfoliov1.Portfolio
	)

	// Retrieve portfolio
	p = svc.portfolio

	// Set up the snapshot
	snap = &portfoliov1.PortfolioSnapshot{
		Time:      req.Msg.Time,
		Positions: make(map[string]*portfoliov1.PortfolioPosition),
	}

	// We need to look at the portfolio events up to the time of the snapshot
	// and calculate the current positions.
	// TODO(oxisto): We need to filter the tx per position
	c := finance.NewCalculation(p.Events)
	snap.Positions["US0378331005"] = &portfoliov1.PortfolioPosition{
		SecurityName: "US0378331005",
		Amount:       c.Amount,
		EntryValue:   c.NetValue(),
	}

	return connect.NewResponse(snap), nil
}
