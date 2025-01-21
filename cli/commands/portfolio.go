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
	"strings"
	"time"

	mcli "github.com/oxisto/money-gopher/cli"
	"github.com/oxisto/money-gopher/currency"
	"github.com/oxisto/money-gopher/models"

	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
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
				&cli.StringSliceFlag{Name: "account-ids", Usage: "The account IDs that should be linked to the portfolio"},
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
func ListPortfolio(ctx context.Context, cmd *cli.Command) (err error) {
	s := mcli.FromContext(ctx)

	var query struct {
		Portfolios []struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
			Snapshot    struct {
				TotalMarketValue  currency.Currency `json:"totalMarketValue"`
				TotalProfitOrLoss currency.Currency `json:"totalProfitOrLoss"`
				TotalGains        float64           `json:"totalGains"`
			} `graphql:"snapshot(when: $when)" json:"snapshot"`
		} `json:"portfolios"`
	}

	err = s.GraphQL.Query(context.Background(), &query, map[string]any{
		"when": time.Now(),
	})
	if err != nil {
		return err
	}

	var in string

	for _, portfolio := range query.Portfolios {
		snapshot := portfolio.Snapshot

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
			15, snapshot.TotalMarketValue.Pretty(),
			15, "Performance",
			15, fmt.Sprintf("%s € (%s %%)",
				greenOrRed(float64(snapshot.TotalProfitOrLoss.Amount/100)),
				greenOrRed(snapshot.TotalGains*100),
			),
		)
	}

	fmt.Fprintln(cmd.Writer, in)

	return nil
}

// CreatePortfolio creates a new portfolio.
func CreatePortfolio(ctx context.Context, cmd *cli.Command) (err error) {
	s := mcli.FromContext(ctx)

	var query struct {
		CreatePortfolio struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
			Accounts    []struct {
				ID          string `json:"id"`
				DisplayName string `json:"displayName"`
				Type        string `json:"type"`
			} `json:"accounts"`
		} `graphql:"createPortfolio(input: $input)" json:"account"`
	}

	err = s.GraphQL.Mutate(context.Background(), &query, map[string]interface{}{
		"input": models.PortfolioInput{
			ID:          cmd.String("id"),
			DisplayName: cmd.String("display-name"),
			AccountIds:  cmd.StringSlice("account-ids"),
		},
	})
	if err != nil {
		return err
	}

	s.WriteJSON(cmd.Writer, query.CreatePortfolio)

	return nil
}

// ShowPortfolio shows details about a portfolio.
func ShowPortfolio(ctx context.Context, cmd *cli.Command) error {
	/*s := mcli.FromContext(ctx)
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

	fmt.Println(res.Msg)*/
	return nil
}

func greenOrRed(f float64) string {
	if f < 0 {
		return color.RedString("%.02f", f)
	} else {
		return color.GreenString("%.02f", f)
	}
}

/*
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
}*/

// ImportTransactions imports transactions from a CSV file
func ImportTransactions(ctx context.Context, cmd *cli.Command) error {
	/*s := mcli.FromContext(ctx)
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

	fmt.Println(res.Msg)*/
	return nil
}

// PredictPortfolios predicts the portfolios for shell completion.
func PredictPortfolios(ctx context.Context, cmd *cli.Command) {
	s := mcli.FromContext(ctx)

	var query struct {
		Portfolios []struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"portfolios"`
	}

	err := s.GraphQL.Query(context.Background(), &query, nil)
	if err != nil {
		return
	}

	for _, p := range query.Portfolios {
		fmt.Fprintf(cmd.Writer, "%s:%s\n", p.ID, p.DisplayName)
	}
}
