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
)

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
