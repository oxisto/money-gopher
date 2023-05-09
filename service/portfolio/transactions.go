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

	portfoliov1 "github.com/oxisto/money-gopher/gen"

	"github.com/bufbuild/connect-go"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (svc *service) CreatePortfolioTransaction(ctx context.Context, req *connect.Request[portfoliov1.CreatePortfolioTransactionRequest]) (res *connect.Response[portfoliov1.PortfolioEvent], err error) {
	var (
		p *portfoliov1.Portfolio = svc.portfolio
	)

	// Increment transaction ID
	req.Msg.Transaction.Id = int32(len(p.Events))

	// Store transaction
	p.Events = append(p.Events, req.Msg.Transaction)

	res = connect.NewResponse(req.Msg.Transaction)

	return
}

func (svc *service) ListPortfolioTransactions(ctx context.Context, req *connect.Request[portfoliov1.ListPortfolioTransactionsRequest]) (res *connect.Response[portfoliov1.ListPortfolioTransactionsResponse], err error) {
	var (
		p *portfoliov1.Portfolio = svc.portfolio
	)

	res = connect.NewResponse(&portfoliov1.ListPortfolioTransactionsResponse{
		Transactions: p.Events,
	})

	return
}

func (svc *service) UpdatePortfolioTransactions(ctx context.Context, req *connect.Request[portfoliov1.UpdatePortfolioTransactionRequest]) (res *connect.Response[portfoliov1.PortfolioEvent], err error) {
	var (
		p   *portfoliov1.Portfolio
		idx int
	)

	// Select portfolio
	p = svc.portfolio

	// Look for transaction by ID
	idx = slices.IndexFunc(p.Events, func(tx *portfoliov1.PortfolioEvent) bool {
		return tx.Id == req.Msg.Transaction.Id
	})

	// Replace the whole transaction; ignore field mask for now
	p.Events[idx] = req.Msg.Transaction

	res = connect.NewResponse(p.Events[idx])

	return
}

func (svc *service) DeletePortfolioTransactions(ctx context.Context, req *connect.Request[portfoliov1.DeletePortfolioTransactionRequest]) (res *connect.Response[emptypb.Empty], err error) {
	var (
		p   *portfoliov1.Portfolio
		idx int
	)

	// Select portfolio
	p = svc.portfolio

	// Look for transaction by ID
	idx = slices.IndexFunc(p.Events, func(tx *portfoliov1.PortfolioEvent) bool {
		return tx.Id == req.Msg.TransactionId
	})

	// Remove it from the portfolio
	slices.Delete(p.Events, idx, idx)

	return
}
