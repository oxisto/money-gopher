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

package commands

import (
	"context"
	"testing"

	"github.com/oxisto/money-gopher/internal"
	"github.com/oxisto/money-gopher/internal/testdata"
	"github.com/oxisto/money-gopher/internal/testing/clitest"
	"github.com/oxisto/money-gopher/internal/testing/servertest"
	"github.com/oxisto/money-gopher/persistence"

	"github.com/oxisto/assert"
	"github.com/urfave/cli/v3"
)

func TestListPortfolio(t *testing.T) {
	srv := servertest.NewServer(internal.NewTestDB(t, func(db *persistence.DB) {
		_, err := db.Queries.CreateAccount(context.Background(), testdata.TestCreateBankAccountParams)
		assert.NoError(t, err)

		_, err = db.Queries.CreateAccount(context.Background(), testdata.TestCreateBrokerageAccountParams)
		assert.NoError(t, err)

		_, err = db.Queries.CreatePortfolio(context.Background(), testdata.TestCreatePortfolioParams)
		assert.NoError(t, err)

		err = db.Queries.AddAccountToPortfolio(context.Background(), testdata.TestAddAccountToPortfolioParams)
		assert.NoError(t, err)

		_, err = db.Queries.CreateTransaction(context.Background(), testdata.TestCreateDepositTransactionParams)
		assert.NoError(t, err)

		_, err = db.Queries.CreateTransaction(context.Background(), testdata.TestCreateBuyTransactionParams)
		assert.NoError(t, err)
	}))
	defer srv.Close()

	type args struct {
		ctx context.Context
		cmd *cli.Command
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				ctx: clitest.NewSessionContext(t, srv),
				cmd: clitest.MockCommand(t,
					PortfolioCmd.Command("list").Flags,
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ListPortfolio(tt.args.ctx, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("ListPortfolio() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreatePortfolio(t *testing.T) {
	srv := servertest.NewServer(internal.NewTestDB(t, func(db *persistence.DB) {
		_, err := db.Queries.CreateAccount(context.Background(), testdata.TestCreateBankAccountParams)
		assert.NoError(t, err)
	}))
	defer srv.Close()

	type args struct {
		ctx context.Context
		cmd *cli.Command
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				ctx: clitest.NewSessionContext(t, srv),
				cmd: clitest.MockCommand(t,
					PortfolioCmd.Command("create").Flags,
					"--id", "mynewportfolio",
					"--display-name", "My New Portfolio",
					"--account-ids", testdata.TestCreateBankAccountParams.ID,
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreatePortfolio(tt.args.ctx, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("CreatePortfolio() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestShowPortfolio(t *testing.T) {
	srv := servertest.NewServer(internal.NewTestDB(t))
	defer srv.Close()

	type args struct {
		ctx context.Context
		cmd *cli.Command
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantRec assert.Want[*clitest.CommandRecorder]
	}{
		{
			name: "happy path",
			args: args{
				ctx: clitest.NewSessionContext(t, srv),
				cmd: clitest.MockCommand(t,
					PortfolioCmd.Commands[2].Flags,
					"--portfolio-id", "myportfolio",
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := clitest.NewCommandRecorder()
			tt.args.cmd.Writer = rec
			if err := ShowPortfolio(tt.args.ctx, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("ShowPortfolio() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantRec != nil {
				tt.wantRec(t, rec)
			}
		})
	}
}

func TestImportTransactions(t *testing.T) {
	srv := servertest.NewServer(internal.NewTestDB(t))
	defer srv.Close()

	type args struct {
		ctx context.Context
		cmd *cli.Command
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				ctx: clitest.NewSessionContext(t, srv),
				cmd: clitest.MockCommand(t,
					PortfolioCmd.Command("transactions").Command("import").Flags,
					"--portfolio-id", "myportfolio",
					"--csv-file", "../../internal/testdata/transactions.csv",
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ImportTransactions(tt.args.ctx, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("ImportTransactions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPredictPortfolios(t *testing.T) {
	srv := servertest.NewServer(internal.NewTestDB(t, func(db *persistence.DB) {
		_, err := db.Queries.CreateAccount(context.Background(), testdata.TestCreateBankAccountParams)
		assert.NoError(t, err)

		_, err = db.Queries.CreatePortfolio(context.Background(), persistence.CreatePortfolioParams{
			ID:          "mybank/myportfolio",
			DisplayName: "My Portfolio",
		})
		assert.NoError(t, err)
	}))
	defer srv.Close()

	type args struct {
		ctx context.Context
		cmd *cli.Command
	}
	tests := []struct {
		name    string
		args    args
		wantRec assert.Want[*clitest.CommandRecorder]
	}{
		{
			name: "happy path",
			args: args{
				ctx: clitest.NewSessionContext(t, srv),
				cmd: &cli.Command{},
			},
			wantRec: func(t *testing.T, r *clitest.CommandRecorder) bool {
				return assert.Equals(t, "mybank/myportfolio:My Portfolio\n", r.String())
			},
		},
	}
	for _, tt := range tests {
		rec := clitest.Record(tt.args.cmd)
		t.Run(tt.name, func(t *testing.T) {
			PredictPortfolios(tt.args.ctx, tt.args.cmd)
		})

		if tt.wantRec != nil {
			tt.wantRec(t, rec)
		}
	}
}
