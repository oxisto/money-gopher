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
	"errors"
	"fmt"
	"net/http"
	"time"

	portfoliov1 "github.com/oxisto/money-gopher/gen"
)

var ErrEmptyResult = errors.New("empty result")

type yf struct {
	http.Client
}

type chart struct {
	Chart struct {
		Results []struct {
			Meta struct {
				RegularMarketPrice float32 `json:"regularMarketPrice"`
				RegularMarketTime  int64   `json:"regularMarketTime"`
			} `json:"meta"`
		} `json:"result"`
	} `json:"chart"`
}

func (yf *yf) LatestQuote(ctx context.Context, ls *portfoliov1.ListedSecurity) (quote float32, t time.Time, err error) {
	var (
		res *http.Response
		ch  chart
	)

	res, err = yf.Get(fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=1mo", ls.Ticker))
	if err != nil {
		return 0, t, fmt.Errorf("could not fetch quote: %w", err)
	}

	err = json.NewDecoder(res.Body).Decode(&ch)
	if err != nil {
		return 0, t, fmt.Errorf("could not decode JSON: %w", err)
	}

	if len(ch.Chart.Results) == 0 {
		return 0, t, ErrEmptyResult
	}

	return ch.Chart.Results[0].Meta.RegularMarketPrice,
		time.Unix(ch.Chart.Results[0].Meta.RegularMarketTime, 0), nil
}
