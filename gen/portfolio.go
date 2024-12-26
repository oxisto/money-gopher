// Copyright 2023 Christian Banse
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This file is part of The Money Gopher.

package portfoliov1

import (
	"hash/fnv"
	"log/slog"
	"strconv"
	"time"
)

func (p *Portfolio) EventMap() (m map[string][]*PortfolioEvent) {
	m = make(map[string][]*PortfolioEvent)

	for _, tx := range p.Events {
		name := tx.GetSecurityId()
		if name != "" {
			m[name] = append(m[name], tx)
		} else {
			// a little bit of a hack
			m["cash"] = append(m["cash"], tx)
		}
	}

	return
}

func EventsBefore(txs []*PortfolioEvent, t time.Time) (out []*PortfolioEvent) {
	out = make([]*PortfolioEvent, 0, len(txs))

	for _, tx := range txs {
		if tx.GetTime().AsTime().After(t) {
			continue
		}

		out = append(out, tx)
	}

	return
}

func (tx *PortfolioEvent) MakeUniqueName() {
	// Create a unique name based on a hash containing:
	//  - security name
	//  - portfolio name
	//  - date
	//  - amount
	h := fnv.New64a()
	h.Write([]byte(tx.SecurityId))
	h.Write([]byte(tx.PortfolioId))
	h.Write([]byte(tx.Time.AsTime().Local().Format(time.DateTime)))
	h.Write([]byte(strconv.FormatInt(int64(tx.Type), 10)))
	h.Write([]byte(strconv.FormatInt(int64(tx.Amount), 10)))

	tx.Id = strconv.FormatUint(h.Sum64(), 16)
}

// LogValue implements slog.LogValuer.
func (tx *PortfolioEvent) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", tx.Id),
		slog.String("security.id", tx.SecurityId),
		slog.Float64("amount", float64(tx.Amount)),
		slog.String("price", tx.Price.Pretty()),
		slog.String("fees", tx.Fees.Pretty()),
		slog.String("taxes", tx.Taxes.Pretty()),
	)
}

// LogValue implements slog.LogValuer.
func (ls *ListedSecurity) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", ls.SecurityId),
		slog.String("ticker", ls.Ticker),
	)
}
