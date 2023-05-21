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
	"io"
	"testing"
	"time"

	"github.com/oxisto/assert"
	moneygopher "github.com/oxisto/money-gopher"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/internal"
	"github.com/oxisto/money-gopher/persistence"
	"golang.org/x/text/currency"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/bufbuild/connect-go"
)

var mockSecuritiesClientWithData = &mockSecuritiesClient{
	securities: []*portfoliov1.Security{
		{
			Name:        "US0378331005",
			DisplayName: "Apple, Inc.",
			ListedOn: []*portfoliov1.ListedSecurity{
				{
					SecurityName:         "US0378331005",
					Ticker:               "APC.F",
					Currency:             currency.EUR.String(),
					LatestQuote:          moneygopher.Ref(float32(100.0)),
					LatestQuoteTimestamp: timestamppb.Now(),
				},
			},
		},
	},
}

func Test_service_GetPortfolioSnapshot(t *testing.T) {
	type fields struct {
		portfolios persistence.StorageOperations[*portfoliov1.Portfolio]
		events     persistence.StorageOperations[*portfoliov1.PortfolioEvent]
		securities portfoliov1connect.SecuritiesServiceClient
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.GetPortfolioSnapshotRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.Want[*connect.Response[portfoliov1.PortfolioSnapshot]]
		wantErr bool
	}{
		{
			name: "happy path, now",
			fields: fields{
				portfolios: myPortfolio(t),
				securities: mockSecuritiesClientWithData,
			},
			args: args{req: connect.NewRequest(&portfoliov1.GetPortfolioSnapshotRequest{
				PortfolioName: "bank/myportfolio",
			})},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.PortfolioSnapshot]) bool {
				return true &&
					assert.Equals(t, "US0378331005", r.Msg.Positions["US0378331005"].Security.Name) &&
					assert.Equals(t, 10, r.Msg.Positions["US0378331005"].Amount) &&
					assert.Equals(t, 1070.8, r.Msg.Positions["US0378331005"].PurchaseValue) &&
					assert.Equals(t, 107.08, r.Msg.Positions["US0378331005"].PurchasePrice) &&
					assert.Equals(t, 1000.0, r.Msg.TotalMarketValue)
			},
		},
		{
			name: "happy path, before sell",
			fields: fields{
				portfolios: myPortfolio(t),
				securities: mockSecuritiesClientWithData,
			},
			args: args{req: connect.NewRequest(&portfoliov1.GetPortfolioSnapshotRequest{
				PortfolioName: "bank/myportfolio",
				Time:          timestamppb.New(time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC)),
			})},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.PortfolioSnapshot]) bool {
				pos := r.Msg.Positions["US0378331005"]

				return true &&
					assert.Equals(t, "US0378331005", pos.Security.Name) &&
					assert.Equals(t, 20, pos.Amount) &&
					assert.Equals(t, 2141.6, pos.PurchaseValue) &&
					assert.Equals(t, 107.08, pos.PurchasePrice) &&
					assert.Equals(t, 100.0, pos.MarketPrice) &&
					assert.Equals(t, 2000.0, pos.MarketValue)
			},
		},
		{
			name: "happy path, position zero'd out",
			fields: fields{
				portfolios: zeroPositions(t),
				securities: mockSecuritiesClientWithData,
			},
			args: args{req: connect.NewRequest(&portfoliov1.GetPortfolioSnapshotRequest{
				PortfolioName: "bank/myportfolio",
				Time:          timestamppb.New(time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC)),
			})},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.PortfolioSnapshot]) bool {
				return true &&
					len(r.Msg.Positions) == 0
			},
		},
		{
			name: "events list error",
			fields: fields{
				portfolios: emptyPortfolio(t),
				events:     internal.ErrOps[*portfoliov1.PortfolioEvent](io.EOF),
				securities: &mockSecuritiesClient{listSecuritiesError: io.EOF},
			},
			args: args{req: connect.NewRequest(&portfoliov1.GetPortfolioSnapshotRequest{
				PortfolioName: "bank/myportfolio",
				Time:          timestamppb.New(time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC)),
			})},
			wantErr: true,
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.PortfolioSnapshot]) bool {
				return true
			},
		},
		{
			name: "securities list error",
			fields: fields{
				portfolios: myPortfolio(t),
				securities: &mockSecuritiesClient{listSecuritiesError: io.EOF},
			},
			args: args{req: connect.NewRequest(&portfoliov1.GetPortfolioSnapshotRequest{
				PortfolioName: "bank/myportfolio",
				Time:          timestamppb.New(time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC)),
			})},
			wantErr: true,
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.PortfolioSnapshot]) bool {
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				portfolios: tt.fields.portfolios,
				events:     persistence.Relationship[*portfoliov1.PortfolioEvent](tt.fields.portfolios),
				securities: tt.fields.securities,
			}

			if tt.fields.events != nil {
				svc.events = tt.fields.events
			}

			gotRes, err := svc.GetPortfolioSnapshot(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.GetPortfolioSnapshot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantRes(t, gotRes)
		})
	}
}

type mockSecuritiesClient struct {
	securities          []*portfoliov1.Security
	listSecuritiesError error
}

func (m *mockSecuritiesClient) ListSecurities(context.Context, *connect.Request[portfoliov1.ListSecuritiesRequest]) (*connect.Response[portfoliov1.ListSecuritiesResponse], error) {
	return connect.NewResponse(&portfoliov1.ListSecuritiesResponse{
		Securities: m.securities,
	}), m.listSecuritiesError
}

func (*mockSecuritiesClient) GetSecurity(context.Context, *connect.Request[portfoliov1.GetSecurityRequest]) (*connect.Response[portfoliov1.Security], error) {
	return nil, nil
}

func (*mockSecuritiesClient) CreateSecurity(context.Context, *connect.Request[portfoliov1.CreateSecurityRequest]) (*connect.Response[portfoliov1.Security], error) {
	return nil, nil
}

func (*mockSecuritiesClient) UpdateSecurity(context.Context, *connect.Request[portfoliov1.UpdateSecurityRequest]) (*connect.Response[portfoliov1.Security], error) {
	return nil, nil
}

func (*mockSecuritiesClient) DeleteSecurity(context.Context, *connect.Request[portfoliov1.DeleteSecurityRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, nil
}

func (*mockSecuritiesClient) TriggerSecurityQuoteUpdate(context.Context, *connect.Request[portfoliov1.TriggerQuoteUpdateRequest]) (*connect.Response[portfoliov1.TriggerQuoteUpdateResponse], error) {
	return nil, nil
}
