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
	"github.com/oxisto/money-gopher/internal/assert"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name string
		want assert.Want[*service]
	}{
		{
			name: "Default",
			want: func(t *testing.T, s *service) bool {
				return assert.Equals(t, 1, len(s.sec))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewService()
			tt.want(t, assert.Is[*service](t, got))
		})
	}
}

func Test_service_ListSecurities(t *testing.T) {
	type fields struct {
		sec map[string]*portfoliov1.Security
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
				sec: map[string]*portfoliov1.Security{
					"My Security": {
						Name: "My Security",
					},
				},
			},
			wantRes: func(t *testing.T, r *connect.Response[portfoliov1.ListSecuritiesResponse]) bool {
				return assert.Equals(t, "My Security", r.Msg.Securities[0].Name)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				sec: tt.fields.sec,
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
