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
	"net/http"

	"connectrpc.com/connect"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
)

// Session holds all necessary information about the current CLI session.
type Session struct {
	PortfolioClient portfoliov1connect.PortfolioServiceClient
}

func NewSession() *Session {
	var s Session

	s.PortfolioClient = portfoliov1connect.NewPortfolioServiceClient(
		http.DefaultClient, "http://localhost:8080",
		connect.WithHTTPGet(),
	)

	return &s
}
