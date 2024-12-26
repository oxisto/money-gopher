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

	"github.com/oxisto/assert"
	"github.com/oxisto/money-gopher/internal"
	"github.com/oxisto/money-gopher/internal/testing/clitest"
	"github.com/oxisto/money-gopher/internal/testing/servertest"
	"github.com/urfave/cli/v3"
)

func TestListPortfolio(t *testing.T) {
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
				cmd: &cli.Command{},
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
				cmd: &cli.Command{},
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

func TestCreateTransaction(t *testing.T) {
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
					PortfolioCmd.Commands[3].Commands[0].Flags,
					"--portfolio-id", "myportfolio",
					"--security-id", "mysecurity",
					"--type", "buy",
					"--amount", "10",
					"--price", "10",
					"--fees", "0",
					"--taxes", "0",
					"--time", "2023-01-01",
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
	srv := servertest.NewServer(internal.NewTestDB(t))
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
				return assert.Equals(t, "mybank-myportfolio:My Portfolio\n", r.String())
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
