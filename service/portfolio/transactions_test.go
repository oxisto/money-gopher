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
	"testing"

	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/persistence"

	"connectrpc.com/connect"
	"github.com/oxisto/assert"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func Test_service_CreatePortfolioTransaction(t *testing.T) {
	type fields struct {
		portfolios persistence.StorageOperations[*portfoliov1.Portfolio]
		securities portfoliov1connect.SecuritiesServiceClient
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.CreatePortfolioTransactionRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.Want[*connect.Response[portfoliov1.PortfolioEvent]]
		wantSvc assert.Want[*service]
		wantErr bool
	}{
		{
			name: "happy path buy",
			fields: fields{
				portfolios: myPortfolio(t),
			},
			args: args{
				req: connect.NewRequest(&portfoliov1.CreatePortfolioTransactionRequest{
					Transaction: &portfoliov1.PortfolioEvent{
						PortfolioName: "mybank-myportfolio",
						Type:          portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY,
						SecurityId:    "My Security",
						Amount:        1,
						Price:         portfoliov1.Value(2000),
					},
				}),
			},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.PortfolioEvent]) bool {
				return assert.Equals(t, "My Security", r.Msg.GetSecurityId())
			},
			wantSvc: func(t *testing.T, s *service) bool {
				list, _ := s.events.List("mybank-myportfolio")
				return assert.Equals(t, 3, len(list))
			},
		},
		{
			name: "happy path sell",
			fields: fields{
				portfolios: myPortfolio(t),
			},
			args: args{
				req: connect.NewRequest(&portfoliov1.CreatePortfolioTransactionRequest{
					Transaction: &portfoliov1.PortfolioEvent{
						PortfolioName: "mybank-myportfolio",
						Type:          portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_SELL,
						SecurityId:    "My Security",
						Amount:        1,
						Price:         portfoliov1.Value(2000),
					},
				}),
			},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.PortfolioEvent]) bool {
				return assert.Equals(t, "My Security", r.Msg.GetSecurityId())
			},
			wantSvc: func(t *testing.T, s *service) bool {
				list, _ := s.events.List("mybank-myportfolio")
				return assert.Equals(t, 3, len(list))
			},
		},
		{
			name: "missing security name",
			fields: fields{
				portfolios: myPortfolio(t),
			},
			args: args{
				req: connect.NewRequest(&portfoliov1.CreatePortfolioTransactionRequest{
					Transaction: &portfoliov1.PortfolioEvent{
						PortfolioName: "mybank-myportfolio",
						Type:          portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_SELL,
						Amount:        1,
						Price:         portfoliov1.Value(2000),
					},
				}),
			},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.PortfolioEvent]) bool {
				if r != nil {
					t.Fatal("not nil")
				}

				return true
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				portfolios: tt.fields.portfolios,
				events:     persistence.Relationship[*portfoliov1.PortfolioEvent](tt.fields.portfolios),
				securities: tt.fields.securities,
			}
			gotRes, err := svc.CreatePortfolioTransaction(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.CreatePortfolioTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tt.wantRes(t, gotRes)
			if tt.wantSvc != nil {
				tt.wantSvc(t, svc)
			}
		})
	}
}

func Test_service_GetPortfolioTransaction(t *testing.T) {
	type fields struct {
		portfolios persistence.StorageOperations[*portfoliov1.Portfolio]
		securities portfoliov1connect.SecuritiesServiceClient
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.GetPortfolioTransactionRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.Want[*connect.Response[portfoliov1.PortfolioEvent]]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				portfolios: myPortfolio(t),
			},
			args: args{
				req: connect.NewRequest(&portfoliov1.GetPortfolioTransactionRequest{
					Name: "buy",
				}),
			},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.PortfolioEvent]) bool {
				return assert.Equals(t, "buy", r.Msg.Name) && assert.Equals(t, "mybank-myportfolio", r.Msg.PortfolioName)
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
			gotRes, err := svc.GetPortfolioTransaction(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.GetPortfolioTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantRes(t, gotRes)
		})
	}
}

