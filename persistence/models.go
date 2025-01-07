// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package persistence

import (
	"database/sql"
	"time"

	currency "github.com/oxisto/money-gopher/currency"
	"github.com/oxisto/money-gopher/portfolio/accounts"
	"github.com/oxisto/money-gopher/portfolio/events"
)

// Accounts represents an account, such as a brokerage account or a bank
type Account struct {
	// ID is the primary identifier for a brokerage account.
	ID string
	// DisplayName is the human-readable name of the brokerage account.
	DisplayName string
	// Type is the type of the account.
	Type accounts.AccountType
	// ReferenceAccountID is the ID of the account that this account is related to. For example, if this is a brokerage account, the reference account could be a bank account.
	ReferenceAccountID sql.NullInt64
}

type BankAccount struct {
	// ID is the primary identifier for a bank account.
	ID string
	// DisplayName is the human-readable name of the bank account.
	DisplayName string
}

// ListedSecurity represents a security that is listed on a particular exchange.
type ListedSecurity struct {
	// SecurityID is the ID of the security.
	SecurityID string
	// Ticker is the symbol used to identify the security on the exchange.
	Ticker string
	// Currency is the currency in which the security is traded.
	Currency string
	// LatestQuote is the latest quote for the security as a [currency.Currency].
	LatestQuote *currency.Currency
	// LatestQuoteTimestamp is the timestamp of the latest quote.
	LatestQuoteTimestamp sql.NullTime
}

// Portfolios represent a collection of securities and other positions
type Portfolio struct {
	// ID is the primary identifier for a portfolio.
	ID string
	// DisplayName is the human-readable name of the portfolio.
	DisplayName string
	// BankAccountID is the ID of the bank account that holds the portfolio.
	BankAccountID string
}

// PortfolioAccounts represents the relationship between portfolios and accounts.
type PortfolioAccount struct {
	PortfolioID string
	AccountID   string
}

type PortfolioEvent struct {
	ID          string
	Type        events.PortfolioEventType
	Time        time.Time
	PortfolioID string
	SecurityID  string
	Amount      float64
	Price       *currency.Currency
	Fees        *currency.Currency
	Taxes       *currency.Currency
}

// Security represents a security that can be traded on an exchange.
type Security struct {
	// ID is the primary identifier for a security.
	ID string
	// DisplayName is the human-readable name of the security.
	DisplayName string
	// QuoteProvider is the name of the provider that provides quotes for this security.
	QuoteProvider sql.NullString
}

// Transactions represents a transaction in an account.
type Transaction struct {
	// ID is the primary identifier for a transaction.
	ID string
	// SourceAccountID is the ID of the account that the transaction originated from.
	SourceAccountID sql.NullString
	// DestinationAccountID is the ID of the account that the transaction is destined for.
	DestinationAccountID sql.NullString
	// Time is the time that the transaction occurred.
	Time time.Time
	// Type is the type of the transaction. Depending on the type, different fields (source, destination) will be used.
	Type events.PortfolioEventType
	// SecurityID is the ID of the security that the transaction is related to. Can be empty if the transaction is not related to a security.
	SecurityID sql.NullString
	// Amount is the amount of the transaction.
	Amount float64
	// Price is the price of the transaction.
	Price *currency.Currency
	// Fees is the fees of the transaction.
	Fees *currency.Currency
	// Taxes is the taxes of the transaction.
	Taxes *currency.Currency
}
