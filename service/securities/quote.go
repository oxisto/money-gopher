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

package securities

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/persistence"

	"connectrpc.com/connect"
	"github.com/lmittmann/tint"
)

// UpdateQuotes triggers an update of the quotes for the given securities.
func (svc *service) UpdateQuotes(ctx context.Context, IDs []string) (err error) {
	var (
		sec    *persistence.Security
		listed []*persistence.ListedSecurity
		qp     QuoteProvider
		ok     bool
	)

	for _, id := range IDs {
		// Fetch security
		sec, err = svc.db.GetSecurity(ctx, id)
		if err != nil {
			return err
		}

		if !sec.QuoteProvider.Valid {
			slog.Warn("No quote provider configured for security", "security", sec.ID)
			return
		}

		qp, ok = providers[sec.QuoteProvider.String]
		if !ok {
			return
		}

		listed, err = sec.ListedAs(ctx, svc.db)
		if err != nil {
			return err
		}

		// Trigger update from quote provider in separate go-routine
		// TODO(oxisto): Use sync/errgroup instead
		for _, ls := range listed {
			go func() {
				slog.Debug("Triggering quote update", "security", ls, "provider", sec.QuoteProvider)

				err = svc.updateQuote(qp, ls)
				if err != nil {
					slog.Error("An error occurred during quote update", tint.Err(err), "ls", ls)
				}
			}()
		}
	}

	return
}

func (svc *service) TriggerSecurityQuoteUpdate(ctx context.Context, req *connect.Request[portfoliov1.TriggerQuoteUpdateRequest]) (res *connect.Response[portfoliov1.TriggerQuoteUpdateResponse], err error) {
	err = svc.UpdateQuotes(ctx, req.Msg.SecurityIds)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res = connect.NewResponse(&portfoliov1.TriggerQuoteUpdateResponse{})

	return
}

func (svc *service) updateQuote(qp QuoteProvider, ls *persistence.ListedSecurity) (err error) {
	var (
		quote  *persistence.Currency
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

	ls.LatestQuote = sql.NullInt64{Int64: int64(quote.Value), Valid: true}
	ls.LatestQuoteTimestamp = sql.NullTime{Time: t, Valid: true}

	_, err = svc.db.UpsertListedSecurity(ctx, persistence.UpsertListedSecurityParams{
		SecurityID: ls.SecurityID,
		Ticker:     ls.Ticker,
		Currency:   ls.Currency,
	})
	if err != nil {
		return err
	}

	return
}