func Test_service_ListPortfolioTransactions(t *testing.T) {
	type fields struct {
		portfolios persistence.StorageOperations[*portfoliov1.Portfolio]
		securities portfoliov1connect.SecuritiesServiceClient
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.ListPortfolioTransactionsRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.Want[*connect.Response[portfoliov1.ListPortfolioTransactionsResponse]]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				portfolios: myPortfolio(t),
			},
			args: args{
				req: connect.NewRequest(&portfoliov1.ListPortfolioTransactionsRequest{
					PortfolioName: "mybank-myportfolio",
				}),
			},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.ListPortfolioTransactionsResponse]) bool {
				return assert.Equals(t, 2, len(r.Msg.Transactions))
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
			gotRes, err := svc.ListPortfolioTransactions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.ListPortfolioTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantRes(t, gotRes)
		})
	}
}

func Test_service_UpdatePortfolioTransaction(t *testing.T) {
	type fields struct {
		portfolios persistence.StorageOperations[*portfoliov1.Portfolio]
		securities portfoliov1connect.SecuritiesServiceClient
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.UpdatePortfolioTransactionRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.Want[*connect.Response[portfoliov1.PortfolioEvent]]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				portfolios: myPortfolio(t),
			},
			args: args{
				req: connect.NewRequest(&portfoliov1.UpdatePortfolioTransactionRequest{
					Transaction: &portfoliov1.PortfolioEvent{
						Name:       "buy",
						Type:       portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY,
						SecurityId: "My Second Security",
					},
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"security_id"}},
				}),
			},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.PortfolioEvent]) bool {
				return assert.Equals(t, "My Second Security", r.Msg.SecurityId)
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
			gotRes, err := svc.UpdatePortfolioTransaction(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.UpdatePortfolioTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantRes(t, gotRes)
		})
	}
}

func Test_service_DeletePortfolioTransactions(t *testing.T) {
	type fields struct {
		portfolios persistence.StorageOperations[*portfoliov1.Portfolio]
		securities portfoliov1connect.SecuritiesServiceClient
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.DeletePortfolioTransactionRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantSvc assert.Want[*service]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				portfolios: myPortfolio(t),
			},
			args: args{
				req: connect.NewRequest(&portfoliov1.DeletePortfolioTransactionRequest{
					TransactionId: 1,
				}),
			},
			wantSvc: func(t *testing.T, s *service) bool {
				list, _ := s.portfolios.List("mybank-myportfolio")
				return assert.Equals(t, 0, len(list))
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
			_, err := svc.DeletePortfolioTransactions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.DeletePortfolioTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantSvc(t, svc)
		})
	}
}

func Test_service_ImportTransactions(t *testing.T) {
	type fields struct {
		portfolios persistence.StorageOperations[*portfoliov1.Portfolio]
		securities portfoliov1connect.SecuritiesServiceClient
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.ImportTransactionsRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantSvc assert.Want[*service]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				portfolios: emptyPortfolio(t),
				securities: &mockSecuritiesClient{},
			},
			args: args{
				req: connect.NewRequest(&portfoliov1.ImportTransactionsRequest{
					PortfolioName: "mybank-myportfolio",
					FromCsv: `Date;Type;Value;Transaction Currency;Gross Amount;Currency Gross Amount;Exchange Rate;Fees;Taxes;Shares;ISIN;WKN;Ticker Symbol;Security Name;Note
2021-06-05T00:00;Buy;2.151,85;EUR;;;;10,25;0,00;20;US0378331005;865985;APC.F;Apple Inc.;
2021-06-05T00:00;Sell;-2.151,85;EUR;;;;10,25;0,00;20;US0378331005;865985;APC.F;Apple Inc.;
2021-06-18T00:00;Buy;912,66;EUR;;;;7,16;0,00;5;US09075V1026;A2PSR2;22UA.F;BioNTech SE;`,
				}),
			},
			wantSvc: func(t *testing.T, s *service) bool {
				txs, err := s.events.List("mybank-myportfolio")
				return true &&
					assert.NoError(t, err) &&
					assert.Equals(t, 3, len(txs))
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
			_, err := svc.ImportTransactions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.ImportTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantSvc(t, svc)
		})
	}
}
