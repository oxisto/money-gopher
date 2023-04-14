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

// package securities contains the code for the SecuritiesService implementation.
package securities

import (
	"context"

	"github.com/bufbuild/connect-go"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"golang.org/x/text/currency"
)

type service struct {
	// TODO(oxisto): convert this to sqlite
	sec []*portfoliov1.Security
}

func NewService() portfoliov1connect.SecuritiesServiceHandler {
	return &service{
		// Add some static data for testing
		sec: []*portfoliov1.Security{
			{
				Name: "Apple Inc.",
				Isin: "US0378331005",
				ListedOn: []*portfoliov1.ListedSecurity{
					{
						Ticker:   "APC.F",
						Currency: currency.EUR.String(),
					},
					{
						Ticker:   "AAPL",
						Currency: currency.USD.String(),
					},
				},
			},
		},
	}
}

func (svc *service) ListSecurities(ctx context.Context, req *connect.Request[portfoliov1.ListSecuritiesRequest]) (res *connect.Response[portfoliov1.ListSecuritiesResponse], err error) {
	res = connect.NewResponse(&portfoliov1.ListSecuritiesResponse{
		Securities: svc.sec,
	})

	return
}
