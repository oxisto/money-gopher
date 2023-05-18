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

// package portfolio contains the code for the PortfolioService implementation.
package portfolio

import (
	"net/http"

	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/persistence"
)

const DefaultSecuritiesServiceURL = "http://localhost:8080"

// service is the main struct fo the [PortfolioService] implementation.
type service struct {
	// a simple portfolio for testing, will be replaced by database later
	portfolio  *portfoliov1.Portfolio
	portfolios persistence.StorageOperations[*portfoliov1.Portfolio]
	events     persistence.StorageOperations[*portfoliov1.PortfolioEvent]
	securities portfoliov1connect.SecuritiesServiceClient

	portfoliov1connect.UnimplementedPortfolioServiceHandler
}

type Options struct {
	SecuritiesClient portfoliov1connect.SecuritiesServiceClient
	DB               *persistence.DB
}

func NewService(opts Options) portfoliov1connect.PortfolioServiceHandler {
	var s service

	s.portfolios = persistence.Ops[*portfoliov1.Portfolio](opts.DB)
	s.events = persistence.Relationship[*portfoliov1.PortfolioEvent](s.portfolios)

	s.securities = opts.SecuritiesClient
	if s.securities == nil {
		s.securities = portfoliov1connect.NewSecuritiesServiceClient(http.DefaultClient, DefaultSecuritiesServiceURL)
	}

	return &s
}
