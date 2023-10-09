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
	"log"
	"time"

	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"google.golang.org/protobuf/types/known/timestamppb"

	"connectrpc.com/connect"
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
			log.Printf("No quote provider configured for %s\n", sec.Name)
			return
		}

		qp, ok = providers[*sec.QuoteProvider]
		if !ok {
			return
		}

		// Trigger update from quote provider in separate go-routine
		for _, ls := range sec.ListedOn {
			go svc.updateQuote(qp, ls)
		}
	}

	return
}

func (svc *service) updateQuote(qp QuoteProvider, ls *portfoliov1.ListedSecurity) (err error) {
	var (
		quote  float32
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

	ls.LatestQuote = &quote
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
