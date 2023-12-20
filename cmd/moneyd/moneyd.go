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
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/persistence"
	"github.com/oxisto/money-gopher/service/portfolio"
	"github.com/oxisto/money-gopher/service/securities"
	"github.com/oxisto/money-gopher/ui"
)

var cmd moneydCmd

type moneydCmd struct {
	Debug bool `help:"Enable debug mode."`
}

func main() {
	ctx := kong.Parse(&cmd)

	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

func (cmd *moneydCmd) Run() error {
	var (
		w     = os.Stdout
		level = slog.LevelInfo
	)

	if cmd.Debug {
		level = slog.LevelDebug
	}

	logger := slog.New(
		tint.NewHandler(colorable.NewColorable(w), &tint.Options{
			TimeFormat: time.TimeOnly,
			Level:      level,
			NoColor:    !isatty.IsTerminal(w.Fd()),
		}),
	)

	slog.SetDefault(logger)
	slog.Info("Welcome to the Money Gopher")

	db, err := persistence.OpenDB(persistence.Options{})
	if err != nil {
		slog.Error("Error while opening database", tint.Err(err))
	}

	mux := http.NewServeMux()
	// The generated constructors return a path and a plain net/http
	// handler.
	mux.Handle(portfoliov1connect.NewPortfolioServiceHandler(portfolio.NewService(
		portfolio.Options{
			DB:               db,
			SecuritiesClient: portfoliov1connect.NewSecuritiesServiceClient(http.DefaultClient, portfolio.DefaultSecuritiesServiceURL),
		},
	)))
	mux.Handle(portfoliov1connect.NewSecuritiesServiceHandler(securities.NewService(db)))
	mux.Handle("/", ui.SvelteKitHandler("/"))

	err = http.ListenAndServe(
		":8080",
		h2c.NewHandler(handleCORS(mux), &http2.Server{}),
	)

	slog.Error("listen failed", tint.Err(err))

	return err
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
