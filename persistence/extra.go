package persistence

import (
	"context"
	"hash/fnv"
	"strconv"
	"time"
)

// ListedAs returns the listed securities for the security.
func (s *Security) ListedAs(ctx context.Context, db *DB) ([]*ListedSecurity, error) {
	return db.ListListedSecuritiesBySecurityID(ctx, s.ID)

}

// MakeUniqueID creates a unique ID for the portfolio event based on a hash containing:
//   - security ID
//   - portfolio ID
//   - date
//   - amount
func (tx *PortfolioEvent) MakeUniqueID() {
	h := fnv.New64a()
	h.Write([]byte(tx.SecurityID))
	h.Write([]byte(tx.PortfolioID))
	h.Write([]byte(tx.Time.Format(time.DateTime)))
	h.Write([]byte(strconv.FormatInt(int64(tx.Type), 10)))
	h.Write([]byte(strconv.FormatInt(int64(tx.Amount), 10)))

	tx.ID = strconv.FormatUint(h.Sum64(), 16)
}
