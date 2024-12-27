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

// package commands contains commands that can be executed by the CLI.
package commands

import (
	"context"
	"testing"

	moneygopher "github.com/oxisto/money-gopher"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/internal"
	"github.com/oxisto/money-gopher/internal/testing/clitest"
	"github.com/oxisto/money-gopher/internal/testing/servertest"
	"github.com/oxisto/money-gopher/persistence"

	"github.com/oxisto/assert"
	"github.com/urfave/cli/v3"
)

func TestUpdateQuote(t *testing.T) {
	srv := servertest.NewServer(internal.NewTestDB(t, func(db *persistence.DB) {
		ops := persistence.Ops[*portfoliov1.Security](db)
		ops.Replace(&portfoliov1.Security{
			Id:            "mysecurity",
			QuoteProvider: moneygopher.Ref("mock"),
		})
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
					SecuritiesCmd.Command("update-quote").Flags,
					"--security-ids", "mysecurity",
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UpdateQuote(tt.args.ctx, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("UpdateQuote() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateAllQuotes(t *testing.T) {
	srv := servertest.NewServer(internal.NewTestDB(t, func(db *persistence.DB) {
		ops := persistence.Ops[*portfoliov1.Security](db)
		ops.Replace(&portfoliov1.Security{
			Id:            "mysecurity",
			QuoteProvider: moneygopher.Ref("mock"),
		})
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UpdateAllQuotes(tt.args.ctx, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("UpdateAllQuotes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListSecurities(t *testing.T) {
	srv := servertest.NewServer(internal.NewTestDB(t, func(db *persistence.DB) {
		_, err := db.CreateSecurity(context.Background(), persistence.CreateSecurityParams{
			ID:          "1234",
			DisplayName: "One Two Three Four",
		})
		assert.NoError(t, err)

		_, err = db.UpsertListedSecurity(context.Background(), persistence.UpsertListedSecurityParams{
			SecurityID: "1234",
			Ticker:     "ONE",
			Currency:   "USD",
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
		wantErr bool
		wantRec assert.Want[*clitest.CommandRecorder]
	}{
		{
			name: "happy path",
			args: args{
				ctx: clitest.NewSessionContext(t, srv),
				cmd: clitest.MockCommand(t, SecuritiesCmd.Command("list").Flags),
			},
			wantRec: func(t *testing.T, rec *clitest.CommandRecorder) bool {
				return assert.Equals(t, `{
  "securities": [
    {
      "id": "1234",
      "displayName": "One Two Three Four"
    },
    {
      "id": "US0378331005",
      "displayName": "Apple Inc."
    }
  ]
}
`, rec.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := clitest.NewCommandRecorder()
			tt.args.cmd.Writer = rec
			if err := ListSecurities(tt.args.ctx, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("ListSecurities() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantRec != nil {
				tt.wantRec(t, rec)
			}
		})
	}
}

func TestShowSecurity(t *testing.T) {
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
				cmd: clitest.MockCommand(t, SecuritiesCmd.Command("show").Flags, "--security-id", "US0378331005"),
			},
			wantRec: func(t *testing.T, rec *clitest.CommandRecorder) bool {
				return assert.Equals(t, `{
  "security": {
    "id": "US0378331005",
    "displayName": "Apple Inc.",
    "listedAs": [
      {
        "ticker": "AAPL"
      },
      {
        "ticker": "APC.F"
      }
    ]
  }
}
`, rec.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := clitest.Record(tt.args.cmd)
			if err := ShowSecurity(tt.args.ctx, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("ShowSecurity() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantRec != nil {
				tt.wantRec(t, rec)
			}
		})
	}
}

func TestPredictSecurities(t *testing.T) {
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
			wantRec: func(t *testing.T, rec *clitest.CommandRecorder) bool {
				return assert.Equals(t, "US0378331005:Apple Inc.\n", rec.String())
			},
		},
	}

	for _, tt := range tests {
		rec := clitest.Record(tt.args.cmd)
		t.Run(tt.name, func(t *testing.T) {
			PredictSecurities(tt.args.ctx, tt.args.cmd)
		})

		if tt.wantRec != nil {
			tt.wantRec(t, rec)
		}
	}
}
