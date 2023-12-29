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
	"time"

	"connectrpc.com/connect"
	"github.com/oxisto/assert"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/internal"
	"github.com/oxisto/money-gopher/persistence"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func myPortfolio(t *testing.T) persistence.StorageOperations[*portfoliov1.Portfolio] {
	return internal.NewTestDBOps(t, func(ops persistence.StorageOperations[*portfoliov1.Portfolio]) {
		assert.NoError(t, ops.Replace(&portfoliov1.Portfolio{
			Name:        "bank/myportfolio",
			DisplayName: "My Portfolio",
		}))
		rel := persistence.Relationship[*portfoliov1.PortfolioEvent](ops)
		assert.NoError(t, rel.Replace(&portfoliov1.PortfolioEvent{
			Name:          "buy",
			Type:          portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY,
			PortfolioName: "bank/myportfolio",
			SecurityName:  "US0378331005",
			Amount:        20,
			Price:         portfoliov1.Value(10708),
			Fees:          portfoliov1.Value(1025),
			Time:          timestamppb.New(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
		}))
		assert.NoError(t, rel.Replace(&portfoliov1.PortfolioEvent{
			Name:          "sell",
			Type:          portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_SELL,
			PortfolioName: "bank/myportfolio",
			SecurityName:  "US0378331005",
			Amount:        10,
			Price:         portfoliov1.Value(14588),
			Fees:          portfoliov1.Value(855),
			Time:          timestamppb.New(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)),
		}))
	})
}

func myCash(t *testing.T) persistence.StorageOperations[*portfoliov1.BankAccount] {
	return internal.NewTestDBOps(t, func(ops persistence.StorageOperations[*portfoliov1.BankAccount]) {
		assert.NoError(t, ops.Replace(&portfoliov1.BankAccount{
			Name:        "bank/mycash",
			DisplayName: "My Cash",
		}))
	})
}

func zeroPositions(t *testing.T) persistence.StorageOperations[*portfoliov1.Portfolio] {
	return internal.NewTestDBOps(t, func(ops persistence.StorageOperations[*portfoliov1.Portfolio]) {
		assert.NoError(t, ops.Replace(&portfoliov1.Portfolio{
			Name:        "bank/myportfolio",
			DisplayName: "My Portfolio",
		}))
		rel := persistence.Relationship[*portfoliov1.PortfolioEvent](ops)
		assert.NoError(t, rel.Replace(&portfoliov1.PortfolioEvent{
			Name:          "buy",
			Type:          portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY,
			PortfolioName: "bank/myportfolio",
			SecurityName:  "sec123",
			Amount:        10,
			Price:         portfoliov1.Value(10000),
			Fees:          portfoliov1.Zero(),
			Time:          timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
		}))
		assert.NoError(t, rel.Replace(&portfoliov1.PortfolioEvent{
			Name:          "sell",
			Type:          portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_SELL,
			PortfolioName: "bank/myportfolio",
			SecurityName:  "sec123",
			Amount:        10,
			Price:         portfoliov1.Value(10000),
			Fees:          portfoliov1.Zero(),
			Time:          timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
		}))
	})
}

func emptyPortfolio(t *testing.T) persistence.StorageOperations[*portfoliov1.Portfolio] {
	return internal.NewTestDBOps(t, func(ops persistence.StorageOperations[*portfoliov1.Portfolio]) {
		assert.NoError(t, ops.Replace(&portfoliov1.Portfolio{
			Name:        "bank/myportfolio",
			DisplayName: "My Portfolio",
		}))
	})
}

func Test_service_CreatePortfolio(t *testing.T) {
	type fields struct {
		portfolios persistence.StorageOperations[*portfoliov1.Portfolio]
		securities portfoliov1connect.SecuritiesServiceClient
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.CreatePortfolioRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.Want[*connect.Response[portfoliov1.Portfolio]]
		wantSvc assert.Want[*service]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				portfolios: internal.NewTestDBOps[*portfoliov1.Portfolio](t),
			},
			args: args{
				req: connect.NewRequest(&portfoliov1.CreatePortfolioRequest{
					Portfolio: &portfoliov1.Portfolio{
						Name:        "bank/myportfolio",
						DisplayName: "My Portfolio",
					},
				}),
			},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.Portfolio]) bool {
				return true &&
					assert.Equals(t, "bank/myportfolio", r.Msg.Name) &&
					assert.Equals(t, "My Portfolio", r.Msg.DisplayName)
			},
			wantSvc: func(t *testing.T, s *service) bool {
				list, _ := s.portfolios.List()
				return assert.Equals(t, 1, len(list))
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
			gotRes, err := svc.CreatePortfolio(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.CreatePortfolio() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantRes(t, gotRes)
			tt.wantSvc(t, svc)
		})
	}
}

