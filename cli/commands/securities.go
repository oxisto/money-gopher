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

	"github.com/oxisto/money-gopher/cli"
	portfoliov1 "github.com/oxisto/money-gopher/gen"

	"connectrpc.com/connect"
	kongcompletion "github.com/jotaen/kong-completion"
	"github.com/posener/complete"
)

type SecurityCmd struct {
	List            ListSecuritiesCmd  `cmd:"" help:"Lists all securities."`
	UpdateQuote     UpdateQuoteCmd     `cmd:"" help:"Triggers an update of one or more securities' quotes."`
	UpdateAllQuotes UpdateAllQuotesCmd `cmd:"" help:"Triggers an update of all quotes."`
}

type ListSecuritiesCmd struct{}

// Exec implements [repl.Command]
func (cmd *ListSecuritiesCmd) Run(s *cli.Session) error {
	res, err := s.SecuritiesClient.ListSecurities(context.Background(), connect.NewRequest(&portfoliov1.ListSecuritiesRequest{}))
	if err != nil {
		return err
	}

	fmt.Println(res.Msg.Securities)
	return nil
}

type UpdateQuoteCmd struct {
	SecurityNames []string `arg:""`
}

// Exec implements [cli.Command]
func (cmd *UpdateQuoteCmd) Run(s *cli.Session) error {
	_, err := s.SecuritiesClient.TriggerSecurityQuoteUpdate(
		context.Background(),
		connect.NewRequest(&portfoliov1.TriggerQuoteUpdateRequest{
			SecurityNames: cmd.SecurityNames,
		}),
	)

	return err
}

type UpdateAllQuotesCmd struct{}

// Exec implements [cli.Command]
func (cmd *UpdateAllQuotesCmd) Run(s *cli.Session) error {
	res, err := s.SecuritiesClient.ListSecurities(context.Background(), connect.NewRequest(&portfoliov1.ListSecuritiesRequest{}))
	if err != nil {
		return err
	}

	var names []string

	for _, sec := range res.Msg.Securities {
		names = append(names, sec.Name)
	}

	_, err = s.SecuritiesClient.TriggerSecurityQuoteUpdate(
		context.Background(),
		connect.NewRequest(&portfoliov1.TriggerQuoteUpdateRequest{
			SecurityNames: names,
		}),
	)

	return err
}

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
			names = append(names, p.Name)
		}

		return
	}
}
