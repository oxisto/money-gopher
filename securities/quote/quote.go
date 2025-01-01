// Copyright 2023 Christian Banse
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This file is part of The Money Gopher.

// package quote contains the logic to update quotes for securities. Its main
// way to interface is the [QuoteUpdater] interface. A default implementation
// for the interface can be created using [NewQuoteUpdater].
package quote

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/oxisto/money-gopher/currency"
	"github.com/oxisto/money-gopher/persistence"

	"github.com/lmittmann/tint"
)

// QuoteProvider is an interface that retrieves quotes for a [ListedSecurity]. They
// can either be historical quotes or the latest quote.
type QuoteUpdater interface {
	UpdateQuotes(ctx context.Context, IDs []string) (err error)
}

// qu is the internal default implementation of the [QuoteUpdater] interface.
type qu struct {
	db *persistence.DB
}

// NewQuoteUpdater creates a new instance of the [QuoteUpdater] interface.
func NewQuoteUpdater(db *persistence.DB) QuoteUpdater {
	return &qu{
		db: db,
	}
}

// UpdateQuotes triggers an update of the quotes for the given securities' IDs.
func (qu *qu) UpdateQuotes(ctx context.Context, secIDs []string) (err error) {
	var (
		secs   []*persistence.Security
		listed []*persistence.ListedSecurity
		qp     QuoteProvider
		ok     bool
	)

	// Fetch all securities if no IDs are given
	if len(secIDs) == 0 {
		secs, err = qu.db.ListSecurities(ctx)
	} else {
		secs, err = qu.db.ListSecuritiesByIDs(ctx, secIDs)
	}
	if err != nil {
		return err
	}

	for _, sec := range secs {
		if !sec.QuoteProvider.Valid {
			slog.Warn("No quote provider configured for security", "security", sec.ID)
			return
		}

		qp, ok = providers[sec.QuoteProvider.String]
		if !ok {
			return
		}

		listed, err = sec.ListedAs(ctx, qu.db)
		if err != nil {
			return err
		}

		// Trigger update from quote provider in separate go-routine
		// TODO(oxisto): Use sync/errgroup instead
		for _, ls := range listed {
			go func() {
				slog.Debug("Triggering quote update", "security", ls, "provider", sec.QuoteProvider)

				err = qu.updateQuote(qp, ls)
				if err != nil {
					slog.Error("An error occurred during quote update", tint.Err(err), "ls", ls)
				}
			}()
		}
	}

	return
}

// updateQuote updates the quote for the given [ListedSecurity] using the given [QuoteProvider].
func (qu *qu) updateQuote(qp QuoteProvider, ls *persistence.ListedSecurity) (err error) {
	var (
		quote  *currency.Currency
		t      time.Time
		ctx    context.Context
		cancel context.CancelFunc
	)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	quote, t, err = qp.LatestQuote(ctx, ls)
	if err != nil {
		return err
	}

	ls.LatestQuote = quote
	ls.LatestQuoteTimestamp = sql.NullTime{Time: t, Valid: true}

	_, err = qu.db.UpsertListedSecurity(ctx, persistence.UpsertListedSecurityParams{
		SecurityID: ls.SecurityID,
		Ticker:     ls.Ticker,
		Currency:   ls.Currency,
	})
	if err != nil {
		return err
	}

	return
}
