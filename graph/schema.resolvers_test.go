package graph

import (
	"context"
	"reflect"
	"testing"

	"github.com/oxisto/assert"
	"github.com/oxisto/money-gopher/internal/testdata"
	"github.com/oxisto/money-gopher/internal/testing/persistencetest"
	"github.com/oxisto/money-gopher/models"
	"github.com/oxisto/money-gopher/persistence"
)

func Test_mutationResolver_CreateSecurity(t *testing.T) {
	type fields struct {
		Resolver *Resolver
	}
	type args struct {
		ctx   context.Context
		input models.SecurityInput
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *persistence.Security
		wantDB  assert.Want[*persistence.DB]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				Resolver: &Resolver{
					DB: persistencetest.NewTestDB(t),
				},
			},
			args: args{
				ctx: context.TODO(),
				input: models.SecurityInput{
					ID:          "DE1234567890",
					DisplayName: "My Security",
					ListedAs: []*models.ListedSecurityInput{
						{
							Ticker:   "TICK",
							Currency: "USD",
						},
					},
				},
			},
			want: &persistence.Security{
				ID:          "DE1234567890",
				DisplayName: "My Security",
			},
			wantDB: func(t *testing.T, db *persistence.DB) bool {
				_, err := db.Queries.GetSecurity(context.Background(), "DE1234567890")
				assert.NoError(t, err)

				ls, err := db.Queries.ListListedSecuritiesBySecurityID(context.Background(), "DE1234567890")
				return len(ls) == 1 && assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &mutationResolver{
				Resolver: tt.fields.Resolver,
			}
			got, err := r.CreateSecurity(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("mutationResolver.CreateSecurity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equals(t, tt.want, got)
			tt.wantDB(t, tt.fields.Resolver.DB)
		})
	}
}

func Test_queryResolver_Accounts(t *testing.T) {
	type fields struct {
		Resolver *Resolver
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*persistence.Account
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				Resolver: &Resolver{
					DB: persistencetest.NewTestDB(t, func(db *persistence.DB) {
						_, err := db.Queries.CreateAccount(context.Background(), testdata.TestCreateBankAccountParams)
						assert.NoError(t, err)
					}),
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			want: []*persistence.Account{
				testdata.TestBankAccount,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &queryResolver{
				Resolver: tt.fields.Resolver,
			}
			got, err := r.Accounts(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("queryResolver.Accounts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("queryResolver.Accounts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_queryResolver_Transactions(t *testing.T) {
	type fields struct {
		Resolver *Resolver
	}
	type args struct {
		ctx       context.Context
		accountID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*persistence.Transaction
		wantErr bool
	}{
		{
			name: "list brokerage transactions",
			fields: fields{
				Resolver: &Resolver{
					DB: persistencetest.NewTestDB(t, func(db *persistence.DB) {
						_, err := db.Queries.CreateAccount(context.Background(), testdata.TestCreateBankAccountParams)
						assert.NoError(t, err)

						_, err = db.Queries.CreateAccount(context.Background(), testdata.TestCreateBrokerageAccountParams)
						assert.NoError(t, err)

						_, err = db.Queries.CreateTransaction(context.Background(), testdata.TestCreateBuyTransactionParams)
						assert.NoError(t, err)
					}),
				},
			},
			args: args{
				ctx:       context.TODO(),
				accountID: testdata.TestBrokerageAccount.ID,
			},
			want: []*persistence.Transaction{
				testdata.TestBuyTransaction,
			},
		},
		{
			name: "list bank transactions",
			fields: fields{
				Resolver: &Resolver{
					DB: persistencetest.NewTestDB(t, func(db *persistence.DB) {
						_, err := db.Queries.CreateAccount(context.Background(), testdata.TestCreateBankAccountParams)
						assert.NoError(t, err)

						_, err = db.Queries.CreateAccount(context.Background(), testdata.TestCreateBrokerageAccountParams)
						assert.NoError(t, err)

						_, err = db.Queries.CreateTransaction(context.Background(), testdata.TestCreateBuyTransactionParams)
						assert.NoError(t, err)

						_, err = db.Queries.CreateTransaction(context.Background(), testdata.TestCreateDepositTransactionParams)
						assert.NoError(t, err)
					}),
				},
			},
			args: args{
				ctx:       context.TODO(),
				accountID: testdata.TestBankAccount.ID,
			},
			want: []*persistence.Transaction{
				testdata.TestBuyTransaction,
				testdata.TestDepositTransaction,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &queryResolver{
				Resolver: tt.fields.Resolver,
			}
			got, err := r.Transactions(tt.args.ctx, tt.args.accountID)
			if (err != nil) != tt.wantErr {
				t.Errorf("queryResolver.Transactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equals(t, tt.want, got)
		})
	}
}

func Test_mutationResolver_CreatePortfolio(t *testing.T) {
	type fields struct {
		Resolver *Resolver
	}
	type args struct {
		ctx   context.Context
		input models.PortfolioInput
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *persistence.Portfolio
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				Resolver: &Resolver{
					DB: persistencetest.NewTestDB(t, func(db *persistence.DB) {
						_, err := db.Queries.CreateAccount(context.Background(), testdata.TestCreateBankAccountParams)
						assert.NoError(t, err)

						_, err = db.Queries.CreateAccount(context.Background(), testdata.TestCreateBrokerageAccountParams)
						assert.NoError(t, err)

						_, err = db.Queries.CreateTransaction(context.Background(), testdata.TestCreateBuyTransactionParams)
						assert.NoError(t, err)
					}),
				},
			},
			args: args{
				ctx: context.TODO(),
				input: models.PortfolioInput{
					ID:          "mybank/myportfolio",
					DisplayName: "My Portfolio",
					AccountIds: []string{
						testdata.TestCreateBankAccountParams.ID,
						testdata.TestCreateBrokerageAccountParams.ID,
					},
				},
			},
			want: &persistence.Portfolio{
				ID:          "mybank/myportfolio",
				DisplayName: "My Portfolio",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &mutationResolver{
				Resolver: tt.fields.Resolver,
			}
			got, err := r.CreatePortfolio(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("mutationResolver.CreatePortfolio() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mutationResolver.CreatePortfolio() = %v, want %v", got, tt.want)
			}
		})
	}
}
