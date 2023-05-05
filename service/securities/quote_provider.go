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
	"time"

	portfoliov1 "github.com/oxisto/money-gopher/gen"
)

// providers contains a map of all quote providers
var providers map[string]QuoteProvider = make(map[string]QuoteProvider)

func init() {
	RegisterQuoteProvider("yf", &yf{})
}

// AddCommand adds a command using the specific symbol.
func RegisterQuoteProvider(name string, qp QuoteProvider) {
	providers[name] = qp
}

// QuoteProvider is an interface that retrieves quotes for a [ListedSecurity]. They
// can either be historical quotes or the latest quote.
type QuoteProvider interface {
	LatestQuote(ctx context.Context, ls *portfoliov1.ListedSecurity) (quote float32, t time.Time, err error)
}
