// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package models

import (
	"time"

	"github.com/oxisto/money-gopher/currency"
	"github.com/oxisto/money-gopher/persistence"
	"github.com/oxisto/money-gopher/portfolio/accounts"
	"github.com/oxisto/money-gopher/portfolio/events"
)

type AccountInput struct {
	ID          string               `json:"id"`
	DisplayName string               `json:"displayName"`
	Type        accounts.AccountType `json:"type"`
}

type CurrencyInput struct {
	Amount int    `json:"amount"`
	Symbol string `json:"symbol"`
}

type ListedSecurityInput struct {
	Ticker   string `json:"ticker"`
	Currency string `json:"currency"`
}

type Mutation struct {
}

type PortfolioEvent struct {
	Time     time.Time                 `json:"time"`
	Type     events.PortfolioEventType `json:"type"`
	Security *persistence.Security     `json:"security,omitempty"`
}

type PortfolioInput struct {
	ID          string   `json:"id"`
	DisplayName string   `json:"displayName"`
	AccountIds  []string `json:"accountIds"`
}

type PortfolioPosition struct {
	Security *persistence.Security `json:"security"`
	Amount   float64               `json:"amount"`
	// PurchaseValue was the market value of this position when it was bought (net;
	// exclusive of any fees).
	PurchaseValue *currency.Currency `json:"purchaseValue"`
	// PurchasePrice was the market price of this position when it was bought (net;
	// exclusive of any fees).
	PurchasePrice *currency.Currency `json:"purchasePrice"`
	// MarketValue is the current market value of this position, as retrieved from
	// the securities service.
	MarketValue *currency.Currency `json:"marketValue"`
	// MarketPrice is the current market price of this position, as retrieved from
	// the securities service.
	MarketPrice *currency.Currency `json:"marketPrice"`
	// TotalFees is the total amount of fees accumulating in this position through
	// various transactions.
	TotalFees *currency.Currency `json:"totalFees"`
	// ProfitOrLoss contains the absolute amount of profit or loss in this position.
	ProfitOrLoss *currency.Currency `json:"profitOrLoss"`
	// Gains contains the relative amount of profit or loss in this position.
	Gains float64 `json:"gains"`
}

type PortfolioSnapshot struct {
	Time                 time.Time            `json:"time"`
	Positions            []*PortfolioPosition `json:"positions"`
	FirstTransactionTime time.Time            `json:"firstTransactionTime"`
	TotalPurchaseValue   *currency.Currency   `json:"totalPurchaseValue"`
	TotalMarketValue     *currency.Currency   `json:"totalMarketValue"`
	TotalProfitOrLoss    *currency.Currency   `json:"totalProfitOrLoss"`
	TotalGains           float64              `json:"totalGains"`
	TotalPortfolioValue  *currency.Currency   `json:"totalPortfolioValue,omitempty"`
	Cash                 *currency.Currency   `json:"cash"`
}

type Query struct {
}

type SecurityInput struct {
	ID          string                 `json:"id"`
	DisplayName string                 `json:"displayName"`
	ListedAs    []*ListedSecurityInput `json:"listedAs,omitempty"`
}

type TransactionInput struct {
	Time                 time.Time                 `json:"time"`
	SourceAccountID      string                    `json:"sourceAccountID"`
	DestinationAccountID string                    `json:"destinationAccountID"`
	SecurityID           string                    `json:"securityID"`
	Amount               float64                   `json:"amount"`
	Price                *CurrencyInput            `json:"price"`
	Fees                 *CurrencyInput            `json:"fees"`
	Taxes                *CurrencyInput            `json:"taxes"`
	Type                 events.PortfolioEventType `json:"type"`
}
