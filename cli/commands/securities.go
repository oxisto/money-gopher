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
	"fmt"

	mcli "github.com/oxisto/money-gopher/cli"
	portfoliov1 "github.com/oxisto/money-gopher/gen"

	"connectrpc.com/connect"
	"github.com/shurcooL/graphql"
	"github.com/urfave/cli/v3"
)

// SecuritiesCmd is the command for security related commands.
var SecuritiesCmd = &cli.Command{
	Name:   "securities",
	Usage:  "Securities commands",
	Before: mcli.InjectSession,
	Commands: []*cli.Command{
		{
			Name:   "list",
			Usage:  "Lists all securities",
			Action: ListSecurities,
		},
		{
			Name:   "show",
			Usage:  "Shows information about a security",
			Action: ShowSecurity,
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "security-id", Usage: "The security ID", Required: true},
			},
		},
		{
			Name:   "update-quote",
			Usage:  "Triggers an update of one or more securities' quotes",
			Action: UpdateQuote,
			Flags: []cli.Flag{
				&cli.StringSliceFlag{Name: "security-ids", Usage: "The security IDs to update", Required: true},
			},
			EnableShellCompletion: true,
			ShellComplete:         PredictSecurities,
		},
		{
			Name:   "update-all-quotes",
			Usage:  "Triggers an update of all quotes",
			Action: UpdateAllQuotes,
		},
	},
}

// ListSecurities lists all securities.
func ListSecurities(ctx context.Context, cmd *cli.Command) (err error) {
	s := mcli.FromContext(ctx)

	var query struct {
		Securities []struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"securities"`
	}

	err = s.GraphQL.Query(context.Background(), &query, nil)
	if err != nil {
		return err
	}

	s.WriteJSON(cmd.Writer, query)

	return nil
}

// ShowSecurity shows information about a security.
func ShowSecurity(ctx context.Context, cmd *cli.Command) (err error) {
	s := mcli.FromContext(ctx)

	var query struct {
		Security struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
			ListedAs    []struct {
				Ticker string `json:"ticker"`
			} `json:"listedAs"`
		} `graphql:"security(id: $id)" json:"security"`
	}

	err = s.GraphQL.Query(context.Background(), &query, map[string]interface{}{
		"id": graphql.String(cmd.String("security-id")),
	})
	if err != nil {
		return err
	}

	s.WriteJSON(cmd.Writer, query)

	return nil
}

// UpdateQuote triggers an update of one or more securities' quotes.
func UpdateQuote(ctx context.Context, cmd *cli.Command) (err error) {
	s := mcli.FromContext(ctx)

	var query struct {
		TriggerQuoteUpdate bool `graphql:"triggerQuoteUpdate(securityIDs: $IDs)" json:"security"`
	}

	var ids []graphql.String
	for _, id := range cmd.StringSlice("security-ids") {
		ids = append(ids, graphql.String(id))
	}

	err = s.GraphQL.Mutate(context.Background(), &query, map[string]interface{}{
		"IDs": ids,
	})
	if err != nil {
		return err
	}
	return err
}

// UpdateAllQuotes triggers an update of all quotes.
func UpdateAllQuotes(ctx context.Context, cmd *cli.Command) error {
	s := mcli.FromContext(ctx)
	res, err := s.SecuritiesClient.ListSecurities(context.Background(), connect.NewRequest(&portfoliov1.ListSecuritiesRequest{}))
	if err != nil {
		return err
	}

	var names []string

	for _, sec := range res.Msg.Securities {
		names = append(names, sec.Id)
	}

	_, err = s.SecuritiesClient.TriggerSecurityQuoteUpdate(
		context.Background(),
		connect.NewRequest(&portfoliov1.TriggerQuoteUpdateRequest{
			SecurityIds: names,
		}),
	)

	return err
}

// PredictSecurities predicts the securities for shell completion.
func PredictSecurities(ctx context.Context, cmd *cli.Command) {
	s := mcli.FromContext(ctx)
	res, err := s.SecuritiesClient.ListSecurities(
		context.Background(),
		connect.NewRequest(&portfoliov1.ListSecuritiesRequest{}),
	)
	if err != nil {
		return
	}

	for _, p := range res.Msg.Securities {
		fmt.Fprintf(cmd.Root().Writer, "%s:%s\n", p.Id, p.DisplayName)
	}
}
