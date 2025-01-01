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
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/oxisto/money-gopher/currency"
	"github.com/oxisto/money-gopher/persistence"
	"github.com/oxisto/money-gopher/portfolio/events"
	"github.com/oxisto/money-gopher/securities/quote"

	"github.com/lmittmann/tint"
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
func Import(r io.Reader, pname string) (
	txs []*persistence.PortfolioEvent,
	secs []*persistence.Security,
	lss []*persistence.ListedSecurity,
) {
	cr := csv.NewReader(r)
	cr.Comma = ';'

	// Skip header line
	cr.Read()

	// Read until EOF
	for {
		tx, sec, ls, err := readLine(cr, pname)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			// Skip this transaction
			slog.Warn("Could not parse line", tint.Err(err))
			continue
		}

		txs = append(txs, tx)
		secs = append(secs, sec)
		lss = append(lss, ls...)
	}

	// Make (listed) securities unique
	secs = slices.CompactFunc(secs, func(a *persistence.Security, b *persistence.Security) bool {
		return a.ID == b.ID
	})
	lss = slices.CompactFunc(lss, func(a *persistence.ListedSecurity, b *persistence.ListedSecurity) bool {
		return a.SecurityID == b.SecurityID && a.Ticker == b.Ticker
	})

	return
}

func readLine(cr *csv.Reader, pname string) (
	tx *persistence.PortfolioEvent,
	sec *persistence.Security,
	ls []*persistence.ListedSecurity,
	err error) {
	var (
		record []string
		value  *currency.Currency
	)

	record, err = cr.Read()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: %w", ErrReadingCSV, err)
	}

	tx = new(persistence.PortfolioEvent)
	tx.Time, err = txTime(record[0])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: %w", ErrParsingTime, err)
	}

	tx.Type = txType(record[1])
	if tx.Type == events.PortfolioEventTypeUnknown {
		return nil, nil, nil, ErrParsingType
	}

	value, err = parseFloatCurrency(record[2])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: %w", ErrParsingValue, err)
	}

	tx.Fees, err = parseFloatCurrency(record[7])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: %w", ErrParsingFees, err)
	}

	tx.Taxes, err = parseFloatCurrency(record[8])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: %w", ErrParsingTaxes, err)
	}

	tx.Amount, err = parseFloat64(record[9])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: %w", ErrParsingAmount, err)
	}

	// Calculate the price
	if tx.Type == events.PortfolioEventTypeBuy ||
		tx.Type == events.PortfolioEventTypeDeliveryInbound {
		tx.Price = currency.Divide(currency.Minus(value, tx.Fees), tx.Amount)
	} else if tx.Type == events.PortfolioEventTypeSell ||
		tx.Type == events.PortfolioEventTypeDeliveryOutbound {
		tx.Price = currency.Times(currency.Divide(currency.Minus(currency.Minus(value, tx.Fees), tx.Taxes), tx.Amount), -1)
	}

	sec = new(persistence.Security)
	sec.ID = record[10]
	sec.DisplayName = record[13]

	ls = append(ls, &persistence.ListedSecurity{
		SecurityID: sec.ID,
		Ticker:     record[12],
		Currency:   lsCurrency(record[3], record[5]),
	})

	// Default to YF, but only if we have a ticker symbol, otherwise, let's try ING
	if len(ls) >= 0 && len(ls[0].Ticker) > 0 {
		sec.QuoteProvider = sql.NullString{String: quote.QuoteProviderYF, Valid: true}
	} else {
		sec.QuoteProvider = sql.NullString{String: quote.QuoteProviderING, Valid: true}
	}

	tx.PortfolioID = pname
	tx.SecurityID = sec.ID
	tx.MakeUniqueID()

	return
}

func txType(typ string) events.PortfolioEventType {
	switch typ {
	case "Buy":
		return events.PortfolioEventTypeBuy
	case "Sell":
		return events.PortfolioEventTypeSell
	case "Delivery (Inbound)":
		return events.PortfolioEventTypeDeliveryInbound
	case "Delivery (Outbound)":
		return events.PortfolioEventTypeDeliveryOutbound
	default:
		return events.PortfolioEventTypeUnknown
	}
}

func txTime(s string) (t time.Time, err error) {
	// First try without seconds
	t, err = time.ParseInLocation("2006-01-02T15:04", s, time.Local)
	if err != nil {
		// Then with seconds
		t, err = time.ParseInLocation("2006-01-02T15:04:05", s, time.Local)
		if err != nil {
			return time.Time{}, err
		}
	}

	return t, nil
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

func parseFloatCurrency(s string) (c *currency.Currency, err error) {
	// Get rid of all , and .
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")

	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return currency.Zero(), err
	}

	return currency.Value(int32(i)), nil
}

func lsCurrency(txCurrency string, tickerCurrency string) string {
	if tickerCurrency == "" {
		return txCurrency
	} else {
		return tickerCurrency
	}
}
