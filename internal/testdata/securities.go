package testdata

import (
	"database/sql"

	"github.com/oxisto/money-gopher/internal/testing/quotetest"
	"github.com/oxisto/money-gopher/persistence"
)

// TestSecurity is a test security.
var TestSecurity = &persistence.Security{
	ID:            "DE1234567890",
	DisplayName:   "My Security",
	QuoteProvider: sql.NullString{String: quotetest.QuoteProviderStatic, Valid: true},
}

// TestListedSecurity is a listed security for [TestSecurity] that has a ticker
// "TICK" and currency "USD".
var TestListedSecurity = &persistence.ListedSecurity{
	SecurityID: TestSecurity.ID,
	Ticker:     "TICK",
	Currency:   "USD",
}

// TestCreateSecurityParams is a test security creation parameter.
var TestCreateSecurityParams = persistence.CreateSecurityParams{
	ID:            TestSecurity.ID,
	DisplayName:   TestSecurity.DisplayName,
	QuoteProvider: TestSecurity.QuoteProvider,
}

// TestUpsertListedSecurityParams is a test listed security upsert parameter.
var TestUpsertListedSecurityParams = persistence.UpsertListedSecurityParams{
	SecurityID: TestSecurity.ID,
	Ticker:     TestListedSecurity.Ticker,
	Currency:   TestListedSecurity.Currency,
}
