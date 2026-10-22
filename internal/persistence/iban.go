package persistence

import (
	"context"
	"database/sql"
	"strings"
)

// NormalizeIBAN brings an IBAN into the form it is stored and compared in:
// upper case, no spaces. People type "DE46 5001 0517 5802 2452 63", bank
// statements print it that way too, and the importer strips the spaces —
// one form for all of them, or the statement never finds its account.
func NormalizeIBAN(iban string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(iban), " ", ""))
}

// NormalizeIBANs rewrites every stored IBAN into the normalized form. It
// is cheap when there is nothing to do; moneyd runs it at startup, for
// accounts created before IBANs were normalized on the way in.
func (db *DB) NormalizeIBANs(ctx context.Context) (int, error) {
	accounts, err := db.ListCashAccountsWithIBAN(ctx)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, a := range accounts {
		normalized := NormalizeIBAN(a.Iban.String)
		if normalized == a.Iban.String {
			continue
		}
		if err := db.SetCashAccountIBAN(ctx, SetCashAccountIBANParams{
			Iban: sql.NullString{String: normalized, Valid: true},
			ID:   a.ID,
		}); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
