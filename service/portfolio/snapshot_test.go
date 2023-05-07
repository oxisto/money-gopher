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
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/bufbuild/connect-go"
)

func Test_service_GetPortfolioSnapshot(t *testing.T) {
	type fields struct {
		// portfolio *portfoliov1.Portfolio
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
			args: args{req: connect.NewRequest(&portfoliov1.GetPortfolioSnapshotRequest{})},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.PortfolioSnapshot]) bool {
				return true &&
					assert.Equals(t, "US0378331005", r.Msg.Positions["US0378331005"].SecurityName) &&
					assert.Equals(t, 10, r.Msg.Positions["US0378331005"].Amount) &&
					assert.Equals(t, 1070.8, r.Msg.Positions["US0378331005"].PurchaseValue) &&
					assert.Equals(t, 107.08, r.Msg.Positions["US0378331005"].PurchasePrice)
			},
		},
		{
			name: "happy path, before sell",
			args: args{req: connect.NewRequest(&portfoliov1.GetPortfolioSnapshotRequest{
				Time: timestamppb.New(time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC)),
			})},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.PortfolioSnapshot]) bool {
				return true &&
					assert.Equals(t, "US0378331005", r.Msg.Positions["US0378331005"].SecurityName) &&
					assert.Equals(t, 20, r.Msg.Positions["US0378331005"].Amount) &&
					assert.Equals(t, 2141.6, r.Msg.Positions["US0378331005"].PurchaseValue) &&
					assert.Equals(t, 107.08, r.Msg.Positions["US0378331005"].PurchasePrice)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For now, we just use the portfolio of NewService
			// svc := &service{
			// 	portfolio: tt.fields.portfolio,
			// }
			svc := NewService()
			gotRes, err := svc.GetPortfolioSnapshot(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.GetPortfolioSnapshot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantRes(t, gotRes)
		})
	}
}
