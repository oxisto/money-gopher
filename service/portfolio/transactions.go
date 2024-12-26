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
	"errors"
	"log/slog"

	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/import/csv"
	"github.com/oxisto/money-gopher/service/internal/crud"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

var portfolioEventSetter = func(obj *portfoliov1.PortfolioEvent) *portfoliov1.PortfolioEvent {
	return obj
}

var (
	ErrMissingSecurityId = errors.New("the specified transaction type requires a security ID")
	ErrMissingPrice      = errors.New("a transaction requires a price")
	ErrMissingAmount     = errors.New("the specified transaction type requires an amount")
)

func (svc *service) CreatePortfolioTransaction(ctx context.Context, req *connect.Request[portfoliov1.CreatePortfolioTransactionRequest]) (res *connect.Response[portfoliov1.PortfolioEvent], err error) {
	var (
		tx *portfoliov1.PortfolioEvent = req.Msg.Transaction
	)

	// Do some basic validation depending on the type
	switch tx.Type {
	case portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_SELL:
		fallthrough
	case portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY:
		if tx.SecurityId == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrMissingSecurityId)
		} else if tx.Amount == 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrMissingAmount)
		}
	}

	// We always need a price
	if tx.Price.IsZero() {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrMissingPrice)
	}

	// Create a unique name for the transaction
	tx.MakeUniqueID()

	slog.Info(
		"Creating transaction",
		"transaction", req.Msg.Transaction,
	)

	return crud.Create(
		req.Msg.Transaction,
		svc.events,
		portfolioEventSetter,
	)
}

func (svc *service) GetPortfolioTransaction(ctx context.Context, req *connect.Request[portfoliov1.GetPortfolioTransactionRequest]) (res *connect.Response[portfoliov1.PortfolioEvent], err error) {
	return crud.Get(
		req.Msg.Id,
		svc.events,
		func(obj *portfoliov1.PortfolioEvent) *portfoliov1.PortfolioEvent {
			return obj
		},
	)
}

func (svc *service) ListPortfolioTransactions(ctx context.Context, req *connect.Request[portfoliov1.ListPortfolioTransactionsRequest]) (res *connect.Response[portfoliov1.ListPortfolioTransactionsResponse], err error) {
	return crud.List(
		svc.events,
		func(
			res *connect.Response[portfoliov1.ListPortfolioTransactionsResponse],
			list []*portfoliov1.PortfolioEvent,
		) error {
			res.Msg.Transactions = list
			return nil
		},
		req.Msg.PortfolioId,
	)
}

func (svc *service) UpdatePortfolioTransaction(ctx context.Context, req *connect.Request[portfoliov1.UpdatePortfolioTransactionRequest]) (res *connect.Response[portfoliov1.PortfolioEvent], err error) {
	slog.Info(
		"Updating transaction",
		"tx", req.Msg.Transaction,
		"update-mask", req.Msg.UpdateMask.Paths,
	)

	return crud.Update(
		req.Msg.Transaction.Id,
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

	txs, secs = csv.Import(bytes.NewReader([]byte(req.Msg.FromCsv)), req.Msg.PortfolioId)

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
