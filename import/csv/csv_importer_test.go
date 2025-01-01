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

package csv

import (
	"bytes"
	"encoding/csv"
	"io"
	"testing"
	"time"

	moneygopher "github.com/oxisto/money-gopher"
	"github.com/oxisto/money-gopher/securities/quote"

	"github.com/oxisto/assert"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestImport(t *testing.T) {
	type args struct {
		r     io.Reader
		pname string
	}
	tests := []struct {
		name     string
		args     args
		wantTxs  assert.Want[[]*portfoliov1.PortfolioEvent]
		wantSecs assert.Want[[]*portfoliov1.Security]
	}{
		{
			name: "happy path",
			args: args{
				r: bytes.NewReader([]byte(`Date;Type;Value;Transaction Currency;Gross Amount;Currency Gross Amount;Exchange Rate;Fees;Taxes;Shares;ISIN;WKN;Ticker Symbol;Security Name;Note
2021-06-05T00:00;Buy;2.151,85;EUR;;;;10,25;0,00;20;US0378331005;865985;APC.F;Apple Inc.;
2021-06-05T00:00;Sell;-2.151,85;EUR;;;;10,25;0,00;20;US0378331005;865985;APC.F;Apple Inc.;
2021-06-18T00:00;Delivery (Inbound);912,66;EUR;;;;7,16;0,00;5;US09075V1026;A2PSR2;22UA.F;BioNTech SE;`)),
			},
			wantTxs: func(t *testing.T, txs []*portfoliov1.PortfolioEvent) bool {
				return assert.Equals(t, 3, len(txs))
			},
			wantSecs: func(t *testing.T, secs []*portfoliov1.Security) bool {
				return assert.Equals(t, 2, len(secs))
			},
		},
		{
			name: "error",
			args: args{
				r: bytes.NewReader([]byte(`Date;Type;Value;Transaction Currency;Gross Amount;Currency Gross Amount;Exchange Rate;Fees;Taxes;Shares;ISIN;WKN;Ticker Symbol;Security Name;Note
this;will;be;an;error`)),
			},
			wantTxs: func(t *testing.T, txs []*portfoliov1.PortfolioEvent) bool {
				return assert.Equals(t, 0, len(txs))
			},
			wantSecs: func(t *testing.T, secs []*portfoliov1.Security) bool {
				return assert.Equals(t, 0, len(secs))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTxs, gotSecs := Import(tt.args.r, tt.args.pname)
			tt.wantTxs(t, gotTxs)
			tt.wantSecs(t, gotSecs)
		})
	}
}

func Test_readLine(t *testing.T) {
	type args struct {
		cr    *csv.Reader
		pname string
	}
	tests := []struct {
		name    string
		args    args
		wantTx  *portfoliov1.PortfolioEvent
		wantSec *portfoliov1.Security
		wantErr assert.Want[error]
	}{
		{
			name: "buy",
			args: args{
				cr: func() *csv.Reader {
					cr := csv.NewReader(bytes.NewReader([]byte(`2020-01-01T00:00;Buy;2.151,85;EUR;;;;10,25;0,00;20;US0378331005;865985;APC.F;Apple Inc.;`)))
					cr.Comma = ';'
					return cr
				}(),
			},
			wantTx: &portfoliov1.PortfolioEvent{
				Id:         "9e7b470b7566beca",
				SecurityId: "US0378331005",
				Type:       portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY,
				Time:       timestamppb.New(time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)),
				Amount:     20,
				Fees:       portfoliov1.Value(1025),
				Taxes:      portfoliov1.Zero(),
				Price:      portfoliov1.Value(10708),
			},
			wantSec: &portfoliov1.Security{
				Id:            "US0378331005",
				DisplayName:   "Apple Inc.",
				QuoteProvider: moneygopher.Ref(quote.QuoteProviderYF),
				ListedOn: []*portfoliov1.ListedSecurity{
					{
						SecurityId: "US0378331005",
						Ticker:     "APC.F",
						Currency:   "EUR",
					},
				},
			},
		},
		{
			name: "buy cross-currency",
			args: args{
				cr: func() *csv.Reader {
					cr := csv.NewReader(bytes.NewReader([]byte(`2022-01-01T09:00;Buy;1.207,90;EUR;1.413,24;USD;0,8547;0,00;0,00;20;US00827B1061;A2QL1G;AFRM;Affirm Holdings Inc.;`)))
					cr.Comma = ';'
					return cr
				}(),
			},
			wantTx: &portfoliov1.PortfolioEvent{
				Id:         "1070dafc882785a0",
				SecurityId: "US00827B1061",
				Type:       portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY,
				Time:       timestamppb.New(time.Date(2022, 1, 1, 9, 0, 0, 0, time.Local)),
				Amount:     20,
				Price:      portfoliov1.Value(6040),
				Fees:       portfoliov1.Zero(),
				Taxes:      portfoliov1.Zero(),
			},
			wantSec: &portfoliov1.Security{
				Id:            "US00827B1061",
				DisplayName:   "Affirm Holdings Inc.",
				QuoteProvider: moneygopher.Ref(quote.QuoteProviderYF),
				ListedOn: []*portfoliov1.ListedSecurity{
					{
						SecurityId: "US00827B1061",
						Ticker:     "AFRM",
						Currency:   "USD",
					},
				},
			},
		},
		{
			name: "sell",
			args: args{
				cr: func() *csv.Reader {
					cr := csv.NewReader(bytes.NewReader([]byte(`2022-01-01T08:00:06;Sell;-1.580,26;EUR;;;;0,00;18,30;103;DE0005557508;;DTE.F;Deutsche Telekom AG;`)))
					cr.Comma = ';'
					return cr
				}(),
			},
			wantTx: &portfoliov1.PortfolioEvent{
				Id:         "8bb43fed65b35685",
				SecurityId: "DE0005557508",
				Type:       portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_SELL,
				Time:       timestamppb.New(time.Date(2022, 1, 1, 8, 0, 6, 0, time.Local)),
				Amount:     103,
				Fees:       portfoliov1.Zero(),
				Taxes:      portfoliov1.Value(1830),
				Price:      portfoliov1.Value(1552),
			},
			wantSec: &portfoliov1.Security{
				Id:            "DE0005557508",
				DisplayName:   "Deutsche Telekom AG",
				QuoteProvider: moneygopher.Ref(quote.QuoteProviderYF),
				ListedOn: []*portfoliov1.ListedSecurity{
					{
						SecurityId: "DE0005557508",
						Ticker:     "DTE.F",
						Currency:   "EUR",
					},
				},
			},
		},
		{
			name: "type-error",
			args: args{
				cr: func() *csv.Reader {
					cr := csv.NewReader(bytes.NewReader([]byte(`2022-01-01T09:00;Sel;-1.580,26;EUR;;;;0,00;taxes;103;DE0005557508;;DTE.F;Deutsche Telekom AG;`)))
					cr.Comma = ';'
					return cr
				}(),
			},
			wantErr: func(t *testing.T, err error) bool {
				return assert.ErrorIs(t, ErrParsingType, err)
			},
		},
		{
			name: "value-error",
			args: args{
				cr: func() *csv.Reader {
					cr := csv.NewReader(bytes.NewReader([]byte(`2022-01-01T09:00;Sell;value;EUR;;;;0,00;taxes;103;DE0005557508;;DTE.F;Deutsche Telekom AG;`)))
					cr.Comma = ';'
					return cr
				}(),
			},
			wantErr: func(t *testing.T, err error) bool {
				return assert.ErrorIs(t, ErrParsingValue, err)
			},
		},
		{
			name: "amount-error",
			args: args{
				cr: func() *csv.Reader {
					cr := csv.NewReader(bytes.NewReader([]byte(`2022-01-01T09:00;Sell;-1234,00;EUR;;;;0,00;0,00;amount;DE0005557508;;DTE.F;Deutsche Telekom AG;`)))
					cr.Comma = ';'
					return cr
				}(),
			},
			wantErr: func(t *testing.T, err error) bool {
				return assert.ErrorIs(t, ErrParsingAmount, err)
			},
		},
		{
			name: "time-error",
			args: args{
				cr: func() *csv.Reader {
					cr := csv.NewReader(bytes.NewReader([]byte(`not_time;Sell;-1.580,26;EUR;;;;0,00;18,30;103;DE0005557508;;DTE.F;Deutsche Telekom AG;`)))
					cr.Comma = ';'
					return cr
				}(),
			},
			wantErr: func(t *testing.T, err error) bool {
				return assert.ErrorIs(t, ErrParsingTime, err)
			},
		},
		{
			name: "fees-error",
			args: args{
				cr: func() *csv.Reader {
					cr := csv.NewReader(bytes.NewReader([]byte(`2022-01-01T09:00;Sell;-1.580,26;EUR;;;;fee;0,0;103;DE0005557508;;DTE.F;Deutsche Telekom AG;`)))
					cr.Comma = ';'
					return cr
				}(),
			},
			wantErr: func(t *testing.T, err error) bool {
				return assert.ErrorIs(t, ErrParsingFees, err)
			},
		},
		{
			name: "taxes-error",
			args: args{
				cr: func() *csv.Reader {
					cr := csv.NewReader(bytes.NewReader([]byte(`2022-01-01T09:00;Sell;-1.580,26;EUR;;;;0,00;taxes;103;DE0005557508;;DTE.F;Deutsche Telekom AG;`)))
					cr.Comma = ';'
					return cr
				}(),
			},
			wantErr: func(t *testing.T, err error) bool {
				return assert.ErrorIs(t, ErrParsingTaxes, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTx, gotSec, err := readLine(tt.args.cr, tt.args.pname)
			if err != nil {
				tt.wantErr(t, err)
				return
			}
			assert.Equals(t, tt.wantTx, gotTx, protocmp.Transform())
			assert.Equals(t, tt.wantSec, gotSec, protocmp.Transform())
		})
	}
}
