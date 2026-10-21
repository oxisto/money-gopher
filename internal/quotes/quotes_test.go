package quotes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestYahoo_LatestQuote(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v8/finance/chart/VWCE.DE" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("expected a User-Agent header")
		}
		w.Write([]byte(`{"chart":{"result":[{"meta":{
			"regularMarketPrice": 123.45,
			"regularMarketTime": 1767225600
		}}]}}`))
	}))
	defer srv.Close()

	y := NewYahoo()
	y.BaseURL = srv.URL

	quote, err := y.LatestQuote(context.Background(), Instrument{Ticker: "VWCE.DE"})
	if err != nil {
		t.Fatalf("LatestQuote() error = %v", err)
	}
	if quote.Price.StringPrec(2) != "123.45" {
		t.Errorf("price = %s, want 123.45", quote.Price.StringPrec(2))
	}
	if want := time.Unix(1767225600, 0).UTC(); !quote.Time.Equal(want) {
		t.Errorf("time = %v, want %v", quote.Time, want)
	}

	// An empty chart result is an error, not a zero quote.
	empty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"chart":{"result":[]}}`))
	}))
	defer empty.Close()
	y.BaseURL = empty.URL

	if _, err := y.LatestQuote(context.Background(), Instrument{Ticker: "VWCE.DE"}); err == nil {
		t.Error("expected an error for an empty result")
	}
}

func TestING_LatestQuote(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/instrument-header" {
			http.NotFound(w, r)
			return
		}
		if got := r.URL.Query().Get("isinOrSearchTerm"); got != "IE00BK5BQT80" {
			t.Errorf("isinOrSearchTerm = %q, want IE00BK5BQT80", got)
		}
		// Bid/ask are entirely absent for instruments without a spread (many
		// ETFs) — price/priceChangeDate is the only field relied on.
		w.Write([]byte(`{
			"price": 111.11, "priceChangeDate": "2026-07-03T15:30:00Z"
		}`))
	}))
	defer srv.Close()

	i := NewING()
	i.BaseURL = srv.URL

	quote, err := i.LatestQuote(context.Background(), Instrument{ISIN: "IE00BK5BQT80"})
	if err != nil {
		t.Fatalf("LatestQuote() error = %v", err)
	}
	if quote.Price.StringPrec(2) != "111.11" {
		t.Errorf("price = %s, want 111.11", quote.Price.StringPrec(2))
	}

	// Without an ISIN the provider must refuse instead of querying nonsense.
	if _, err := i.LatestQuote(context.Background(), Instrument{Ticker: "VWCE"}); err == nil {
		t.Error("expected an error without an ISIN")
	}
}
