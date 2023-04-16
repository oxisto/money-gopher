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

	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/persistence"

	"github.com/bufbuild/connect-go"
	"golang.org/x/text/currency"
	"google.golang.org/protobuf/types/known/emptypb"
)

type service struct {
	// TODO(oxisto): convert this to sqlite
	sec map[string]*portfoliov1.Security

	securities persistence.StorageOperations[*portfoliov1.Security]

	portfoliov1connect.UnimplementedSecuritiesServiceHandler
}

func NewService(db *persistence.DB) portfoliov1connect.SecuritiesServiceHandler {
	securities := persistence.Ops[*portfoliov1.Security](db)
	secs := []*portfoliov1.Security{
		{
			Name:        "US0378331005",
			DisplayName: "Apple Inc.",
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
	}
	for _, sec := range secs {
		securities.Replace(sec)
	}

	return &service{
		securities: securities,
	}
}

func (svc *service) CreateSecurity(ctx context.Context, req *connect.Request[portfoliov1.CreateSecurityRequest]) (res *connect.Response[portfoliov1.Security], err error) {
	svc.sec[req.Msg.Security.Name] = req.Msg.Security
	return
}

func (svc *service) GetSecurity(ctx context.Context, req *connect.Request[portfoliov1.GetSecurityRequest]) (res *connect.Response[portfoliov1.Security], err error) {
	res = connect.NewResponse(&portfoliov1.Security{})
	res.Msg, err = svc.securities.Get(req.Msg.Name)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return
}

func (svc *service) ListSecurities(ctx context.Context, req *connect.Request[portfoliov1.ListSecuritiesRequest]) (res *connect.Response[portfoliov1.ListSecuritiesResponse], err error) {
	res = connect.NewResponse(&portfoliov1.ListSecuritiesResponse{})
	res.Msg.Securities, err = svc.securities.List()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return
}

func (svc *service) UpdateSecurity(ctx context.Context, req *connect.Request[portfoliov1.UpdateSecurityRequest]) (res *connect.Response[portfoliov1.Security], err error) {
	res = connect.NewResponse(&portfoliov1.Security{})
	res.Msg, err = svc.securities.Update(req.Msg.Security.Name, req.Msg.Security, req.Msg.UpdateMask.Paths)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return
}

func (svc *service) DeleteSecurityRequest(ctx context.Context, req *connect.Request[portfoliov1.DeleteSecurityRequest]) (res *connect.Response[emptypb.Empty], err error) {
	res = connect.NewResponse(&emptypb.Empty{})
	err = svc.securities.Delete(req.Msg.Name)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return
}
