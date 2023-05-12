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
	"github.com/oxisto/money-gopher/service/common"

	"github.com/bufbuild/connect-go"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (svc *service) CreatePortfolioTransaction(ctx context.Context, req *connect.Request[portfoliov1.CreatePortfolioTransactionRequest]) (res *connect.Response[portfoliov1.PortfolioEvent], err error) {
	err = svc.events.Replace(req.Msg.Transaction)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res = connect.NewResponse(req.Msg.Transaction)

	return
}

func (svc *service) ListPortfolioTransactions(ctx context.Context, req *connect.Request[portfoliov1.ListPortfolioTransactionsRequest]) (res *connect.Response[portfoliov1.ListPortfolioTransactionsResponse], err error) {
	res = connect.NewResponse(&portfoliov1.ListPortfolioTransactionsResponse{})
	res.Msg.Transactions, err = svc.events.List(req.Msg.PortfolioName)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return
}

func (svc *service) UpdatePortfolioTransactions(ctx context.Context, req *connect.Request[portfoliov1.UpdatePortfolioTransactionRequest]) (res *connect.Response[portfoliov1.PortfolioEvent], err error) {
	res = connect.NewResponse(&portfoliov1.PortfolioEvent{})
	res.Msg, err = svc.events.Update(
		req.Msg.Transaction.Id,
		req.Msg.Transaction,
		req.Msg.UpdateMask.Paths,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return
}

func (svc *service) DeletePortfolioTransactions(ctx context.Context, req *connect.Request[portfoliov1.DeletePortfolioTransactionRequest]) (res *connect.Response[emptypb.Empty], err error) {
	return common.Delete(req.Msg.TransactionId, svc.events)
}
