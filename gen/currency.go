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
	"fmt"
	"math"
)

func Zero() *Currency {
	// TODO(oxisto): Somehow make it possible to change default currency
	return &Currency{Symbol: "EUR"}
}

func Value(v int32) *Currency {
	// TODO(oxisto): Somehow make it possible to change default currency
	return &Currency{Symbol: "EUR", Value: v}
}

func (c *Currency) PlusAssign(o *Currency) {
	if o != nil {
		c.Value += o.Value
	}
}

func (c *Currency) MinusAssign(o *Currency) {
	if o != nil {
		c.Value -= o.Value
	}
}

func Plus(a *Currency, b *Currency) *Currency {
	return &Currency{
		Value:  a.Value + b.Value,
		Symbol: a.Symbol,
	}
}

func (a *Currency) Plus(b *Currency) *Currency {
	if b == nil {
		return &Currency{
			Value:  a.Value,
			Symbol: a.Symbol,
		}
	}

	return &Currency{
		Value:  a.Value + b.Value,
		Symbol: a.Symbol,
	}
}

func Minus(a *Currency, b *Currency) *Currency {
	return &Currency{
		Value:  a.Value - b.Value,
		Symbol: a.Symbol,
	}
}

func Divide(a *Currency, b float64) *Currency {
	return &Currency{
		Value:  int32(math.Round((float64(a.Value) / b))),
		Symbol: a.Symbol,
	}
}

func Times(a *Currency, b float64) *Currency {
	return &Currency{
		Value:  int32(math.Round((float64(a.Value) * b))),
		Symbol: a.Symbol,
	}
}

func (c *Currency) Pretty() string {
	return fmt.Sprintf("%.0f %s", float32(c.Value)/100, c.Symbol)
}

func (c *Currency) IsZero() bool {
	return c == nil || c.Value == 0
}
