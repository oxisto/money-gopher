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
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/oxisto/assert"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/internal"
	"github.com/oxisto/money-gopher/persistence"
	"golang.org/x/text/currency"
)

func Test_service_TriggerSecurityQuoteUpdate(t *testing.T) {
	type fields struct {
		securities persistence.StorageOperations[*portfoliov1.Security]
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.TriggerQuoteUpdateRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.Want[*portfoliov1.TriggerQuoteUpdateResponse]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				securities: internal.NewTestDBOps(t, func(ops persistence.StorageOperations[*portfoliov1.Security]) {
					ops.Replace(&portfoliov1.Security{Name: "My Security"})
					rel := persistence.Relationship[*portfoliov1.ListedSecurity](ops)
					assert.NoError(t, rel.Replace(&portfoliov1.ListedSecurity{SecurityName: "My Security", Ticker: "SEC", Currency: currency.EUR.String()}))
				}),
			},
			args: args{
				req: connect.NewRequest(&portfoliov1.TriggerQuoteUpdateRequest{
					SecurityName: "My Security",
				}),
			},
			wantRes: func(t *testing.T, tqur *portfoliov1.TriggerQuoteUpdateResponse) bool {
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				securities:       tt.fields.securities,
				listedSecurities: persistence.Relationship[*portfoliov1.ListedSecurity](tt.fields.securities),
			}
			gotRes, err := svc.TriggerSecurityQuoteUpdate(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.TriggerSecurityQuoteUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantRes(t, gotRes.Msg)
		})
	}
}

type mockQuoteProvider struct{}

func (mockQuoteProvider) LatestQuote(_ context.Context, _ *portfoliov1.ListedSecurity) (quote float32, t time.Time, err error) {
	return 100, time.Date(1, 0, 0, 0, 0, 0, 0, time.UTC), nil
}

func Test_service_updateQuote(t *testing.T) {
	type fields struct {
		securities persistence.StorageOperations[*portfoliov1.Security]
	}
	type args struct {
		qp QuoteProvider
		ls *portfoliov1.ListedSecurity
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.Want[*portfoliov1.ListedSecurity]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				securities: internal.NewTestDBOps(t, func(ops persistence.StorageOperations[*portfoliov1.Security]) {
					ops.Replace(&portfoliov1.Security{Name: "My Security"})
					rel := persistence.Relationship[*portfoliov1.ListedSecurity](ops)
					assert.NoError(t, rel.Replace(&portfoliov1.ListedSecurity{SecurityName: "My Security", Ticker: "SEC", Currency: currency.EUR.String()}))
				}),
			},
			args: args{
				qp: &mockQuoteProvider{},
				ls: &portfoliov1.ListedSecurity{SecurityName: "My Security", Ticker: "SEC", Currency: currency.EUR.String()},
			},
			want: func(t *testing.T, ls *portfoliov1.ListedSecurity) bool {
				return assert.Equals(t, 100, *ls.LatestQuote)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				securities:       tt.fields.securities,
				listedSecurities: persistence.Relationship[*portfoliov1.ListedSecurity](tt.fields.securities),
			}
			if err := svc.updateQuote(tt.args.qp, tt.args.ls); (err != nil) != tt.wantErr {
				t.Errorf("updateQuote() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.want(t, tt.args.ls)
		})
	}
}
