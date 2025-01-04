package persistence_test

import (
	"context"
	"testing"

	"github.com/oxisto/assert"
	"github.com/oxisto/money-gopher/internal/testing/persistencetest"
	"github.com/oxisto/money-gopher/persistence"
)

func TestQueries_ListListedSecuritiesBySecurityID(t *testing.T) {
	type fields struct {
		db persistence.DBTX
	}
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*persistence.ListedSecurity
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				db: persistencetest.NewTestDB(t, func(db *persistence.DB) {
					_, err := db.Queries.CreateSecurity(context.Background(), persistence.CreateSecurityParams{
						ID:          "DE1234567890",
						DisplayName: "My Security",
					})
					assert.NoError(t, err)

					_, err = db.Queries.UpsertListedSecurity(context.Background(), persistence.UpsertListedSecurityParams{
						SecurityID: "DE1234567890",
						Ticker:     "TICK",
						Currency:   "USD",
					})
					assert.NoError(t, err)
				}),
			},
			args: args{
				ctx: context.TODO(),
				id:  "DE1234567890",
			},
			want: []*persistence.ListedSecurity{
				{
					SecurityID: "DE1234567890",
					Ticker:     "TICK",
					Currency:   "USD",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := persistence.New(tt.fields.db)
			got, err := q.ListListedSecuritiesBySecurityID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Queries.ListListedSecuritiesBySecurityID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equals(t, tt.want, got)
		})
	}
}
