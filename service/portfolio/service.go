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
	"time"

	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const DefaultSecuritiesServiceURL = "http://localhost:8080"

// service is the main struct fo the [PortfolioService] implementation.
type service struct {
	// a simple portfolio for testing, will be replaced by database later
	portfolio *portfoliov1.Portfolio
	//portfolios persistence.StorageOperations[*portfoliov1.Portfolio]

	portfoliov1connect.UnimplementedPortfolioServiceHandler

	securities portfoliov1connect.SecuritiesServiceClient
}

type Options struct {
	Securities portfoliov1connect.SecuritiesServiceClient
}

func NewService(opts Options) portfoliov1connect.PortfolioServiceHandler {
	var s service

	s.portfolio = &portfoliov1.Portfolio{
		Name: "My Portfolio",
		Events: []*portfoliov1.PortfolioEvent{
			{
				EventOneof: &portfoliov1.PortfolioEvent_Buy{
					Buy: &portfoliov1.BuySecurityTransaction{
						SecurityName: "US0378331005",
						Amount:       20,
						Price:        107.08,
						Fees:         10.25,
						Time:         timestamppb.New(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
					},
				},
			},
			{
				EventOneof: &portfoliov1.PortfolioEvent_Sell{
					Sell: &portfoliov1.SellSecurityTransaction{
						SecurityName: "US0378331005",
						Amount:       10,
						Price:        145.88,
						Fees:         8.55,
						Time:         timestamppb.New(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)),
					},
				},
			},
		},
	}
	s.securities = opts.Securities
	if s.securities == nil {
		s.securities = portfoliov1connect.NewSecuritiesServiceClient(http.DefaultClient, DefaultSecuritiesServiceURL)
	}

	return &s
}
