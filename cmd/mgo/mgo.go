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
	"log"
	"net/http"
	"strings"

	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/persistence"
	"github.com/oxisto/money-gopher/repl"
	_ "github.com/oxisto/money-gopher/repl/commands"
	"github.com/oxisto/money-gopher/service/portfolio"
	"github.com/oxisto/money-gopher/service/securities"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	log.SetPrefix("[🤑] ")
	log.SetFlags(log.Lmsgprefix | log.Ltime)
	log.Print("Welcome to The Money Gopher")

	db, err := persistence.OpenDB(persistence.Options{})
	if err != nil {
		log.Fatalf("Error while opening database: %v", err)
	}

	mux := http.NewServeMux()
	// The generated constructors return a path and a plain net/http
	// handler.
	mux.Handle(portfoliov1connect.NewPortfolioServiceHandler(portfolio.NewService()))
	mux.Handle(portfoliov1connect.NewSecuritiesServiceHandler(securities.NewService(db)))

	go func() {
		err = http.ListenAndServe(
			"localhost:8080",
			h2c.NewHandler(handleCORS(mux), &http2.Server{}),
		)
		log.Fatalf("listen failed: %v", err)
	}()

	r := repl.REPL{}
	r.Run()
}

func handleCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Add("Vary", "Origin")

		if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join([]string{
				"Connect-Protocol-Version",
				"Content-Type",
				"Accept",
				"Authorization",
			}, ","))
			w.Header().Set("Access-Control-Allow-Methods", strings.Join([]string{
				"GET",
				"POST",
				"PUT",
				"DELETE",
			}, ","))
		} else {
			h.ServeHTTP(w, r)
		}
	})
}
