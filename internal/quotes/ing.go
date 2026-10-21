package quotes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/oxisto/money-gopher/internal/persistence"
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

// instrumentHeader is the part of the API response we care about. Bid/ask
// aren't included: the API omits those keys entirely for instruments that
// don't trade with a spread (e.g. many ETFs), and price/priceChangeDate is
// present uniformly across stocks, bonds and ETFs, matching what ING's own
// site shows as "the" price.
type instrumentHeader struct {
	Price            float64   `json:"price"`
	PriceChangedDate time.Time `json:"priceChangeDate"`
}

func (i *ING) LatestQuote(ctx context.Context, ins Instrument) (Quote, error) {
	if ins.ISIN == "" {
		return Quote{}, fmt.Errorf("the ING provider needs an ISIN identifier")
	}

	// As of 2026, this is a query-string endpoint (isinOrSearchTerm), not a
	// path segment — ING moved it from the old
	// /api/v1/components/instrumentheader/{isin} some time after 2024,
	// which now 404s unconditionally, even for highly liquid instruments.
	u := fmt.Sprintf("%s/api/v1/instrument-header?%s", i.BaseURL, url.Values{
		"isinOrSearchTerm": {ins.ISIN},
		"isKnownIsin":      {"true"},
	}.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
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

	return Quote{Price: persistence.DecimalFromFloat(h.Price), Time: h.PriceChangedDate.UTC()}, nil
}
