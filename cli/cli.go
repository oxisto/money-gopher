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

// cli provides the commands for a simple CLI.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"github.com/lmittmann/tint"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	oauth2 "github.com/oxisto/oauth2go"
)

// Session holds all necessary information about the current CLI session.
type Session struct {
	PortfolioClient  portfoliov1connect.PortfolioServiceClient  `json:"-"`
	SecuritiesClient portfoliov1connect.SecuritiesServiceClient `json:"-"`

	Config *oauth2.Config
	Token  *oauth2.Token
}

func NewSession(config *oauth2.Config, token *oauth2.Token) (s *Session) {
	s = &Session{
		Config: config,
		Token:  token,
	}

	s.initClients()

	return s
}

func ContinueSession() (s *Session, err error) {
	var (
		file *os.File
	)

	file, err = os.OpenFile("session.json", os.O_RDONLY, 0600)
	if err != nil {
		return
	}

	s = new(Session)
	err = json.NewDecoder(file).Decode(&s)
	if err != nil {
		return nil, fmt.Errorf("could not parse session file: %w", err)
	}

	s.initClients()

	return
}

func (s *Session) Save() (err error) {
	var (
		file *os.File
	)

	file, err = os.OpenFile("session.json", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}

	err = json.NewEncoder(file).Encode(s)
	if err != nil {
		return fmt.Errorf("could not save session file: %w", err)
	}

	return nil
}

func (s *Session) initClients() {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			if req.Spec().IsClient {
				var t, err = s.Config.TokenSource(context.Background(), s.Token).Token()
				if err != nil {
					slog.Error("Could not retrieve token", tint.Err(err))
				} else {
					req.Header().Set("Authorization", "Bearer "+t.AccessToken)
				}
			}
			return next(ctx, req)
		})
	}

	s.PortfolioClient = portfoliov1connect.NewPortfolioServiceClient(
		http.DefaultClient, "http://localhost:8080",
		connect.WithHTTPGet(),
		connect.WithInterceptors(connect.UnaryInterceptorFunc(interceptor)),
	)

	s.SecuritiesClient = portfoliov1connect.NewSecuritiesServiceClient(
		http.DefaultClient, "http://localhost:8080",
		connect.WithHTTPGet(),
		connect.WithInterceptors(connect.UnaryInterceptorFunc(interceptor)),
	)
}
