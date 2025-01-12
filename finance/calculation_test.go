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

// package finance contains all kinds of different finance calculations.
package finance

import (
	"testing"

	"github.com/oxisto/assert"
	"github.com/oxisto/money-gopher/currency"
	"github.com/oxisto/money-gopher/persistence"
	"github.com/oxisto/money-gopher/portfolio/events"
)

func TestNewCalculation(t *testing.T) {
	type args struct {
		txs []*persistence.Transaction
	}
	tests := []struct {
		name string
		args args
		want assert.Want[*calculation]
	}{
		{
			name: "buy and sell",
			args: args{
				txs: []*persistence.Transaction{
					{
						Type:  events.PortfolioEventTypeDepositCash,
						Price: currency.Value(500000),
					},
					{
						Type:   events.PortfolioEventTypeBuy,
						Amount: 5,
						Price:  currency.Value(18110),
						Fees:   currency.Value(716),
					},
					{
						Type:   events.PortfolioEventTypeSell,
						Amount: 2,
						Price:  currency.Value(30430),
						Fees:   currency.Value(642),
						Taxes:  currency.Value(1632),
					},
					{
						Type:   events.PortfolioEventTypeBuy,
						Amount: 5,
						Price:  currency.Value(29000),
						Fees:   currency.Value(853),
					},
					{
						Type:   events.PortfolioEventTypeSell,
						Amount: 3,
						Price:  currency.Value(22000),
						Fees:   currency.Value(845),
					},
					{
						Type:   events.PortfolioEventTypeBuy,
						Amount: 5,
						Price:  currency.Value(20330),
						Fees:   currency.Value(744),
					},
					{
						Type:   events.PortfolioEventTypeBuy,
						Amount: 5,
						Price:  currency.Value(19645),
						Fees:   currency.Value(736),
					},
					{
						Type:   events.PortfolioEventTypeBuy,
						Amount: 10,
						Price:  currency.Value(14655),
						Fees:   currency.Value(856),
					},
				},
			},
			want: func(t *testing.T, c *calculation) bool {
				return true &&
					assert.Equals(t, 25, c.Amount) &&
					assert.Equals(t, 491425, int(c.NetValue().Amount)) &&
					assert.Equals(t, 494614, int(c.GrossValue().Amount)) &&
					assert.Equals(t, 19657, int(c.NetPrice().Amount)) &&
					assert.Equals(t, 19785, int(c.GrossPrice().Amount)) &&
					assert.Equals(t, 44099, int(c.Cash.Amount))
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
