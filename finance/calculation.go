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

// package finance contains all kinds of different finance calculations.
package finance

import (
	"math"

	"github.com/oxisto/money-gopher/currency"
	"github.com/oxisto/money-gopher/persistence"
	"github.com/oxisto/money-gopher/portfolio/events"
)

// fifoTx is a helper struct to store transaction-related information in a FIFO
// list. We basically need to copy the values from the original transaction,
// since we need to modify it.
type fifoTx struct {
	amount float64            // amount of shares in this transaction
	value  *currency.Currency // value contains the net value of this transaction, i.e., without taxes and fees
	fees   *currency.Currency // fees contain any fees associated to this transaction
	ppu    *currency.Currency // ppu is the price per unit (amount)
}

// calculation is a helper struct to calculate the net and gross value of a
// portfolio (snapshot).
type calculation struct {
	Amount float64
	Fees   *currency.Currency
	Taxes  *currency.Currency

	Cash *currency.Currency

	fifo []*fifoTx
}

// NewCalculation creates a new calculation struct and applies all events
func NewCalculation(events []*persistence.PortfolioEvent) *calculation {
	var c calculation
	c.Fees = currency.Zero()
	c.Taxes = currency.Zero()
	c.Cash = currency.Zero()

	for _, tx := range events {
		c.Apply(tx)
	}

	return &c
}

func (c *calculation) Apply(tx *persistence.PortfolioEvent) {
	switch tx.Type {
	case events.PortfolioEventTypeDeliveryInbound:
		fallthrough
	case events.PortfolioEventTypeBuy:
		var (
			total *currency.Currency
		)

		// Increase the amount of shares and the fees by the value stored in the
		// transaction
		c.Fees.PlusAssign(tx.Fees())
		c.Amount += tx.Amount.Float64

		total = currency.Times(tx.Price(), tx.Amount.Float64).Plus(tx.Fees()).Plus(tx.Taxes())

		// Decrease our cash
		c.Cash.MinusAssign(total)

		// Add the transaction to the FIFO list. We need to have a list because
		// sold shares are sold according to the FIFO principle. We therefore
		// need to store this information to reduce the amount in the items
		// later when a sell transaction occurs.
		c.fifo = append(c.fifo, &fifoTx{
			amount: tx.Amount.Float64,
			ppu:    tx.Price(),
			value:  currency.Times(tx.Price(), tx.Amount.Float64),
			fees:   tx.Fees(),
		})
	case events.PortfolioEventTypeDeliveryOutbound:
		fallthrough
	case events.PortfolioEventTypeSell:
		var (
			sold  float64
			total *currency.Currency
		)

		// Increase the fees and taxes by the value stored in the
		// transaction
		c.Fees.PlusAssign(tx.Fees())
		c.Taxes.PlusAssign(tx.Taxes())

		total = currency.Times(tx.Price(), tx.Amount.Float64).Plus(tx.Fees()).Plus(tx.Taxes())

		// Increase our cash
		c.Cash.PlusAssign(total)

		// Store the amount of shares sold in this variable, since we later need
		// to decrease it while looping through the FIFO list
		sold = tx.Amount.Float64

		// Calculate the remaining shares (if any)
		c.Amount -= sold
		if c.Amount < 0 {
			// TODO(oxisto): some kind of warning would probably be nice here
			c.Amount = 0
		}

		// We need to loop through our FIFO list and reduce the amount of sold
		// shares until it is 0.
		for _, item := range c.fifo {
			if sold <= 0 {
				// All sold shares accounted for. We are done
				break
			}

			// FIFO items could already be empty since we sold those shares
			// already; we cannot really remove them from the list properly, so
			// we can just skip them
			if item.amount == 0 {
				continue
			}

			// Reduce the number of shares in this entry by the sold amount (but
			// max it to the item's amount).
			n := math.Min(float64(sold), float64(item.amount))
			item.amount -= n

			// Adjust the value with the new amount
			item.value = currency.Times(item.ppu, item.amount)

			// If no shares are left in this FIFO transaction, also remove the
			// fees, because they are now associated to the sale and not part of
			// the price calculation anymore.
			if item.amount <= 0 {
				item.fees = currency.Zero()
			}

			sold -= n
		}
	case events.PortfolioEventTypeDepositCash:
		// Add to the cash
		c.Cash.PlusAssign(tx.Price())
	case events.PortfolioEventTypeWithdrawCash:
		// Remove from the cash
		c.Cash.MinusAssign(tx.Price())
	}
}

func (c *calculation) NetValue() (f *currency.Currency) {
	f = currency.Zero()

	for _, item := range c.fifo {
		f.PlusAssign(item.value)
	}

	return
}

func (c *calculation) GrossValue() (f *currency.Currency) {
	f = currency.Zero()

	for _, item := range c.fifo {
		f.PlusAssign(currency.Plus(item.value, item.fees))
	}

	return
}

func (c *calculation) NetPrice() (f *currency.Currency) {
	return currency.Divide(c.NetValue(), c.Amount)
}

func (c *calculation) GrossPrice() (f *currency.Currency) {
	return currency.Divide(c.GrossValue(), c.Amount)
}
