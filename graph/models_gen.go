// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package graph

import (
	"github.com/oxisto/money-gopher/persistence"
)

type ListedSecurityInput struct {
	Ticker   string `json:"ticker"`
	Currency string `json:"currency"`
}

type Mutation struct {
}

type PortfolioPosition struct {
	Security *persistence.Security `json:"security"`
	Quantity int                   `json:"quantity"`
	// PurchaseValue was the market value of this position when it was bought (net;
	// exclusive of any fees).
	PurchaseValue *persistence.Currency `json:"purchaseValue"`
	// PurchasePrice was the market price of this position when it was bought (net;
	// exclusive of any fees).
	PurchasePrice *persistence.Currency `json:"purchasePrice"`
	// MarketValue is the current market value of this position, as retrieved from
	// the securities service.
	MarketValue *persistence.Currency `json:"marketValue"`
	// MarketPrice is the current market price of this position, as retrieved from
	// the securities service.
	MarketPrice *persistence.Currency `json:"marketPrice"`
	// TotalFees is the total amount of fees accumulating in this position through
	// various transactions.
	TotalFees *persistence.Currency `json:"totalFees"`
	// ProfitOrLoss contains the absolute amount of profit or loss in this position.
	ProfitOrLoss *persistence.Currency `json:"profitOrLoss"`
	// Gains contains the relative amount of profit or loss in this position.
	Gains float64 `json:"gains"`
}

type PortfolioSnapshot struct {
	Time     string               `json:"time"`
	Position []*PortfolioPosition `json:"position"`
}

type Query struct {
}

type SecurityInput struct {
	ID          string                 `json:"id"`
	DisplayName string                 `json:"displayName"`
	ListedAs    []*ListedSecurityInput `json:"listedAs,omitempty"`
}
