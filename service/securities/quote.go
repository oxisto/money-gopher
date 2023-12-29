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
	"log/slog"
	"time"

	portfoliov1 "github.com/oxisto/money-gopher/gen"

	"connectrpc.com/connect"
	"github.com/lmittmann/tint"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (svc *service) TriggerSecurityQuoteUpdate(ctx context.Context, req *connect.Request[portfoliov1.TriggerQuoteUpdateRequest]) (res *connect.Response[portfoliov1.TriggerQuoteUpdateResponse], err error) {
	var (
		sec *portfoliov1.Security
		qp  QuoteProvider
		ok  bool
	)

	// TODO(oxisto): Support a "list" with filtered values instead
	for _, name := range req.Msg.SecurityNames {
		// Fetch security
		sec, err = svc.fetchSecurity(name)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}

		res = connect.NewResponse(&portfoliov1.TriggerQuoteUpdateResponse{})

		if sec.QuoteProvider == nil {
			slog.Warn("No quote provider configured for security", "security", sec.Name)
			return
		}

		qp, ok = providers[*sec.QuoteProvider]
		if !ok {
			return
		}

		// Trigger update from quote provider in separate go-routine
		// TODO(oxisto): Use sync/errgroup instead
		for idx := range sec.ListedOn {
			idx := idx
			go func() {
				ls := sec.ListedOn[idx]

				slog.Debug("Triggering quote update", "security", ls, "provider", *sec.QuoteProvider)

				err = svc.updateQuote(qp, ls)
				if err != nil {
					slog.Error("An error occurred during quote update", tint.Err(err), "ls", ls)
				}
			}()
		}
	}

	return
}

func (svc *service) updateQuote(qp QuoteProvider, ls *portfoliov1.ListedSecurity) (err error) {
	var (
		quote  *portfoliov1.Currency
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
	ls.LatestQuoteTimestamp = timestamppb.New(t)

	_, err = svc.listedSecurities.Update(
		[]any{ls.SecurityName, ls.Ticker},
		ls, []string{"latest_quote", "latest_quote_timestamp"},
	)
	if err != nil {
		return err
	}

	return
}