func Test_service_ListPortfolios(t *testing.T) {
	type fields struct {
		portfolios persistence.StorageOperations[*portfoliov1.Portfolio]
		securities portfoliov1connect.SecuritiesServiceClient
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.ListPortfoliosRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.Want[*connect.Response[portfoliov1.ListPortfoliosResponse]]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				portfolios: myPortfolio(t),
			},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.ListPortfoliosResponse]) bool {
				return true &&
					assert.Equals(t, "bank/myportfolio", r.Msg.Portfolios[0].Name) &&
					assert.Equals(t, "My Portfolio", r.Msg.Portfolios[0].DisplayName) &&
					assert.Equals(t, 2, len(r.Msg.Portfolios[0].Events))
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
			gotRes, err := svc.ListPortfolios(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.ListPortfolios() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantRes(t, gotRes)
		})
	}
}

func Test_service_GetPortfolio(t *testing.T) {
	type fields struct {
		portfolios persistence.StorageOperations[*portfoliov1.Portfolio]
		securities portfoliov1connect.SecuritiesServiceClient
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.GetPortfolioRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.Want[*connect.Response[portfoliov1.Portfolio]]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				portfolios: myPortfolio(t),
			},
			args: args{
				req: connect.NewRequest(&portfoliov1.GetPortfolioRequest{
					Name: "bank/myportfolio",
				}),
			},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.Portfolio]) bool {
				return true &&
					assert.Equals(t, 2, len(r.Msg.Events))
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
			gotRes, err := svc.GetPortfolio(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.GetPortfolio() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantRes(t, gotRes)
		})
	}
}

func Test_service_UpdatePortfolio(t *testing.T) {
	type fields struct {
		portfolios persistence.StorageOperations[*portfoliov1.Portfolio]
		securities portfoliov1connect.SecuritiesServiceClient
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.UpdatePortfolioRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.Want[*connect.Response[portfoliov1.Portfolio]]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				portfolios: myPortfolio(t),
			},
			args: args{
				req: connect.NewRequest(&portfoliov1.UpdatePortfolioRequest{
					Portfolio: &portfoliov1.Portfolio{
						Name:        "bank/myportfolio",
						DisplayName: "My Second Portfolio",
					},
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"display_name"}},
				}),
			},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.Portfolio]) bool {
				return assert.Equals(t, "My Second Portfolio", r.Msg.DisplayName)
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
			gotRes, err := svc.UpdatePortfolio(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.UpdatePortfolio() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantRes(t, gotRes)
		})
	}
}

func Test_service_DeletePortfolio(t *testing.T) {
	type fields struct {
		portfolios                           persistence.StorageOperations[*portfoliov1.Portfolio]
		events                               persistence.StorageOperations[*portfoliov1.PortfolioEvent]
		securities                           portfoliov1connect.SecuritiesServiceClient
		UnimplementedPortfolioServiceHandler portfoliov1connect.UnimplementedPortfolioServiceHandler
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.DeletePortfolioRequest]
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
				req: connect.NewRequest(&portfoliov1.DeletePortfolioRequest{
					Name: "bank/myportfolio",
				}),
			},
			wantSvc: func(t *testing.T, s *service) bool {
				list, _ := s.portfolios.List()
				return assert.Equals(t, 0, len(list))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				portfolios:                           tt.fields.portfolios,
				events:                               tt.fields.events,
				securities:                           tt.fields.securities,
				UnimplementedPortfolioServiceHandler: tt.fields.UnimplementedPortfolioServiceHandler,
			}
			_, err := svc.DeletePortfolio(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.DeletePortfolio() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantSvc(t, svc)
		})
	}
}
