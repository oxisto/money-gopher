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

type fifoTx struct {
	// amount of shares in this transaction
	amount int32
	// value contains the net value of this transaction, i.e., without taxes and fees
	value float32
	// fees contain any fees associated to this transaction
	fees float32
	// ppu is the price per unit (amount)
	ppu float32
}

type calculation struct {
	Amount     int32
	BuyerFees  float32
	SellerFees float32
	TotalFees  float32
	Taxes      float32

	fifo []*fifoTx
}

func NewCalculation(txs []*portfoliov1.PortfolioEvent) *calculation {
	var c calculation

	for _, tx := range txs {
		c.Apply(tx)
	}

	return &c
}

func (c *calculation) Apply(tx *portfoliov1.PortfolioEvent) {
	switch v := tx.EventOneof.(type) {
	case *portfoliov1.PortfolioEvent_Buy:
		if v.Buy == nil {
			return
		}

		// Increase the amount of shares and the fees by the value stored in the
		// transaction
		c.BuyerFees += v.Buy.Fees
		c.TotalFees += v.Buy.Fees
		c.Amount += v.Buy.Amount

		// Add the transaction to the FIFO queue. We need to have a queue
		// because sold shares are sold according to the FIFO principle. We
		// therefore need to store this information to reduce the amount in the
		// items later when a sell transaction occurs.
		c.fifo = append(c.fifo, &fifoTx{
			amount: v.Buy.Amount,
			ppu:    v.Buy.Price,
			value:  v.Buy.Price * float32(v.Buy.Amount),
			fees:   v.Buy.Fees,
		})
	case *portfoliov1.PortfolioEvent_Sell:
		var (
			sold int32
		)

		if v.Sell == nil {
			return
		}

		// Increase the fees and taxes by the value stored in the
		// transaction
		c.SellerFees += v.Sell.Fees
		c.TotalFees += v.Sell.Fees
		c.Taxes += v.Sell.Taxes

		// Store the amount of shares sold in this variable, since we later need
		// to decrease it while looping through the FIFO queue
		sold = v.Sell.Amount

		// Calculate the remaining shares (if any)
		c.Amount -= sold
		if c.Amount < 0 {
			// TODO(oxisto): some kind of warning would probably be nice here
			c.Amount = 0
		}

		// We need to loop through our FIFO queue and reduce the amount of sold
		// shares until it is 0.
		for _, item := range c.fifo {
			if sold <= 0 {
				// All sold shares accounted for. We are done
				break
			}

			// FIFO items could already be empty since we sold those shares
			// already; we cannot really remove them from the queue properly, so
			// we can just skip them
			if item.amount == 0 {
				continue
			}

			// Reduce the number of shares in this entry by the sold amount (but
			// max it to the item's amount).
			n := int32(math.Min(float64(sold), float64(item.amount)))
			item.amount -= n

			// Adjust the value with the new amount
			item.value = item.ppu * float32(item.amount)

			// If no shares are left in this FIFO transaction, also remove the
			// fees, because they are now associated to the sale and not part of
			// the price calculation anymore.
			if item.amount <= 0 {
				item.fees = 0
			}

			sold -= n
		}
	}
}

func (c *calculation) NetValue() (f float32) {
	for _, item := range c.fifo {
		f += item.value
	}

	return
}

func (c *calculation) GrossValue() (f float32) {
	for _, item := range c.fifo {
		f += (item.value + item.fees)
	}

	return
}

func (c *calculation) NetPrice() (f float32) {
	return c.NetValue() / float32(c.Amount)
}

func (c *calculation) GrossPrice() (f float32) {
	return c.GrossValue() / float32(c.Amount)
}
