// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"database/sql"
)

type ListedSecurity struct {
	SecurityID           string
	Ticker               string
	Currency             string
	LatestQuote          sql.NullInt64
	LatestQuoteTimestamp sql.NullTime
}

type Security struct {
	ID            string
	DisplayName   string
	QuoteProvider sql.NullString
}
