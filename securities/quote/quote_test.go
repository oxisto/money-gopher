package quote

import (
	"context"
	"database/sql"
	"testing"

	"github.com/oxisto/assert"
	"github.com/oxisto/money-gopher/currency"
	"github.com/oxisto/money-gopher/internal"
	"github.com/oxisto/money-gopher/internal/testing/quotetest"
	"github.com/oxisto/money-gopher/persistence"
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
						QuoteProvider: sql.NullString{String: quotetest.QuoteProviderStatic, Valid: true},
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
				qp: quotetest.NewStaticQuoteProvider(currency.Value(100)),
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
