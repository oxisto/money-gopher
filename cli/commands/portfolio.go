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
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
	kongcompletion "github.com/jotaen/kong-completion"
	"github.com/oxisto/money-gopher/cli"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/posener/complete"
	"google.golang.org/protobuf/types/known/timestamppb"

	"connectrpc.com/connect"
)

type PortfolioCmd struct {
	Create             CreatePortfolioCmd    `cmd:"" help:"Creates a new portfolio."`
	List               ListPortfolioCmd      `cmd:"" help:"Lists all portfolios."`
	Show               ShowPortfolioCmd      `cmd:"" help:"Shows details about one portfolio."`
	ImportTransactions ImportTransactionsCmd `cmd:"" help:"Imports transactions from CSV."`
}

type ListPortfolioCmd struct{}

func (l *ListPortfolioCmd) Run(s *cli.Session) error {
	client := portfoliov1connect.NewPortfolioServiceClient(
		http.DefaultClient, "http://localhost:8080",
		connect.WithHTTPGet(),
	)
	res, err := client.ListPortfolios(
		context.Background(),
		connect.NewRequest(&portfoliov1.ListPortfoliosRequest{}),
	)
	if err != nil {
		log.Println(err)
	} else {
		in := `This is a list of all portfolios.
`

		for _, portfolio := range res.Msg.Portfolios {
			snapshot, _ := client.GetPortfolioSnapshot(
				context.Background(),
				connect.NewRequest(&portfoliov1.GetPortfolioSnapshotRequest{
					PortfolioName: portfolio.Name,
				}),
			)

			in += fmt.Sprintf(`
| %-*s |
| %s | %s |
| %-*s | %*s |
| %-*s | %*s |
`,
				15+15+3, color.New(color.FgWhite, color.Bold).Sprint(portfolio.DisplayName),
				strings.Repeat("-", 15),
				strings.Repeat("-", 15),
				15, "Market Value",
				15, fmt.Sprintf("%.02f €", snapshot.Msg.TotalMarketValue),
				15, "Performance",
				15, fmt.Sprintf("%s € (%s %%)",
					greenOrRed(snapshot.Msg.TotalProfitOrLoss),
					greenOrRed(snapshot.Msg.TotalGains*100),
				),
			)
		}

		//out, _ := glamour.Render(in, "dark")
		fmt.Println(in)
	}

	return nil
}

type CreatePortfolioCmd struct {
	Name        string `help:"The identifier of the portfolio, e.g. mybank/myportfolio" required:""`
	DisplayName string `help:"The display name of the portfolio"`
}

func (cmd *CreatePortfolioCmd) Run(s *cli.Session) error {
	client := portfoliov1connect.NewPortfolioServiceClient(
		http.DefaultClient, "http://localhost:8080",
		connect.WithHTTPGet(),
	)
	res, err := client.CreatePortfolio(
		context.Background(),
		connect.NewRequest(&portfoliov1.CreatePortfolioRequest{
			Portfolio: &portfoliov1.Portfolio{
				Name:        cmd.Name,
				DisplayName: cmd.DisplayName,
			},
		}),
	)
	if err != nil {
		return err
	}

	log.Println(res.Msg)
	return nil
}

type ShowPortfolioCmd struct {
	PortfolioName string `help:"The identifier of the portfolio, e.g. mybank/myportfolio" required:"" predictor:"portfolio"`
}

func (cmd *ShowPortfolioCmd) Run(s *cli.Session) error {
	client := portfoliov1connect.NewPortfolioServiceClient(
		http.DefaultClient, "http://localhost:8080",
		connect.WithHTTPGet(),
	)
	res, err := client.GetPortfolioSnapshot(
		context.Background(),
		connect.NewRequest(&portfoliov1.GetPortfolioSnapshotRequest{
			PortfolioName: cmd.PortfolioName,
			Time:          timestamppb.Now(),
		}),
	)
	if err != nil {
		return err
	}

	log.Println(res.Msg)
	return nil
}

func greenOrRed(f float32) string {
	if f < 0 {
		return color.RedString("%.02f", f)
	} else {
		return color.GreenString("%.02f", f)
	}
}

type ImportTransactionsCmd struct {
	PortfolioName string `required:"" predictor:"portfolio"`
	CsvFile       string `arg:"" help:"The path to the CSV file to import"`
}

// Exec implements [cli.Command]
func (cmd *ImportTransactionsCmd) Run(s *cli.Session) error {
	// Read from args[1]
	f, err := os.Open(cmd.CsvFile)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	client := portfoliov1connect.NewPortfolioServiceClient(
		http.DefaultClient, "http://localhost:8080",
		connect.WithHTTPGet(),
	)
	res, err := client.ImportTransactions(
		context.Background(),
		connect.NewRequest(&portfoliov1.ImportTransactionsRequest{
			PortfolioName: cmd.PortfolioName,
			FromCsv:       string(b),
		}),
	)
	if err != nil {
		return err
	}

	log.Println(res.Msg)
	return nil
}

type PortfolioPredictor struct{}

func (p *PortfolioPredictor) Predict(complete.Args) (names []string) {
	client := portfoliov1connect.NewPortfolioServiceClient(
		http.DefaultClient, "http://localhost:8080",
		connect.WithHTTPGet(),
	)
	res, err := client.ListPortfolios(
		context.Background(),
		connect.NewRequest(&portfoliov1.ListPortfoliosRequest{}),
	)
	if err != nil {
		return nil
	}

	for _, p := range res.Msg.Portfolios {
		names = append(names, p.Name)
	}

	return
}

var PredictPortfolios = kongcompletion.WithPredictor(
	"portfolio",
	&PortfolioPredictor{},
)
