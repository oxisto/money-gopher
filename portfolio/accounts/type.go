package accounts

// AccountType is the type of an account.
type AccountType int

const (
	// AccountTypeBrokerage represents a brokerage account.
	AccountTypeBrokerage AccountType = iota + 1
	// AccountTypeBank represents a bank account.
	AccountTypeBank
	// AccountTypeLoan represents a loan account.
	AccountTypeLoan
	// AccountTypeUnknown represents an unknown account type.
	AccountTypeUnknown
)
