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

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/oxisto/money-gopher/graph"
	"github.com/oxisto/money-gopher/persistence"
	"github.com/oxisto/money-gopher/securities/quote"

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
func StartServer(pdb *persistence.DB, opts Options) (err error) {
	var (
		authSrv *oauth2.AuthorizationServer
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

	// Create a quote updater
	qu := quote.NewQuoteUpdater(pdb)

	// Configure serve mux
	mux := http.NewServeMux()
	ConfigureGraphQL(mux, pdb, qu)

	err = http.ListenAndServe(
		":8080",
		h2c.NewHandler(handleCORS(mux), &http2.Server{}),
	)

	slog.Error("listen failed", tint.Err(err))
	return err
}

// ConfigureGraphQL configures the GraphQL server for a [http.ServeMux].
func ConfigureGraphQL(
	mux *http.ServeMux,
	db *persistence.DB,
	qu quote.QuoteUpdater,
) (err error) {
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		DB:           db,
		QuoteUpdater: qu,
	}}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.Use(extension.Introspection{})

	mux.Handle("/graphql", playground.Handler("GraphQL playground", "/graphql/query"))
	mux.Handle("/graphql/query", srv)

	return err
}
