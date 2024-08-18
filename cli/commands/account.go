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

	"github.com/oxisto/money-gopher/cli"
	portfoliov1 "github.com/oxisto/money-gopher/gen"

	"connectrpc.com/connect"
)

type BankAccountCmd struct {
	Create CreateBankAccountCmd `cmd:"" help:"Creates a new bank account."`
}

type CreateBankAccountCmd struct {
	Name        string `help:"The identifier of the portfolio, e.g. mybank-myportfolio" required:""`
	DisplayName string `help:"The display name of the portfolio"`
}

func (cmd *CreateBankAccountCmd) Run(s *cli.Session) error {
	res, err := s.PortfolioClient.CreateBankAccount(
		context.Background(),
		connect.NewRequest(&portfoliov1.CreateBankAccountRequest{
			BankAccount: &portfoliov1.BankAccount{
				Name:        cmd.Name,
				DisplayName: cmd.DisplayName,
			},
		}),
	)
	if err != nil {
		return err
	}

	fmt.Println(res.Msg)
	return nil
}
