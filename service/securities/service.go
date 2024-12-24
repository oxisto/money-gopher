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

// package securities contains the code for the SecuritiesService implementation.
package securities

import (
	"time"

	moneygopher "github.com/oxisto/money-gopher"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/persistence"

	"golang.org/x/text/currency"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type service struct {
	securities       persistence.StorageOperations[*portfoliov1.Security]
	listedSecurities persistence.StorageOperations[*portfoliov1.ListedSecurity]

	portfoliov1connect.UnimplementedSecuritiesServiceHandler
}

func NewService(db *persistence.DB) portfoliov1connect.SecuritiesServiceHandler {
	securities := persistence.Ops[*portfoliov1.Security](db)
	listedSecurities := persistence.Relationship[*portfoliov1.ListedSecurity](securities)
	secs := []*portfoliov1.Security{
		{
			Id:          "US0378331005",
			DisplayName: "Apple Inc.",
			ListedOn: []*portfoliov1.ListedSecurity{
				{
					SecurityId:           "US0378331005",
					Ticker:               "APC.F",
					Currency:             currency.EUR.String(),
					LatestQuote:          portfoliov1.Value(15016),
					LatestQuoteTimestamp: timestamppb.New(time.Date(2023, 4, 21, 0, 0, 0, 0, time.Local)),
				},
				{
					SecurityId:           "US0378331005",
					Ticker:               "AAPL",
					Currency:             currency.USD.String(),
					LatestQuote:          portfoliov1.Value(16502),
					LatestQuoteTimestamp: timestamppb.New(time.Date(2023, 4, 21, 0, 0, 0, 0, time.Local)),
				},
			},
			QuoteProvider: moneygopher.Ref(QuoteProviderYF),
		},
	}
	for _, sec := range secs {
		securities.Replace(sec)

		// TODO: in the future, we might do this automatically
		for _, ls := range sec.ListedOn {
			listedSecurities.Replace(ls)
		}
	}

	return &service{
		securities:       securities,
		listedSecurities: listedSecurities,
	}
}
