// Copyright 2024 Christian Banse
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

package server

import (
	"crypto/ecdsa"
	"log/slog"
	"net/http"

	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/persistence"
	"github.com/oxisto/money-gopher/service/portfolio"
	"github.com/oxisto/money-gopher/service/securities"

	"connectrpc.com/connect"
	"connectrpc.com/vanguard"
	"github.com/lmittmann/tint"
	oauth2 "github.com/oxisto/oauth2go"
	"github.com/oxisto/oauth2go/login"
	"github.com/oxisto/oauth2go/storage"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Options holds all options to configure the server.
type Options struct {
	Debug bool

	EmbeddedOAuth2ServerDashboardCallback string

	PrivateKeyFile     string
	PrivateKeyPassword string
}

// StartServer starts the server.
func StartServer(pdb *persistence.DB, q *persistence.Queries, opts Options) (err error) {
	var (
		authSrv    *oauth2.AuthorizationServer
		transcoder *vanguard.Transcoder
	)

	authSrv = oauth2.NewServer(
		":8000",
		oauth2.WithClient("dashboard", "", opts.EmbeddedOAuth2ServerDashboardCallback),
		oauth2.WithClient("cli", "", "http://localhost:10000/callback"),
		oauth2.WithPublicURL("http://localhost:8000"),
		login.WithLoginPage(
			login.WithUser("money", "money"),
		),
		oauth2.WithAllowedOrigins("*"),
		oauth2.WithSigningKeysFunc(func() map[int]*ecdsa.PrivateKey {
			return storage.LoadSigningKeys(opts.PrivateKeyFile, opts.PrivateKeyPassword, true)
		}),
	)
	go authSrv.ListenAndServe()

	interceptors := connect.WithInterceptors(
		NewSimpleLoggingInterceptor(),
		NewAuthInterceptor(),
	)

	portfolioService := vanguard.NewService(
		portfoliov1connect.NewPortfolioServiceHandler(portfolio.NewService(
			portfolio.Options{
				DB:               pdb,
				SecuritiesClient: portfoliov1connect.NewSecuritiesServiceClient(http.DefaultClient, portfolio.DefaultSecuritiesServiceURL),
			},
		), interceptors))
	securitiesService := vanguard.NewService(
		portfoliov1connect.NewSecuritiesServiceHandler(securities.NewService(pdb), interceptors),
	)

	transcoder, err = vanguard.NewTranscoder([]*vanguard.Service{
		portfolioService,
		securitiesService,
	}, vanguard.WithCodec(func(tr vanguard.TypeResolver) vanguard.Codec {
		codec := vanguard.NewJSONCodec(tr)
		codec.MarshalOptions.EmitDefaultValues = true
		return codec
	}))
	if err != nil {
		slog.Error("transcoder failed", tint.Err(err))
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", transcoder)

	err = http.ListenAndServe(
		":8080",
		h2c.NewHandler(handleCORS(mux), &http2.Server{}),
	)

	slog.Error("listen failed", tint.Err(err))
	return err
}
