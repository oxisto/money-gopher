package testdata

import (
	"database/sql"

	"github.com/oxisto/money-gopher/persistence"
)

// TestSecurity is a test security.
var TestSecurity = &persistence.Security{
	ID:            "DE1234567890",
	DisplayName:   "My Security",
	QuoteProvider: sql.NullString{String: "mock", Valid: true},
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
