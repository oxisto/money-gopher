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
	"github.com/oxisto/money-gopher/service/internal/crud"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/bufbuild/connect-go"
)

func (svc *service) CreatePortfolio(ctx context.Context, req *connect.Request[portfoliov1.CreatePortfolioRequest]) (res *connect.Response[portfoliov1.Portfolio], err error) {
	return crud.Create(
		req.Msg.Portfolio,
		svc.portfolios,
		func(obj *portfoliov1.Portfolio) *portfoliov1.Portfolio {
			return obj
		})
}

func (svc *service) ListPortfolios(ctx context.Context, req *connect.Request[portfoliov1.ListPortfolioRequest]) (res *connect.Response[portfoliov1.ListPortfolioResponse], err error) {
	res = connect.NewResponse(&portfoliov1.ListPortfolioResponse{})
	res.Msg.Portfolios, err = svc.portfolios.List()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	for _, p := range res.Msg.Portfolios {
		p.Events, err = svc.events.List(p.Name)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	return
}

func (svc *service) UpdatePortfolio(ctx context.Context, req *connect.Request[portfoliov1.UpdatePortfolioRequest]) (res *connect.Response[portfoliov1.Portfolio], err error) {
	res = connect.NewResponse(&portfoliov1.Portfolio{})
	res.Msg, err = svc.portfolios.Update(req.Msg.Portfolio.Name, req.Msg.Portfolio, req.Msg.UpdateMask.Paths)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if slices.Contains(req.Msg.UpdateMask.Paths, "events") {
		for _, ls := range req.Msg.Portfolio.Events {
			svc.events.Replace(ls)
		}
	}

	return
}

func (svc *service) DeletePortfolio(ctx context.Context, req *connect.Request[portfoliov1.DeletePortfolioRequest]) (res *connect.Response[emptypb.Empty], err error) {
	return crud.Delete(req.Msg.Name, svc.portfolios)
}
