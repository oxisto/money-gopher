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
	"fmt"

	mcli "github.com/oxisto/money-gopher/cli"
	portfoliov1 "github.com/oxisto/money-gopher/gen"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"
)

var BankAccountCmd = &cli.Command{
	Name:   "bank-account",
	Usage:  "Manage bank accounts",
	Before: mcli.InjectSession,
	Commands: []*cli.Command{
		{
			Name:   "create",
			Usage:  "Creates a new bank account",
			Action: CreateBankAccount,
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "name", Usage: "The identifier of the portfolio, e.g. mybank-myportfolio", Required: true},
				&cli.StringFlag{Name: "display-name", Usage: "The display name of the portfolio"},
			},
		},
	},
}

func CreateBankAccount(ctx context.Context, cmd *cli.Command) error {
	s := mcli.FromContext(ctx)
	res, err := s.PortfolioClient.CreateBankAccount(
		context.Background(),
		connect.NewRequest(&portfoliov1.CreateBankAccountRequest{
			BankAccount: &portfoliov1.BankAccount{
				Name:        cmd.String("name"),
				DisplayName: cmd.String("display-name"),
			},
		}),
	)
	if err != nil {
		return err
	}

	fmt.Fprint(cmd.Writer, res.Msg)
	return nil
}
