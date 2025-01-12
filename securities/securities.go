package securities

import (
	"context"

	"github.com/oxisto/money-gopher/persistence"
)

// SecuritiesLister is the interface for listing securities.
type SecuritiesLister interface {
	ListSecurities(ctx context.Context) ([]*persistence.Security, error)
	ListSecuritiesByIDs(ctx context.Context, ids []string) ([]*persistence.Security, error)
}
