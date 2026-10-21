package fx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/oxisto/money-gopher/internal/persistence"
)

// ecbXML is the shape of every ECB file: the full history has many days,
// the daily one a single day. Namespaces as in the real feed.
const ecbXML = `<?xml version="1.0" encoding="UTF-8"?>
<gesmes:Envelope xmlns:gesmes="http://www.gesmes.org/xml/2002-08-01" xmlns="http://www.ecb.int/vocabulary/2002-08-01/eurofxref">
	<gesmes:subject>Reference rates</gesmes:subject>
	<gesmes:Sender><gesmes:name>European Central Bank</gesmes:name></gesmes:Sender>
	<Cube>
		<Cube time="2026-10-21">
			<Cube currency="USD" rate="1.1578"/>
			<Cube currency="JPY" rate="175.12"/>
			<Cube currency="GBP" rate="0.86950"/>
		</Cube>
		<Cube time="2026-10-20">
			<Cube currency="USD" rate="1.1601"/>
			<Cube currency="JPY" rate="174.80"/>
			<Cube currency="GBP" rate="0.86880"/>
		</Cube>
	</Cube>
</gesmes:Envelope>`

func TestParseECB(t *testing.T) {
	rates, err := ParseECB(strings.NewReader(ecbXML))
	if err != nil {
		t.Fatalf("ParseECB() error = %v", err)
	}
	if len(rates) != 6 {
		t.Fatalf("len(rates) = %d, want 6", len(rates))
	}

	usd := rates[0]
	if usd.Currency != "USD" || usd.Rate.StringPrec(4) != "1.1578" {
		t.Errorf("rates[0] = %s %s, want USD 1.1578", usd.Currency, usd.Rate)
	}
	if want := time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC); !usd.Time.Equal(want) {
		t.Errorf("rates[0].Time = %v, want %v", usd.Time, want)
	}
}

// fakeECB serves the fixture under both file names and records which one
// was asked for.
func fakeECB(t *testing.T) (*ECB, *[]string) {
	t.Helper()
	var requested []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = append(requested, strings.TrimPrefix(r.URL.Path, "/"))
		_, _ = w.Write([]byte(ecbXML))
	}))
	t.Cleanup(srv.Close)
	return &ECB{BaseURL: srv.URL, Client: srv.Client()}, &requested
}

func TestUpdater(t *testing.T) {
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	source, requested := fakeECB(t)
	u := &Updater{DB: db, Source: source}

	// Nothing needs converting yet: no fetch at all.
	if n, err := u.UpdateAll(ctx); err != nil || n != 0 {
		t.Fatalf("UpdateAll() with no listings = %d, %v; want 0, nil", n, err)
	}
	if len(*requested) != 0 {
		t.Fatalf("fetched %v for an empty database", *requested)
	}

	// A USD listing appears: the first fetch is the full history, and only
	// USD is kept — JPY and GBP are of no use.
	if _, err := db.CreateSecurity(ctx, persistence.CreateSecurityParams{ID: "tsy", DisplayName: "4,125 % United States of America"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateListing(ctx, persistence.CreateListingParams{ID: "l1", SecurityID: "tsy", Ticker: "US91282CFV81", Currency: "USD"}); err != nil {
		t.Fatal(err)
	}
	if n, err := u.UpdateAll(ctx); err != nil || n != 2 {
		t.Fatalf("UpdateAll() = %d, %v; want 2 USD rates stored", n, err)
	}
	if got := *requested; len(got) != 1 || got[0] != "eurofxref-hist.xml" {
		t.Errorf("first fetch requested %v, want the full history", got)
	}
	if n, _ := db.CountFxRates(ctx, "JPY"); n != 0 {
		t.Errorf("stored %d JPY rates, want none", n)
	}

	// From then on the recent file does, and re-storing is idempotent.
	if n, err := u.UpdateAll(ctx); err != nil || n != 2 {
		t.Fatalf("second UpdateAll() = %d, %v; want 2, nil", n, err)
	}
	if got := *requested; got[len(got)-1] != "eurofxref-hist-90d.xml" {
		t.Errorf("second fetch requested %v, want the 90-day file", got)
	}
	if n, _ := db.CountFxRates(ctx, "USD"); n != 2 {
		t.Errorf("USD rates = %d, want 2 (upsert, not duplicates)", n)
	}
}

func TestConvert(t *testing.T) {
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	day := func(d int) time.Time { return time.Date(2026, 10, d, 0, 0, 0, 0, time.UTC) }
	for _, r := range []struct {
		currency string
		day      int
		rate     string
	}{
		{"USD", 20, "1.1601"}, {"USD", 21, "1.1578"}, {"GBP", 21, "0.8695"},
	} {
		rate, _ := persistence.NewDecimal(r.rate)
		if err := db.CreateFxRate(ctx, persistence.CreateFxRateParams{Currency: r.currency, Time: day(r.day), Rate: rate}); err != nil {
			t.Fatal(err)
		}
	}
	amount, _ := persistence.NewDecimal("95.552")

	tests := []struct {
		name     string
		from, to string
		at       time.Time
		want     string
		wantErr  bool
	}{
		{"same currency is untouched", "USD", "USD", day(21), "95.55", false},
		{"into euros at the day's rate", "USD", "EUR", day(21).Add(15 * time.Hour), "82.53", false},
		{"an earlier time uses the rate in force then", "USD", "EUR", day(20), "82.37", false},
		{"out of euros", "EUR", "USD", day(21), "110.63", false},
		{"between two foreign currencies via the euro", "USD", "GBP", day(21), "71.76", false},
		{"before the first known rate", "USD", "EUR", day(1), "", true},
		{"unknown currency", "CHF", "EUR", day(21), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convert(ctx, db, amount, tt.from, tt.to, tt.at)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Convert() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && got.StringPrec(2) != tt.want {
				t.Errorf("Convert() = %s, want %s", got.StringPrec(2), tt.want)
			}
		})
	}
}
