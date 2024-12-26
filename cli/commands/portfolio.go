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
	"os"
	"strings"
	"time"

	mcli "github.com/oxisto/money-gopher/cli"
	portfoliov1 "github.com/oxisto/money-gopher/gen"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PortfolioCmd is the command for portfolio related commands.
var PortfolioCmd = &cli.Command{
	Name:   "portfolio",
	Usage:  "Manage portfolios and transactions",
	Before: mcli.InjectSession,
	Commands: []*cli.Command{
		{
			Name:   "create",
			Usage:  "Creates a new portfolio",
			Action: CreatePortfolio,
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "id", Usage: "The identifier of the portfolio, e.g. mybank-myportfolio", Required: true},
				&cli.StringFlag{Name: "display-name", Usage: "The display name of the portfolio"},
			},
		},
		{
			Name:   "list",
			Usage:  "Lists all portfolios",
			Action: ListPortfolio,
			Flags:  []cli.Flag{},
		},
		{
			Name:   "show",
			Usage:  "Shows details about one portfolio",
			Action: ShowPortfolio,
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "portfolio-id", Usage: "The identifier of the portfolio, e.g. mybank-myportfolio", Required: true},
			},
		},
		{
			Name:  "transactions",
			Usage: "Subcommands supporting transactions within one portfolio",
			Commands: []*cli.Command{
				{
					Name:   "create",
					Usage:  "Creates a transaction. Defaults to a \"buy\" transaction",
					Action: CreateTransaction,
					Flags: []cli.Flag{
						&cli.StringFlag{Name: "portfolio-id", Usage: "The name of the portfolio where the transaction will be created in", Required: true},
						&cli.StringFlag{Name: "security-id", Usage: "The ID of the security this transaction belongs to (its ISIN)", Required: true},
						&cli.StringFlag{Name: "type", Usage: "The type of the transaction", Required: true, DefaultText: "buy"},
						&cli.FloatFlag{Name: "amount", Usage: "The amount of securities involved in the transaction", Required: true},
						&cli.FloatFlag{Name: "price", Usage: "The price without fees or taxes", Required: true},
						&cli.FloatFlag{Name: "fees", Usage: "Any fees that applied to the transaction"},
						&cli.FloatFlag{Name: "taxes", Usage: "Any taxes that applied to the transaction"},
						&cli.StringFlag{Name: "time", Usage: "The time of the transaction. Defaults to 'now'", DefaultText: "now"},
					},
				},
				{
					Name:   "import",
					Usage:  "Imports transactions from CSV",
					Action: ImportTransactions,
					Flags: []cli.Flag{
						&cli.StringFlag{Name: "portfolio-id", Usage: "The name of the portfolio where the transaction will be created in", Required: true},
						&cli.StringFlag{Name: "csv-file", Usage: "The path to the CSV file to import", Required: true},
					},
				},
			},
		},
	},
}

// ListPortfolio lists all portfolios.
func ListPortfolio(ctx context.Context, cmd *cli.Command) error {
	s := mcli.FromContext(ctx)
	res, err := s.PortfolioClient.ListPortfolios(
		context.Background(),
		connect.NewRequest(&portfoliov1.ListPortfoliosRequest{}),
	)
	if err != nil {
		return err
	} else {
		in := `This is a list of all portfolios.
`

		for _, portfolio := range res.Msg.Portfolios {
			snapshot, _ := s.PortfolioClient.GetPortfolioSnapshot(
				context.Background(),
				connect.NewRequest(&portfoliov1.GetPortfolioSnapshotRequest{
					PortfolioId: portfolio.Id,
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
				15, snapshot.Msg.TotalMarketValue.Pretty(),
				15, "Performance",
				15, fmt.Sprintf("%s € (%s %%)",
					greenOrRed(float64(snapshot.Msg.TotalProfitOrLoss.Value/100)),
					greenOrRed(snapshot.Msg.TotalGains*100),
				),
			)
		}

		//out, _ := glamour.Render(in, "dark")
		fmt.Println(in)
	}

	return nil
}

// CreatePortfolio creates a new portfolio.
func CreatePortfolio(ctx context.Context, cmd *cli.Command) error {
	s := mcli.FromContext(ctx)
	res, err := s.PortfolioClient.CreatePortfolio(
		context.Background(),
		connect.NewRequest(&portfoliov1.CreatePortfolioRequest{
			Portfolio: &portfoliov1.Portfolio{
				Id:          cmd.String("id"),
				DisplayName: cmd.String("display-name"),
			},
		}),
	)
	if err != nil {
		return err
	}

	fmt.Println(res.Msg)
	return nil
}

// ShowPortfolio shows details about a portfolio.
func ShowPortfolio(ctx context.Context, cmd *cli.Command) error {
	s := mcli.FromContext(ctx)
	res, err := s.PortfolioClient.GetPortfolioSnapshot(
		context.Background(),
		connect.NewRequest(&portfoliov1.GetPortfolioSnapshotRequest{
			PortfolioId: cmd.String("portfolio-id"),
			Time:        timestamppb.Now(),
		}),
	)
	if err != nil {
		return err
	}

	fmt.Println(res.Msg)
	return nil
}

func greenOrRed(f float64) string {
	if f < 0 {
		return color.RedString("%.02f", f)
	} else {
		return color.GreenString("%.02f", f)
	}
}

// CreateTransaction creates a transaction.
func CreateTransaction(ctx context.Context, cmd *cli.Command) error {
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
	)

	return nil
}

func eventTypeFrom(typ string) portfoliov1.PortfolioEventType {
	if typ == "buy" {
		return portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY
	} else if typ == "sell" {
		return portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_SELL
	} else if typ == "delivery-inbound" {
		return portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_DELIVERY_INBOUND
	} else if typ == "delivery-outbound" {
		return portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_DELIVERY_OUTBOUND
	}

	return portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_UNSPECIFIED
}

func timeOrNow(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return timestamppb.Now()
	}

	return timestamppb.New(t)
}

// ImportTransactions imports transactions from a CSV file
func ImportTransactions(ctx context.Context, cmd *cli.Command) error {
	s := mcli.FromContext(ctx)
	// Read from args[1]
	f, err := os.Open(cmd.String("csv-file"))
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	res, err := s.PortfolioClient.ImportTransactions(
		context.Background(),
		connect.NewRequest(&portfoliov1.ImportTransactionsRequest{
			PortfolioId: cmd.String("portfolio-id"),
			FromCsv:     string(b),
		}),
	)
	if err != nil {
		return err
	}

	fmt.Println(res.Msg)
	return nil
}

// PredictPortfolios predicts the portfolios for shell completion.
func PredictPortfolios(ctx context.Context, cmd *cli.Command) {
	s := mcli.FromContext(ctx)
	res, err := s.PortfolioClient.ListPortfolios(
		context.Background(),
		connect.NewRequest(&portfoliov1.ListPortfoliosRequest{}),
	)
	if err != nil {
		return
	}

	for _, p := range res.Msg.Portfolios {
		fmt.Fprintf(cmd.Root().Writer, "%s:%s\n", p.Id, p.DisplayName)
	}
}
