package quote

import (
	"context"
	"time"

	"github.com/oxisto/money-gopher/currency"
	"github.com/oxisto/money-gopher/persistence"
)

const QuoteProviderMock = "mock"

type mockQP struct {
}

func (m *mockQP) LatestQuote(ctx context.Context, ls *persistence.ListedSecurity) (quote *currency.Currency, t time.Time, err error) {
	return currency.Value(100), time.Now(), nil
}

type mockQuoteProvider struct{}

func (mockQuoteProvider) LatestQuote(_ context.Context, _ *persistence.ListedSecurity) (quote *currency.Currency, t time.Time, err error) {
	return currency.Value(100), time.Date(1, 0, 0, 0, 0, 0, 0, time.UTC), nil
}

func init() {
	RegisterQuoteProvider(QuoteProviderMock, &mockQP{})
}
