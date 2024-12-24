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
	"crypto/ecdsa"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/persistence"
	"github.com/oxisto/money-gopher/service/portfolio"
	"github.com/oxisto/money-gopher/service/securities"

	"connectrpc.com/connect"
	"connectrpc.com/vanguard"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/alecthomas/kong"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	oauth2 "github.com/oxisto/oauth2go"
	"github.com/oxisto/oauth2go/login"
	"github.com/oxisto/oauth2go/storage"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var cmd moneydCmd

type moneydCmd struct {
	Debug bool `help:"Enable debug mode."`

	EmbeddedOAuth2ServerDashboardCallback string `default:"http://localhost:3000/api/auth/callback/money-gopher" help:"Specifies the callback URL for the dashboard, if the embedded oauth2 server is used."`

	PrivateKeyFile     string `default:"private.key"`
	PrivateKeyPassword string `default:"moneymoneymoney"`
}

func main() {
	ctx := kong.Parse(&cmd)

	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

type query struct{}

func (query) Hello() string { return "Hello, world!" }

func (cmd *moneydCmd) Run() error {
	var (
		w       = os.Stdout
		level   = slog.LevelInfo
		authSrv *oauth2.AuthorizationServer
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
	slog.Info("Welcome to the Money Gopher", "money", "🤑")

	db, err := persistence.OpenDB(persistence.Options{})
	if err != nil {
		slog.Error("Error while opening database", tint.Err(err))
		return err
	}

	authSrv = oauth2.NewServer(
		":8000",
		oauth2.WithClient("dashboard", "", cmd.EmbeddedOAuth2ServerDashboardCallback),
		oauth2.WithClient("cli", "", "http://localhost:10000/callback"),
		oauth2.WithPublicURL("http://localhost:8000"),
		login.WithLoginPage(
			login.WithUser("money", "money"),
		),
		oauth2.WithAllowedOrigins("*"),
		oauth2.WithSigningKeysFunc(func() map[int]*ecdsa.PrivateKey {
			return storage.LoadSigningKeys(cmd.PrivateKeyFile, cmd.PrivateKeyPassword, true)
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
				DB:               db,
				SecuritiesClient: portfoliov1connect.NewSecuritiesServiceClient(http.DefaultClient, portfolio.DefaultSecuritiesServiceURL),
			},
		), interceptors))
	securitiesService := vanguard.NewService(
		portfoliov1connect.NewSecuritiesServiceHandler(securities.NewService(db), interceptors),
	)

	transcoder, err := vanguard.NewTranscoder([]*vanguard.Service{
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

func NewSimpleLoggingInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			slog.Debug("Handling RPC Request",
				slog.Group("req",
					"procedure", req.Spec().Procedure,
					"httpmethod", req.HTTPMethod(),
				))
			return next(ctx, req)
		})
	}

	return connect.UnaryInterceptorFunc(interceptor)
}

func NewAuthInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		k, err := keyfunc.NewDefault([]string{"http://localhost:8000/certs"})
		if err != nil {
			slog.Error("Error while setting up JWKS", tint.Err(err))
		}

		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			var (
				claims jwt.RegisteredClaims
				auth   string
				token  string
				err    error
				ok     bool
			)
			auth = req.Header().Get("Authorization")
			if auth == "" {
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					errors.New("no token provided"),
				)
			}

			token, ok = strings.CutPrefix(auth, "Bearer ")
			if !ok {
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					errors.New("no token provided"),
				)
			}

			_, err = jwt.ParseWithClaims(token, &claims, k.Keyfunc)
			if err != nil {
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					err,
				)
			}

			ctx = context.WithValue(ctx, "claims", claims)
			return next(ctx, req)
		})
	}
	return connect.UnaryInterceptorFunc(interceptor)
}
