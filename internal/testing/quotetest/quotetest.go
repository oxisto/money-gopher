package quotetest

import (
	"context"
	"time"

	"github.com/oxisto/money-gopher/currency"
	"github.com/oxisto/money-gopher/persistence"
)

const QuoteProviderStatic = "static"

type StaticQuoteProvider struct {
	Quote *currency.Currency
}

// NewStaticQuoteProvider creates a new static quote provider that always returns the same quote.
func NewStaticQuoteProvider(quote *currency.Currency) *StaticQuoteProvider {
	return &StaticQuoteProvider{Quote: quote}
}

func (p *StaticQuoteProvider) LatestQuote(ctx context.Context, ls *persistence.ListedSecurity) (quote *currency.Currency, t time.Time, err error) {
	return p.Quote, time.Now(), nil
}
