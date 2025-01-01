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

package quote

import (
	"context"
	"database/sql"
	"testing"

	"github.com/oxisto/money-gopher/internal"
	"github.com/oxisto/money-gopher/persistence"

	"github.com/oxisto/assert"
)

func Test_qu_updateQuote(t *testing.T) {
	type fields struct {
		db *persistence.DB
	}
	type args struct {
		qp QuoteProvider
		ls *persistence.ListedSecurity
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.Want[*persistence.ListedSecurity]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				db: internal.NewTestDB(t, func(db *persistence.DB) {
					_, err := db.CreateSecurity(context.Background(), persistence.CreateSecurityParams{
						ID:            "My Security",
						QuoteProvider: sql.NullString{String: QuoteProviderMock, Valid: true},
					})
					assert.NoError(t, err)
					_, err = db.UpsertListedSecurity(context.Background(), persistence.UpsertListedSecurityParams{
						SecurityID: "My Security",
						Ticker:     "SEC",
						Currency:   "EUR",
					})
					assert.NoError(t, err)
				}),
			},
			args: args{
				qp: &mockQuoteProvider{},
				ls: &persistence.ListedSecurity{SecurityID: "My Security", Ticker: "SEC", Currency: "EUR"},
			},
			want: func(t *testing.T, ls *persistence.ListedSecurity) bool {
				return assert.Equals(t, 100, ls.LatestQuote.Amount)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &qu{
				db: tt.fields.db,
			}
			if err := svc.updateQuote(tt.args.qp, tt.args.ls); (err != nil) != tt.wantErr {
				t.Errorf("updateQuote() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.want(t, tt.args.ls)
		})
	}
}
