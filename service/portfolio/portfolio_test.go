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

	"github.com/oxisto/assert"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/internal"
	"github.com/oxisto/money-gopher/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/bufbuild/connect-go"
)

func Test_service_ListPortfolios(t *testing.T) {
	type fields struct {
		portfolios persistence.StorageOperations[*portfoliov1.Portfolio]
		securities portfoliov1connect.SecuritiesServiceClient
	}
	type args struct {
		ctx context.Context
		req *connect.Request[portfoliov1.ListPortfolioRequest]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.Want[*connect.Response[portfoliov1.ListPortfolioResponse]]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				portfolios: internal.NewTestDBOps(t, func(ops persistence.StorageOperations[*portfoliov1.Portfolio]) {
					assert.NoError(t, ops.Replace(&portfoliov1.Portfolio{
						Name:        "bank/myportfolio",
						DisplayName: "My Portfolio",
					}))
					rel := persistence.Relationship[*portfoliov1.PortfolioEvent](ops)
					assert.NoError(t, rel.Replace(&portfoliov1.PortfolioEvent{
						Id:            1,
						PortfolioName: "bank/myportfolio",
						SecurityName:  "My Security",
						Type:          portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY,
						Time:          timestamppb.New(time.Date(2022, 1, 0, 0, 0, 0, 0, time.UTC)),
						Amount:        10,
						Price:         100.0,
					}))
				}),
			},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.ListPortfolioResponse]) bool {
				return true &&
					assert.Equals(t, "bank/myportfolio", r.Msg.Portfolios[0].Name) &&
					assert.Equals(t, "My Portfolio", r.Msg.Portfolios[0].DisplayName) &&
					assert.Equals(t, 1, len(r.Msg.Portfolios[0].Events))
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
