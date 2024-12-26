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

// package csv contains a CSV importer for securities and portfolios.
//
// It supports CSV files containing transactions with the following header
// structure:
//
//	Date;Type;Value;Transaction Currency;Gross Amount;Currency Gross Amount;Exchange Rate;Fees;Taxes;Shares;ISIN;WKN;Ticker Symbol;Security Name;Note
//
// Values must be separated using a semi-colon and numbers are formatted using a
// German locale.
//
// This structure is intentionally compatible with the export functionality of
// [Portfolio Performance](https://github.com/buchen/portfolio).
package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	moneygopher "github.com/oxisto/money-gopher"
	portfoliov1 "github.com/oxisto/money-gopher/gen"

	"github.com/lmittmann/tint"
	"github.com/oxisto/money-gopher/service/securities"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrReadingCSV    = errors.New("could not read CSV line")
	ErrParsingType   = errors.New("could not parse type")
	ErrParsingTime   = errors.New("could not parse time")
	ErrParsingTaxes  = errors.New("could not parse taxes")
	ErrParsingFees   = errors.New("could not parse fees")
	ErrParsingAmount = errors.New("could not parse amount")
	ErrParsingValue  = errors.New("could not parse value")
)

// Import imports CSV records from a [io.Reader] containing portfolio
// transactions.
func Import(r io.Reader, pname string) (txs []*portfoliov1.PortfolioEvent, secs []*portfoliov1.Security) {
	cr := csv.NewReader(r)
	cr.Comma = ';'

	// Skip header line
	cr.Read()

	// Read until EOF
	for {
		tx, sec, err := readLine(cr, pname)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			// Skip this transaction
			slog.Warn("Could not parse line", tint.Err(err))
			continue
		}

		txs = append(txs, tx)
		secs = append(secs, sec)
	}

	// Compact securities
	secs = slices.CompactFunc(secs, func(a *portfoliov1.Security, b *portfoliov1.Security) bool {
		return a.Id == b.Id
	})

	return
}

func readLine(cr *csv.Reader, pname string) (tx *portfoliov1.PortfolioEvent, sec *portfoliov1.Security, err error) {
	var (
		record []string
		value  *portfoliov1.Currency
	)

	record, err = cr.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrReadingCSV, err)
	}

	tx = new(portfoliov1.PortfolioEvent)
	tx.Time, err = txTime(record[0])
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrParsingTime, err)
	}

	tx.Type = txType(record[1])
	if tx.Type == portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_UNSPECIFIED {
		return nil, nil, ErrParsingType
	}

	value, err = parseFloatCurrency(record[2])
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrParsingValue, err)
	}

	tx.Fees, err = parseFloatCurrency(record[7])
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrParsingFees, err)
	}

	tx.Taxes, err = parseFloatCurrency(record[8])
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrParsingTaxes, err)
	}

	tx.Amount, err = parseFloat64(record[9])
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrParsingAmount, err)
	}

	// Calculate the price
	if tx.Type == portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY ||
		tx.Type == portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_DELIVERY_INBOUND {
		tx.Price = portfoliov1.Divide(portfoliov1.Minus(value, tx.Fees), tx.Amount)
	} else if tx.Type == portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_SELL ||
		tx.Type == portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_DELIVERY_OUTBOUND {
		tx.Price = portfoliov1.Times(portfoliov1.Divide(portfoliov1.Minus(portfoliov1.Minus(value, tx.Fees), tx.Taxes), tx.Amount), -1)
	}

	sec = new(portfoliov1.Security)
	sec.Id = record[10]
	sec.DisplayName = record[13]
	sec.ListedOn = []*portfoliov1.ListedSecurity{
		{
			SecurityId: sec.Id,
			Ticker:     record[12],
			Currency:   lsCurrency(record[3], record[5]),
		},
	}

	// Default to YF, but only if we have a ticker symbol, otherwise, let's try ING
	if len(sec.ListedOn) >= 0 && len(sec.ListedOn[0].Ticker) > 0 {
		sec.QuoteProvider = moneygopher.Ref(securities.QuoteProviderYF)
	} else {
		sec.QuoteProvider = moneygopher.Ref(securities.QuoteProviderING)
	}

	tx.PortfolioId = pname
	tx.SecurityId = sec.Id
	tx.MakeUniqueName()

	return
}

func txType(typ string) portfoliov1.PortfolioEventType {
	switch typ {
	case "Buy":
		return portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY
	case "Sell":
		return portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_SELL
	case "Delivery (Inbound)":
		return portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_DELIVERY_INBOUND
	case "Delivery (Outbound)":
		return portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_DELIVERY_OUTBOUND
	default:
		return portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_UNSPECIFIED
	}
}

func txTime(s string) (ts *timestamppb.Timestamp, err error) {
	var (
		t time.Time
	)
	// First try without seconds
	t, err = time.ParseInLocation("2006-01-02T15:04", s, time.Local)
	if err != nil {
		// Then with seconds
		t, err = time.ParseInLocation("2006-01-02T15:04:05", s, time.Local)
		if err != nil {
			return nil, err
		}
	}

	return timestamppb.New(t), nil
}

func parseFloat64(s string) (f float64, err error) {
	// We assume that the float is in German locale (this might not be true for
	// all users), so we need to convert it
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")

	f, err = strconv.ParseFloat(s, 32)
	if err != nil {
		return 0, err
	}

	return
}

func parseFloatCurrency(s string) (c *portfoliov1.Currency, err error) {
	// Get rid of all , and .
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")

	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return portfoliov1.Zero(), err
	}

	return portfoliov1.Value(int32(i)), nil
}

func lsCurrency(txCurrency string, tickerCurrency string) string {
	if tickerCurrency == "" {
		return txCurrency
	} else {
		return tickerCurrency
	}
}
