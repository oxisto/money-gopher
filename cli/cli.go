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
	"io"
	"net/http"
	"os"

	"github.com/shurcooL/graphql"

	oauth2 "github.com/oxisto/oauth2go"
	"github.com/urfave/cli/v3"
)

type sessionKeyType struct{}

// SessionKey is the key for the session in the context.
var SessionKey sessionKeyType

// Session holds all necessary information about the current CLI session.
type Session struct {
	GraphQL *graphql.Client

	opts *SessionOptions
}

// SessionOptions holds all options to configure a [Session].
type SessionOptions struct {
	OAuth2Config *oauth2.Config
	HttpClient   *http.Client `json:"-"`
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

// NewSession creates a new session.
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

// ContinueSession continues a session from a file.
func ContinueSession() (s *Session, err error) {
	var (
		file *os.File
	)

	file, err = os.OpenFile("session.json", os.O_RDONLY, 0600)
	if err != nil {
		return
	}

	s = new(Session)
	err = json.NewDecoder(file).Decode(&s.opts)
	if err != nil {
		return nil, fmt.Errorf("could not parse session file: %w", err)
	}

	s.initClients()

	return
}

// Save saves the session to a file.
func (s *Session) Save() (err error) {
	var (
		file *os.File
	)

	file, err = os.OpenFile("session.json", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}

	// We don't want to save the clients, so we only save the options.
	err = json.NewEncoder(file).Encode(s.opts)
	if err != nil {
		return fmt.Errorf("could not save session file: %w", err)
	}

	return nil
}

// initClients initializes the clients for the session.
func (s *Session) initClients() {
	if s.opts.HttpClient == nil {
		s.opts.HttpClient = http.DefaultClient
	}

	s.GraphQL = graphql.NewClient(s.opts.BaseURL+"/graphql/query", s.opts.HttpClient)
}

func (s *Session) WriteJSON(w io.Writer, v any) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

// FromContext extracts the session from the context.
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
