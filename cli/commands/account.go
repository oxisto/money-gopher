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
	"errors"
	"fmt"
	"time"

	mcli "github.com/oxisto/money-gopher/cli"
	"github.com/oxisto/money-gopher/models"
	"github.com/oxisto/money-gopher/portfolio/accounts"
	"github.com/oxisto/money-gopher/portfolio/events"

	"github.com/shurcooL/graphql"
	"github.com/urfave/cli/v3"
)

// AccountCmd is the command for account related commands.
var AccountCmd = &cli.Command{
	Name:   "account",
	Usage:  "Manage accounts",
	Before: mcli.InjectSession,
	Commands: []*cli.Command{
		{
			Name:   "create",
			Usage:  "Creates a new account",
			Action: CreateAccount,
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "id", Usage: "The unique ID for the account", Required: true},
				&cli.StringFlag{Name: "display-name", Usage: "The display name of the account", Required: true},
				&cli.GenericFlag{Name: "type", Usage: "The type of bank account", Value: func() *accounts.AccountType {
					var typ accounts.AccountType = accounts.AccountTypeBrokerage
					return &typ
				}()},
			},
		},
		{
			Name:   "list",
			Usage:  "Lists all accounts",
			Action: ListAccounts,
		},
		{
			Name:   "delete",
			Usage:  "Deletes an account",
			Action: DeleteAccount,
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "id", Usage: "The unique ID for the account", Required: true},
				&cli.BoolFlag{Name: "confirm", Usage: "Confirm account deletion", Required: true},
			},
		},
		{
			Name:  "transactions",
			Usage: "Subcommands supporting transactions within one portfolio",
			Commands: []*cli.Command{
				{
					Name:   "list",
					Usage:  "Lists all transactions for an account",
					Action: ListTransactions,
					Flags: []cli.Flag{
						&cli.StringFlag{Name: "account-id", Usage: "The ID of the account the transaction is coming from or is destined to", Required: true},
					},
				},
				{
					Name:   "create",
					Usage:  "Creates a transaction. Defaults to a \"buy\" transaction",
					Action: CreateTransaction,
					Flags: []cli.Flag{
						&cli.StringFlag{Name: "source-account-id", Usage: "The ID of the account the transaction is coming from", Required: true},
						&cli.StringFlag{Name: "destination-account-id", Usage: "The ID of the account the transaction is destined to", Required: true},
						&cli.StringFlag{Name: "security-id", Usage: "The ID of the security this transaction belongs to (its ISIN)", Required: true},
						&cli.GenericFlag{Name: "type", Usage: "The type of the transaction", Required: true, DefaultText: "BUY", Value: func() *events.PortfolioEventType {
							var typ events.PortfolioEventType = events.PortfolioEventTypeBuy
							return &typ
						}()},
						&cli.FloatFlag{Name: "amount", Usage: "The amount of securities involved in the transaction", Required: true},
						&cli.FloatFlag{Name: "price", Usage: "The price without fees or taxes", Required: true},
						&cli.FloatFlag{Name: "fees", Usage: "Any fees that applied to the transaction"},
						&cli.FloatFlag{Name: "taxes", Usage: "Any taxes that applied to the transaction"},
						&cli.StringFlag{Name: "time", Usage: "The time of the transaction. Defaults to 'now'", DefaultText: "now"},
					},
				},
			},
		},
	},
}

// CreateAccount creates a new bank account.
func CreateAccount(ctx context.Context, cmd *cli.Command) (err error) {
	s := mcli.FromContext(ctx)

	var query struct {
		CreateAccount struct {
			ID          string               `json:"id"`
			DisplayName string               `json:"displayName"`
			Type        accounts.AccountType `json:"type"`
		} `graphql:"createAccount(input: $input)" json:"account"`
	}

	err = s.GraphQL.Mutate(context.Background(), &query, map[string]interface{}{
		"input": models.AccountInput{
			ID:          cmd.String("id"),
			DisplayName: cmd.String("display-name"),
			Type:        *cmd.Generic("type").(*accounts.AccountType),
		},
	})
	if err != nil {
		return err
	}

	s.WriteJSON(cmd.Writer, query.CreateAccount)

	return nil
}

