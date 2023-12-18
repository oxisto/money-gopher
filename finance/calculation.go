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

// package portfolio contains all kinds of different finance calculations.
package finance

import (
	"math"

	portfoliov1 "github.com/oxisto/money-gopher/gen"
)

// fifoTx is a helper struct to store transaction-related information in a FIFO
// list. We basically need to copy the values from the original transaction,
// since we need to modify it.
type fifoTx struct {
	// amount of shares in this transaction
	amount float32
	// value contains the net value of this transaction, i.e., without taxes and fees
	value *portfoliov1.Currency
	// fees contain any fees associated to this transaction
	fees *portfoliov1.Currency
	// ppu is the price per unit (amount)
	ppu *portfoliov1.Currency
}

type calculation struct {
	Amount float32
	Fees   *portfoliov1.Currency
	Taxes  *portfoliov1.Currency

	fifo []*fifoTx
}

func NewCalculation(txs []*portfoliov1.PortfolioEvent) *calculation {
	var c calculation
	c.Fees = portfoliov1.Zero()
	c.Taxes = portfoliov1.Zero()

	for _, tx := range txs {
		c.Apply(tx)
	}

	return &c
}

func (c *calculation) Apply(tx *portfoliov1.PortfolioEvent) {
	switch tx.Type {
	case portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_DELIVERY_INBOUND:
		fallthrough
	case portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY:
		// Increase the amount of shares and the fees by the value stored in the
		// transaction
		c.Fees.PlusAssign(tx.Fees)
		c.Amount += tx.Amount

		// Add the transaction to the FIFO list. We need to have a list because
		// sold shares are sold according to the FIFO principle. We therefore
		// need to store this information to reduce the amount in the items
		// later when a sell transaction occurs.
		c.fifo = append(c.fifo, &fifoTx{
			amount: tx.Amount,
			ppu:    tx.Price,
			value:  portfoliov1.Times(tx.Price, tx.Amount),
			fees:   tx.Fees,
		})
	case portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_DELIVERY_OUTBOUND:
		fallthrough
	case portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_SELL:
		var (
			sold float32
		)

		// Increase the fees and taxes by the value stored in the
		// transaction
		c.Fees.PlusAssign(tx.Fees)
		c.Taxes.PlusAssign(tx.Taxes)

		// Store the amount of shares sold in this variable, since we later need
		// to decrease it while looping through the FIFO list
		sold = tx.Amount

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
			n := float32(math.Min(float64(sold), float64(item.amount)))
			item.amount -= n

			// Adjust the value with the new amount
			item.value = portfoliov1.Times(item.ppu, item.amount)

			// If no shares are left in this FIFO transaction, also remove the
			// fees, because they are now associated to the sale and not part of
			// the price calculation anymore.
			if item.amount <= 0 {
				item.fees = portfoliov1.Zero()
			}

			sold -= n
		}
	}
}

func (c *calculation) NetValue() (f *portfoliov1.Currency) {
	f = &portfoliov1.Currency{}

	for _, item := range c.fifo {
		f.PlusAssign(item.value)
	}

	return
}

func (c *calculation) GrossValue() (f *portfoliov1.Currency) {
	f = &portfoliov1.Currency{}

	for _, item := range c.fifo {
		f.PlusAssign(portfoliov1.Plus(item.value, item.fees))
	}

	return
}

func (c *calculation) NetPrice() (f *portfoliov1.Currency) {
	return portfoliov1.Divide(c.NetValue(), c.Amount)
}

func (c *calculation) GrossPrice() (f *portfoliov1.Currency) {
	return portfoliov1.Divide(c.GrossValue(), c.Amount)
}
