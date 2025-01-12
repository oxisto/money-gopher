package events

import "github.com/oxisto/money-gopher/internal/enum"

//go:generate stringer -linecomment -type=PortfolioEventType -output=type_string.go

// PortfolioEventType is the type of a portfolio event.
type PortfolioEventType int

const (
	// PortfolioEventTypeBuy represents a buy event.
	PortfolioEventTypeBuy PortfolioEventType = iota + 1 // BUY
	// PortfolioEventTypeSell represents a sell event.
	PortfolioEventTypeSell // SELL
	// PortfolioEventTypeDividend represents a dividend event.
	PortfolioEventTypeDividend // DIVIDEND
	// PortfolioEventTypeTax represents the inbound delivery of a security or cash from another account.
	PortfolioEventTypeDeliveryInbound // DELIVERY_INBOUND
	// PortfolioEventTypeTax represents the outbound delivery of a security or cash to another account.
	PortfolioEventTypeDeliveryOutbound // DELIVERY_OUTBOUND
	// PortfolioEventTypeDepositCash represents a deposit of cash.
	PortfolioEventTypeDepositCash // DEPOSIT_CASH
	// PortfolioEventTypeWithdrawCash represents a withdrawal of cash.
	PortfolioEventTypeWithdrawCash // WITHDRAW_CASH
	// PortfolioEventTypeUnknown represents an unknown event type.
	PortfolioEventTypeUnknown // UNKNOWN
)

// Get implements [flag.Getter].
func (t *PortfolioEventType) Get() any {
	return t
}

// Set implements [flag.Value].
func (t *PortfolioEventType) Set(v string) error {
	return enum.Set(t, v, _PortfolioEventType_name, _PortfolioEventType_index[:])
}

// MarshalJSON marshals the account type to JSON using the string
// representation.
func (t PortfolioEventType) MarshalJSON() ([]byte, error) {
	return enum.MarshalJSON(t)
}

// UnmarshalJSON unmarshals the account type from JSON. It expects a string
// representation.
func (t *PortfolioEventType) UnmarshalJSON(data []byte) error {
	return enum.UnmarshalJSON(t, data)
}
