package persistence

import "context"

// ListedAs returns the listed securities for the security.
func (s *Security) ListedAs(ctx context.Context, db *DB) ([]*ListedSecurity, error) {
	return db.ListListedSecuritiesBySecurityID(ctx, s.ID)
}
