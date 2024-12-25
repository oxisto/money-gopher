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
	"github.com/urfave/cli/v3"
)

var SecurityCmd = &cli.Command{
	Name:   "security",
	Usage:  "Security commands",
	Before: mcli.InjectSession,
	Commands: []*cli.Command{
		{
			Name:   "list",
			Usage:  "Lists all securities",
			Action: ListSecuritiesCmd,
		},
		{
			Name:   "update-quote",
			Usage:  "Triggers an update of one or more securities' quotes",
			Action: UpdateQuoteCmd,
			Flags: []cli.Flag{
				&cli.StringSliceFlag{Name: "security-ids", Usage: "The security IDs to update", Required: true},
			},
			EnableShellCompletion: true,
			ShellComplete:         PredictSecurities,
		},
		{
			Name:   "update-all-quotes",
			Usage:  "Triggers an update of all quotes",
			Action: UpdateAllQuotesCmd,
		},
	},
}

func ListSecuritiesCmd(ctx context.Context, cmd *cli.Command) error {
	s := mcli.FromContext(ctx)
	res, err := s.SecuritiesClient.ListSecurities(context.Background(), connect.NewRequest(&portfoliov1.ListSecuritiesRequest{}))
	if err != nil {
		return err
	}

	fmt.Println(res.Msg.Securities)
	return nil
}

func UpdateQuoteCmd(ctx context.Context, cmd *cli.Command) error {
	s := mcli.FromContext(ctx)
	_, err := s.SecuritiesClient.TriggerSecurityQuoteUpdate(
		context.Background(),
		connect.NewRequest(&portfoliov1.TriggerQuoteUpdateRequest{
			SecurityIds: cmd.StringSlice("security-ids"),
		}),
	)

	return err
}

func UpdateAllQuotesCmd(ctx context.Context, cmd *cli.Command) error {
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

/*
func WithPredictSecurities(s *cli.Session) kongcompletion.Option {
	return kongcompletion.WithPredictor(
		"security",
		PredictSecurities(s),
	)
}

func PredictSecurities(s *cli.Session) complete.PredictFunc {
	return func(complete.Args) (names []string) {
		res, err := s.SecuritiesClient.ListSecurities(
			context.Background(),
			connect.NewRequest(&portfoliov1.ListSecuritiesRequest{}),
		)
		if err != nil {
			return nil
		}

		for _, p := range res.Msg.Securities {
			names = append(names, p.Id)
		}

		return
	}
}
*/

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
