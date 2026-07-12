package quotes

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"
)

// ProviderING is the name of the ING provider.
const ProviderING = "ing"

// ING fetches quotes from the instrument header API of ING Germany's
// securities site, going by the security's ISIN.
type ING struct {
	Client *http.Client
	// BaseURL of the API, overridable for tests.
	BaseURL string
}

// NewING returns the ING provider with production defaults.
func NewING() *ING {
	return &ING{
		Client:  &http.Client{Timeout: 30 * time.Second},
		BaseURL: "https://component-api.wertpapiere.ing.de",
	}
}

func (i *ING) Name() string { return ProviderING }

// instrumentHeader is the part of the API response we care about.
type instrumentHeader struct {
	Ask              float64   `json:"ask"`
	AskDate          time.Time `json:"askDate"`
	Bid              float64   `json:"bid"`
	BidDate          time.Time `json:"bidDate"`
	HasBidAsk        bool      `json:"hasBidAsk"`
	Price            float64   `json:"price"`
	PriceChangedDate time.Time `json:"priceChangeDate"`
}

func (i *ING) LatestQuote(ctx context.Context, ins Instrument) (Quote, error) {
	if ins.ISIN == "" {
		return Quote{}, fmt.Errorf("the ING provider needs an ISIN identifier")
	}

	url := fmt.Sprintf("%s/api/v1/components/instrumentheader/%s", i.BaseURL, ins.ISIN)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Quote{}, err
	}

	res, err := i.Client.Do(req)
	if err != nil {
		return Quote{}, fmt.Errorf("could not fetch quote: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Quote{}, fmt.Errorf("quote request failed with status %s", res.Status)
	}

	var h instrumentHeader
	if err := json.NewDecoder(res.Body).Decode(&h); err != nil {
		return Quote{}, fmt.Errorf("could not decode JSON: %w", err)
	}

	// Prefer the bid when the instrument trades with bid/ask.
	if h.HasBidAsk {
		return Quote{Price: int64(math.Round(h.Bid * 100)), Time: h.BidDate.UTC()}, nil
	}
	return Quote{Price: int64(math.Round(h.Price * 100)), Time: h.PriceChangedDate.UTC()}, nil
}
