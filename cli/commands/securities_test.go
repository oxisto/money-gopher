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

	"github.com/oxisto/money-gopher/currency"
	"github.com/oxisto/money-gopher/internal/testdata"
	"github.com/oxisto/money-gopher/internal/testing/clitest"
	"github.com/oxisto/money-gopher/internal/testing/persistencetest"
	"github.com/oxisto/money-gopher/internal/testing/quotetest"
	"github.com/oxisto/money-gopher/internal/testing/servertest"
	"github.com/oxisto/money-gopher/persistence"
	"github.com/oxisto/money-gopher/securities/quote"

	"github.com/oxisto/assert"
	"github.com/urfave/cli/v3"
)

func TestUpdateQuote(t *testing.T) {
	srv := servertest.NewServer(persistencetest.NewTestDB(t, func(db *persistence.DB) {
		_, err := db.CreateSecurity(context.Background(), testdata.TestCreateSecurityParams)
		assert.NoError(t, err)

		_, err = db.UpsertListedSecurity(context.Background(), testdata.TestUpsertListedSecurityParams)
		assert.NoError(t, err)

		quote.RegisterQuoteProvider(quotetest.QuoteProviderStatic, quotetest.NewStaticQuoteProvider(currency.Value(100)))
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
				cmd: clitest.MockCommand(t,
					SecuritiesCmd.Command("update-quote").Flags,
					"--security-ids", testdata.TestCreateSecurityParams.ID,
				),
			},
			wantRec: func(t *testing.T, rec *clitest.CommandRecorder) bool {
				return assert.Equals(t, `{
  "updated": [
    {
      "latestQuote": {
        "value": 100,
        "symbol": "EUR"
      }
    }
  ]
}
`, rec.String(),
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := clitest.Record(tt.args.cmd)
			if err := UpdateQuote(tt.args.ctx, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("UpdateQuote() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantRec != nil {
				tt.wantRec(t, rec)
			}
		})
	}
}

func TestUpdateAllQuotes(t *testing.T) {
	srv := servertest.NewServer(persistencetest.NewTestDB(t, func(db *persistence.DB) {
		_, err := db.CreateSecurity(context.Background(), testdata.TestCreateSecurityParams)
		assert.NoError(t, err)

		_, err = db.UpsertListedSecurity(context.Background(), testdata.TestUpsertListedSecurityParams)
		assert.NoError(t, err)

		quote.RegisterQuoteProvider(quotetest.QuoteProviderStatic, quotetest.NewStaticQuoteProvider(currency.Value(100)))
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
	srv := servertest.NewServer(persistencetest.NewTestDB(t, func(db *persistence.DB) {
		_, err := db.CreateSecurity(context.Background(), testdata.TestCreateSecurityParams)
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
      "id": "DE1234567890",
      "displayName": "My Security"
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
	srv := servertest.NewServer(persistencetest.NewTestDB(t, func(db *persistence.DB) {
		_, err := db.CreateSecurity(context.Background(), testdata.TestCreateSecurityParams)
		assert.NoError(t, err)

		_, err = db.UpsertListedSecurity(context.Background(), testdata.TestUpsertListedSecurityParams)
		assert.NoError(t, err)

		quote.RegisterQuoteProvider(quotetest.QuoteProviderStatic, quotetest.NewStaticQuoteProvider(currency.Value(100)))
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
				cmd: clitest.MockCommand(t, SecuritiesCmd.Command("show").Flags, "--security-id", "DE1234567890"),
			},
			wantRec: func(t *testing.T, rec *clitest.CommandRecorder) bool {
				return assert.Equals(t, `{
  "security": {
    "id": "DE1234567890",
    "displayName": "My Security",
    "listedAs": [
      {
        "ticker": "TICK"
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
	srv := servertest.NewServer(persistencetest.NewTestDB(t, func(db *persistence.DB) {
		_, err := db.CreateSecurity(context.Background(), testdata.TestCreateSecurityParams)
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
			wantRec: func(t *testing.T, rec *clitest.CommandRecorder) bool {
				return assert.Equals(t, "DE1234567890:My Security\n", rec.String())
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
