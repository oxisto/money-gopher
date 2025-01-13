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
	"time"

	"github.com/oxisto/assert"
	"github.com/oxisto/money-gopher/internal"
	"github.com/oxisto/money-gopher/internal/testdata"
	"github.com/oxisto/money-gopher/internal/testing/clitest"
	"github.com/oxisto/money-gopher/internal/testing/servertest"
	"github.com/oxisto/money-gopher/persistence"
	"github.com/urfave/cli/v3"
)

func TestCreateAccount(t *testing.T) {
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
					AccountCmd.Command("create").Flags,
					"--id", "myaccount",
					"--display-name", "My Account",
					"--type", "BANK",
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateAccount(tt.args.ctx, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("CreateAccount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListAccounts(t *testing.T) {
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
		wantRec assert.Want[*clitest.CommandRecorder]
	}{
		{
			name: "happy path",
			args: args{
				ctx: clitest.NewSessionContext(t, srv),
				cmd: clitest.MockCommand(t, AccountCmd.Command("list").Flags),
			},
			wantRec: func(t *testing.T, rec *clitest.CommandRecorder) bool {
				return assert.Equals(t, `{
  "accounts": [
    {
      "id": "myaccount",
      "displayName": "My Account",
      "type": "BANK"
    }
  ]
}
`, rec.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := clitest.Record(tt.args.cmd)

			if err := ListAccounts(tt.args.ctx, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("ListAccounts() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.wantRec(t, rec)
		})
	}
}

func TestDeleteAccount(t *testing.T) {
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
		wantRec assert.Want[*clitest.CommandRecorder]
	}{
		{
			name: "happy path",
			args: args{
				ctx: clitest.NewSessionContext(t, srv),
				cmd: clitest.MockCommand(t, AccountCmd.Command("delete").Flags, "--id", "myaccount", "--confirm"),
			},
			wantRec: func(t *testing.T, rec *clitest.CommandRecorder) bool {
				return assert.Equals(t, "Account \"myaccount\" deleted.\n", rec.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := clitest.Record(tt.args.cmd)

			if err := DeleteAccount(tt.args.ctx, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("DeleteAccount() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.wantRec(t, rec)
		})
	}
}

func TestListTransactions(t *testing.T) {
	srv := servertest.NewServer(internal.NewTestDB(t, func(db *persistence.DB) {
		_, err := db.Queries.CreateAccount(context.Background(), testdata.TestCreateBankAccountParams)
		assert.NoError(t, err)

		_, err = db.Queries.CreateAccount(context.Background(), testdata.TestCreateBrokerageAccountParams)
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
				cmd: clitest.MockCommand(t, AccountCmd.Command("transactions").Command("list").Flags, "--account-id", testdata.TestBrokerageAccount.ID),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ListTransactions(tt.args.ctx, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("ListTransactions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateTransaction(t *testing.T) {
	srv := servertest.NewServer(internal.NewTestDB(t, func(db *persistence.DB) {
		_, err := db.Queries.CreateAccount(context.Background(), testdata.TestCreateBankAccountParams)
		assert.NoError(t, err)

		_, err = db.Queries.CreateAccount(context.Background(), testdata.TestCreateBrokerageAccountParams)
		assert.NoError(t, err)

		_, err = db.Queries.CreateSecurity(context.Background(), testdata.TestCreateSecurityParams)
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
					AccountCmd.Command("transactions").Command("create").Flags,
					"--source-account-id", testdata.TestBankAccount.ID,
					"--destination-account-id", testdata.TestBrokerageAccount.ID,
					"--security-id", testdata.TestSecurity.ID,
					"--type", "BUY",
					"--amount", "10",
					"--price", "10",
					"--fees", "0",
					"--taxes", "0",
					"--time", time.Now().Format(time.RFC3339),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateTransaction(tt.args.ctx, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("CreateTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}