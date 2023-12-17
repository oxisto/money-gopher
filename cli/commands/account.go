package commands

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/oxisto/money-gopher/cli"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
)

type BankAccountCmd struct {
	Create CreateBankAccountCmd `cmd:"" help:"Creates a new bank account."`
}

type CreateBankAccountCmd struct {
	Name        string `help:"The identifier of the portfolio, e.g. mybank/myportfolio" required:""`
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
