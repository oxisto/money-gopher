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
	"github.com/oxisto/money-gopher/internal"
	"github.com/oxisto/money-gopher/persistence"

	"connectrpc.com/connect"
	"github.com/oxisto/assert"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func Test_service_CreateBankAccount(t *testing.T) {
	type fields struct {
		portfolios   persistence.StorageOperations[*portfoliov1.Portfolio]
		bankAccounts persistence.StorageOperations[*portfoliov1.BankAccount]
		securities   portfoliov1connect.SecuritiesServiceClient
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.CreateBankAccountRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.Want[*connect.Response[portfoliov1.BankAccount]]
		wantSvc assert.Want[*service]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				portfolios:   internal.NewTestDBOps[*portfoliov1.Portfolio](t),
				bankAccounts: internal.NewTestDBOps[*portfoliov1.BankAccount](t),
			},
			args: args{
				req: connect.NewRequest(&portfoliov1.CreateBankAccountRequest{
					BankAccount: &portfoliov1.BankAccount{
						Name:        "bank/mycash",
						DisplayName: "My Cash Account",
					},
				}),
			},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.BankAccount]) bool {
				return true &&
					assert.Equals(t, "bank/mycash", r.Msg.Name) &&
					assert.Equals(t, "My Cash Account", r.Msg.DisplayName)
			},
			wantSvc: func(t *testing.T, s *service) bool {
				list, _ := s.bankAccounts.List()
				return assert.Equals(t, 1, len(list))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				portfolios:   tt.fields.portfolios,
				events:       persistence.Relationship[*portfoliov1.PortfolioEvent](tt.fields.portfolios),
				bankAccounts: tt.fields.bankAccounts,
				securities:   tt.fields.securities,
			}
			gotRes, err := svc.CreateBankAccount(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.CreateBankAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantRes(t, gotRes)
			tt.wantSvc(t, svc)
		})
	}
}

func Test_service_UpdateBankAccount(t *testing.T) {
	type fields struct {
		portfolios   persistence.StorageOperations[*portfoliov1.Portfolio]
		bankAccounts persistence.StorageOperations[*portfoliov1.BankAccount]
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.UpdateBankAccountRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.Want[*connect.Response[portfoliov1.BankAccount]]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				portfolios:   myPortfolio(t),
				bankAccounts: myCash(t),
			},
			args: args{
				req: connect.NewRequest(&portfoliov1.UpdateBankAccountRequest{
					Account: &portfoliov1.BankAccount{
						Name:        "bank/mycash",
						DisplayName: "My Cash",
					},
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"display_name"}},
				}),
			},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.BankAccount]) bool {
				return assert.Equals(t, "My Cash", r.Msg.DisplayName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				portfolios:   tt.fields.portfolios,
				bankAccounts: tt.fields.bankAccounts,
			}
			gotRes, err := svc.UpdateBankAccount(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.UpdateBankAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantRes(t, gotRes)
		})
	}
}

func Test_service_DeleteBankAccount(t *testing.T) {
	type fields struct {
		portfolios   persistence.StorageOperations[*portfoliov1.Portfolio]
		bankAccounts persistence.StorageOperations[*portfoliov1.BankAccount]
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.DeleteBankAccountRequest]
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
				portfolios:   myPortfolio(t),
				bankAccounts: myCash(t),
			},
			args: args{
				req: connect.NewRequest(&portfoliov1.DeleteBankAccountRequest{
					Name: "bank/mycash",
				}),
			},
			wantSvc: func(t *testing.T, s *service) bool {
				list, _ := s.bankAccounts.List()
				return assert.Equals(t, 0, len(list))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				portfolios:   tt.fields.portfolios,
				bankAccounts: tt.fields.bankAccounts,
			}
			_, err := svc.DeleteBankAccount(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.DeleteBankAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantSvc(t, svc)
		})
	}
}
