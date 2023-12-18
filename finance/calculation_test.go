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

// package portfolio contains all kinds of different finance calculations.
package finance

import (
	"testing"

	"github.com/oxisto/assert"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
)

func TestNewCalculation(t *testing.T) {
	type args struct {
		txs []*portfoliov1.PortfolioEvent
	}
	tests := []struct {
		name string
		args args
		want assert.Want[*calculation]
	}{
		{
			name: "buy and sell",
			args: args{
				txs: []*portfoliov1.PortfolioEvent{
					{
						Type:   portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY,
						Amount: 5,
						Price:  portfoliov1.Value(18110),
						Fees:   portfoliov1.Value(716),
					},
					{
						Type:   portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_SELL,
						Amount: 2,
						Price:  portfoliov1.Value(30430),
						Fees:   portfoliov1.Value(642),
						Taxes:  portfoliov1.Value(1632),
					},
					{
						Type:   portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY,
						Amount: 5,
						Price:  portfoliov1.Value(29000),
						Fees:   portfoliov1.Value(853),
					},
					{
						Type:   portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_SELL,
						Amount: 3,
						Price:  portfoliov1.Value(22000),
						Fees:   portfoliov1.Value(845),
					},
					{
						Type:   portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY,
						Amount: 5,
						Price:  portfoliov1.Value(20330),
						Fees:   portfoliov1.Value(744),
					},
					{
						Type:   portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY,
						Amount: 5,
						Price:  portfoliov1.Value(19645),
						Fees:   portfoliov1.Value(736),
					},
					{
						Type:   portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY,
						Amount: 10,
						Price:  portfoliov1.Value(14655),
						Fees:   portfoliov1.Value(856),
					},
				},
			},
			want: func(t *testing.T, c *calculation) bool {
				return true &&
					assert.Equals(t, 25, c.Amount) &&
					assert.Equals(t, 491425, int(c.NetValue().Value)) &&
					assert.Equals(t, 494614, int(c.GrossValue().Value)) &&
					assert.Equals(t, 19657, int(c.NetPrice().Value)) &&
					assert.Equals(t, 19784, int(c.GrossPrice().Value))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCalculation(tt.args.txs)
			tt.want(t, got)
		})
	}
}
