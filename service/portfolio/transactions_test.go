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

	"github.com/bufbuild/connect-go"
	"github.com/oxisto/assert"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/persistence"
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
			name: "happy path",
			fields: fields{
				portfolios: myPortfolio(t),
			},
			args: args{
				req: connect.NewRequest(&portfoliov1.CreatePortfolioTransactionRequest{
					Transaction: &portfoliov1.PortfolioEvent{
						PortfolioName: "bank/myportfolio",
						Type:          portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY,
						SecurityName:  "My Security",
					},
				}),
			},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.PortfolioEvent]) bool {
				return assert.Equals(t, "My Security", r.Msg.GetSecurityName())
			},
			wantSvc: func(t *testing.T, s *service) bool {
				list, _ := s.events.List("bank/myportfolio")
				return assert.Equals(t, 2, len(list))
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
			gotRes, err := svc.CreatePortfolioTransaction(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.CreatePortfolioTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantRes(t, gotRes)
			tt.wantSvc(t, svc)
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
					PortfolioName: "bank/myportfolio",
				}),
			},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.ListPortfolioTransactionsResponse]) bool {
				return assert.Equals(t, 1, len(r.Msg.Transactions))
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

func Test_service_UpdatePortfolioTransactions(t *testing.T) {
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
						Id:           1,
						Type:         portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY,
						SecurityName: "My Second Security",
					},
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"security_name"}},
				}),
			},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.PortfolioEvent]) bool {
				return assert.Equals(t, "My Second Security", r.Msg.SecurityName)
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
			gotRes, err := svc.UpdatePortfolioTransactions(tt.args.ctx, tt.args.req)
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
				list, _ := s.portfolios.List("bank/myportfolio")
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