// ListAccounts lists all accounts.
func ListAccounts(ctx context.Context, cmd *cli.Command) (err error) {
	s := mcli.FromContext(ctx)

	var query struct {
		Accounts []struct {
			ID          string               `json:"id"`
			DisplayName string               `json:"displayName"`
			Type        accounts.AccountType `json:"type"`
		} `json:"accounts"`
	}

	err = s.GraphQL.Query(context.Background(), &query, nil)
	if err != nil {
		return err
	}

	s.WriteJSON(cmd.Writer, query)

	return nil
}

// DeleteAccount deletes an account.
func DeleteAccount(ctx context.Context, cmd *cli.Command) (err error) {
	s := mcli.FromContext(ctx)

	// Confirm deletion
	if !cmd.Bool("confirm") {
		return errors.New("please confirm delete with --confirm")
	}

	var query struct {
		DeleteAccount struct {
			ID string `json:"id"`
		} `graphql:"deleteAccount(id: $id)" json:"account"`
	}

	err = s.GraphQL.Mutate(context.Background(), &query, map[string]interface{}{
		"id": graphql.String(cmd.String("id")),
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.Writer, "Account %q deleted.\n", query.DeleteAccount.ID)

	return nil
}

// ListTransactions lists all transactions for an account.
func ListTransactions(ctx context.Context, cmd *cli.Command) (err error) {
	s := mcli.FromContext(ctx)

	var query struct {
		Transactions []struct {
			ID   string                    `json:"id"`
			Time time.Time                 `json:"time"`
			Type events.PortfolioEventType `json:"type"`
		} `graphql:"transactions(accountID: $accountID)" json:"transactions"`
	}

	err = s.GraphQL.Query(context.Background(), &query, map[string]interface{}{
		"accountID": graphql.String(cmd.String("account-id")),
	})
	if err != nil {
		return err
	}

	s.WriteJSON(cmd.Writer, query)

	return nil
}

// CreateTransaction creates a transaction.
func CreateTransaction(ctx context.Context, cmd *cli.Command) (err error) {
	s := mcli.FromContext(ctx)

	var query struct {
		CreateTransaction struct {
			ID   string `json:"id"`
			Time string `json:"time"`
		} `graphql:"createTransaction(input: $input)" json:"account"`
	}

	err = s.GraphQL.Mutate(context.Background(), &query, map[string]interface{}{
		"input": models.TransactionInput{
			Time:                 time.Now(),
			SourceAccountID:      cmd.String("source-account-id"),
			DestinationAccountID: cmd.String("destination-account-id"),
			Type:                 *cmd.Generic("type").(*events.PortfolioEventType),
			SecurityID:           cmd.String("security-id"),
			Price:                &models.CurrencyInput{Amount: int(cmd.Float("price") * 100), Symbol: "EUR"},
			Fees:                 &models.CurrencyInput{Amount: int(cmd.Float("fees") * 100), Symbol: "EUR"},
			Taxes:                &models.CurrencyInput{Amount: int(cmd.Float("taxes") * 100), Symbol: "EUR"},
			Amount:               cmd.Float("amount"),
		},
	})
	if err != nil {
		return err
	}
	/*

		s := mcli.FromContext(ctx)
		var req = connect.NewRequest(&portfoliov1.CreatePortfolioTransactionRequest{
			Transaction: &portfoliov1.PortfolioEvent{
				PortfolioId: cmd.String("portfolio-id"),
				SecurityId:  cmd.String("security-id"),
				Type:        eventTypeFrom(cmd.String("type")),
				Amount:      cmd.Float("amount"),
				Time:        timeOrNow(cmd.Timestamp("time")),
				Price:       portfoliov1.Value(int32(cmd.Float("price") * 100)),
				Fees:        portfoliov1.Value(int32(cmd.Float("fees") * 100)),
				Taxes:       portfoliov1.Value(int32(cmd.Float("taxes") * 100)),
			},
		})

		res, err := s.PortfolioClient.CreatePortfolioTransaction(context.Background(), req)
		if err != nil {
			return err
		}

		fmt.Printf("Successfully created a %s transaction (%s) for security %s in %s.\n",
			color.CyanString(cmd.String("type")),
			color.GreenString(res.Msg.Id),
			color.CyanString(res.Msg.SecurityId),
			color.CyanString(res.Msg.PortfolioId),
		)*/

	return nil
}
