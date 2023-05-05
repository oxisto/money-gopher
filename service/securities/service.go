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
	"time"

	moneygopher "github.com/oxisto/money-gopher"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/persistence"

	"github.com/bufbuild/connect-go"
	"golang.org/x/exp/slices"
	"golang.org/x/text/currency"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type service struct {
	// TODO(oxisto): convert this to sqlite
	sec map[string]*portfoliov1.Security

	securities       persistence.StorageOperations[*portfoliov1.Security]
	listedSecurities persistence.StorageOperations[*portfoliov1.ListedSecurity]

	portfoliov1connect.UnimplementedSecuritiesServiceHandler
}

func NewService(db *persistence.DB) portfoliov1connect.SecuritiesServiceHandler {
	securities := persistence.Ops[*portfoliov1.Security](db)
	listedSecurities := persistence.Relationship[*portfoliov1.ListedSecurity](securities)
	secs := []*portfoliov1.Security{
		{
			Name:        "US0378331005",
			DisplayName: "Apple Inc.",
			ListedOn: []*portfoliov1.ListedSecurity{
				{
					SecurityName:         "US0378331005",
					Ticker:               "APC.F",
					Currency:             currency.EUR.String(),
					LatestQuote:          moneygopher.Ref(float32(150.16)),
					LatestQuoteTimestamp: timestamppb.New(time.Date(2023, 4, 21, 0, 0, 0, 0, time.Local)),
				},
				{
					SecurityName:         "US0378331005",
					Ticker:               "AAPL",
					Currency:             currency.USD.String(),
					LatestQuote:          moneygopher.Ref(float32(165.02)),
					LatestQuoteTimestamp: timestamppb.New(time.Date(2023, 4, 21, 0, 0, 0, 0, time.Local)),
				},
			},
			QuoteProvider: moneygopher.Ref(QuoteProviderYF),
		},
	}
	for _, sec := range secs {
		securities.Replace(sec)

		// TODO: in the future, we might do this automatically
		for _, ls := range sec.ListedOn {
			listedSecurities.Replace(ls)
		}
	}

	return &service{
		securities:       securities,
		listedSecurities: listedSecurities,
	}
}

func (svc *service) CreateSecurity(ctx context.Context, req *connect.Request[portfoliov1.CreateSecurityRequest]) (res *connect.Response[portfoliov1.Security], err error) {
	svc.sec[req.Msg.Security.Name] = req.Msg.Security
	return
}

func (svc *service) GetSecurity(ctx context.Context, req *connect.Request[portfoliov1.GetSecurityRequest]) (res *connect.Response[portfoliov1.Security], err error) {
	res = connect.NewResponse(&portfoliov1.Security{})
	res.Msg, err = svc.fetchSecurity(req.Msg.Name)
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

	for _, sec := range res.Msg.Securities {
		sec.ListedOn, err = svc.listedSecurities.List(sec.Name)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	return
}

func (svc *service) UpdateSecurity(ctx context.Context, req *connect.Request[portfoliov1.UpdateSecurityRequest]) (res *connect.Response[portfoliov1.Security], err error) {
	res = connect.NewResponse(&portfoliov1.Security{})
	res.Msg, err = svc.securities.Update(req.Msg.Security.Name, req.Msg.Security, req.Msg.UpdateMask.Paths)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if slices.Contains(req.Msg.UpdateMask.Paths, "listed_on") {
		for _, ls := range req.Msg.Security.ListedOn {
			svc.listedSecurities.Replace(ls)
		}
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

func (svc *service) fetchSecurity(name string) (sec *portfoliov1.Security, err error) {
	sec, err = svc.securities.Get(name)
	if err != nil {
		return nil, err
	}

	sec.ListedOn, err = svc.listedSecurities.List(name)
	if err != nil {
		return nil, err
	}

	return
}
