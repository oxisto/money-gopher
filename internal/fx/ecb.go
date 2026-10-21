// Package fx keeps the euro reference rates of the European Central Bank
// and converts amounts between currencies with them. A listing quoted in
// one currency is valued in the currency its portfolio settles in; the
// bridge is always the euro, which is what the ECB publishes against.
package fx

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/oxisto/money-gopher/internal/persistence"
)

// Rate is the ECB reference rate of one currency on one day: how many units
// of Currency one euro bought.
type Rate struct {
	Currency string
	// Time is the reference date at midnight UTC.
	Time time.Time
	Rate persistence.Decimal
}

// Source provides reference rates. With history it returns everything it
// has back to the start; otherwise the recent weeks suffice.
type Source interface {
	Rates(ctx context.Context, history bool) ([]Rate, error)
}

// ECB fetches the rates from the ECB's own feed: a 90-day file for keeping
// up, and the full history (back to 1999) for a currency seen for the first
// time, so that past snapshots can be valued as well.
type ECB struct {
	BaseURL string
	Client  *http.Client
}

// NewECB returns the public feed.
func NewECB() *ECB {
	return &ECB{BaseURL: "https://www.ecb.europa.eu/stats/eurofxref", Client: http.DefaultClient}
}

// Rates implements Source.
func (e *ECB) Rates(ctx context.Context, history bool) ([]Rate, error) {
	name := "eurofxref-hist-90d.xml"
	if history {
		name = "eurofxref-hist.xml"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.BaseURL+"/"+name, nil)
	if err != nil {
		return nil, err
	}
	res, err := e.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch %s: %w", name, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %s failed with status %s", name, res.Status)
	}

	return ParseECB(res.Body)
}

// ecbEnvelope is the shape of the feed: one Cube per day, one Cube per
// currency inside it. The namespaces are ignored on purpose.
type ecbEnvelope struct {
	Days []struct {
		Time  string `xml:"time,attr"`
		Rates []struct {
			Currency string `xml:"currency,attr"`
			Rate     string `xml:"rate,attr"`
		} `xml:"Cube"`
	} `xml:"Cube>Cube"`
}

// ParseECB reads the ECB's XML format (daily, 90-day and full history
// share it).
func ParseECB(r io.Reader) ([]Rate, error) {
	var env ecbEnvelope
	if err := xml.NewDecoder(r).Decode(&env); err != nil {
		return nil, fmt.Errorf("could not decode ECB rates: %w", err)
	}

	var rates []Rate
	for _, day := range env.Days {
		t, err := time.Parse("2006-01-02", day.Time)
		if err != nil {
			return nil, fmt.Errorf("invalid reference date %q: %w", day.Time, err)
		}
		for _, r := range day.Rates {
			rate, err := persistence.NewDecimal(r.Rate)
			if err != nil {
				return nil, fmt.Errorf("invalid rate %q for %s on %s: %w", r.Rate, r.Currency, day.Time, err)
			}
			rates = append(rates, Rate{Currency: r.Currency, Time: t.UTC(), Rate: rate})
		}
	}

	return rates, nil
}
