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
	"github.com/oxisto/money-gopher/cli"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type createPortfolio struct{}

func (cmd *createPortfolio) Exec(s *cli.Session, args ...string) {
	client := portfoliov1connect.NewPortfolioServiceClient(
		http.DefaultClient, "http://localhost:8080",
		connect.WithHTTPGet(),
	)
	res, err := client.CreatePortfolio(
		context.Background(),
		connect.NewRequest(&portfoliov1.CreatePortfolioRequest{
			Portfolio: &portfoliov1.Portfolio{
				Name:        args[1],
				DisplayName: args[2],
			},
		}),
	)
	if err != nil {
		log.Println(err)
	} else {
		log.Println(res)
	}
}

type listPortfolio struct{}

func (cmd *listPortfolio) Exec(s *cli.Session, args ...string) {
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
}

func greenOrRed(f float32) string {
	if f < 0 {
		return color.RedString("%.02f", f)
	} else {
		return color.GreenString("%.02f", f)
	}
}

type portfolioSnapshot struct{}

// Exec implements [cli.Command]
func (cmd *portfolioSnapshot) Exec(s *cli.Session, args ...string) {
	client := portfoliov1connect.NewPortfolioServiceClient(
		http.DefaultClient, "http://localhost:8080",
		connect.WithHTTPGet(),
	)
	res, err := client.GetPortfolioSnapshot(
		context.Background(),
		connect.NewRequest(&portfoliov1.GetPortfolioSnapshotRequest{
			PortfolioName: "My Portfolio",
			Time:          timestamppb.Now(),
		}),
	)
	if err != nil {
		log.Println(err)
	} else {
		log.Println(res.Msg)
	}
}

type importTransactions struct{}

// Exec implements [cli.Command]
func (cmd *importTransactions) Exec(s *cli.Session, args ...string) {
	// Read from args[1]
	f, err := os.Open(args[1])
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		log.Println(err)
		return
	}

	client := portfoliov1connect.NewPortfolioServiceClient(
		http.DefaultClient, "http://localhost:8080",
		connect.WithHTTPGet(),
	)
	res, err := client.ImportTransactions(
		context.Background(),
		connect.NewRequest(&portfoliov1.ImportTransactionsRequest{
			PortfolioName: args[0],
			FromCsv:       string(b),
		}),
	)
	if err != nil {
		log.Println(err)
	} else {
		log.Println(res.Msg)
	}
}
