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
					{EventOneof: &portfoliov1.PortfolioEvent_Buy{
						Buy: &portfoliov1.BuySecurityTransaction{
							Amount: 5,
							Price:  181.10,
							Fees:   7.16,
						},
					}},
					{EventOneof: &portfoliov1.PortfolioEvent_Sell{
						Sell: &portfoliov1.SellSecurityTransaction{
							Amount: 2,
							Price:  304.30,
							Fees:   6.42,
							Taxes:  16.32,
						},
					}},
					{EventOneof: &portfoliov1.PortfolioEvent_Buy{
						Buy: &portfoliov1.BuySecurityTransaction{
							Amount: 5,
							Price:  290,
							Fees:   8.53,
						},
					}},
					{EventOneof: &portfoliov1.PortfolioEvent_Sell{
						Sell: &portfoliov1.SellSecurityTransaction{
							Amount: 3,
							Price:  220,
							Fees:   8.45,
						},
					}},
					{EventOneof: &portfoliov1.PortfolioEvent_Buy{
						Buy: &portfoliov1.BuySecurityTransaction{
							Amount: 5,
							Price:  203.30,
							Fees:   7.44,
						},
					}},
					{EventOneof: &portfoliov1.PortfolioEvent_Buy{
						Buy: &portfoliov1.BuySecurityTransaction{
							Amount: 5,
							Price:  196.45,
							Fees:   7.36,
						},
					}},
					{EventOneof: &portfoliov1.PortfolioEvent_Buy{
						Buy: &portfoliov1.BuySecurityTransaction{
							Amount: 10,
							Price:  146.55,
							Fees:   8.56,
						},
					}},
				},
			},
			want: func(t *testing.T, c *calculation) bool {
				return true &&
					assert.Equals(t, 25, c.Amount) &&
					assert.Equals(t, 4914.25, c.NetValue()) &&
					assert.Equals(t, 4946.14, c.GrossValue()) &&
					assert.Equals(t, 196.57, c.NetPrice()) &&
					assert.Equals(t, 197.84561, c.GrossPrice())
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
