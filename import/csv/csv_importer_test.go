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
	"github.com/oxisto/money-gopher/currency"
	"github.com/oxisto/money-gopher/internal/testdata"
	"github.com/oxisto/money-gopher/persistence"
	"github.com/oxisto/money-gopher/portfolio/events"
	"github.com/oxisto/money-gopher/securities/quote"

	"github.com/oxisto/assert"
)

func TestImport(t *testing.T) {
	type args struct {
		r                  io.Reader
		bankAccountID      string
		brokerageAccountID string
	}
	tests := []struct {
		name     string
		args     args
		wantTxs  assert.Want[[]*persistence.Transaction]
		wantSecs assert.Want[[]*persistence.Security]
		wantLss  assert.Want[[]*persistence.ListedSecurity]
	}{
		{
			name: "happy path",
			args: args{
				r: bytes.NewReader([]byte(`Date;Type;Value;Transaction Currency;Gross Amount;Currency Gross Amount;Exchange Rate;Fees;Taxes;Shares;ISIN;WKN;Ticker Symbol;Security Name;Note
2021-06-05T00:00;Buy;2.151,85;EUR;;;;10,25;0,00;20;US0378331005;865985;APC.F;Apple Inc.;
2021-06-05T00:00;Sell;-2.151,85;EUR;;;;10,25;0,00;20;US0378331005;865985;APC.F;Apple Inc.;
2021-06-18T00:00;Delivery (Inbound);912,66;EUR;;;;7,16;0,00;5;US09075V1026;A2PSR2;22UA.F;BioNTech SE;`)),
				bankAccountID:      testdata.TestBankAccount.ID,
				brokerageAccountID: testdata.TestBrokerageAccount.ID,
			},
			wantTxs: func(t *testing.T, txs []*persistence.Transaction) bool {
				return assert.Equals(t, 3, len(txs))
			},
			wantSecs: func(t *testing.T, secs []*persistence.Security) bool {
				return assert.Equals(t, 2, len(secs))
			},
			wantLss: func(t *testing.T, lss []*persistence.ListedSecurity) bool {
				return assert.Equals(t, 2, len(lss))
			},
		},
		{
			name: "error",
			args: args{
				r: bytes.NewReader([]byte(`Date;Type;Value;Transaction Currency;Gross Amount;Currency Gross Amount;Exchange Rate;Fees;Taxes;Shares;ISIN;WKN;Ticker Symbol;Security Name;Note
this;will;be;an;error`)),
			},
			wantTxs: func(t *testing.T, txs []*persistence.Transaction) bool {
				return assert.Equals(t, 0, len(txs))
			},
			wantSecs: func(t *testing.T, secs []*persistence.Security) bool {
				return assert.Equals(t, 0, len(secs))
			},
			wantLss: func(t *testing.T, lss []*persistence.ListedSecurity) bool {
				return assert.Equals(t, 0, len(lss))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTxs, gotSecs, gotLss := Import(tt.args.r, tt.args.bankAccountID, tt.args.brokerageAccountID)
			tt.wantTxs(t, gotTxs)
			tt.wantSecs(t, gotSecs)
			tt.wantLss(t, gotLss)
		})
	}
}

func Test_readLine(t *testing.T) {
	type args struct {
		cr                 *csv.Reader
		bankAccountID      string
		brokerageAccountID string
	}
	tests := []struct {
		name    string
		args    args
		wantTx  *persistence.Transaction
		wantSec *persistence.Security
		wantLs  []*persistence.ListedSecurity
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
				bankAccountID:      testdata.TestBankAccount.ID,
				brokerageAccountID: testdata.TestBrokerageAccount.ID,
			},
			wantTx: &persistence.Transaction{
				ID:                   "670240c1f8373a3f",
				SecurityID:           moneygopher.Ref("US0378331005"),
				Type:                 events.PortfolioEventTypeBuy,
				SourceAccountID:      &testdata.TestBankAccount.ID,
				DestinationAccountID: &testdata.TestBrokerageAccount.ID,
				Time:                 time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
				Amount:               20,
				Fees:                 currency.Value(1025),
				Taxes:                currency.Zero(),
				Price:                currency.Value(10708),
			},
			wantSec: &persistence.Security{
				ID:            "US0378331005",
				DisplayName:   "Apple Inc.",
				QuoteProvider: moneygopher.Ref(quote.QuoteProviderYF),
			},
			wantLs: []*persistence.ListedSecurity{
				{
					SecurityID: "US0378331005",
					Ticker:     "APC.F",
					Currency:   "EUR",
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
				bankAccountID:      testdata.TestBankAccount.ID,
				brokerageAccountID: testdata.TestBrokerageAccount.ID,
			},
			wantTx: &persistence.Transaction{
				ID:                   "dbddb1c7b1ce1375",
				SecurityID:           moneygopher.Ref("US00827B1061"),
				Type:                 events.PortfolioEventTypeBuy,
				SourceAccountID:      &testdata.TestBankAccount.ID,
				DestinationAccountID: &testdata.TestBrokerageAccount.ID,
				Time:                 time.Date(2022, 1, 1, 9, 0, 0, 0, time.Local),
				Amount:               20,
				Price:                currency.Value(6040),
				Fees:                 currency.Zero(),
				Taxes:                currency.Zero(),
			},
			wantSec: &persistence.Security{
				ID:            "US00827B1061",
				DisplayName:   "Affirm Holdings Inc.",
				QuoteProvider: moneygopher.Ref(quote.QuoteProviderYF),
			},
			wantLs: []*persistence.ListedSecurity{
				{
					SecurityID: "US00827B1061",
					Ticker:     "AFRM",
					Currency:   "USD",
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
				bankAccountID:      testdata.TestBankAccount.ID,
				brokerageAccountID: testdata.TestBrokerageAccount.ID,
			},
			wantTx: &persistence.Transaction{
				ID:                   "4201924709e1f078",
				SecurityID:           moneygopher.Ref("DE0005557508"),
				SourceAccountID:      &testdata.TestBrokerageAccount.ID,
				DestinationAccountID: &testdata.TestBankAccount.ID,
				Type:                 events.PortfolioEventTypeSell,
				Time:                 time.Date(2022, 1, 1, 8, 0, 6, 0, time.Local),
				Amount:               103,
				Fees:                 currency.Zero(),
				Taxes:                currency.Value(1830),
				Price:                currency.Value(1552),
			},
			wantSec: &persistence.Security{
				ID:            "DE0005557508",
				DisplayName:   "Deutsche Telekom AG",
				QuoteProvider: moneygopher.Ref(quote.QuoteProviderYF),
			},
			wantLs: []*persistence.ListedSecurity{
				{
					SecurityID: "DE0005557508",
					Ticker:     "DTE.F",
					Currency:   "EUR",
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
			gotTx, gotSec, gotLs, err := readLine(tt.args.cr, tt.args.bankAccountID, tt.args.brokerageAccountID)
			if err != nil {
				tt.wantErr(t, err)
				return
			}
			assert.Equals(t, tt.wantTx, gotTx)
			assert.Equals(t, tt.wantSec, gotSec)
			assert.Equals(t, tt.wantLs, gotLs)
		})
	}
}
