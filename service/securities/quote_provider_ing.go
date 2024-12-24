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

package securities

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	portfoliov1 "github.com/oxisto/money-gopher/gen"
)

const QuoteProviderING = "ing"

type ing struct {
	http.Client
}

type header struct {
	Ask              float32   `json:"ask"`
	AskDate          time.Time `json:"askDate"`
	Bid              float32   `json:"bid"`
	BidDate          time.Time `json:"bidDate"`
	Currency         string    `json:"currency"`
	ISIN             string    `json:"isin"`
	HasBidAsk        bool      `json:"hasBidAsk"`
	Price            float32   `json:"price"`
	PriceChangedDate time.Time `json:"priceChangeDate"`
	WKN              string    `json:"wkn"`
}

func (ing *ing) LatestQuote(ctx context.Context, ls *portfoliov1.ListedSecurity) (quote *portfoliov1.Currency, t time.Time, err error) {
	var (
		res *http.Response
		h   header
	)

	res, err = ing.Get(fmt.Sprintf("https://component-api.wertpapiere.ing.de/api/v1/components/instrumentheader/%s", ls.SecurityId))
	if err != nil {
		return nil, t, fmt.Errorf("could not fetch quote: %w", err)
	}

	err = json.NewDecoder(res.Body).Decode(&h)
	if err != nil {
		return nil, t, fmt.Errorf("could not decode JSON: %w", err)
	}

	if h.HasBidAsk {
		return portfoliov1.Value(int32(h.Bid * 100)), h.BidDate, nil
	} else {
		return portfoliov1.Value(int32(h.Price * 100)), h.PriceChangedDate, nil
	}
}
