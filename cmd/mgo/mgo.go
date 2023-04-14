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

package main

import (
	"context"
	"log"
	"net/http"

	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/persistence"
	"github.com/oxisto/money-gopher/repl"
	_ "github.com/oxisto/money-gopher/repl/commands"
	"github.com/oxisto/money-gopher/service/securities"

	"github.com/bufbuild/connect-go"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type PortfolioService struct {
	portfoliov1connect.UnimplementedPortfolioServiceHandler
}

func (ps *PortfolioService) CreatePortfolio(ctx context.Context, req *connect.Request[portfoliov1.PortfolioCreateMessage]) (res *connect.Response[portfoliov1.Portfolio], err error) {
	res = connect.NewResponse(&portfoliov1.Portfolio{Name: req.Msg.Name})
	res.Header().Set("X-Money", "true")

	return
}

func main() {
	log.SetPrefix("[🤑] ")
	log.SetFlags(log.Lmsgprefix | log.Ltime)
	log.Print("Welcome to The Money Gopher")

	_, err := persistence.OpenDB(persistence.Options{})
	if err != nil {
		log.Fatalf("Error while opening database: %v", err)
	}

	mux := http.NewServeMux()
	// The generated constructors return a path and a plain net/http
	// handler.
	mux.Handle(portfoliov1connect.NewPortfolioServiceHandler(&PortfolioService{}))
	mux.Handle(portfoliov1connect.NewSecuritiesServiceHandler(securities.NewService()))

	go func() {
		err = http.ListenAndServe(
			"localhost:8080",
			h2c.NewHandler(mux, &http2.Server{}),
		)
		log.Fatalf("listen failed: %v", err)
	}()

	r := repl.REPL{}
	r.Run()
}
