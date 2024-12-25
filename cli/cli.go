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

	"github.com/oxisto/money-gopher/gen/portfoliov1connect"

	"connectrpc.com/connect"
	"github.com/lmittmann/tint"
	oauth2 "github.com/oxisto/oauth2go"
	"github.com/urfave/cli/v3"
)

type sessionKeyType struct{}

var SessionKey sessionKeyType

// Session holds all necessary information about the current CLI session.
type Session struct {
	PortfolioClient  portfoliov1connect.PortfolioServiceClient  `json:"-"`
	SecuritiesClient portfoliov1connect.SecuritiesServiceClient `json:"-"`

	opts *SessionOptions
}

// SessionOptions holds all options to configure a [Session].
type SessionOptions struct {
	OAuth2Config *oauth2.Config
	HttpClient   *http.Client
	Token        *oauth2.Token
	BaseURL      string
}

// MergeWith can be used to merge two [SessionOptions] structs.
func (opts *SessionOptions) MergeWith(other *SessionOptions) *SessionOptions {
	if other.BaseURL != "" {
		opts.BaseURL = other.BaseURL
	}

	if other.HttpClient != nil {
		opts.HttpClient = other.HttpClient
	}

	if other.OAuth2Config != nil {
		opts.OAuth2Config = other.OAuth2Config
	}

	if other.Token != nil {
		opts.Token = other.Token
	}

	return opts
}

// DefaultBaseURL is the default base URL for all services.
const DefaultBaseURL = "http://localhost:8080"

func NewSession(opts *SessionOptions) (s *Session) {
	def := &SessionOptions{
		HttpClient: opts.HttpClient,
		BaseURL:    DefaultBaseURL,
	}

	s = &Session{
		opts: def.MergeWith(opts),
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
				var t, err = s.opts.OAuth2Config.TokenSource(context.Background(), s.opts.Token).Token()
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
		s.opts.HttpClient, s.opts.BaseURL,
		connect.WithHTTPGet(),
		connect.WithInterceptors(connect.UnaryInterceptorFunc(interceptor)),
	)

	s.SecuritiesClient = portfoliov1connect.NewSecuritiesServiceClient(
		s.opts.HttpClient, s.opts.BaseURL,
		connect.WithHTTPGet(),
		connect.WithInterceptors(connect.UnaryInterceptorFunc(interceptor)),
	)
}

func FromContext(ctx context.Context) (s *Session) {
	s = ctx.Value(SessionKey).(*Session)
	return
}

// InjectSession is a pre-hook that injects the session into the context.
func InjectSession(ctx context.Context, cmd *cli.Command) (newCtx context.Context, err error) {
	if cmd.NArg() != 0 {
		s, err := ContinueSession()
		if err != nil {
			fmt.Println("Could not continue with existing session or session is missing. Please use `mgo login`.")
			return ctx, err
		}

		newCtx = context.WithValue(ctx, SessionKey, s)
	} else {
		newCtx = ctx
	}

	return
}
