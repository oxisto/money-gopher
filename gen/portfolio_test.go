// Copyright 2023 Christian Banse
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This file is part of The Money Gopher.

package portfoliov1

import (
	"testing"
	"time"

	"github.com/oxisto/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestPortfolioEvent_MakeUniqueName(t *testing.T) {
	type fields struct {
		Name          string
		Type          PortfolioEventType
		Time          *timestamppb.Timestamp
		PortfolioName string
		SecurityId    string
		Amount        float64
		Price         *Currency
		Fees          *Currency
		Taxes         *Currency
	}
	tests := []struct {
		name   string
		fields fields
		want   assert.Want[*PortfolioEvent]
	}{
		{
			name: "happy path",
			fields: fields{
				SecurityId:    "stock",
				PortfolioName: "mybank-myportfolio",
				Amount:        10,
				Type:          PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY,
				Time:          timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.Local)),
			},
			want: func(t *testing.T, tx *PortfolioEvent) bool {
				return assert.Equals(t, "a04f32c39c6b9086", tx.Id)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &PortfolioEvent{
				Id:          tt.fields.Name,
				Type:        tt.fields.Type,
				Time:        tt.fields.Time,
				PortfolioId: tt.fields.PortfolioName,
				SecurityId:  tt.fields.SecurityId,
				Amount:      tt.fields.Amount,
				Price:       tt.fields.Price,
				Fees:        tt.fields.Fees,
				Taxes:       tt.fields.Taxes,
			}
			tx.MakeUniqueID()
			tt.want(t, tx)
		})
	}
}
