package events

// PortfolioEventType is the type of a portfolio event.
type PortfolioEventType int

const (
	// PortfolioEventTypeBuy represents a buy event.
	PortfolioEventTypeBuy PortfolioEventType = iota + 1
	// PortfolioEventTypeSell represents a sell event.
	PortfolioEventTypeSell
	// PortfolioEventTypeDividend represents a dividend event.
	PortfolioEventTypeDividend
	// PortfolioEventTypeTax represents the inbound delivery of a security or cash from another account.
	PortfolioEventTypeDeliveryInbound
	// PortfolioEventTypeTax represents the outbound delivery of a security or cash to another account.
	PortfolioEventTypeDeliveryOutbound
	// PortfolioEventTypeDepositCash represents a deposit of cash.
	PortfolioEventTypeDepositCash
	// PortfolioEventTypeWithdrawCash represents a withdrawal of cash.
	PortfolioEventTypeWithdrawCash
	// PortfolioEventTypeUnknown represents an unknown event type.
	PortfolioEventTypeUnknown
)
