package accounts

import (
	"encoding/json"
	"fmt"
)

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

func (t *AccountType) Get() any {
	return t
}

func (t *AccountType) Set(v string) error {
	switch v {
	case "BROKERAGE":
		*t = AccountTypeBrokerage
	case "BANK":
		*t = AccountTypeBank
	case "LOAN":
		*t = AccountTypeLoan
	case "UNKNOWN":
		*t = AccountTypeUnknown
	default:
		return fmt.Errorf("unknown account type: %s", v)
	}

	return nil
}

// String returns the string representation of the account type. This matches
// the enum value of the GraphQL schema.
func (t AccountType) String() string {
	switch t {
	case AccountTypeBrokerage:
		return "BROKERAGE"
	case AccountTypeBank:
		return "BANK"
	case AccountTypeLoan:
		return "LOAN"
	default:
		return "UNKNOWN"
	}
}

// MarshalJSON marshals the account type to JSON using the string
// representation.
func (t AccountType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, t.String())), nil
}

// UnmarshalJSON unmarshals the account type from JSON. It expects a string
// representation.
func (t *AccountType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	return t.Set(s)
}
