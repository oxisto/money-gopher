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
	"io"
	"log"
	"net/http"
	"os"

	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/repl"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/bufbuild/connect-go"
)

type portfolioSnapshot struct{}

// Exec implements [repl.Command]
func (cmd *portfolioSnapshot) Exec(r *repl.REPL, args ...string) {
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

// Exec implements [repl.Command]
func (cmd *importTransactions) Exec(r *repl.REPL, args ...string) {
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
