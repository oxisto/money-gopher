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
	"bytes"
	"context"

	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/import/csv"
	"github.com/oxisto/money-gopher/service/internal/crud"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

var portfolioEventSetter = func(obj *portfoliov1.PortfolioEvent) *portfoliov1.PortfolioEvent {
	return obj
}

func (svc *service) CreatePortfolioTransaction(ctx context.Context, req *connect.Request[portfoliov1.CreatePortfolioTransactionRequest]) (res *connect.Response[portfoliov1.PortfolioEvent], err error) {
	// Create a unique name for the transaction
	req.Msg.Transaction.MakeUniqueName()

	return crud.Create(
		req.Msg.Transaction,
		svc.events,
		portfolioEventSetter,
	)
}

func (svc *service) ListPortfolioTransactions(ctx context.Context, req *connect.Request[portfoliov1.ListPortfolioTransactionsRequest]) (res *connect.Response[portfoliov1.ListPortfolioTransactionsResponse], err error) {
	return crud.List(
		svc.events,
		func(
			res *connect.Response[portfoliov1.ListPortfolioTransactionsResponse],
			list []*portfoliov1.PortfolioEvent,
		) {
			res.Msg.Transactions = list
		},
		req.Msg.PortfolioName,
	)
}

func (svc *service) UpdatePortfolioTransactions(ctx context.Context, req *connect.Request[portfoliov1.UpdatePortfolioTransactionRequest]) (res *connect.Response[portfoliov1.PortfolioEvent], err error) {
	return crud.Update(
		req.Msg.Transaction.Name,
		req.Msg.Transaction,
		req.Msg.UpdateMask.Paths,
		svc.events,
		portfolioEventSetter,
	)
}

func (svc *service) DeletePortfolioTransactions(ctx context.Context, req *connect.Request[portfoliov1.DeletePortfolioTransactionRequest]) (res *connect.Response[emptypb.Empty], err error) {
	return crud.Delete(
		req.Msg.TransactionId,
		svc.events,
	)
}

func (svc *service) ImportTransactions(ctx context.Context, req *connect.Request[portfoliov1.ImportTransactionsRequest]) (res *connect.Response[emptypb.Empty], err error) {
	var (
		txs  []*portfoliov1.PortfolioEvent
		secs []*portfoliov1.Security
	)

	txs, secs = csv.Import(bytes.NewReader([]byte(req.Msg.FromCsv)), req.Msg.PortfolioName)

	for _, sec := range secs {
		// TODO(oxisto): Once "Create" is really create and not replace, we need
		//  to change this to something else.
		svc.securities.CreateSecurity(
			context.Background(),
			connect.NewRequest(&portfoliov1.CreateSecurityRequest{
				Security: sec,
			}),
		)
	}

	for _, tx := range txs {
		err = svc.events.Replace(tx)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	res = connect.NewResponse(&emptypb.Empty{})

	return
}
