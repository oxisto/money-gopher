package quotes

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"
)

// ProviderYahoo is the name of the Yahoo Finance provider.
const ProviderYahoo = "yf"

// Yahoo fetches quotes from the (unofficial) Yahoo Finance chart API, going
// by the listing's ticker.
type Yahoo struct {
	Client *http.Client
	// BaseURL of the API, overridable for tests.
	BaseURL string
}

// NewYahoo returns the Yahoo provider with production defaults.
func NewYahoo() *Yahoo {
	return &Yahoo{
		Client:  &http.Client{Timeout: 30 * time.Second},
		BaseURL: "https://query1.finance.yahoo.com",
	}
}

func (y *Yahoo) Name() string { return ProviderYahoo }

// chart is the part of the chart API response we care about.
type chart struct {
	Chart struct {
		Result []struct {
			Meta struct {
				RegularMarketPrice float64 `json:"regularMarketPrice"`
				RegularMarketTime  int64   `json:"regularMarketTime"`
			} `json:"meta"`
		} `json:"result"`
	} `json:"chart"`
}

func (y *Yahoo) LatestQuote(ctx context.Context, ins Instrument) (Quote, error) {
	url := fmt.Sprintf("%s/v8/finance/chart/%s?interval=1d&range=1d", y.BaseURL, ins.Ticker)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Quote{}, err
	}
	// Yahoo rejects requests without a browser-ish user agent.
	req.Header.Set("User-Agent", "Mozilla/5.0 (money-gopher)")

	res, err := y.Client.Do(req)
	if err != nil {
		return Quote{}, fmt.Errorf("could not fetch quote: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Quote{}, fmt.Errorf("quote request failed with status %s", res.Status)
	}

	var ch chart
	if err := json.NewDecoder(res.Body).Decode(&ch); err != nil {
		return Quote{}, fmt.Errorf("could not decode JSON: %w", err)
	}

	if len(ch.Chart.Result) == 0 {
		return Quote{}, ErrEmptyResult
	}

	meta := ch.Chart.Result[0].Meta
	return Quote{
		Price: int64(math.Round(meta.RegularMarketPrice * 100)),
		Time:  time.Unix(meta.RegularMarketTime, 0).UTC(),
	}, nil
}
