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

package securities

import (
	"context"
	"testing"

	"github.com/bufbuild/connect-go"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/internal"
	"github.com/oxisto/money-gopher/internal/assert"
	"github.com/oxisto/money-gopher/persistence"
	"golang.org/x/text/currency"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name string
		want assert.Want[*service]
	}{
		{
			name: "Default",
			want: func(t *testing.T, s *service) bool {
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewService(internal.NewTestDB(t))
			tt.want(t, assert.Is[*service](t, got))
		})
	}
}

func Test_service_ListSecurities(t *testing.T) {
	type fields struct {
		sec        map[string]*portfoliov1.Security
		securities persistence.StorageOperations[*portfoliov1.Security]
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.ListSecuritiesRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.Want[*connect.Response[portfoliov1.ListSecuritiesResponse]]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				securities: internal.NewTestDBOps(t, func(ops persistence.StorageOperations[*portfoliov1.Security]) {
					assert.NoError(t, ops.Replace(&portfoliov1.Security{Name: "My Security"}))
					rel := persistence.Relationship[*portfoliov1.ListedSecurity](ops)
					assert.NoError(t, rel.Replace(&portfoliov1.ListedSecurity{SecurityName: "My Security", Ticker: "SEC", Currency: currency.EUR.String()}))
				}),
			},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.ListSecuritiesResponse]) bool {
				return true &&
					assert.Equals(t, "My Security", r.Msg.Securities[0].Name) &&
					assert.Equals(t, 1, len(r.Msg.Securities[0].ListedOn))
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				sec:              tt.fields.sec,
				securities:       tt.fields.securities,
				listedSecurities: persistence.Relationship[*portfoliov1.ListedSecurity](tt.fields.securities),
			}
			gotRes, err := svc.ListSecurities(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.ListSecurities() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantRes(t, gotRes)
		})
	}
}

func Test_service_UpdateSecurity(t *testing.T) {
	type fields struct {
		securities persistence.StorageOperations[*portfoliov1.Security]
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.UpdateSecurityRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes *connect.Response[portfoliov1.Security]
		wantErr bool
	}{
		{
			name: "change display_name",
			fields: fields{
				securities: internal.NewTestDBOps(t, func(ops persistence.StorageOperations[*portfoliov1.Security]) {
					ops.Replace(&portfoliov1.Security{Name: "My Stock"})
				}),
			},
			args: args{req: connect.NewRequest(&portfoliov1.UpdateSecurityRequest{
				Security:   &portfoliov1.Security{Name: "My Stock", DisplayName: "Test"},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"display_name"}},
			})},
			wantRes: connect.NewResponse(&portfoliov1.Security{Name: "My Stock", DisplayName: "Test"}),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				securities: tt.fields.securities,
			}
			gotRes, err := svc.UpdateSecurity(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.UpdateSecurity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !proto.Equal(gotRes.Msg, tt.wantRes.Msg) {
				t.Errorf("service.UpdateSecurity() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func Test_service_GetSecurity(t *testing.T) {
	type fields struct {
		sec        map[string]*portfoliov1.Security
		securities persistence.StorageOperations[*portfoliov1.Security]
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.GetSecurityRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.Want[*portfoliov1.Security]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				securities: internal.NewTestDBOps(t, func(ops persistence.StorageOperations[*portfoliov1.Security]) {
					ops.Replace(&portfoliov1.Security{Name: "My Security"})
				}),
			},
			args: args{
				req: connect.NewRequest(&portfoliov1.GetSecurityRequest{Name: "My Security"}),
			},
			wantRes: func(t *testing.T, s *portfoliov1.Security) bool {
				return assert.Equals(t, &portfoliov1.Security{
					Name: "My Security",
				}, s)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				sec:        tt.fields.sec,
				securities: tt.fields.securities,
			}
			gotRes, err := svc.GetSecurity(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.GetSecurity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantRes(t, gotRes.Msg)
		})
	}
}

func Test_service_DeleteSecurityRequest(t *testing.T) {
	type fields struct {
		sec                                   map[string]*portfoliov1.Security
		securities                            persistence.StorageOperations[*portfoliov1.Security]
		UnimplementedSecuritiesServiceHandler portfoliov1connect.UnimplementedSecuritiesServiceHandler
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.DeleteSecurityRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.Want[*emptypb.Empty]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				securities: internal.NewTestDBOps(t, func(ops persistence.StorageOperations[*portfoliov1.Security]) {
					ops.Replace(&portfoliov1.Security{Name: "My Stock"})
				}),
			},
			args: args{req: connect.NewRequest(&portfoliov1.DeleteSecurityRequest{
				Name: "My Stock",
			})},
			wantRes: func(t *testing.T, e *emptypb.Empty) bool {
				return assert.Equals(t, &emptypb.Empty{}, e)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				sec:        tt.fields.sec,
				securities: tt.fields.securities,
			}
			gotRes, err := svc.DeleteSecurityRequest(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.DeleteSecurityRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantRes(t, gotRes.Msg)
		})
	}
}
