package accounts

import (
	"github.com/oxisto/money-gopher/internal/enum"
)

//go:generate stringer -linecomment -type=AccountType -output=type_string.go

// AccountType is the type of an account.
type AccountType int

const (
	// AccountTypeBrokerage represents a brokerage account.
	AccountTypeBrokerage AccountType = iota + 1 // BROKERAGE
	// AccountTypeBank represents a bank account.
	AccountTypeBank // BANK
	// AccountTypeLoan represents a loan account.
	AccountTypeLoan // LOAN
)

// Get implements [flag.Getter].
func (t *AccountType) Get() any {
	return t
}

// Set implements [flag.Value].
func (t *AccountType) Set(v string) error {
	return enum.Set(t, v, _AccountType_name, _AccountType_index[:])
}

// MarshalJSON marshals the account type to JSON using the string
// representation.
func (t AccountType) MarshalJSON() ([]byte, error) {
	return enum.MarshalJSON(t)
}

// UnmarshalJSON unmarshals the account type from JSON. It expects a string
// representation.
func (t *AccountType) UnmarshalJSON(data []byte) error {
	return enum.UnmarshalJSON(t, data)
}
